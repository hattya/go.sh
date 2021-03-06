//
// go.sh/ast :: ast_test.go
//
//   Copyright (c) 2018-2020 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
//

package ast_test

import (
	"testing"

	"github.com/hattya/go.sh/ast"
)

func TestList(t *testing.T) {
	var c ast.Command = new(ast.List)
	if g, e := c.Pos(), ast.NewPos(0, 0); g != e {
		t.Errorf("List.Pos() = %v, expected %v", g, e)
	}
	if g, e := c.End(), ast.NewPos(0, 0); g != e {
		t.Errorf("List.End() = %v, expected %v", g, e)
	}

	c = ast.List{
		&ast.AndOrList{
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
		},
	}
	if g, e := c.Pos(), ast.NewPos(1, 1); g != e {
		t.Errorf("List.Pos() = %v, expected %v", g, e)
	}
	if g, e := c.End(), ast.NewPos(1, 4); g != e {
		t.Errorf("List.End() = %v, expected %v", g, e)
	}
}

func TestAndOrList(t *testing.T) {
	var c ast.Command = new(ast.AndOrList)
	if g, e := c.Pos(), ast.NewPos(0, 0); g != e {
		t.Errorf("AndOrList.Pos() = %v, expected %v", g, e)
	}
	if g, e := c.End(), ast.NewPos(0, 0); g != e {
		t.Errorf("AndOrList.End() = %v, expected %v", g, e)
	}

	c = &ast.AndOrList{
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
	if g, e := c.Pos(), ast.NewPos(1, 1); g != e {
		t.Errorf("AndOrList.Pos() = %v, expected %v", g, e)
	}
	if g, e := c.End(), ast.NewPos(1, 4); g != e {
		t.Errorf("AndOrList.End() = %v, expected %v", g, e)
	}

	c = &ast.AndOrList{
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
	if g, e := c.Pos(), ast.NewPos(1, 1); g != e {
		t.Errorf("AndOrList.Pos() = %v, expected %v", g, e)
	}
	if g, e := c.End(), ast.NewPos(1, 11); g != e {
		t.Errorf("AndOrList.End() = %v, expected %v", g, e)
	}
}

func TestAndOr(t *testing.T) {
	var n ast.Node = new(ast.AndOr)
	if g, e := n.Pos(), ast.NewPos(0, 0); g != e {
		t.Errorf("AndOr.Pos() = %v, expected %v", g, e)
	}
	if g, e := n.End(), ast.NewPos(0, 0); g != e {
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
	if g, e := n.Pos(), ast.NewPos(1, 1); g != e {
		t.Errorf("AndOr.Pos() = %v, expected %v", g, e)
	}
	if g, e := n.End(), ast.NewPos(1, 7); g != e {
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
	if g, e := n.Pos(), ast.NewPos(1, 1); g != e {
		t.Errorf("AndOr.Pos() = %v, expected %v", g, e)
	}
	if g, e := n.End(), ast.NewPos(1, 7); g != e {
		t.Errorf("AndOr.End() = %v, expected %v", g, e)
	}
}

func TestPipeline(t *testing.T) {
	var c ast.Command = new(ast.Pipeline)
	if g, e := c.Pos(), ast.NewPos(0, 0); g != e {
		t.Errorf("Pipeline.Pos() = %v, expected %v", g, e)
	}
	if g, e := c.End(), ast.NewPos(0, 0); g != e {
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
	if g, e := c.Pos(), ast.NewPos(1, 1); g != e {
		t.Errorf("Pipeline.Pos() = %v, expected %v", g, e)
	}
	if g, e := c.End(), ast.NewPos(1, 5); g != e {
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
	if g, e := c.Pos(), ast.NewPos(1, 1); g != e {
		t.Errorf("Pipeline.Pos() = %v, expected %v", g, e)
	}
	if g, e := c.End(), ast.NewPos(1, 10); g != e {
		t.Errorf("Pipeline.End() = %v, expected %v", g, e)
	}
}

func TestPipe(t *testing.T) {
	var n ast.Node = new(ast.Pipe)
	if g, e := n.Pos(), ast.NewPos(0, 0); g != e {
		t.Errorf("Pipe.Pos() = %v, expected %v", g, e)
	}
	if g, e := n.End(), ast.NewPos(0, 0); g != e {
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
	if g, e := n.Pos(), ast.NewPos(1, 1); g != e {
		t.Errorf("Pipe.Pos() = %v, expected %v", g, e)
	}
	if g, e := n.End(), ast.NewPos(1, 6); g != e {
		t.Errorf("Pipe.End() = %v, expected %v", g, e)
	}
}

func TestCmd(t *testing.T) {
	var c ast.Command = new(ast.Cmd)
	if g, e := c.Pos(), ast.NewPos(0, 0); g != e {
		t.Errorf("Cmd.Pos() = %v, expected %v", g, e)
	}
	if g, e := c.End(), ast.NewPos(0, 0); g != e {
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
	if g, e := c.Pos(), ast.NewPos(1, 1); g != e {
		t.Errorf("Cmd.Pos() = %v, expected %v", g, e)
	}
	if g, e := c.End(), ast.NewPos(1, 5); g != e {
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
	if g, e := c.Pos(), ast.NewPos(1, 1); g != e {
		t.Errorf("Cmd.Pos() = %v, expected %v", g, e)
	}
	if g, e := c.End(), ast.NewPos(1, 9); g != e {
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
	if g, e := c.Pos(), ast.NewPos(1, 1); g != e {
		t.Errorf("Cmd.Pos() = %v, expected %v", g, e)
	}
	if g, e := c.End(), ast.NewPos(1, 9); g != e {
		t.Errorf("Cmd.End() = %v, expected %v", g, e)
	}
}

func TestSimpleCmd(t *testing.T) {
	var x ast.CmdExpr = new(ast.SimpleCmd)
	if g, e := x.Pos(), ast.NewPos(0, 0); g != e {
		t.Errorf("SimpleCmd.Pos() = %v, expected %v", g, e)
	}
	if g, e := x.End(), ast.NewPos(0, 0); g != e {
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
	if g, e := x.Pos(), ast.NewPos(1, 1); g != e {
		t.Errorf("SimpleCmd.Pos() = %v, expected %v", g, e)
	}
	if g, e := x.End(), ast.NewPos(1, 4); g != e {
		t.Errorf("SimpleCmd.End() = %v, expected %v", g, e)
	}

	x = &ast.SimpleCmd{
		Assigns: []*ast.Assign{
			{
				Name: &ast.Lit{
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
	if g, e := x.Pos(), ast.NewPos(1, 1); g != e {
		t.Errorf("SimpleCmd.Pos() = %v, expected %v", g, e)
	}
	if g, e := x.End(), ast.NewPos(1, 8); g != e {
		t.Errorf("SimpleCmd.End() = %v, expected %v", g, e)
	}

	x = &ast.SimpleCmd{
		Assigns: []*ast.Assign{
			{
				Name: &ast.Lit{
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
	if g, e := x.Pos(), ast.NewPos(1, 1); g != e {
		t.Errorf("SimpleCmd.Pos() = %v, expected %v", g, e)
	}
	if g, e := x.End(), ast.NewPos(1, 12); g != e {
		t.Errorf("SimpleCmd.End() = %v, expected %v", g, e)
	}

	var c ast.Command = &ast.Cmd{Expr: new(ast.SimpleCmd)}
	if g, e := c.Pos(), ast.NewPos(0, 0); g != e {
		t.Errorf("Cmd.Pos() = %v, expected %v", g, e)
	}
	if g, e := c.End(), ast.NewPos(0, 0); g != e {
		t.Errorf("Cmd.End() = %v, expected %v", g, e)
	}

	c = &ast.Cmd{Expr: x}
	if g, e := c.Pos(), ast.NewPos(1, 1); g != e {
		t.Errorf("Cmd.Pos() = %v, expected %v", g, e)
	}
	if g, e := c.End(), ast.NewPos(1, 12); g != e {
		t.Errorf("Cmd.End() = %v, expected %v", g, e)
	}
}

func TestAssign(t *testing.T) {
	var n ast.Node = new(ast.Assign)
	if g, e := n.Pos(), ast.NewPos(0, 0); g != e {
		t.Errorf("Assign.Pos() = %v, expected %v", g, e)
	}
	if g, e := n.End(), ast.NewPos(0, 0); g != e {
		t.Errorf("Assign.End() = %v, expected %v", g, e)
	}

	n = &ast.Assign{
		Name: &ast.Lit{
			ValuePos: ast.NewPos(1, 1),
			Value:    "lit",
		},
		Op: "=",
	}
	if g, e := n.Pos(), ast.NewPos(1, 1); g != e {
		t.Errorf("Assign.Pos() = %v, expected %v", g, e)
	}
	if g, e := n.End(), ast.NewPos(1, 5); g != e {
		t.Errorf("Assign.End() = %v, expected %v", g, e)
	}

	n = &ast.Assign{
		Name: &ast.Lit{
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
	if g, e := n.Pos(), ast.NewPos(1, 1); g != e {
		t.Errorf("Assign.Pos() = %v, expected %v", g, e)
	}
	if g, e := n.End(), ast.NewPos(1, 8); g != e {
		t.Errorf("Assign.End() = %v, expected %v", g, e)
	}
}

func TestSubshell(t *testing.T) {
	var x ast.CmdExpr = new(ast.Subshell)
	if g, e := x.Pos(), ast.NewPos(0, 0); g != e {
		t.Errorf("Subshell.Pos() = %v, expected %v", g, e)
	}
	if g, e := x.End(), ast.NewPos(0, 0); g != e {
		t.Errorf("Subshell.End() = %v, expected %v", g, e)
	}

	x = &ast.Subshell{
		Lparen: ast.NewPos(1, 1),
		Rparen: ast.NewPos(1, 3),
	}
	if g, e := x.Pos(), ast.NewPos(1, 1); g != e {
		t.Errorf("Subshell.Pos() = %v, expected %v", g, e)
	}
	if g, e := x.End(), ast.NewPos(1, 4); g != e {
		t.Errorf("Subshell.End() = %v, expected %v", g, e)
	}
}

func TestGroup(t *testing.T) {
	var x ast.CmdExpr = new(ast.Group)
	if g, e := x.Pos(), ast.NewPos(0, 0); g != e {
		t.Errorf("Group.Pos() = %v, expected %v", g, e)
	}
	if g, e := x.End(), ast.NewPos(0, 0); g != e {
		t.Errorf("Group.End() = %v, expected %v", g, e)
	}

	x = &ast.Group{
		Lbrace: ast.NewPos(1, 1),
		Rbrace: ast.NewPos(1, 3),
	}
	if g, e := x.Pos(), ast.NewPos(1, 1); g != e {
		t.Errorf("Group.Pos() = %v, expected %v", g, e)
	}
	if g, e := x.End(), ast.NewPos(1, 4); g != e {
		t.Errorf("Group.End() = %v, expected %v", g, e)
	}
}

func TestArithEval(t *testing.T) {
	var x ast.CmdExpr = new(ast.ArithEval)
	if g, e := x.Pos(), ast.NewPos(0, 0); g != e {
		t.Errorf("ArithEval.Pos() = %v, expected %v", g, e)
	}
	if g, e := x.End(), ast.NewPos(0, 0); g != e {
		t.Errorf("ArithEval.End() = %v, expected %v", g, e)
	}

	x = &ast.ArithEval{
		Left:  ast.NewPos(1, 1),
		Right: ast.NewPos(1, 4),
	}
	if g, e := x.Pos(), ast.NewPos(1, 1); g != e {
		t.Errorf("ArithEval.Pos() = %v, expected %v", g, e)
	}
	if g, e := x.End(), ast.NewPos(1, 6); g != e {
		t.Errorf("ArithEval.End() = %v, expected %v", g, e)
	}
}

func TestForClause(t *testing.T) {
	var x ast.CmdExpr = new(ast.ForClause)
	if g, e := x.Pos(), ast.NewPos(0, 0); g != e {
		t.Errorf("ForClause.Pos() = %v, expected %v", g, e)
	}
	if g, e := x.End(), ast.NewPos(0, 0); g != e {
		t.Errorf("ForClause.End() = %v, expected %v", g, e)
	}

	x = &ast.ForClause{
		For:  ast.NewPos(1, 1),
		Done: ast.NewPos(4, 1),
	}
	if g, e := x.Pos(), ast.NewPos(1, 1); g != e {
		t.Errorf("ForClause.Pos() = %v, expected %v", g, e)
	}
	if g, e := x.End(), ast.NewPos(4, 5); g != e {
		t.Errorf("ForClause.End() = %v, expected %v", g, e)
	}
}

func TestCaseClause(t *testing.T) {
	var x ast.CmdExpr = new(ast.CaseClause)
	if g, e := x.Pos(), ast.NewPos(0, 0); g != e {
		t.Errorf("CaseClause.Pos() = %v, expected %v", g, e)
	}
	if g, e := x.End(), ast.NewPos(0, 0); g != e {
		t.Errorf("CaseClause.End() = %v, expected %v", g, e)
	}

	x = &ast.CaseClause{
		Case: ast.NewPos(1, 1),
		Esac: ast.NewPos(3, 1),
	}
	if g, e := x.Pos(), ast.NewPos(1, 1); g != e {
		t.Errorf("CaseClause.Pos() = %v, expected %v", g, e)
	}
	if g, e := x.End(), ast.NewPos(3, 5); g != e {
		t.Errorf("CaseClause.End() = %v, expected %v", g, e)
	}
}

func TestCaseItem(t *testing.T) {
	var n ast.Node = new(ast.CaseItem)
	if g, e := n.Pos(), ast.NewPos(0, 0); g != e {
		t.Errorf("CaseItem.Pos() = %v, expected %v", g, e)
	}
	if g, e := n.End(), ast.NewPos(0, 0); g != e {
		t.Errorf("CaseItem.End() = %v, expected %v", g, e)
	}

	n = &ast.CaseItem{
		Patterns: []ast.Word{
			{
				&ast.Lit{
					ValuePos: ast.NewPos(2, 1),
					Value:    "lit",
				},
			},
		},
		Break: ast.NewPos(2, 6),
	}
	if g, e := n.Pos(), ast.NewPos(2, 1); g != e {
		t.Errorf("CaseItem.Pos() = %v, expected %v", g, e)
	}
	if g, e := n.End(), ast.NewPos(2, 8); g != e {
		t.Errorf("CaseItem.End() = %v, expected %v", g, e)
	}

	n = &ast.CaseItem{
		Lparen: ast.NewPos(2, 1),
		Patterns: []ast.Word{
			{
				&ast.Lit{
					ValuePos: ast.NewPos(2, 2),
					Value:    "lit",
				},
			},
		},
		Break: ast.NewPos(2, 7),
	}
	if g, e := n.Pos(), ast.NewPos(2, 1); g != e {
		t.Errorf("CaseItem.Pos() = %v, expected %v", g, e)
	}
	if g, e := n.End(), ast.NewPos(2, 9); g != e {
		t.Errorf("CaseItem.End() = %v, expected %v", g, e)
	}

	n = &ast.CaseItem{
		Patterns: []ast.Word{
			{
				&ast.Lit{
					ValuePos: ast.NewPos(2, 1),
					Value:    "lit",
				},
			},
		},
		List: []ast.Command{
			&ast.Cmd{
				Expr: &ast.SimpleCmd{
					Args: []ast.Word{
						{
							&ast.Lit{
								ValuePos: ast.NewPos(2, 6),
								Value:    "lit",
							},
						},
					},
				},
			},
		},
	}
	if g, e := n.Pos(), ast.NewPos(2, 1); g != e {
		t.Errorf("CaseItem.Pos() = %v, expected %v", g, e)
	}
	if g, e := n.End(), ast.NewPos(2, 9); g != e {
		t.Errorf("CaseItem.End() = %v, expected %v", g, e)
	}

	n = &ast.CaseItem{
		Lparen: ast.NewPos(2, 1),
		Patterns: []ast.Word{
			{
				&ast.Lit{
					ValuePos: ast.NewPos(2, 2),
					Value:    "lit",
				},
			},
		},
		List: []ast.Command{
			&ast.Cmd{
				Expr: &ast.SimpleCmd{
					Args: []ast.Word{
						{
							&ast.Lit{
								ValuePos: ast.NewPos(2, 7),
								Value:    "lit",
							},
						},
					},
				},
			},
		},
	}
	if g, e := n.Pos(), ast.NewPos(2, 1); g != e {
		t.Errorf("CaseItem.Pos() = %v, expected %v", g, e)
	}
	if g, e := n.End(), ast.NewPos(2, 10); g != e {
		t.Errorf("CaseItem.End() = %v, expected %v", g, e)
	}
}

func TestIfClause(t *testing.T) {
	var x ast.CmdExpr = new(ast.IfClause)
	if g, e := x.Pos(), ast.NewPos(0, 0); g != e {
		t.Errorf("IfClause.Pos() = %v, expected %v", g, e)
	}
	if g, e := x.End(), ast.NewPos(0, 0); g != e {
		t.Errorf("IfClause.End() = %v, expected %v", g, e)
	}

	x = &ast.IfClause{
		If: ast.NewPos(1, 1),
		Fi: ast.NewPos(4, 1),
	}
	if g, e := x.Pos(), ast.NewPos(1, 1); g != e {
		t.Errorf("IfClause.Pos() = %v, expected %v", g, e)
	}
	if g, e := x.End(), ast.NewPos(4, 3); g != e {
		t.Errorf("IfClause.End() = %v, expected %v", g, e)
	}
}

func TestElifClause(t *testing.T) {
	var e ast.ElsePart = new(ast.ElifClause)
	if g, e := e.Pos(), ast.NewPos(0, 0); g != e {
		t.Errorf("ElifClause.Pos() = %v, expected %v", g, e)
	}
	if g, e := e.End(), ast.NewPos(0, 0); g != e {
		t.Errorf("ElifClause.End() = %v, expected %v", g, e)
	}

	e = &ast.ElifClause{
		Elif: ast.NewPos(4, 1),
		List: []ast.Command{
			&ast.Cmd{
				Expr: &ast.SimpleCmd{
					Args: []ast.Word{
						{
							&ast.Lit{
								ValuePos: ast.NewPos(6, 3),
								Value:    "lit",
							},
						},
					},
				},
			},
		},
	}
	if g, e := e.Pos(), ast.NewPos(4, 1); g != e {
		t.Errorf("ElifClause.Pos() = %v, expected %v", g, e)
	}
	if g, e := e.End(), ast.NewPos(6, 6); g != e {
		t.Errorf("ElifClause.End() = %v, expected %v", g, e)
	}
}

func TestElseClause(t *testing.T) {
	var e ast.ElsePart = new(ast.ElseClause)
	if g, e := e.Pos(), ast.NewPos(0, 0); g != e {
		t.Errorf("ElseClause.Pos() = %v, expected %v", g, e)
	}
	if g, e := e.End(), ast.NewPos(0, 0); g != e {
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
	if g, e := e.Pos(), ast.NewPos(4, 1); g != e {
		t.Errorf("ElseClause.Pos() = %v, expected %v", g, e)
	}
	if g, e := e.End(), ast.NewPos(5, 6); g != e {
		t.Errorf("ElseClause.End() = %v, expected %v", g, e)
	}
}

func TestWhileClause(t *testing.T) {
	var x ast.CmdExpr = new(ast.WhileClause)
	if g, e := x.Pos(), ast.NewPos(0, 0); g != e {
		t.Errorf("WhileClause.Pos() = %v, expected %v", g, e)
	}
	if g, e := x.End(), ast.NewPos(0, 0); g != e {
		t.Errorf("WhileClause.End() = %v, expected %v", g, e)
	}

	x = &ast.WhileClause{
		While: ast.NewPos(1, 1),
		Done:  ast.NewPos(4, 1),
	}
	if g, e := x.Pos(), ast.NewPos(1, 1); g != e {
		t.Errorf("WhileClause.Pos() = %v, expected %v", g, e)
	}
	if g, e := x.End(), ast.NewPos(4, 5); g != e {
		t.Errorf("WhileClause.End() = %v, expected %v", g, e)
	}
}

func TestUntilClause(t *testing.T) {
	var x ast.CmdExpr = new(ast.UntilClause)
	if g, e := x.Pos(), ast.NewPos(0, 0); g != e {
		t.Errorf("UntilClause.Pos() = %v, expected %v", g, e)
	}
	if g, e := x.End(), ast.NewPos(0, 0); g != e {
		t.Errorf("UntilClause.End() = %v, expected %v", g, e)
	}

	x = &ast.UntilClause{
		Until: ast.NewPos(1, 1),
		Done:  ast.NewPos(4, 1),
	}
	if g, e := x.Pos(), ast.NewPos(1, 1); g != e {
		t.Errorf("UntilClause.Pos() = %v, expected %v", g, e)
	}
	if g, e := x.End(), ast.NewPos(4, 5); g != e {
		t.Errorf("UntilClause.End() = %v, expected %v", g, e)
	}
}

func TestFuncDef(t *testing.T) {
	var x ast.CmdExpr = new(ast.FuncDef)
	if g, e := x.Pos(), ast.NewPos(0, 0); g != e {
		t.Errorf("FuncDef.Pos() = %v, expected %v", g, e)
	}
	if g, e := x.End(), ast.NewPos(0, 0); g != e {
		t.Errorf("FuncDef.End() = %v, expected %v", g, e)
	}

	x = &ast.FuncDef{
		Name: &ast.Lit{
			ValuePos: ast.NewPos(1, 1),
			Value:    "lit",
		},
		Body: &ast.Cmd{
			Expr: &ast.Group{
				Lbrace: ast.NewPos(1, 7),
				Rbrace: ast.NewPos(1, 10),
			},
		},
	}
	if g, e := x.Pos(), ast.NewPos(1, 1); g != e {
		t.Errorf("FuncDef.Pos() = %v, expected %v", g, e)
	}
	if g, e := x.End(), ast.NewPos(1, 11); g != e {
		t.Errorf("FuncDef.End() = %v, expected %v", g, e)
	}
}

func TestRedir(t *testing.T) {
	var n ast.Node = new(ast.Redir)
	if g, e := n.Pos(), ast.NewPos(0, 0); g != e {
		t.Errorf("Redir.Pos() = %v, expected %v", g, e)
	}
	if g, e := n.End(), ast.NewPos(0, 0); g != e {
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
	if g, e := n.Pos(), ast.NewPos(1, 1); g != e {
		t.Errorf("Redir.Pos() = %v, expected %v", g, e)
	}
	if g, e := n.End(), ast.NewPos(1, 6); g != e {
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
	if g, e := n.Pos(), ast.NewPos(1, 1); g != e {
		t.Errorf("Redir.Pos() = %v, expected %v", g, e)
	}
	if g, e := n.End(), ast.NewPos(1, 7); g != e {
		t.Errorf("Redir.End() = %v, expected %v", g, e)
	}

	n = &ast.Redir{
		OpPos: ast.NewPos(1, 1),
		Op:    "<<",
		Delim: ast.Word{
			&ast.Lit{
				ValuePos: ast.NewPos(3, 1),
				Value:    "lit",
			},
		},
	}
	if g, e := n.Pos(), ast.NewPos(1, 1); g != e {
		t.Errorf("Redir.Pos() = %v, expected %v", g, e)
	}
	if g, e := n.End(), ast.NewPos(3, 4); g != e {
		t.Errorf("Redir.End() = %v, expected %v", g, e)
	}
}

func TestWord(t *testing.T) {
	var n ast.Node = ast.Word{}
	if g, e := n.Pos(), ast.NewPos(0, 0); g != e {
		t.Errorf("Word.Pos() = %v, expected %v", g, e)
	}
	if g, e := n.End(), ast.NewPos(0, 0); g != e {
		t.Errorf("Word.End() = %v, expected %v", g, e)
	}
	if g, e := len(n.(ast.Word)), 0; g != e {
		t.Errorf("len(Word) = %v, expected %v", g, e)
	}

	n = ast.Word{
		&ast.Lit{
			ValuePos: ast.NewPos(1, 1),
			Value:    "lit",
		},
	}
	if g, e := n.Pos(), ast.NewPos(1, 1); g != e {
		t.Errorf("Word.Pos() = %v, expected %v", g, e)
	}
	if g, e := n.End(), ast.NewPos(1, 4); g != e {
		t.Errorf("Word.End() = %v, expected %v", g, e)
	}
	if g, e := len(n.(ast.Word)), 1; g != e {
		t.Errorf("len(Word) = %v, expected %v", g, e)
	}
}

func TestLit(t *testing.T) {
	var w ast.WordPart = new(ast.Lit)
	if g, e := w.Pos(), ast.NewPos(0, 0); g != e {
		t.Errorf("Lit.Pos() = %v, expected %v", g, e)
	}
	if g, e := w.End(), ast.NewPos(0, 0); g != e {
		t.Errorf("Lit.End() = %v, expected %v", g, e)
	}

	w = &ast.Lit{
		ValuePos: ast.NewPos(1, 1),
		Value:    "lit",
	}
	if g, e := w.Pos(), ast.NewPos(1, 1); g != e {
		t.Errorf("Lit.Pos() = %v, expected %v", g, e)
	}
	if g, e := w.End(), ast.NewPos(1, 4); g != e {
		t.Errorf("Lit.End() = %v, expected %v", g, e)
	}

	w = &ast.Lit{
		ValuePos: ast.NewPos(1, 1),
		Value:    "1\n2",
	}
	if g, e := w.Pos(), ast.NewPos(1, 1); g != e {
		t.Errorf("Lit.Pos() = %v, expected %v", g, e)
	}
	if g, e := w.End(), ast.NewPos(2, 2); g != e {
		t.Errorf("Lit.End() = %v, expected %v", g, e)
	}
}

func TestQuote(t *testing.T) {
	var w ast.WordPart = new(ast.Quote)
	if g, e := w.Pos(), ast.NewPos(0, 0); g != e {
		t.Errorf("Quote.Pos() = %v, expected %v", g, e)
	}
	if g, e := w.End(), ast.NewPos(0, 0); g != e {
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
	if g, e := w.Pos(), ast.NewPos(1, 1); g != e {
		t.Errorf("Quote.Pos() = %v, expected %v", g, e)
	}
	if g, e := w.End(), ast.NewPos(1, 3); g != e {
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
	if g, e := w.Pos(), ast.NewPos(1, 1); g != e {
		t.Errorf("Quote.Pos() = %v, expected %v", g, e)
	}
	if g, e := w.End(), ast.NewPos(1, 8); g != e {
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
	if g, e := w.Pos(), ast.NewPos(1, 1); g != e {
		t.Errorf("Quote.Pos() = %v, expected %v", g, e)
	}
	if g, e := w.End(), ast.NewPos(2, 7); g != e {
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
	if g, e := w.Pos(), ast.NewPos(1, 1); g != e {
		t.Errorf("Quote.Pos() = %v, expected %v", g, e)
	}
	if g, e := w.End(), ast.NewPos(1, 8); g != e {
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
	if g, e := w.Pos(), ast.NewPos(1, 1); g != e {
		t.Errorf("Quote.Pos() = %v, expected %v", g, e)
	}
	if g, e := w.End(), ast.NewPos(2, 7); g != e {
		t.Errorf("Quote.End() = %v, expected %v", g, e)
	}
}

func TestParamExp(t *testing.T) {
	var w ast.WordPart = new(ast.ParamExp)
	if g, e := w.Pos(), ast.NewPos(0, 0); g != e {
		t.Errorf("ParamExp.Pos() = %v, expected %v", g, e)
	}
	if g, e := w.End(), ast.NewPos(0, 0); g != e {
		t.Errorf("ParamExp.End() = %v, expected %v", g, e)
	}

	w = &ast.ParamExp{
		Dollar: ast.NewPos(1, 1),
		Name: &ast.Lit{
			ValuePos: ast.NewPos(1, 2),
			Value:    "lit",
		},
	}
	if g, e := w.Pos(), ast.NewPos(1, 1); g != e {
		t.Errorf("ParamExp.Pos() = %v, expected %v", g, e)
	}
	if g, e := w.End(), ast.NewPos(1, 5); g != e {
		t.Errorf("ParamExp.End() = %v, expected %v", g, e)
	}

	w = &ast.ParamExp{
		Dollar: ast.NewPos(1, 1),
		Braces: true,
		Name: &ast.Lit{
			ValuePos: ast.NewPos(1, 2),
			Value:    "lit",
		},
	}
	if g, e := w.Pos(), ast.NewPos(1, 1); g != e {
		t.Errorf("ParamExp.Pos() = %v, expected %v", g, e)
	}
	if g, e := w.End(), ast.NewPos(1, 6); g != e {
		t.Errorf("ParamExp.End() = %v, expected %v", g, e)
	}

	w = &ast.ParamExp{
		Dollar: ast.NewPos(1, 1),
		Braces: true,
		Name: &ast.Lit{
			ValuePos: ast.NewPos(1, 3),
			Value:    "lit",
		},
		OpPos: ast.NewPos(1, 6),
		Op:    ":-",
	}
	if g, e := w.Pos(), ast.NewPos(1, 1); g != e {
		t.Errorf("ParamExp.Pos() = %v, expected %v", g, e)
	}
	if g, e := w.End(), ast.NewPos(1, 9); g != e {
		t.Errorf("ParamExp.End() = %v, expected %v", g, e)
	}

	w = &ast.ParamExp{
		Dollar: ast.NewPos(1, 1),
		Braces: true,
		Name: &ast.Lit{
			ValuePos: ast.NewPos(1, 3),
			Value:    "lit",
		},
		OpPos: ast.NewPos(1, 6),
		Op:    ":-",
		Word: ast.Word{
			&ast.Lit{
				ValuePos: ast.NewPos(1, 8),
				Value:    "lit",
			},
		},
	}
	if g, e := w.Pos(), ast.NewPos(1, 1); g != e {
		t.Errorf("ParamExp.Pos() = %v, expected %v", g, e)
	}
	if g, e := w.End(), ast.NewPos(1, 12); g != e {
		t.Errorf("ParamExp.End() = %v, expected %v", g, e)
	}

	w = &ast.ParamExp{
		Dollar: ast.NewPos(1, 1),
		Braces: true,
		Name: &ast.Lit{
			ValuePos: ast.NewPos(1, 4),
			Value:    "lit",
		},
		OpPos: ast.NewPos(1, 3),
		Op:    "#",
	}
	if g, e := w.Pos(), ast.NewPos(1, 1); g != e {
		t.Errorf("ParamExp.Pos() = %v, expected %v", g, e)
	}
	if g, e := w.End(), ast.NewPos(1, 8); g != e {
		t.Errorf("ParamExp.End() = %v, expected %v", g, e)
	}
}

func TestCmdSubst(t *testing.T) {
	var w ast.WordPart = new(ast.CmdSubst)
	if g, e := w.Pos(), ast.NewPos(0, 0); g != e {
		t.Errorf("CmdSubst.Pos() = %v, expected %v", g, e)
	}
	if g, e := w.End(), ast.NewPos(0, 0); g != e {
		t.Errorf("CmdSubst.End() = %v, expected %v", g, e)
	}

	w = &ast.CmdSubst{
		Dollar: true,
		Left:   ast.NewPos(1, 2),
		Right:  ast.NewPos(1, 4),
	}
	if g, e := w.Pos(), ast.NewPos(1, 1); g != e {
		t.Errorf("CmdSubst.Pos() = %v, expected %v", g, e)
	}
	if g, e := w.End(), ast.NewPos(1, 5); g != e {
		t.Errorf("CmdSubst.End() = %v, expected %v", g, e)
	}

	w = &ast.CmdSubst{
		Left:  ast.NewPos(1, 1),
		Right: ast.NewPos(1, 3),
	}
	if g, e := w.Pos(), ast.NewPos(1, 1); g != e {
		t.Errorf("CmdSubst.Pos() = %v, expected %v", g, e)
	}
	if g, e := w.End(), ast.NewPos(1, 4); g != e {
		t.Errorf("CmdSubst.End() = %v, expected %v", g, e)
	}
}

func TestArithExp(t *testing.T) {
	var w ast.WordPart = new(ast.ArithExp)
	if g, e := w.Pos(), ast.NewPos(0, 0); g != e {
		t.Errorf("ArithExp.Pos() = %v, expected %v", g, e)
	}
	if g, e := w.End(), ast.NewPos(0, 0); g != e {
		t.Errorf("ArithExp.End() = %v, expected %v", g, e)
	}

	w = &ast.ArithExp{
		Left:  ast.NewPos(1, 1),
		Right: ast.NewPos(1, 5),
	}
	if g, e := w.Pos(), ast.NewPos(1, 1); g != e {
		t.Errorf("ArithExp.Pos() = %v, expected %v", g, e)
	}
	if g, e := w.End(), ast.NewPos(1, 7); g != e {
		t.Errorf("ArithExp.End() = %v, expected %v", g, e)
	}
}

func TestComment(t *testing.T) {
	var n ast.Node = new(ast.Comment)
	if g, e := n.Pos(), ast.NewPos(0, 0); g != e {
		t.Errorf("Comment.Pos() = %v, expected %v", g, e)
	}
	if g, e := n.End(), ast.NewPos(0, 0); g != e {
		t.Errorf("Comment.End() = %v, expected %v", g, e)
	}

	n = &ast.Comment{
		Hash: ast.NewPos(1, 1),
		Text: " comment",
	}
	if g, e := n.Pos(), ast.NewPos(1, 1); g != e {
		t.Errorf("Comment.Pos() = %v, expected %v", g, e)
	}
	if g, e := n.End(), ast.NewPos(1, 9); g != e {
		t.Errorf("Comment.End() = %v, expected %v", g, e)
	}
}
