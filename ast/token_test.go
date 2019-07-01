//
// go.sh/ast :: token_test.go
//
//   Copyright (c) 2018-2019 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
//

package ast_test

import (
	"testing"

	"github.com/hattya/go.sh/ast"
)

func TestPos(t *testing.T) {
	var p ast.Pos
	if g, e := p.Line(), 0; g != e {
		t.Errorf("Pos.Line() = %v, expected %v", g, e)
	}
	if g, e := p.Col(), 0; g != e {
		t.Errorf("Pos.Col() = %v, expected %v", g, e)
	}
	if !p.IsZero() {
		t.Errorf("Pos.IsZero() = false, expected true")
	}

	p = ast.NewPos(1, 2)
	if g, e := p.Line(), 1; g != e {
		t.Errorf("Pos.Line() = %v, expected %v", g, e)
	}
	if g, e := p.Col(), 2; g != e {
		t.Errorf("Pos.Col() = %v, expected %v", g, e)
	}
	if p.IsZero() {
		t.Errorf("Pos.IsZero() = true, expected false")
	}
}

var posCompareTests = []struct {
	p, q          ast.Pos
	before, after bool
}{
	{},
	{
		p: ast.NewPos(1, 1),
		q: ast.NewPos(1, 1),
	},
	{
		p:      ast.NewPos(1, 1),
		q:      ast.NewPos(1, 2),
		before: true,
	},
	{
		p:      ast.NewPos(1, 1),
		q:      ast.NewPos(2, 1),
		before: true,
	},
	{
		p:     ast.NewPos(1, 2),
		q:     ast.NewPos(1, 1),
		after: true,
	},
	{
		p:     ast.NewPos(2, 1),
		q:     ast.NewPos(1, 1),
		after: true,
	},
}

func TestPosCompare(t *testing.T) {
	for _, tt := range posCompareTests {
		if g, e := tt.p.Before(tt.q), tt.before; g != e {
			t.Errorf("%v.Before(%v) = %v, expected %v", tt.p, tt.q, g, e)
		}
		if g, e := tt.p.After(tt.q), tt.after; g != e {
			t.Errorf("%v.After(%v) = %v, expected %v", tt.p, tt.q, g, e)
		}
	}
}
