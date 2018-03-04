//
// go.sh/parser :: lexer_test.go
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
	"testing"

	"github.com/hattya/go.sh/ast"
)

func TestToken(t *testing.T) {
	var n ast.Node = token{}
	if g, e := n.Pos(), ast.NewPos(0, 0); e != g {
		t.Errorf("token.Pos() = %v, expected %v", g, e)
	}
	if g, e := n.End(), ast.NewPos(0, 0); e != g {
		t.Errorf("token.End() = %v, expected %v", g, e)
	}

	n = token{
		typ: '&',
		pos: ast.NewPos(1, 1),
		val: "&",
	}
	if g, e := n.Pos(), ast.NewPos(1, 1); e != g {
		t.Errorf("token.Pos() = %v, expected %v", g, e)
	}
	if g, e := n.End(), ast.NewPos(1, 2); e != g {
		t.Errorf("token.End() = %v, expected %v", g, e)
	}
}
