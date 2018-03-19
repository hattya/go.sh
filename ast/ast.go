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

type (
	// List represents an AND-OR list.
	List struct {
		Pipeline *Pipeline // pipeline
		List     []*AndOr  // pipelines separated by "&&" or "||" operator; or nil
		SepPos   Pos       // position of "&" or ";" operator (zero if there is no operator)
		Sep      string
	}

	// Pipeline represents a pipeline.
	Pipeline struct {
		Bang Pos     // position of reserved word "!"
		Cmd  *Cmd    // command
		List []*Pipe // commands separated by "|" operator; or nil
	}

	// Cmd represents a simple command or a compound command.
	Cmd struct {
		Expr   CmdExpr
		Redirs []*Redir
	}
)

func (c *List) Pos() Pos {
	if c.Pipeline == nil {
		return Pos{}
	}
	return c.Pipeline.Pos()
}
func (c *Pipeline) Pos() Pos {
	if !c.Bang.IsZero() {
		return c.Bang
	}
	if c.Cmd == nil {
		return Pos{}
	}
	return c.Cmd.Pos()
}
func (c *Cmd) Pos() Pos {
	switch {
	case len(c.Redirs) == 0:
		if c.Expr == nil {
			return Pos{}
		}
		return c.Expr.Pos()
	case c.Expr == nil:
		return c.Redirs[0].Pos()
	default:
		x := c.Expr.Pos()
		r := c.Redirs[0].Pos()
		if x.Before(r) {
			return x
		}
		return r
	}
}

func (c *List) End() Pos {
	if len(c.List) != 0 {
		return c.List[len(c.List)-1].End()
	}
	if c.Pipeline == nil {
		return Pos{}
	}
	return c.Pipeline.End()
}
func (c *Pipeline) End() Pos {
	if len(c.List) != 0 {
		return c.List[len(c.List)-1].End()
	}
	if c.Cmd == nil {
		return Pos{}
	}
	return c.Cmd.End()
}
func (c *Cmd) End() Pos {
	switch {
	case len(c.Redirs) == 0:
		if c.Expr == nil {
			return Pos{}
		}
		return c.Expr.End()
	case c.Expr == nil:
		return c.Redirs[len(c.Redirs)-1].End()
	default:
		x := c.Expr.End()
		r := c.Redirs[len(c.Redirs)-1].End()
		if x.After(r) {
			return x
		}
		return r
	}
}

func (c *List) commandNode()     {}
func (c *Pipeline) commandNode() {}
func (c *Cmd) commandNode()      {}

// AndOr represents a pipeline of the AND-OR list.
type AndOr struct {
	OpPos    Pos       // position of Op
	Op       string    // "&&" or "||" operator
	Pipeline *Pipeline // pipeline
}

func (ao *AndOr) Pos() Pos { return ao.OpPos }
func (ao *AndOr) End() Pos {
	if ao.Pipeline == nil {
		return Pos{}
	}
	return ao.Pipeline.End()
}

// Pipe represents a command of the pipeline.
type Pipe struct {
	OpPos Pos    // position of Op
	Op    string // "|" operator
	Cmd   *Cmd   // command
}

func (p *Pipe) Pos() Pos { return p.OpPos }
func (p *Pipe) End() Pos {
	if p.Cmd == nil {
		return Pos{}
	}
	return p.Cmd.End()
}

// CmdExpr represents a detail of the command.
type CmdExpr interface {
	Node
	cmdExprNode()
}

type (
	// SimpleCmd represents a simple command.
	SimpleCmd struct {
		Assigns []*Assign // variable assignments; or nil
		Args    []Word    // command line arguments
	}

	// Subshell represents a sequence of commands that executes in a subshell
	// environment.
	Subshell struct {
		Lparen Pos       // position of "(" operator
		List   []Command // commands
		Rparen Pos       // position of ")" operator
	}

	// Group represents a sequence of commands that executes in the current
	// process environment.
	Group struct {
		Lbrace Pos       // position of reserved word "{"
		List   []Command // commands
		Rbrace Pos       // position of reserved word "}"
	}
)

func (x *SimpleCmd) Pos() Pos {
	if len(x.Assigns) != 0 {
		return x.Assigns[0].Pos()
	}
	if len(x.Args) == 0 {
		return Pos{}
	}
	return x.Args[0].Pos()
}
func (x *Subshell) Pos() Pos { return x.Lparen }
func (x *Group) Pos() Pos    { return x.Lbrace }

func (x *SimpleCmd) End() Pos {
	if len(x.Args) == 0 {
		if len(x.Assigns) == 0 {
			return Pos{}
		}
		return x.Assigns[len(x.Assigns)-1].End()
	}
	return x.Args[len(x.Args)-1].End()
}
func (x *Subshell) End() Pos {
	if x.Rparen.IsZero() {
		return x.Rparen
	}
	return x.Rparen.shift(1)
}
func (x *Group) End() Pos {
	if x.Rbrace.IsZero() {
		return x.Rbrace
	}
	return x.Rbrace.shift(1)
}

func (x *SimpleCmd) cmdExprNode() {}
func (x *Subshell) cmdExprNode()  {}
func (x *Group) cmdExprNode()     {}

// Assign represents a variable assignment.
type Assign struct {
	Symbol *Lit
	Op     string
	Value  Word
}

func (a *Assign) Pos() Pos {
	if a.Symbol == nil {
		return Pos{}
	}
	return a.Symbol.Pos()
}
func (a *Assign) End() Pos {
	if len(a.Value) == 0 {
		if a.Symbol == nil {
			return Pos{}
		}
		return a.Symbol.End().shift(len(a.Op))
	}
	return a.Value.End()
}

// Redir represents an I/O redirection.
type Redir struct {
	N     *Lit
	OpPos Pos
	Op    string
	Word  Word
}

func (r *Redir) Pos() Pos {
	if r.N != nil {
		return r.N.Pos()
	}
	return r.OpPos
}
func (r *Redir) End() Pos { return r.Word.End() }

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

func (w *Lit) End() Pos {
	line := w.ValuePos.line
	col := w.ValuePos.col
	for _, r := range w.Value {
		if r == '\n' {
			line++
			col = 1
		} else {
			col++
		}
	}
	return Pos{line, col}
}
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
