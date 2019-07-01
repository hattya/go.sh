//
// go.sh/ast :: token.go
//
//   Copyright (c) 2018-2019 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
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

// After reports whether the position is after q.
func (p Pos) After(q Pos) bool {
	return p.line > q.line || p.line == q.line && p.col > q.col
}

func (p Pos) shift(off int) Pos {
	p.col += off
	return p
}
