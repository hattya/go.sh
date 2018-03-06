//
// go.sh/ast :: ast.go
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

// Package ast declares the types used to represent syntax trees for the Shell
// Command Language.
package ast

// Node represents an abstract syntax tree.
type Node interface {
	Pos() Pos // position of the first character of the node
	End() Pos // position of the character immediately after the node
}

// Command represents a command.
type Command interface {
	Node
	commandNode()
}

// List represents an AND-OR list.
type List struct {
	Pipeline []Word   // pipeline
	List     []*AndOr // pipelines separated by "&&" or "||" operator; or nil
	SepPos   Pos      // position of "&" or ";" operator (zero if there is no operator)
	Sep      string
}

func (c *List) Pos() Pos {
	if len(c.Pipeline) == 0 {
		return Pos{}
	}
	return c.Pipeline[0].Pos()
}

func (c *List) End() Pos {
	if len(c.List) != 0 {
		return c.List[len(c.List)-1].End()
	}
	if len(c.Pipeline) == 0 {
		return Pos{}
	}
	return c.Pipeline[len(c.Pipeline)-1].End()
}

func (c *List) commandNode() {}

// AndOr represents a pipeline of the AND-OR list.
type AndOr struct {
	OpPos    Pos    // position of Op
	Op       string // "&&" or "||" operator
	Pipeline []Word // pipeline
}

func (ao *AndOr) Pos() Pos { return ao.OpPos }
func (ao *AndOr) End() Pos {
	if len(ao.Pipeline) == 0 {
		return Pos{}
	}
	return ao.Pipeline[len(ao.Pipeline)-1].End()
}

// Word represents a WORD token.
type Word []WordPart

func (w Word) Pos() Pos {
	if len(w) == 0 {
		return Pos{}
	}
	return w[0].Pos()
}
func (w Word) End() Pos {
	if len(w) == 0 {
		return Pos{}
	}
	return w[len(w)-1].End()
}

// WordPart represents a part of the WORD token.
type WordPart interface {
	Node
	wordPartNode()
}

type (
	// Lit represents a literal string.
	Lit struct {
		ValuePos Pos
		Value    string
	}

	// Quote represents a quoted WORD token.
	Quote struct {
		TokPos Pos    // position of Tok
		Tok    string // quoting character
		Value  Word
	}
)

func (w *Lit) Pos() Pos   { return w.ValuePos }
func (w *Quote) Pos() Pos { return w.TokPos }

func (w *Lit) End() Pos { return w.ValuePos.shift(len(w.Value)) }
func (w *Quote) End() Pos {
	end := w.Value.End()
	if end.IsZero() || w.Tok == "\\" {
		return end
	}
	return end.shift(1)
}

func (w *Lit) wordPartNode()   {}
func (w *Quote) wordPartNode() {}

// Comment represents a comment.
type Comment struct {
	Hash Pos    // position of "#"
	Text string // comment text (excluding "\n")
}

func (c *Comment) Pos() Pos { return c.Hash }
func (c *Comment) End() Pos { return c.Hash.shift(len(c.Text)) }
