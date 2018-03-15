//
// go.sh/ast :: token_test.go
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
