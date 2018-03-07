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

func TestList(t *testing.T) {
	var c ast.Command = new(ast.List)
	if g, e := c.Pos(), ast.NewPos(0, 0); e != g {
		t.Errorf("List.Pos() = %v, expected %v", g, e)
	}
	if g, e := c.End(), ast.NewPos(0, 0); e != g {
		t.Errorf("List.End() = %v, expected %v", g, e)
	}

	c = &ast.List{
		Pipeline: &ast.Pipeline{
			Cmd: []ast.Word{
				{
					&ast.Lit{
						ValuePos: ast.NewPos(1, 1),
						Value:    "lit",
					},
				},
			},
		},
	}
	if g, e := c.Pos(), ast.NewPos(1, 1); e != g {
		t.Errorf("List.Pos() = %v, expected %v", g, e)
	}
	if g, e := c.End(), ast.NewPos(1, 4); e != g {
		t.Errorf("List.End() = %v, expected %v", g, e)
	}

	c = &ast.List{
		Pipeline: &ast.Pipeline{
			Cmd: []ast.Word{
				{
					&ast.Lit{
						ValuePos: ast.NewPos(1, 1),
						Value:    "lit",
					},
				},
			},
		},
		List: []*ast.AndOr{
			{
				OpPos: ast.NewPos(1, 5),
				Op:    "&&",
				Pipeline: &ast.Pipeline{
					Cmd: []ast.Word{
						{
							&ast.Lit{
								ValuePos: ast.NewPos(1, 8),
								Value:    "lit",
							},
						},
					},
				},
			},
		},
	}
	if g, e := c.Pos(), ast.NewPos(1, 1); e != g {
		t.Errorf("List.Pos() = %v, expected %v", g, e)
	}
	if g, e := c.End(), ast.NewPos(1, 11); e != g {
		t.Errorf("List.End() = %v, expected %v", g, e)
	}
}

func TestAndOr(t *testing.T) {
	var n ast.Node = new(ast.AndOr)
	if g, e := n.Pos(), ast.NewPos(0, 0); e != g {
		t.Errorf("AndOr.Pos() = %v, expected %v", g, e)
	}
	if g, e := n.End(), ast.NewPos(0, 0); e != g {
		t.Errorf("AndOr.End() = %v, expected %v", g, e)
	}

	n = &ast.AndOr{
		OpPos: ast.NewPos(1, 1),
		Op:    "&&",
		Pipeline: &ast.Pipeline{
			Cmd: []ast.Word{
				{
					&ast.Lit{
						ValuePos: ast.NewPos(1, 4),
						Value:    "lit",
					},
				},
			},
		},
	}
	if g, e := n.Pos(), ast.NewPos(1, 1); e != g {
		t.Errorf("AndOr.Pos() = %v, expected %v", g, e)
	}
	if g, e := n.End(), ast.NewPos(1, 7); e != g {
		t.Errorf("AndOr.End() = %v, expected %v", g, e)
	}

	n = &ast.AndOr{
		OpPos: ast.NewPos(1, 1),
		Op:    "||",
		Pipeline: &ast.Pipeline{
			Cmd: []ast.Word{
				{
					&ast.Lit{
						ValuePos: ast.NewPos(1, 4),
						Value:    "lit",
					},
				},
			},
		},
	}
	if g, e := n.Pos(), ast.NewPos(1, 1); e != g {
		t.Errorf("AndOr.Pos() = %v, expected %v", g, e)
	}
	if g, e := n.End(), ast.NewPos(1, 7); e != g {
		t.Errorf("AndOr.End() = %v, expected %v", g, e)
	}
}

func TestPipeline(t *testing.T) {
	var c ast.Command = new(ast.Pipeline)
	if g, e := c.Pos(), ast.NewPos(0, 0); e != g {
		t.Errorf("Pipeline.Pos() = %v, expected %v", g, e)
	}
	if g, e := c.End(), ast.NewPos(0, 0); e != g {
		t.Errorf("Pipeline.End() = %v, expected %v", g, e)
	}

	c = &ast.Pipeline{
		Bang: ast.NewPos(1, 1),
		Cmd: []ast.Word{
			{
				&ast.Lit{
					ValuePos: ast.NewPos(1, 2),
					Value:    "lit",
				},
			},
		},
	}
	if g, e := c.Pos(), ast.NewPos(1, 1); e != g {
		t.Errorf("Pipeline.Pos() = %v, expected %v", g, e)
	}
	if g, e := c.End(), ast.NewPos(1, 5); e != g {
		t.Errorf("Pipeline.End() = %v, expected %v", g, e)
	}

	c = &ast.Pipeline{
		Cmd: []ast.Word{
			{
				&ast.Lit{
					ValuePos: ast.NewPos(1, 1),
					Value:    "lit",
				},
			},
		},
		List: []*ast.Pipe{
			{
				OpPos: ast.NewPos(1, 5),
				Op:    "|",
				Cmd: []ast.Word{
					{
						&ast.Lit{
							ValuePos: ast.NewPos(1, 7),
							Value:    "lit",
						},
					},
				},
			},
		},
	}
	if g, e := c.Pos(), ast.NewPos(1, 1); e != g {
		t.Errorf("Pipeline.Pos() = %v, expected %v", g, e)
	}
	if g, e := c.End(), ast.NewPos(1, 10); e != g {
		t.Errorf("Pipeline.End() = %v, expected %v", g, e)
	}
}

func TestPipe(t *testing.T) {
	var n ast.Node = new(ast.Pipe)
	if g, e := n.Pos(), ast.NewPos(0, 0); e != g {
		t.Errorf("Pipe.Pos() = %v, expected %v", g, e)
	}
	if g, e := n.End(), ast.NewPos(0, 0); e != g {
		t.Errorf("Pipe.End() = %v, expected %v", g, e)
	}

	n = &ast.Pipe{
		OpPos: ast.NewPos(1, 1),
		Op:    "|",
		Cmd: []ast.Word{
			{
				&ast.Lit{
					ValuePos: ast.NewPos(1, 3),
					Value:    "lit",
				},
			},
		},
	}
	if g, e := n.Pos(), ast.NewPos(1, 1); e != g {
		t.Errorf("Pipe.Pos() = %v, expected %v", g, e)
	}
	if g, e := n.End(), ast.NewPos(1, 6); e != g {
		t.Errorf("Pipe.End() = %v, expected %v", g, e)
	}
}

func TestCmd(t *testing.T) {
	var c ast.Command = new(ast.Cmd)
	if g, e := c.Pos(), ast.NewPos(0, 0); e != g {
		t.Errorf("Cmd.Pos() = %v, expected %v", g, e)
	}
	if g, e := c.End(), ast.NewPos(0, 0); e != g {
		t.Errorf("Cmd.End() = %v, expected %v", g, e)
	}
}

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

func TestComment(t *testing.T) {
	var n ast.Node = new(ast.Comment)
	if g, e := n.Pos(), ast.NewPos(0, 0); e != g {
		t.Errorf("Comment.Pos() = %v, expected %v", g, e)
	}
	if g, e := n.End(), ast.NewPos(0, 0); e != g {
		t.Errorf("Comment.End() = %v, expected %v", g, e)
	}

	n = &ast.Comment{
		Hash: ast.NewPos(1, 1),
		Text: " comment",
	}
	if g, e := n.Pos(), ast.NewPos(1, 1); e != g {
		t.Errorf("Comment.Pos() = %v, expected %v", g, e)
	}
	if g, e := n.End(), ast.NewPos(1, 9); e != g {
		t.Errorf("Comment.End() = %v, expected %v", g, e)
	}
}
