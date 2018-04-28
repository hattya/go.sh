%{
//
// go.sh/parser :: parser.go
//
//   Copyright (c) 2018 Akinori Hattori <hattya@gmail.com>
//
//   Permission is hereby granted, free of charge, to any person
//   obtaining a copy of this software and associated documentation files
//   (the "Software"), to deal in the Software without restriction,
//   including without limitation the rights to use, copy, modify, merge,
//   publish, distribute, sublicense, and/or sell copies of the Software,
//   and to permit persons to whom the Software is furnished to do so,
//   subject to the following conditions:
//
//   The above copyright notice and this permission notice shall be
//   included in all copies or substantial portions of the Software.
//
//   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
//   EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
//   MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
//   NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS
//   BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN
//   ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
//   CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
//   SOFTWARE.
//

package parser

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"

	"github.com/hattya/go.sh/ast"
)
%}

%union {
	list     *ast.List
	pipeline *ast.Pipeline
	cmd      *ast.Cmd
	expr     ast.CmdExpr
	token    token
	elt      *element
	item     *ast.CaseItem
	items    []*ast.CaseItem
	else_    []ast.ElsePart
	cmds     []ast.Command
	redir    *ast.Redir
	redirs   []*ast.Redir
	word     ast.Word
	words    []ast.Word
}

%token<token> AND OR '|' '(' ')' LAE RAE BREAK '&' ';'
%token<token> '<' '>' CLOBBER APPEND DUPIN DUPOUT RDWR
%token<word>  IO_NUMBER
%token<word>  WORD NAME ASSIGNMENT_WORD
%token<token> Bang Lbrace Rbrace For Case Esac In If Elif Then Else Fi While Until Do Done

%type<list>     and_or
%type<pipeline> pipeline pipe_seq
%type<cmd>      cmd func_def
%type<elt>      simple_cmd cmd_prefix cmd_suffix
%type<expr>     compound_cmd subshell group arith_eval for_clause case_clause if_clause while_clause until_clause func_body
%type<words>    word_list pattern_list
%type<item>     case_item case_item_ns
%type<items>    case_list case_list_ns
%type<else_>    else_part
%type<cmds>     compound_list term
%type<redir>    io_redir io_file
%type<redirs>   redir_list
%type<token>    redir_op sep seq_sep sep_op

%left  '&' ';'
%left  AND OR
%right '|'

%%

cmdline:
		and_or sep_op
		{
			$1.SepPos = $2.pos
			$1.Sep = $2.val
			yylex.(*lexer).cmd = $1
		}
	|	and_or
		{
			yylex.(*lexer).cmd = extract($1)
		}
	|	/* empty */

and_or:
		           pipeline
		{
			$$ = &ast.List{Pipeline: $1}
		}
	|	and_or AND pipeline
		{
			$$.List = append($$.List, &ast.AndOr{
				OpPos:    $2.pos,
				Op:       $2.val,
				Pipeline: $3,
			})
		}
	|	and_or OR  pipeline
		{
			$$.List = append($$.List, &ast.AndOr{
				OpPos:    $2.pos,
				Op:       $2.val,
				Pipeline: $3,
			})
		}

pipeline:
		     pipe_seq
	|	Bang pipe_seq
		{
			$$ = $2
			$$.Bang = $1.pos
		}

pipe_seq:
		             cmd
		{
			$$ = &ast.Pipeline{Cmd: $1}
		}
	|	pipe_seq '|' cmd
		{
			$$.List = append($$.List, &ast.Pipe{
				OpPos: $2.pos,
				Op:    $2.val,
				Cmd:   $3,
			})
		}

cmd:
		simple_cmd
		{
			$$ = &ast.Cmd{
				Expr: &ast.SimpleCmd{
					Assigns: $1.assigns,
					Args:    $1.args,
				},
				Redirs: $1.redirs,
			}
		}
	|	compound_cmd
		{
			$$ = &ast.Cmd{Expr: $1}
		}
	|	compound_cmd redir_list
		{
			$$ = &ast.Cmd{
				Expr:   $1,
				Redirs: $2,
			}
		}
	|	func_def

simple_cmd:
		cmd_prefix WORD cmd_suffix
		{
			$$ = &element{
				redirs:  append($1.redirs, $3.redirs...),
				assigns: $1.assigns,
			}
			$$.args = append($$.args, $2)
			$$.args = append($$.args, $3.args...)
		}
	|	cmd_prefix WORD
		{
			$$ = $1
			$$.args = append($$.args, $2)
		}
	|	cmd_prefix
	|	           WORD cmd_suffix
		{
			$$ = &element{redirs: $2.redirs}
			$$.args = append($$.args, $1)
			$$.args = append($$.args, $2.args...)
		}
	|	           WORD
		{
			$$ = new(element)
			$$.args = append($$.args, $1)
		}

cmd_prefix:
		           io_redir
		{
			$$ = new(element)
			$$.redirs = append($$.redirs, $1)
		}
	|	cmd_prefix io_redir
		{
			$$.redirs = append($$.redirs, $2)
		}
	|	           ASSIGNMENT_WORD
		{
			$$ = new(element)
			$$.assigns = append($$.assigns, assign($1))
		}
	|	cmd_prefix ASSIGNMENT_WORD
		{
			$$.assigns = append($$.assigns, assign($2))
		}

cmd_suffix:
		           io_redir
		{
			$$ = new(element)
			$$.redirs = append($$.redirs, $1)
		}
	|	cmd_suffix io_redir
		{
			$$.redirs = append($$.redirs, $2)
		}
	|	           WORD
		{
			$$ = new(element)
			$$.args = append($$.args, $1)
		}
	|	cmd_suffix WORD
		{
			$$.args = append($$.args, $2)
		}

compound_cmd:
		subshell
	|	group
	|	arith_eval
	|	for_clause
	|	case_clause
	|	if_clause
	|	while_clause
	|	until_clause

subshell:
		'(' compound_list ')'
		{
			$$ = &ast.Subshell{
				Lparen: $1.pos,
				List:   $2,
				Rparen: $3.pos,
			}
		}

group:
		Lbrace compound_list Rbrace
		{
			$$ = &ast.Group{
				Lbrace: $1.pos,
				List:   $2,
				Rbrace: $3.pos,
			}
		}

arith_eval:
		LAE word_list RAE
		{
			$$ = &ast.ArithEval{
				Left:  $1.pos,
				Expr:  $2,
				Right: $3.pos,
			}
		}

for_clause:
		For NAME                                Do compound_list Done
		{
			$$ = &ast.ForClause{
				For:  $1.pos,
				Name: $2[0].(*ast.Lit),
				Do:   $3.pos,
				List: $4,
				Done: $5.pos,
			}
		}
	|	For NAME                        seq_sep Do compound_list Done
		{
			$$ = &ast.ForClause{
				For:       $1.pos,
				Name:      $2[0].(*ast.Lit),
				Semicolon: $3.pos,
				Do:        $4.pos,
				List:      $5,
				Done:      $6.pos,
			}
		}
	|	For NAME linebreak In           seq_sep Do compound_list Done
		{
			$$ = &ast.ForClause{
				For:       $1.pos,
				Name:      $2[0].(*ast.Lit),
				In:        $4.pos,
				Semicolon: $5.pos,
				Do:        $6.pos,
				List:      $7,
				Done:      $8.pos,
			}
		}
	|	For NAME linebreak In word_list seq_sep Do compound_list Done
		{
			$$ = &ast.ForClause{
				For:       $1.pos,
				Name:      $2[0].(*ast.Lit),
				In:        $4.pos,
				Items:     $5,
				Semicolon: $6.pos,
				Do:        $7.pos,
				List:      $8,
				Done:      $9.pos,
			}
		}

word_list:
		          WORD
		{
			$$ = append($$, $1)
		}
	|	word_list WORD
		{
			$$ = append($$, $2)
		}

case_clause:
		Case WORD linebreak In linebreak              Esac
		{
			$$ = &ast.CaseClause{
				Case: $1.pos,
				Word: $2,
				In:   $4.pos,
				Esac: $6.pos,
			}
		}
	|	Case WORD linebreak In linebreak case_list    Esac
		{
			$$ = &ast.CaseClause{
				Case:  $1.pos,
				Word:  $2,
				In:    $4.pos,
				Items: $6,
				Esac:  $7.pos,
			}
		}
	|	Case WORD linebreak In linebreak case_list_ns Esac
		{
			$$ = &ast.CaseClause{
				Case:  $1.pos,
				Word:  $2,
				In:    $4.pos,
				Items: $6,
				Esac:  $7.pos,
			}
		}

case_list:
		          case_item
		{
			$$ = append($$, $1)
		}
	|	case_list case_item
		{
			$$ = append($$, $2)
		}

case_item:
		    pattern_list ')' linebreak     BREAK linebreak
		{
			$$ = &ast.CaseItem{
				Patterns: $1,
				Rparen:   $2.pos,
				Break:    $4.pos,
			}
		}
	|	    pattern_list ')' compound_list BREAK linebreak
		{
			$$ = &ast.CaseItem{
				Patterns: $1,
				Rparen:   $2.pos,
				List:     $3,
				Break:    $4.pos,
			}
		}
	|	'(' pattern_list ')' linebreak     BREAK linebreak
		{
			$$ = &ast.CaseItem{
				Lparen:   $1.pos,
				Patterns: $2,
				Rparen:   $3.pos,
				Break:    $5.pos,
			}
		}
	|	'(' pattern_list ')' compound_list BREAK linebreak
		{
			$$ = &ast.CaseItem{
				Lparen:   $1.pos,
				Patterns: $2,
				Rparen:   $3.pos,
				List:     $4,
				Break:    $5.pos,
			}
		}

case_list_ns:
		          case_item_ns
		{
			$$ = append($$, $1)
		}
	|	case_list case_item_ns
		{
			$$ = append($$, $2)
		}

case_item_ns:
		    pattern_list ')' linebreak
		{
			$$ = &ast.CaseItem{
				Patterns: $1,
				Rparen:   $2.pos,
			}
		}
	|	    pattern_list ')' compound_list
		{
			$$ = &ast.CaseItem{
				Patterns: $1,
				Rparen:   $2.pos,
				List:     $3,
			}
		}
	|	'(' pattern_list ')' linebreak
		{
			$$ = &ast.CaseItem{
				Lparen:   $1.pos,
				Patterns: $2,
				Rparen:   $3.pos,
			}
		}
	|	'(' pattern_list ')' compound_list
		{
			$$ = &ast.CaseItem{
				Lparen:   $1.pos,
				Patterns: $2,
				Rparen:   $3.pos,
				List:     $4,
			}
		}

pattern_list:
		                 WORD
		{
			$$ = append($$, $1)
		}
	|	pattern_list '|' WORD
		{
			$$ = append($$, $3)
		}

if_clause:
		If compound_list Then compound_list           Fi
		{
			$$ = &ast.IfClause{
				If:   $1.pos,
				Cond: $2,
				Then: $3.pos,
				List: $4,
				Fi:   $5.pos,
			}
		}
	|	If compound_list Then compound_list else_part Fi
		{
			$$ = &ast.IfClause{
				If:   $1.pos,
				Cond: $2,
				Then: $3.pos,
				List: $4,
				Else: $5,
				Fi:   $6.pos,
			}
		}

else_part:
		Elif compound_list Then compound_list
		{
			$$ = append($$, &ast.ElifClause{
				Elif: $1.pos,
				Cond: $2,
				Then: $3.pos,
				List: $4,
			})
		}
	|	Elif compound_list Then compound_list else_part
		{
			$$ = append($$, &ast.ElifClause{
				Elif: $1.pos,
				Cond: $2,
				Then: $3.pos,
				List: $4,
			})
			$$ = append($$, $5...)
		}
	|	Else compound_list
		{
			$$ = append($$, &ast.ElseClause{
				Else: $1.pos,
				List: $2,
			})
		}

while_clause:
		While compound_list Do compound_list Done
		{
			$$ = &ast.WhileClause{
				While: $1.pos,
				Cond:  $2,
				Do:    $3.pos,
				List:  $4,
				Done:  $5.pos,
			}
		}

until_clause:
		Until compound_list Do compound_list Done
		{
			$$ = &ast.UntilClause{
				Until: $1.pos,
				Cond:  $2,
				Do:    $3.pos,
				List:  $4,
				Done:  $5.pos,
			}
		}

func_def:
		NAME '(' ')' linebreak func_body
		{
			x := $5.(*ast.FuncDef)
			x.Name = $1[0].(*ast.Lit)
			x.Lparen = $2.pos
			x.Rparen = $3.pos
			$$ = &ast.Cmd{Expr: $5}
		}

func_body:
		compound_cmd
		{
			$$ = &ast.FuncDef{
				Body: &ast.Cmd{
					Expr: $1,
				},
			}
		}
	|	compound_cmd redir_list
		{
			$$ = &ast.FuncDef{
				Body: &ast.Cmd{
					Expr:   $1,
					Redirs: $2,
				},
			}
		}

compound_list:
		linebreak term sep
		{
			cmd := $2[len($2)-1].(*ast.List)
			if $3.typ != '\n' {
				cmd.SepPos = $3.pos
				cmd.Sep = $3.val
			} else {
				$2[len($2)-1] = extract(cmd)
			}
			$$ = $2
		}
	|	linebreak term
		{
			$2[len($2)-1] = extract($2[len($2)-1].(*ast.List))
			$$ = $2
		}

term:
		         and_or
		{
			$$ = []ast.Command{$1}
		}
	|	term sep and_or
		{
			cmd := $$[len($$)-1].(*ast.List)
			if $2.typ != '\n' {
				cmd.SepPos = $2.pos
				cmd.Sep = $2.val
			} else {
				$$[len($$)-1] = extract(cmd)
			}
			$$ = append($$, $3)
		}

redir_list:
		           io_redir
		{
			$$ = append($$, $1)
		}
	|	redir_list io_redir
		{
			$$ = append($$, $2)
		}

io_redir:
		          io_file
	|	IO_NUMBER io_file
		{
			$$ = $2
			$$.N = $1[0].(*ast.Lit)
		}

io_file:
		redir_op WORD
		{
			$$ = &ast.Redir{
				OpPos: $1.pos,
				Op:    $1.val,
				Word:  $2,
			}
		}

redir_op:
		'<'
	|	'>'
	|	CLOBBER
	|	APPEND
	|	DUPIN
	|	DUPOUT
	|	RDWR

sep:
		sep_op linebreak
	|	newline_list
		{
		}

seq_sep:
		';' linebreak
	|	newline_list
		{
			$$.pos = ast.Pos{}
		}

sep_op:
		'&'
	|	';'

linebreak:
		newline_list
	|	/* empty */

newline_list:
		             '\n'
	|	newline_list '\n'

%%

func init() {
	yyErrorVerbose = true

	for i, s := range yyToknames {
		switch s {
		case "$end":
			s = "EOF"
		case "AND":
			s = "'&&'"
		case "OR":
			s = "'||'"
		case "LAE":
			s = "(("
		case "RAE":
			s = "))"
		case "BREAK":
			s = "';;'"
		case "CLOBBER":
			s = "'>|'"
		case "APPEND":
			s = "'>>'"
		case "DUPIN":
			s = "'<&'"
		case "DUPOUT":
			s = "'>&'"
		case "RDWR":
			s = "'<>'"
		case "Bang":
			s = "'!'"
		case "Lbrace":
			s ="'{'"
		case "Rbrace":
			s = "'}'"
		case "For":
			s = "'for'"
		case "Case":
			s = "'case'"
		case "Esac":
			s = "'esac'"
		case "In":
			s = "'in'"
		case "If":
			s = "'if'"
		case "Elif":
			s = "'elif'"
		case "Then":
			s = "'then'"
		case "Else":
			s = "'else'"
		case "Fi":
			s = "'fi'"
		case "While":
			s = "'while'"
		case "Until":
			s = "'until'"
		case "Do":
			s = "'do'"
		case "Done":
			s = "'done'"
		}
		yyToknames[i] = s
	}
}

type element struct {
	redirs  []*ast.Redir
	assigns []*ast.Assign
	args    []ast.Word
}

func extract(cmd *ast.List) ast.Command {
	switch {
	case len(cmd.List) != 0 || !cmd.SepPos.IsZero():
		return cmd
	case !cmd.Pipeline.Bang.IsZero() || len(cmd.Pipeline.List) != 0:
		return cmd.Pipeline
	}
	return cmd.Pipeline.Cmd
}

func assign(w ast.Word) *ast.Assign{
	k := w[0].(*ast.Lit)
	i := strings.IndexRune(k.Value, '=')
	v := w[1:]
	if i < len(k.Value)-1 {
		v = append(ast.Word{
			&ast.Lit{
				ValuePos: ast.NewPos(k.ValuePos.Line(), k.ValuePos.Col()+i+1),
				Value:    k.Value[i+1:],
			},
		}, v...)
		k.Value = k.Value[:i]
	}
	return &ast.Assign{
		Symbol: k,
		Op:     "=",
		Value:  v,
	}
}

// ParseCommand parses src and returns a command.
func ParseCommand(name string, src interface{}) (ast.Command, []*ast.Comment, error) {
	r, err := open(src)
	if err != nil {
		return nil, nil,err
	}

	l := newLexer(name, r)
	yyParse(l)
	return l.cmd, l.comments, l.err
}

func open(src interface{}) (r io.RuneScanner, err error) {
	switch src := src.(type) {
	case []byte:
		r = bytes.NewReader(src)
	case string:
		r = strings.NewReader(src)
	case io.RuneScanner:
		r = src
	case io.Reader:
		r = bufio.NewReader(src)
	default:
		err = errors.New("invalid source")
	}
	return
}
