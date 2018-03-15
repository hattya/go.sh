//
// go.sh/ast :: token.go
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

package ast

// Pos represents a position.
type Pos struct {
	line, col int
}

// NewPos returns a new Pos.
func NewPos(line, col int) Pos {
	return Pos{line, col}
}

// Line returns the line number of the position.
func (p Pos) Line() int { return p.line }

// Col returns the column number of the position.
func (p Pos) Col() int { return p.col }

// IsZero reports whether the position represents no position.
func (p Pos) IsZero() bool {
	return p == Pos{}
}

// Before reports whether the position is before q.
func (p Pos) Before(q Pos) bool {
	return p.line < q.line || p.line == q.line && p.col < q.col
}

func (p Pos) shift(off int) Pos {
	p.col += off
	return p
}
