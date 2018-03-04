//
// go.sh/ast :: ast_test.go
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

package ast_test

import (
	"testing"

	"github.com/hattya/go.sh/ast"
)

func TestWord(t *testing.T) {
	var n ast.Node = ast.Word{}
	if g, e := n.Pos(), ast.NewPos(0, 0); e != g {
		t.Errorf("Word.Pos() = %v, expected %v", g, e)
	}
	if g, e := n.End(), ast.NewPos(0, 0); e != g {
		t.Errorf("Word.End() = %v, expected %v", g, e)
	}
	if g, e := len(n.(ast.Word)), 0; e != g {
		t.Errorf("len(Word) = %v, expected %v", g, e)
	}

	n = ast.Word{
		&ast.Lit{
			ValuePos: ast.NewPos(1, 1),
			Value:    "lit",
		},
	}
	if g, e := n.Pos(), ast.NewPos(1, 1); e != g {
		t.Errorf("Word.Pos() = %v, expected %v", g, e)
	}
	if g, e := n.End(), ast.NewPos(1, 4); e != g {
		t.Errorf("Word.End() = %v, expected %v", g, e)
	}
	if g, e := len(n.(ast.Word)), 1; e != g {
		t.Errorf("len(Word) = %v, expected %v", g, e)
	}
}

func TestLit(t *testing.T) {
	var w ast.WordPart = new(ast.Lit)
	if g, e := w.Pos(), ast.NewPos(0, 0); e != g {
		t.Errorf("Lit.Pos() = %v, expected %v", g, e)
	}
	if g, e := w.End(), ast.NewPos(0, 0); e != g {
		t.Errorf("Lit.End() = %v, expected %v", g, e)
	}

	w = &ast.Lit{
		ValuePos: ast.NewPos(1, 1),
		Value:    "lit",
	}
	if g, e := w.Pos(), ast.NewPos(1, 1); e != g {
		t.Errorf("Lit.Pos() = %v, expected %v", g, e)
	}
	if g, e := w.End(), ast.NewPos(1, 4); e != g {
		t.Errorf("Lit.End() = %v, expected %v", g, e)
	}
}

func TestQuote(t *testing.T) {
	var w ast.WordPart = new(ast.Quote)
	if g, e := w.Pos(), ast.NewPos(0, 0); e != g {
		t.Errorf("Quote.Pos() = %v, expected %v", g, e)
	}
	if g, e := w.End(), ast.NewPos(0, 0); e != g {
		t.Errorf("Quote.End() = %v, expected %v", g, e)
	}

	w = &ast.Quote{
		TokPos: ast.NewPos(1, 1),
		Tok:    `\`,
		Value: ast.Word{
			&ast.Lit{
				ValuePos: ast.NewPos(1, 2),
				Value:    "q",
			},
		},
	}
	if g, e := w.Pos(), ast.NewPos(1, 1); e != g {
		t.Errorf("Quote.Pos() = %v, expected %v", g, e)
	}
	if g, e := w.End(), ast.NewPos(1, 3); e != g {
		t.Errorf("Quote.End() = %v, expected %v", g, e)
	}

	w = &ast.Quote{
		TokPos: ast.NewPos(1, 1),
		Tok:    `'`,
		Value: ast.Word{
			&ast.Lit{
				ValuePos: ast.NewPos(1, 2),
				Value:    "quote",
			},
		},
	}
	if g, e := w.Pos(), ast.NewPos(1, 1); e != g {
		t.Errorf("Quote.Pos() = %v, expected %v", g, e)
	}
	if g, e := w.End(), ast.NewPos(1, 8); e != g {
		t.Errorf("Quote.End() = %v, expected %v", g, e)
	}

	w = &ast.Quote{
		TokPos: ast.NewPos(1, 1),
		Tok:    `"`,
		Value: ast.Word{
			&ast.Lit{
				ValuePos: ast.NewPos(1, 2),
				Value:    "quote",
			},
		},
	}
	if g, e := w.Pos(), ast.NewPos(1, 1); e != g {
		t.Errorf("Quote.Pos() = %v, expected %v", g, e)
	}
	if g, e := w.End(), ast.NewPos(1, 8); e != g {
		t.Errorf("Quote.End() = %v, expected %v", g, e)
	}
}
