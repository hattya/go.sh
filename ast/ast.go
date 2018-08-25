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
	// AndOrList represents an AND-OR list.
	AndOrList struct {
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

	// Cmd represents a simple command, a compound command, or a function
	// definition command.
	Cmd struct {
		Expr   CmdExpr
		Redirs []*Redir
	}
)

func (c *AndOrList) Pos() Pos {
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

func (c *AndOrList) End() Pos {
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

func (c *AndOrList) commandNode() {}
func (c *Pipeline) commandNode()  {}
func (c *Cmd) commandNode()       {}

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

	// ArithEval represents an arithmetic evaluation.
	ArithEval struct {
		Left  Pos    // position of "((" operator
		Expr  []Word // expression
		Right Pos    // position of "))" operator
	}

	// ForClause represents a for loop.
	ForClause struct {
		For       Pos       // position of reserved word "for"
		Name      *Lit      // variable name
		In        Pos       // position of reserved word "in" (zero if there is no "in")
		Items     []Word    // list of words; or nil
		Semicolon Pos       // position of ";" operator (zero if there is no ";" operator)
		Do        Pos       // position of reserved word "do"
		List      []Command // commands
		Done      Pos       // position of reserved word "done"
	}

	// CaseClause represents a case conditional construct.
	CaseClause struct {
		Case  Pos         // position of reserved word "case"
		Word  Word        // word
		In    Pos         // position of reserved word "in"
		Items []*CaseItem // patterns and commands; or nil
		Esac  Pos         // position of reserved word "esac"
	}

	// IfClause represents an if conditional construct.
	IfClause struct {
		If   Pos        // position of reserved word "if"
		Cond []Command  // condition
		Then Pos        // position of reserved word "then"
		List []Command  // commands
		Else []ElsePart // elif clauses and/or an else clause
		Fi   Pos        // position of reserved word "fi"
	}

	// WhileClause represents a while loop.
	WhileClause struct {
		While Pos       // position of reserved word "while"
		Cond  []Command // condition
		Do    Pos       // position of reserved word "do"
		List  []Command // commands
		Done  Pos       // position of reserved word "done"
	}

	// UntilClause represents an until loop.
	UntilClause struct {
		Until Pos       // position of reserved word "until"
		Cond  []Command // condition
		Do    Pos       // position of reserved word "do"
		List  []Command // commands
		Done  Pos       // position of reserved word "done"
	}

	// FuncDef represents a function definition command.
	FuncDef struct {
		Name   *Lit    // function name
		Lparen Pos     // position of '(' operator
		Rparen Pos     // position of ')' operator
		Body   Command // compound command
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
func (x *Subshell) Pos() Pos    { return x.Lparen }
func (x *Group) Pos() Pos       { return x.Lbrace }
func (x *ArithEval) Pos() Pos   { return x.Left }
func (x *ForClause) Pos() Pos   { return x.For }
func (x *CaseClause) Pos() Pos  { return x.Case }
func (x *IfClause) Pos() Pos    { return x.If }
func (x *WhileClause) Pos() Pos { return x.While }
func (x *UntilClause) Pos() Pos { return x.Until }
func (x *FuncDef) Pos() Pos {
	if x.Name == nil {
		return Pos{}
	}
	return x.Name.Pos()
}

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
func (x *ArithEval) End() Pos {
	if x.Right.IsZero() {
		return x.Right
	}
	return x.Right.shift(2)
}
func (x *ForClause) End() Pos {
	if x.Done.IsZero() {
		return x.Done
	}
	return x.Done.shift(4)
}
func (x *CaseClause) End() Pos {
	if x.Esac.IsZero() {
		return x.Esac
	}
	return x.Esac.shift(4)
}
func (x *IfClause) End() Pos {
	if x.Fi.IsZero() {
		return x.Fi
	}
	return x.Fi.shift(2)
}
func (x *WhileClause) End() Pos {
	if x.Done.IsZero() {
		return x.Done
	}
	return x.Done.shift(4)
}
func (x *UntilClause) End() Pos {
	if x.Done.IsZero() {
		return x.Done
	}
	return x.Done.shift(4)
}
func (x *FuncDef) End() Pos {
	if x.Body == nil {
		return Pos{}
	}
	return x.Body.End()
}

func (x *SimpleCmd) cmdExprNode()   {}
func (x *Subshell) cmdExprNode()    {}
func (x *Group) cmdExprNode()       {}
func (x *ArithEval) cmdExprNode()   {}
func (x *ForClause) cmdExprNode()   {}
func (x *CaseClause) cmdExprNode()  {}
func (x *IfClause) cmdExprNode()    {}
func (x *WhileClause) cmdExprNode() {}
func (x *UntilClause) cmdExprNode() {}
func (x *FuncDef) cmdExprNode()     {}

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

// CaseItem represents patterns and commands of the case conditional construct.
type CaseItem struct {
	Lparen   Pos       // position of "(" operator (zero if there is no "(" operator)
	Patterns []Word    // patterns
	Rparen   Pos       // position of ")" operator
	List     []Command // commands
	Break    Pos       // position of ";;" operator (zero if there is no ";;" operator)
}

func (ci *CaseItem) Pos() Pos {
	if !ci.Lparen.IsZero() {
		return ci.Lparen
	}
	if len(ci.Patterns) == 0 {
		return Pos{}
	}
	return ci.Patterns[0].Pos()
}
func (ci *CaseItem) End() Pos {
	if ci.Break.IsZero() {
		if len(ci.List) == 0 {
			return Pos{}
		}
		return ci.List[len(ci.List)-1].End()
	}
	return ci.Break.shift(2)
}

// ElsePart represents an elif clause or an else clause.
type ElsePart interface {
	Node
	elsePartNode()
}

type (
	// ElifClause represents an elif clause of the if conditional construct.
	ElifClause struct {
		Elif Pos       // position of reserved word "elif"
		Cond []Command // condition
		Then Pos       // position of reserved word "then"
		List []Command // commands
	}

	// ElseClause represents an else clause of the if conditional construct.
	ElseClause struct {
		Else Pos       // position of reserved word "else"
		List []Command // commands
	}
)

func (e *ElifClause) Pos() Pos { return e.Elif }
func (e *ElseClause) Pos() Pos { return e.Else }

func (e *ElifClause) End() Pos {
	if len(e.List) == 0 {
		return Pos{}
	}
	return e.List[len(e.List)-1].End()
}
func (e *ElseClause) End() Pos {
	if len(e.List) == 0 {
		return Pos{}
	}
	return e.List[len(e.List)-1].End()
}

func (e *ElifClause) elsePartNode() {}
func (e *ElseClause) elsePartNode() {}

// Redir represents an I/O redirection.
type Redir struct {
	N       *Lit
	OpPos   Pos
	Op      string
	Word    Word
	Heredoc Word // here-document; or nil
	Delim   Word // here-document delimiter; or nil
}

func (r *Redir) Pos() Pos {
	if r.N != nil {
		return r.N.Pos()
	}
	return r.OpPos
}
func (r *Redir) End() Pos {
	switch r.Op {
	case "<<", "<<-":
		return r.Delim.End()
	}
	return r.Word.End()
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

	// ParamExp represents a parameter expansion.
	ParamExp struct {
		Dollar Pos  // position of "$"
		Braces bool // whether this is enclosed in braces
		Name   *Lit // parameter name
		OpPos  Pos  // position of Op
		Op     string
		Word   Word
	}

	// CmdSubst represents a command substisution.
	CmdSubst struct {
		Dollar bool      // whether this is enclosed in "$()".
		Left   Pos       // position of "(" or "`".
		List   []Command // commands
		Right  Pos       // position of ")" or "`".
	}

	// ArithExp represents an arithmetic expansion.
	ArithExp struct {
		Left  Pos    // position of "$(("
		Expr  []Word // expression
		Right Pos    // position of "))"
	}
)

func (w *Lit) Pos() Pos      { return w.ValuePos }
func (w *Quote) Pos() Pos    { return w.TokPos }
func (w *ParamExp) Pos() Pos { return w.Dollar }
func (w *CmdSubst) Pos() Pos {
	if w.Dollar && !w.Left.IsZero() {
		return w.Left.shift(-1)
	}
	return w.Left
}
func (w *ArithExp) Pos() Pos { return w.Left }

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
func (w *ParamExp) End() Pos {
	var end Pos
	switch {
	case w.OpPos.IsZero():
		if w.Name != nil {
			end = w.Name.End()
		}
	case w.OpPos.Before(w.Name.End()):
		end = w.Name.End()
	case len(w.Word) == 0:
		end = w.OpPos.shift(len(w.Op))
	default:
		end = w.Word.End()
	}
	if !w.Braces {
		return end
	}
	return end.shift(1)
}
func (w *CmdSubst) End() Pos {
	if w.Right.IsZero() {
		return w.Right
	}
	return w.Right.shift(1)
}
func (w *ArithExp) End() Pos {
	if w.Right.IsZero() {
		return w.Right
	}
	return w.Right.shift(2)
}

func (w *Lit) wordPartNode()      {}
func (w *Quote) wordPartNode()    {}
func (w *ParamExp) wordPartNode() {}
func (w *CmdSubst) wordPartNode() {}
func (w *ArithExp) wordPartNode() {}

// Comment represents a comment.
type Comment struct {
	Hash Pos    // position of "#"
	Text string // comment text (excluding "\n")
}

func (c *Comment) Pos() Pos { return c.Hash }
func (c *Comment) End() Pos { return c.Hash.shift(len(c.Text)) }
