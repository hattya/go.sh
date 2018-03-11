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
	assigns  []*ast.Assign
	word     ast.Word
	words    []ast.Word
}

%token<token> AND OR '|' '&' ';'
%token<word>  WORD ASSIGNMENT_WORD
%token<token> Bang

%type<list>     and_or
%type<pipeline> pipeline pipe_seq
%type<cmd>      cmd
%type<expr>     simple_cmd
%type<assigns>  cmd_prefix
%type<words>    cmd_suffix
%type<token>    sep_op

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
			l := yylex.(*lexer)
			switch {
			case len($1.List) != 0:
				l.cmd = $1
			case !$1.Pipeline.Bang.IsZero() || len($1.Pipeline.List) != 0:
				l.cmd = $1.Pipeline
			default:
				l.cmd = $1.Pipeline.Cmd
			}
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
			$$ = &ast.Cmd{Expr: $1}
		}

simple_cmd:
		cmd_prefix WORD cmd_suffix
		{
			$$ = &ast.SimpleCmd{
				Assigns: $1,
				Args:    append([]ast.Word{$2}, $3...),
			}
		}
	|	cmd_prefix WORD
		{
			$$ = &ast.SimpleCmd{
				Assigns: $1,
				Args:    []ast.Word{$2},
			}
		}
	|	cmd_prefix
		{
			$$ = &ast.SimpleCmd{Assigns: $1}
		}
	|	           WORD cmd_suffix
		{
			$$ = &ast.SimpleCmd{Args: append([]ast.Word{$1}, $2...)}
		}
	|	           WORD
		{
			$$ = &ast.SimpleCmd{Args: []ast.Word{$1}}
		}

cmd_prefix:
		           ASSIGNMENT_WORD
		{
			$$ = append($$, assign($1))
		}
	|	cmd_prefix ASSIGNMENT_WORD
		{
			$$ = append($$, assign($2))
		}

cmd_suffix:
		           WORD
		{
			$$ = append($$, $1)
		}
	|	cmd_suffix WORD
		{
			$$ = append($$, $2)
		}

sep_op:
		'&'
	|	';'

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
		case "Bang":
			s = "'!'"
		}
		yyToknames[i] = s
	}
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
