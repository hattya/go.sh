//
// go.sh/parser :: lexer_test.go
//
//   Copyright (c) 2018-2019 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
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

func TestWord(t *testing.T) {
	var n ast.Node = word{}
	if g, e := n.Pos(), ast.NewPos(0, 0); e != g {
		t.Errorf("word.Pos() = %v, expected %v", g, e)
	}
	if g, e := n.End(), ast.NewPos(0, 0); e != g {
		t.Errorf("word.End() = %v, expected %v", g, e)
	}

	n = word{
		typ: WORD,
		val: ast.Word{
			&ast.Lit{
				ValuePos: ast.NewPos(1, 1),
				Value:    "lit",
			},
		},
	}
	if g, e := n.Pos(), ast.NewPos(1, 1); e != g {
		t.Errorf("word.Pos() = %v, expected %v", g, e)
	}
	if g, e := n.End(), ast.NewPos(1, 4); e != g {
		t.Errorf("word.End() = %v, expected %v", g, e)
	}
}
