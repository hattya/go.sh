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
	list  *ast.List
	token token
	word  ast.Word
	words []ast.Word
}

%token<token> AND OR '&' ';'
%token<word>  WORD

%type<list>  and_or
%type<token> sep_op
%type<words> pipeline

%left '&' ';'
%left AND OR

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
			yylex.(*lexer).cmd = $1
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
		         WORD
		{
			$$ = append($$, $1)
		}
	|	pipeline WORD
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
		}
		yyToknames[i] = s
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
