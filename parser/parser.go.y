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
	cmds     []ast.Command
	redir    *ast.Redir
	redirs   []*ast.Redir
	word     ast.Word
}

%token<token> AND OR '|' '(' ')' '&' ';'
%token<token> '<' '>' CLOBBER APPEND DUPIN DUPOUT RDWR
%token<word>  IO_NUMBER
%token<word>  WORD ASSIGNMENT_WORD
%token<token> Bang Lbrace Rbrace

%type<list>     and_or
%type<pipeline> pipeline pipe_seq
%type<cmd>      cmd
%type<elt>      simple_cmd cmd_prefix cmd_suffix
%type<expr>     compound_cmd subshell group
%type<cmds>     compound_list term
%type<redir>    io_redir io_file
%type<redirs>   redir_list
%type<token>    redir_op sep sep_op

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
