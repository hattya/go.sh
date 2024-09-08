%{
//
// go.sh/parser :: parser.go
//
//   Copyright (c) 2018-2024 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
//

package parser

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"

	"github.com/hattya/go.sh/ast"
	"github.com/hattya/go.sh/interp"
)
%}

%union {
	list  any
	node  ast.Node
	elt   *element
	word  ast.Word
	token token
}

%token<token> AND OR '|' '(' ')' LAE RAE BREAK '&' ';'
%token<token> '<' '>' CLOBBER APPEND HEREDOC HEREDOCI DUPIN DUPOUT RDWR
%token<word>  IO_NUMBER
%token<word>  WORD NAME ASSIGNMENT_WORD
%token<token> Bang Lbrace Rbrace For Case Esac In If Elif Then Else Fi While Until Do Done

%type<list>  complete_cmd list
%type<node>  and_or
%type<node>  pipeline pipe_seq
%type<node>  cmd func_def
%type<elt>   simple_cmd cmd_prefix cmd_suffix
%type<node>  compound_cmd subshell group arith_eval for_clause case_clause if_clause while_clause until_clause func_body
%type<list>  word_list pattern_list
%type<list>  case_list case_list_ns
%type<node>  case_item case_item_ns
%type<list>  else_part
%type<list>  compound_list term
%type<list>  redir_list
%type<node>  io_redir io_file io_here
%type<token> redir_op here_op sep seq_sep sep_op

%left  '&' ';'
%left  AND OR
%right '|'

%%

cmdline:
		complete_cmds linebreak
	|	/* empty */

complete_cmds:
		                           complete_cmd
		{
			l := $1.(ast.List)
			if len(l) > 1 {
				yylex.(*lexer).cmds = append(yylex.(*lexer).cmds, l)
			} else {
				yylex.(*lexer).cmds = append(yylex.(*lexer).cmds, extract(l[0]))
			}
		}
	|	complete_cmds newline_list complete_cmd
		{
			l := $3.(ast.List)
			if len(l) > 1 {
				yylex.(*lexer).cmds = append(yylex.(*lexer).cmds, l)
			} else {
				yylex.(*lexer).cmds = append(yylex.(*lexer).cmds, extract(l[0]))
			}
		}

complete_cmd:
		list sep_op
		{
			l := $1.(ast.List)
			ao := l[len(l)-1]
			ao.SepPos = $2.pos
			ao.Sep = $2.val
		}
	|	list

list:
		            and_or
		{
			$$ = ast.List{$1.(*ast.AndOrList)}
		}
	|	list sep_op and_or
		{
			l := $$.(ast.List)
			ao := l[len(l)-1]
			ao.SepPos = $2.pos
			ao.Sep = $2.val
			$$ = append(l, $3.(*ast.AndOrList))
		}

and_or:
		           pipeline
		{
			$$ = &ast.AndOrList{Pipeline: $1.(*ast.Pipeline)}
		}
	|	and_or AND pipeline
		{
			$$.(*ast.AndOrList).List = append($$.(*ast.AndOrList).List, &ast.AndOr{
				OpPos:    $2.pos,
				Op:       $2.val,
				Pipeline: $3.(*ast.Pipeline),
			})
		}
	|	and_or OR  pipeline
		{
			$$.(*ast.AndOrList).List = append($$.(*ast.AndOrList).List, &ast.AndOr{
				OpPos:    $2.pos,
				Op:       $2.val,
				Pipeline: $3.(*ast.Pipeline),
			})
		}

pipeline:
		     pipe_seq
	|	Bang pipe_seq
		{
			$$ = $2
			$$.(*ast.Pipeline).Bang = $1.pos
		}

pipe_seq:
		             cmd
		{
			$$ = &ast.Pipeline{Cmd: $1.(*ast.Cmd)}
		}
	|	pipe_seq '|' cmd
		{
			$$.(*ast.Pipeline).List = append($$.(*ast.Pipeline).List, &ast.Pipe{
				OpPos: $2.pos,
				Op:    $2.val,
				Cmd:   $3.(*ast.Cmd),
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
			$$ = &ast.Cmd{Expr: $1.(ast.CmdExpr)}
		}
	|	compound_cmd redir_list
		{
			$$ = &ast.Cmd{
				Expr:   $1.(ast.CmdExpr),
				Redirs: $2.([]*ast.Redir),
			}
		}
	|	func_def

simple_cmd:
		cmd_prefix WORD cmd_suffix
		{
			$$ = &element{
				redirs:  append($1.redirs, $3.redirs...),
				assigns: $1.assigns,
				args:    append([]ast.Word{$2}, $3.args...),
			}
		}
	|	cmd_prefix WORD
		{
			$$ = $1
			$$.args = append($$.args, $2)
		}
	|	cmd_prefix
	|	           WORD cmd_suffix
		{
			$$ = &element{
				redirs: $2.redirs,
				args:   append([]ast.Word{$1}, $2.args...),
			}
		}
	|	           WORD
		{
			$$ = &element{args: []ast.Word{$1}}
		}

cmd_prefix:
		           io_redir
		{
			$$ = &element{redirs: []*ast.Redir{$1.(*ast.Redir)}}
		}
	|	cmd_prefix io_redir
		{
			$$.redirs = append($$.redirs, $2.(*ast.Redir))
		}
	|	           ASSIGNMENT_WORD
		{
			$$ = &element{assigns: []*ast.Assign{assign($1)}}
		}
	|	cmd_prefix ASSIGNMENT_WORD
		{
			$$.assigns = append($$.assigns, assign($2))
		}

cmd_suffix:
		           io_redir
		{
			$$ = &element{redirs: []*ast.Redir{$1.(*ast.Redir)}}
		}
	|	cmd_suffix io_redir
		{
			$$.redirs = append($$.redirs, $2.(*ast.Redir))
		}
	|	           WORD
		{
			$$ = &element{args: []ast.Word{$1}}
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
				List:   $2.([]ast.Command),
				Rparen: $3.pos,
			}
		}

group:
		Lbrace compound_list Rbrace
		{
			$$ = &ast.Group{
				Lbrace: $1.pos,
				List:   $2.([]ast.Command),
				Rbrace: $3.pos,
			}
		}

arith_eval:
		LAE WORD RAE
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
				List: $4.([]ast.Command),
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
				List:      $5.([]ast.Command),
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
				List:      $7.([]ast.Command),
				Done:      $8.pos,
			}
		}
	|	For NAME linebreak In word_list seq_sep Do compound_list Done
		{
			$$ = &ast.ForClause{
				For:       $1.pos,
				Name:      $2[0].(*ast.Lit),
				In:        $4.pos,
				Items:     $5.([]ast.Word),
				Semicolon: $6.pos,
				Do:        $7.pos,
				List:      $8.([]ast.Command),
				Done:      $9.pos,
			}
		}

word_list:
		          WORD
		{
			$$ = []ast.Word{$1}
		}
	|	word_list WORD
		{
			$$ = append($$.([]ast.Word), $2)
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
				Items: $6.([]*ast.CaseItem),
				Esac:  $7.pos,
			}
		}
	|	Case WORD linebreak In linebreak case_list_ns Esac
		{
			$$ = &ast.CaseClause{
				Case:  $1.pos,
				Word:  $2,
				In:    $4.pos,
				Items: $6.([]*ast.CaseItem),
				Esac:  $7.pos,
			}
		}

case_list:
		          case_item
		{
			$$ = []*ast.CaseItem{$1.(*ast.CaseItem)}
		}
	|	case_list case_item
		{
			$$ = append($$.([]*ast.CaseItem), $2.(*ast.CaseItem))
		}

case_item:
		    pattern_list ')' linebreak     BREAK linebreak
		{
			$$ = &ast.CaseItem{
				Patterns: $1.([]ast.Word),
				Rparen:   $2.pos,
				Break:    $4.pos,
			}
		}
	|	    pattern_list ')' compound_list BREAK linebreak
		{
			$$ = &ast.CaseItem{
				Patterns: $1.([]ast.Word),
				Rparen:   $2.pos,
				List:     $3.([]ast.Command),
				Break:    $4.pos,
			}
		}
	|	'(' pattern_list ')' linebreak     BREAK linebreak
		{
			$$ = &ast.CaseItem{
				Lparen:   $1.pos,
				Patterns: $2.([]ast.Word),
				Rparen:   $3.pos,
				Break:    $5.pos,
			}
		}
	|	'(' pattern_list ')' compound_list BREAK linebreak
		{
			$$ = &ast.CaseItem{
				Lparen:   $1.pos,
				Patterns: $2.([]ast.Word),
				Rparen:   $3.pos,
				List:     $4.([]ast.Command),
				Break:    $5.pos,
			}
		}

case_list_ns:
		          case_item_ns
		{
			$$ = []*ast.CaseItem{$1.(*ast.CaseItem)}
		}
	|	case_list case_item_ns
		{
			$$ = append($$.([]*ast.CaseItem), $2.(*ast.CaseItem))
		}

case_item_ns:
		    pattern_list ')' linebreak
		{
			$$ = &ast.CaseItem{
				Patterns: $1.([]ast.Word),
				Rparen:   $2.pos,
			}
		}
	|	    pattern_list ')' compound_list
		{
			$$ = &ast.CaseItem{
				Patterns: $1.([]ast.Word),
				Rparen:   $2.pos,
				List:     $3.([]ast.Command),
			}
		}
	|	'(' pattern_list ')' linebreak
		{
			$$ = &ast.CaseItem{
				Lparen:   $1.pos,
				Patterns: $2.([]ast.Word),
				Rparen:   $3.pos,
			}
		}
	|	'(' pattern_list ')' compound_list
		{
			$$ = &ast.CaseItem{
				Lparen:   $1.pos,
				Patterns: $2.([]ast.Word),
				Rparen:   $3.pos,
				List:     $4.([]ast.Command),
			}
		}

pattern_list:
		                 WORD
		{
			$$ = []ast.Word{$1}
		}
	|	pattern_list '|' WORD
		{
			$$ = append($$.([]ast.Word), $3)
		}

if_clause:
		If compound_list Then compound_list           Fi
		{
			$$ = &ast.IfClause{
				If:   $1.pos,
				Cond: $2.([]ast.Command),
				Then: $3.pos,
				List: $4.([]ast.Command),
				Fi:   $5.pos,
			}
		}
	|	If compound_list Then compound_list else_part Fi
		{
			$$ = &ast.IfClause{
				If:   $1.pos,
				Cond: $2.([]ast.Command),
				Then: $3.pos,
				List: $4.([]ast.Command),
				Else: $5.([]ast.ElsePart),
				Fi:   $6.pos,
			}
		}

else_part:
		Elif compound_list Then compound_list
		{
			$$ = []ast.ElsePart{&ast.ElifClause{
				Elif: $1.pos,
				Cond: $2.([]ast.Command),
				Then: $3.pos,
				List: $4.([]ast.Command),
			}}
		}
	|	Elif compound_list Then compound_list else_part
		{
			$$ = append([]ast.ElsePart{&ast.ElifClause{
				Elif: $1.pos,
				Cond: $2.([]ast.Command),
				Then: $3.pos,
				List: $4.([]ast.Command),
			}}, $5.([]ast.ElsePart)...)
		}
	|	Else compound_list
		{
			$$ = []ast.ElsePart{&ast.ElseClause{
				Else: $1.pos,
				List: $2.([]ast.Command),
			}}
		}

while_clause:
		While compound_list Do compound_list Done
		{
			$$ = &ast.WhileClause{
				While: $1.pos,
				Cond:  $2.([]ast.Command),
				Do:    $3.pos,
				List:  $4.([]ast.Command),
				Done:  $5.pos,
			}
		}

until_clause:
		Until compound_list Do compound_list Done
		{
			$$ = &ast.UntilClause{
				Until: $1.pos,
				Cond:  $2.([]ast.Command),
				Do:    $3.pos,
				List:  $4.([]ast.Command),
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
			$$ = &ast.Cmd{Expr: $5.(ast.CmdExpr)}
		}

func_body:
		compound_cmd
		{
			$$ = &ast.FuncDef{
				Body: &ast.Cmd{
					Expr: $1.(ast.CmdExpr),
				},
			}
		}
	|	compound_cmd redir_list
		{
			$$ = &ast.FuncDef{
				Body: &ast.Cmd{
					Expr:   $1.(ast.CmdExpr),
					Redirs: $2.([]*ast.Redir),
				},
			}
		}

compound_list:
		linebreak term sep
		{
			cmds := $2.([]ast.Command)
			l := cmds[len(cmds)-1].(ast.List)
			if $3.typ != '\n' {
				ao := l[len(l)-1]
				ao.SepPos = $3.pos
				ao.Sep = $3.val
			}
			if len(l) == 1 {
				cmds[len(cmds)-1] = extract(l[0])
			}
			$$ = $2
		}
	|	linebreak term
		{
			cmds := $2.([]ast.Command)
			l := cmds[len(cmds)-1].(ast.List)
			if len(l) == 1 {
				cmds[len(cmds)-1] = extract(l[0])
			}
			$$ = $2
		}

term:
		         and_or
		{
			$$ = []ast.Command{ast.List{$1.(*ast.AndOrList)}}
		}
	|	term sep and_or
		{
			cmds := $$.([]ast.Command)
			l := cmds[len(cmds)-1].(ast.List)
			if $2.typ != '\n' {
				ao := l[len(l)-1]
				ao.SepPos = $2.pos
				ao.Sep = $2.val
				cmds[len(cmds)-1] = append(l, $3.(*ast.AndOrList))
			} else {
				if len(l) == 1 {
					cmds[len(cmds)-1] = extract(l[0])
				}
				$$ = append(cmds, ast.List{$3.(*ast.AndOrList)})
			}
		}

redir_list:
		           io_redir
		{
			$$ = []*ast.Redir{$1.(*ast.Redir)}
		}
	|	redir_list io_redir
		{
			$$ = append($$.([]*ast.Redir), $2.(*ast.Redir))
		}

io_redir:
		          io_file
	|	IO_NUMBER io_file
		{
			$$ = $2
			$$.(*ast.Redir).N = $1[0].(*ast.Lit)
		}
	|	          io_here
	|	IO_NUMBER io_here
		{
			$$ = $2
			$$.(*ast.Redir).N = $1[0].(*ast.Lit)
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

io_here:
		here_op WORD
		{
			$$ = &ast.Redir{
				OpPos:   $1.pos,
				Op:      $1.val,
				Word:    $2,
			}
			yylex.(*lexer).heredoc.push($$.(*ast.Redir))
		}

here_op:
		HEREDOC
	|	HEREDOCI

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
		case "HEREDOC":
			s = "'<<'"
		case "HEREDOCI":
			s = "'<<-'"
		case "DUPIN":
			s = "'<&'"
		case "DUPOUT":
			s = "'>&'"
		case "RDWR":
			s = "'<>'"
		case "Bang":
			s = "'!'"
		case "Lbrace":
			s = "'{'"
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

func extract(cmd *ast.AndOrList) ast.Command {
	switch {
	case len(cmd.List) != 0 || !cmd.SepPos.IsZero():
		return cmd
	case !cmd.Pipeline.Bang.IsZero() || len(cmd.Pipeline.List) != 0:
		return cmd.Pipeline
	}
	return cmd.Pipeline.Cmd
}

func assign(w ast.Word) *ast.Assign {
	n := w[0].(*ast.Lit)
	if i := strings.IndexRune(n.Value, '='); 0 < i && i < len(n.Value)-1 {
		w[0] = &ast.Lit{
			ValuePos: ast.NewPos(n.ValuePos.Line(), n.ValuePos.Col()+i+1),
			Value:    n.Value[i+1:],
		}
		n.Value = n.Value[:i]
	} else {
		w = w[1:]
		n.Value = n.Value[:len(n.Value)-1]
	}
	return &ast.Assign{
		Name:  n,
		Op:    "=",
		Value: w,
	}
}

// ParseCommands parses src, including alias substitution, and returns
// commands.
func ParseCommands(env *interp.ExecEnv, name string, src any) ([]ast.Command, []*ast.Comment, error) {
	r, err := open(src)
	if err != nil {
		return nil, nil, err
	}

	l := newLexer(env, name, r)
	yyParse(l)
	return l.cmds, l.comments, l.err
}

// ParseCommand parses src and returns a command.
func ParseCommand(name string, src any) (ast.Command, []*ast.Comment, error) {
	cmds, comments, err := ParseCommands(nil, name, src)
	if len(cmds) == 0 {
		return nil, comments, err
	}
	return cmds[0], comments, err
}

func open(src any) (r io.RuneScanner, err error) {
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
