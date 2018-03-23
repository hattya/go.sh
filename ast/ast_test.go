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
			Cmd: &ast.Cmd{
				Expr: &ast.SimpleCmd{
					Args: []ast.Word{
						{
							&ast.Lit{
								ValuePos: ast.NewPos(1, 1),
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
	if g, e := c.End(), ast.NewPos(1, 4); e != g {
		t.Errorf("List.End() = %v, expected %v", g, e)
	}

	c = &ast.List{
		Pipeline: &ast.Pipeline{
			Cmd: &ast.Cmd{
				Expr: &ast.SimpleCmd{
					Args: []ast.Word{
						{
							&ast.Lit{
								ValuePos: ast.NewPos(1, 1),
								Value:    "lit",
							},
						},
					},
				},
			},
		},
		List: []*ast.AndOr{
			{
				OpPos: ast.NewPos(1, 5),
				Op:    "&&",
				Pipeline: &ast.Pipeline{
					Cmd: &ast.Cmd{
						Expr: &ast.SimpleCmd{
							Args: []ast.Word{
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
			Cmd: &ast.Cmd{
				Expr: &ast.SimpleCmd{
					Args: []ast.Word{
						{
							&ast.Lit{
								ValuePos: ast.NewPos(1, 4),
								Value:    "lit",
							},
						},
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
			Cmd: &ast.Cmd{
				Expr: &ast.SimpleCmd{
					Args: []ast.Word{
						{
							&ast.Lit{
								ValuePos: ast.NewPos(1, 4),
								Value:    "lit",
							},
						},
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
		Cmd: &ast.Cmd{
			Expr: &ast.SimpleCmd{
				Args: []ast.Word{
					{
						&ast.Lit{
							ValuePos: ast.NewPos(1, 2),
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
	if g, e := c.End(), ast.NewPos(1, 5); e != g {
		t.Errorf("Pipeline.End() = %v, expected %v", g, e)
	}

	c = &ast.Pipeline{
		Cmd: &ast.Cmd{
			Expr: &ast.SimpleCmd{
				Args: []ast.Word{
					{
						&ast.Lit{
							ValuePos: ast.NewPos(1, 1),
							Value:    "lit",
						},
					},
				},
			},
		},
		List: []*ast.Pipe{
			{
				OpPos: ast.NewPos(1, 5),
				Op:    "|",
				Cmd: &ast.Cmd{
					Expr: &ast.SimpleCmd{
						Args: []ast.Word{
							{
								&ast.Lit{
									ValuePos: ast.NewPos(1, 7),
									Value:    "lit",
								},
							},
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
		Cmd: &ast.Cmd{
			Expr: &ast.SimpleCmd{
				Args: []ast.Word{
					{
						&ast.Lit{
							ValuePos: ast.NewPos(1, 3),
							Value:    "lit",
						},
					},
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

	c = &ast.Cmd{
		Redirs: []*ast.Redir{
			{
				OpPos: ast.NewPos(1, 1),
				Op:    ">",
				Word: ast.Word{
					&ast.Lit{
						ValuePos: ast.NewPos(1, 2),
						Value:    "lit",
					},
				},
			},
		},
	}
	if g, e := c.Pos(), ast.NewPos(1, 1); e != g {
		t.Errorf("Cmd.Pos() = %v, expected %v", g, e)
	}
	if g, e := c.End(), ast.NewPos(1, 5); e != g {
		t.Errorf("Cmd.End() = %v, expected %v", g, e)
	}

	c = &ast.Cmd{
		Expr: &ast.SimpleCmd{
			Args: []ast.Word{
				{
					&ast.Lit{
						ValuePos: ast.NewPos(1, 1),
						Value:    "lit",
					},
				},
			},
		},
		Redirs: []*ast.Redir{
			{
				OpPos: ast.NewPos(1, 5),
				Op:    ">",
				Word: ast.Word{
					&ast.Lit{
						ValuePos: ast.NewPos(1, 6),
						Value:    "lit",
					},
				},
			},
		},
	}
	if g, e := c.Pos(), ast.NewPos(1, 1); e != g {
		t.Errorf("Cmd.Pos() = %v, expected %v", g, e)
	}
	if g, e := c.End(), ast.NewPos(1, 9); e != g {
		t.Errorf("Cmd.End() = %v, expected %v", g, e)
	}

	c = &ast.Cmd{
		Expr: &ast.SimpleCmd{
			Args: []ast.Word{
				{
					&ast.Lit{
						ValuePos: ast.NewPos(1, 6),
						Value:    "lit",
					},
				},
			},
		},
		Redirs: []*ast.Redir{
			{
				OpPos: ast.NewPos(1, 1),
				Op:    ">",
				Word: ast.Word{
					&ast.Lit{
						ValuePos: ast.NewPos(1, 2),
						Value:    "lit",
					},
				},
			},
		},
	}
	if g, e := c.Pos(), ast.NewPos(1, 1); e != g {
		t.Errorf("Cmd.Pos() = %v, expected %v", g, e)
	}
	if g, e := c.End(), ast.NewPos(1, 9); e != g {
		t.Errorf("Cmd.End() = %v, expected %v", g, e)
	}
}

func TestSimpleCmd(t *testing.T) {
	var x ast.CmdExpr = new(ast.SimpleCmd)
	if g, e := x.Pos(), ast.NewPos(0, 0); e != g {
		t.Errorf("SimpleCmd.Pos() = %v, expected %v", g, e)
	}
	if g, e := x.End(), ast.NewPos(0, 0); e != g {
		t.Errorf("SimpleCmd.End() = %v, expected %v", g, e)
	}

	x = &ast.SimpleCmd{
		Args: []ast.Word{
			{
				&ast.Lit{
					ValuePos: ast.NewPos(1, 1),
					Value:    "lit",
				},
			},
		},
	}
	if g, e := x.Pos(), ast.NewPos(1, 1); e != g {
		t.Errorf("SimpleCmd.Pos() = %v, expected %v", g, e)
	}
	if g, e := x.End(), ast.NewPos(1, 4); e != g {
		t.Errorf("SimpleCmd.End() = %v, expected %v", g, e)
	}

	x = &ast.SimpleCmd{
		Assigns: []*ast.Assign{
			{
				Symbol: &ast.Lit{
					ValuePos: ast.NewPos(1, 1),
					Value:    "lit",
				},
				Op: "=",
				Value: ast.Word{
					&ast.Lit{
						ValuePos: ast.NewPos(1, 5),
						Value:    "lit",
					},
				},
			},
		},
	}
	if g, e := x.Pos(), ast.NewPos(1, 1); e != g {
		t.Errorf("SimpleCmd.Pos() = %v, expected %v", g, e)
	}
	if g, e := x.End(), ast.NewPos(1, 8); e != g {
		t.Errorf("SimpleCmd.End() = %v, expected %v", g, e)
	}

	x = &ast.SimpleCmd{
		Assigns: []*ast.Assign{
			{
				Symbol: &ast.Lit{
					ValuePos: ast.NewPos(1, 1),
					Value:    "lit",
				},
				Op: "=",
				Value: ast.Word{
					&ast.Lit{
						ValuePos: ast.NewPos(1, 5),
						Value:    "lit",
					},
				},
			},
		},
		Args: []ast.Word{
			{
				&ast.Lit{
					ValuePos: ast.NewPos(1, 9),
					Value:    "lit",
				},
			},
		},
	}
	if g, e := x.Pos(), ast.NewPos(1, 1); e != g {
		t.Errorf("SimpleCmd.Pos() = %v, expected %v", g, e)
	}
	if g, e := x.End(), ast.NewPos(1, 12); e != g {
		t.Errorf("SimpleCmd.End() = %v, expected %v", g, e)
	}

	var c ast.Command = &ast.Cmd{Expr: new(ast.SimpleCmd)}
	if g, e := c.Pos(), ast.NewPos(0, 0); e != g {
		t.Errorf("Cmd.Pos() = %v, expected %v", g, e)
	}
	if g, e := c.End(), ast.NewPos(0, 0); e != g {
		t.Errorf("Cmd.End() = %v, expected %v", g, e)
	}

	c = &ast.Cmd{Expr: x}
	if g, e := c.Pos(), ast.NewPos(1, 1); e != g {
		t.Errorf("Cmd.Pos() = %v, expected %v", g, e)
	}
	if g, e := c.End(), ast.NewPos(1, 12); e != g {
		t.Errorf("Cmd.End() = %v, expected %v", g, e)
	}
}

func TestAssign(t *testing.T) {
	var n ast.Node = new(ast.Assign)
	if g, e := n.Pos(), ast.NewPos(0, 0); e != g {
		t.Errorf("Assign.Pos() = %v, expected %v", g, e)
	}
	if g, e := n.End(), ast.NewPos(0, 0); e != g {
		t.Errorf("Assign.End() = %v, expected %v", g, e)
	}

	n = &ast.Assign{
		Symbol: &ast.Lit{
			ValuePos: ast.NewPos(1, 1),
			Value:    "lit",
		},
		Op: "=",
	}
	if g, e := n.Pos(), ast.NewPos(1, 1); e != g {
		t.Errorf("Assign.Pos() = %v, expected %v", g, e)
	}
	if g, e := n.End(), ast.NewPos(1, 5); e != g {
		t.Errorf("Assign.End() = %v, expected %v", g, e)
	}

	n = &ast.Assign{
		Symbol: &ast.Lit{
			ValuePos: ast.NewPos(1, 1),
			Value:    "lit",
		},
		Op: "=",
		Value: ast.Word{
			&ast.Lit{
				ValuePos: ast.NewPos(1, 5),
				Value:    "lit",
			},
		},
	}
	if g, e := n.Pos(), ast.NewPos(1, 1); e != g {
		t.Errorf("Assign.Pos() = %v, expected %v", g, e)
	}
	if g, e := n.End(), ast.NewPos(1, 8); e != g {
		t.Errorf("Assign.End() = %v, expected %v", g, e)
	}
}

func TestSubshell(t *testing.T) {
	var x ast.CmdExpr = new(ast.Subshell)
	if g, e := x.Pos(), ast.NewPos(0, 0); e != g {
		t.Errorf("Subshell.Pos() = %v, expected %v", g, e)
	}
	if g, e := x.End(), ast.NewPos(0, 0); e != g {
		t.Errorf("Subshell.End() = %v, expected %v", g, e)
	}

	x = &ast.Subshell{
		Lparen: ast.NewPos(1, 1),
		Rparen: ast.NewPos(1, 3),
	}
	if g, e := x.Pos(), ast.NewPos(1, 1); e != g {
		t.Errorf("Subshell.Pos() = %v, expected %v", g, e)
	}
	if g, e := x.End(), ast.NewPos(1, 4); e != g {
		t.Errorf("Subshell.End() = %v, expected %v", g, e)
	}
}

func TestGroup(t *testing.T) {
	var x ast.CmdExpr = new(ast.Group)
	if g, e := x.Pos(), ast.NewPos(0, 0); e != g {
		t.Errorf("Group.Pos() = %v, expected %v", g, e)
	}
	if g, e := x.End(), ast.NewPos(0, 0); e != g {
		t.Errorf("Group.End() = %v, expected %v", g, e)
	}

	x = &ast.Group{
		Lbrace: ast.NewPos(1, 1),
		Rbrace: ast.NewPos(1, 3),
	}
	if g, e := x.Pos(), ast.NewPos(1, 1); e != g {
		t.Errorf("Group.Pos() = %v, expected %v", g, e)
	}
	if g, e := x.End(), ast.NewPos(1, 4); e != g {
		t.Errorf("Group.End() = %v, expected %v", g, e)
	}
}

func TestIfClause(t *testing.T) {
	var x ast.CmdExpr = new(ast.IfClause)
	if g, e := x.Pos(), ast.NewPos(0, 0); e != g {
		t.Errorf("IfClause.Pos() = %v, expected %v", g, e)
	}
	if g, e := x.End(), ast.NewPos(0, 0); e != g {
		t.Errorf("IfClause.End() = %v, expected %v", g, e)
	}

	x = &ast.IfClause{
		If: ast.NewPos(1, 1),
		Fi: ast.NewPos(4, 1),
	}
	if g, e := x.Pos(), ast.NewPos(1, 1); e != g {
		t.Errorf("IfClause.Pos() = %v, expected %v", g, e)
	}
	if g, e := x.End(), ast.NewPos(4, 3); e != g {
		t.Errorf("IfClause.End() = %v, expected %v", g, e)
	}
}

func TestElseClause(t *testing.T) {
	var e ast.ElsePart = new(ast.ElseClause)
	if g, e := e.Pos(), ast.NewPos(0, 0); e != g {
		t.Errorf("ElseClause.Pos() = %v, expected %v", g, e)
	}
	if g, e := e.End(), ast.NewPos(0, 0); e != g {
		t.Errorf("ElseClause.End() = %v, expected %v", g, e)
	}

	e = &ast.ElseClause{
		Else: ast.NewPos(4, 1),
		List: []ast.Command{
			&ast.Cmd{
				Expr: &ast.SimpleCmd{
					Args: []ast.Word{
						{
							&ast.Lit{
								ValuePos: ast.NewPos(5, 3),
								Value:    "lit",
							},
						},
					},
				},
			},
		},
	}
	if g, e := e.Pos(), ast.NewPos(4, 1); e != g {
		t.Errorf("ElseClause.Pos() = %v, expected %v", g, e)
	}
	if g, e := e.End(), ast.NewPos(5, 6); e != g {
		t.Errorf("ElseClause.End() = %v, expected %v", g, e)
	}
}

func TestRedir(t *testing.T) {
	var n ast.Node = new(ast.Redir)
	if g, e := n.Pos(), ast.NewPos(0, 0); e != g {
		t.Errorf("Redir.Pos() = %v, expected %v", g, e)
	}
	if g, e := n.End(), ast.NewPos(0, 0); e != g {
		t.Errorf("Redir.End() = %v, expected %v", g, e)
	}

	n = &ast.Redir{
		OpPos: ast.NewPos(1, 1),
		Op:    ">",
		Word: ast.Word{
			&ast.Lit{
				ValuePos: ast.NewPos(1, 3),
				Value:    "lit",
			},
		},
	}
	if g, e := n.Pos(), ast.NewPos(1, 1); e != g {
		t.Errorf("Redir.Pos() = %v, expected %v", g, e)
	}
	if g, e := n.End(), ast.NewPos(1, 6); e != g {
		t.Errorf("Redir.End() = %v, expected %v", g, e)
	}

	n = &ast.Redir{
		N: &ast.Lit{
			ValuePos: ast.NewPos(1, 1),
			Value:    "1",
		},
		OpPos: ast.NewPos(1, 2),
		Op:    ">",
		Word: ast.Word{
			&ast.Lit{
				ValuePos: ast.NewPos(1, 4),
				Value:    "lit",
			},
		},
	}
	if g, e := n.Pos(), ast.NewPos(1, 1); e != g {
		t.Errorf("Redir.Pos() = %v, expected %v", g, e)
	}
	if g, e := n.End(), ast.NewPos(1, 7); e != g {
		t.Errorf("Redir.End() = %v, expected %v", g, e)
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

	w = &ast.Lit{
		ValuePos: ast.NewPos(1, 1),
		Value:    "1\n2",
	}
	if g, e := w.Pos(), ast.NewPos(1, 1); e != g {
		t.Errorf("Lit.Pos() = %v, expected %v", g, e)
	}
	if g, e := w.End(), ast.NewPos(2, 2); e != g {
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
		Tok:    `'`,
		Value: ast.Word{
			&ast.Lit{
				ValuePos: ast.NewPos(1, 2),
				Value:    "single\nquote",
			},
		},
	}
	if g, e := w.Pos(), ast.NewPos(1, 1); e != g {
		t.Errorf("Quote.Pos() = %v, expected %v", g, e)
	}
	if g, e := w.End(), ast.NewPos(2, 7); e != g {
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

	w = &ast.Quote{
		TokPos: ast.NewPos(1, 1),
		Tok:    `"`,
		Value: ast.Word{
			&ast.Lit{
				ValuePos: ast.NewPos(1, 2),
				Value:    "double\nquote",
			},
		},
	}
	if g, e := w.Pos(), ast.NewPos(1, 1); e != g {
		t.Errorf("Quote.Pos() = %v, expected %v", g, e)
	}
	if g, e := w.End(), ast.NewPos(2, 7); e != g {
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
