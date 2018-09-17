//
// go.sh/printer :: printer.go
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

// Package printer implements printing of AST nodes.
package printer

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/hattya/go.sh/ast"
)

// Fprint pretty-prints an AST node to w.
func Fprint(w io.Writer, n ast.Node) error {
	c := &Config{
		Indent: Tab,
		Redir:  After,
		Assign: Before,
	}
	return c.Fprint(w, n)
}

// Config controls the output of Fprint.
type Config struct {
	// Indent controls the indentation style:
	//   - indents with tabs (default)
	//   - indents with spaces
	Indent Style

	// Width controls the indentation width for spaces.
	Width int

	// Redir controls the output of redirections:
	//   - before commands
	//   - after commands (default)
	//   - space after operators, except for "<&" and ">&" operators
	Redir Style

	// Assign controls the output of assignments in simple commands:
	//   - before redirections (default)
	//   - after redirections
	Assign Style

	// Then controls the output of the if conditional constract:
	//   - newline before the reserved keyword "then"
	Then Style
}

// Fprint pretty-prints an AST node to w with the specified configuration.
func (c *Config) Fprint(w io.Writer, n ast.Node) error {
	p := &printer{
		cfg: *c,
		w:   bufio.NewWriter(w),
	}
	return p.print(n)
}

// Style represents a coding style.
type Style uint

// List of styles.
const (
	Tab = 1 << iota
	Space
	Newline
	Before
	After
)

type printer struct {
	cfg Config
	w   *bufio.Writer

	lv    int
	stack [][]*ast.Redir
}

func (p *printer) indent() {
	if p.cfg.Indent&Space == 0 {
		p.w.Write(bytes.Repeat([]byte{'\t'}, p.lv))
	} else {
		p.w.Write(bytes.Repeat([]byte{' '}, p.lv*p.cfg.Width))
	}
}

func (p *printer) space() {
	p.w.WriteByte(' ')
}

func (p *printer) newline() {
	p.w.WriteByte('\n')
}

func (p *printer) print(n ast.Node) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = errors.New(e.(string))
		}
	}()

	p.push()
	switch n := n.(type) {
	case ast.Command:
		p.command(n)
	case ast.Word:
		p.word(n)
	case ast.WordPart:
		p.wordPart(n)
	case *ast.Comment:
		p.comment(n)
	default:
		return fmt.Errorf("sh/printer: unsupported node: %T", n)
	}
	p.heredoc()
	return p.w.Flush()
}

func (p *printer) command(c ast.Command) {
	switch c := c.(type) {
	case ast.List:
		p.list(c)
	case *ast.AndOrList:
		p.andOrList(c)
	case *ast.Pipeline:
		p.pipeline(c)
	case *ast.Cmd:
		p.cmd(c)
	default:
		panic("sh/printer: unsupported ast.Command")
	}
}

func (p *printer) list(c ast.List) {
	for i, ao := range c {
		if i > 0 {
			p.space()
		}
		p.andOrList(ao)
	}
}

func (p *printer) andOrList(c *ast.AndOrList) {
	p.pipeline(c.Pipeline)
	for _, ao := range c.List {
		p.w.WriteString(" " + ao.Op + " ")
		p.pipeline(ao.Pipeline)
	}
	switch c.Sep {
	case "&":
		p.w.WriteString(" " + c.Sep)
	case ";":
		p.w.WriteString(c.Sep)
	}
}

func (p *printer) pipeline(c *ast.Pipeline) {
	if !c.Bang.IsZero() {
		p.w.WriteString("! ")
	}
	p.cmd(c.Cmd)
	for _, c := range c.List {
		p.w.WriteString(" " + c.Op + " ")
		p.cmd(c.Cmd)
	}
}

func (p *printer) cmd(c *ast.Cmd) {
	if x, ok := c.Expr.(*ast.SimpleCmd); ok {
		p.simpleCmd(x, c.Redirs)
	} else {
		switch x := c.Expr.(type) {
		case *ast.Subshell:
			p.subshell(x)
		case *ast.Group:
			p.group(x)
		case *ast.IfClause:
			p.ifClause(x)
		default:
			panic("sh/printer: unsupported ast.CmdExpr")
		}
		for _, r := range c.Redirs {
			p.space()
			p.redir(r)
		}
	}
}

func (p *printer) simpleCmd(x *ast.SimpleCmd, redirs []*ast.Redir) (err error) {
	var order [3]string
	if p.cfg.Redir&Before == 0 {
		order[0] = "assign"
		order[1] = "args"
		order[2] = "redir"
	} else {
		if p.cfg.Assign&After == 0 {
			order[0] = "assign"
			order[1] = "redir"
		} else {
			order[0] = "redir"
			order[1] = "assign"
		}
		order[2] = "args"
	}
	var sp bool
	for _, s := range order {
		switch s {
		case "assign":
			for _, a := range x.Assigns {
				if sp {
					p.space()
				}
				p.w.WriteString(a.Symbol.Value + a.Op)
				p.word(a.Value)
				sp = true
			}
		case "args":
			for _, w := range x.Args {
				if sp {
					p.space()
				}
				p.word(w)
				sp = true
			}
		case "redir":
			for _, r := range redirs {
				if sp {
					p.space()
				}
				p.redir(r)
				sp = true
			}
		}
	}
	return
}

func (p *printer) redir(r *ast.Redir) {
	if r.Heredoc != nil {
		p.stack[len(p.stack)-1] = append(p.stack[len(p.stack)-1], r)
	}

	if r.N != nil {
		p.w.WriteString(r.N.Value)
	}
	p.w.WriteString(r.Op)
	if p.cfg.Redir&Space != 0 {
		switch r.Op {
		case "<&", ">&":
		default:
			p.space()
		}
	}
	p.word(r.Word)
}

func (p *printer) subshell(x *ast.Subshell) {
	p.w.WriteByte('(')
	if len(x.List) > 1 || x.Lparen.Line() != x.Rparen.Line() {
		p.compoundList(x.List)
		p.newline()
		p.indent()
	} else {
		p.command(x.List[0])
	}
	p.w.WriteByte(')')
}

func (p *printer) group(x *ast.Group) {
	p.w.WriteByte('{')
	if len(x.List) > 1 || x.Lbrace.Line() != x.Rbrace.Line() {
		p.compoundList(x.List)
		p.newline()
		p.indent()
	} else {
		p.space()
		p.command(x.List[0])
		p.space()
	}
	p.w.WriteByte('}')
}

func (p *printer) ifClause(x *ast.IfClause) {
	list := x.If.Line() == x.Fi.Line()
	ifPart := func(word string, cond []ast.Command, cmds []ast.Command) {
		p.w.WriteString(word)
		sep := "_"
		if !list {
			if p.cfg.Then&Newline != 0 {
				if undo := p.trim(cond[len(cond)-1]); undo != nil {
					defer undo()
				}
			}
			p.push()
		}
		for _, c := range cond {
			if sep != "" {
				p.space()
			} else {
				p.w.WriteString("; ")
			}
			p.command(c)
			sep = p.sepOf(c)
		}
		if !list {
			if p.cfg.Then&Newline == 0 {
				if sep != "" {
					p.w.WriteString(" then")
				} else {
					p.w.WriteString("; then")
				}
				p.heredoc()
			} else {
				p.heredoc()
				p.newline()
				p.indent()
				p.w.WriteString("then")
			}
			p.compoundList(cmds)
		} else {
			p.w.WriteString(" then ")
			p.command(cmds[0])
		}
	}

	ifPart("if", x.Cond, x.List)
	for _, e := range x.Else {
		if !list {
			p.newline()
			p.indent()
		} else {
			p.space()
		}
		switch e := e.(type) {
		case *ast.ElifClause:
			ifPart("elif", e.Cond, e.List)
		case *ast.ElseClause:
			p.w.WriteString("else")
			if !list {
				p.compoundList(e.List)
			} else {
				p.space()
				p.command(e.List[0])
			}
		}
	}
	if !list {
		p.newline()
		p.indent()
	} else {
		p.space()
	}
	p.w.WriteString("fi")
}

func (p *printer) compoundList(cmds []ast.Command) {
	p.lv++
	for _, c := range cmds {
		p.newline()
		p.indent()
		p.push()
		p.command(c)
		p.heredoc()
	}
	p.lv--
}

func (p *printer) sepOf(c ast.Command) string {
	switch c := c.(type) {
	case ast.List:
		return c[len(c)-1].Sep
	case *ast.AndOrList:
		return c.Sep
	}
	return ""
}

func (p *printer) trim(c ast.Command) func() {
	switch c := c.(type) {
	case ast.List:
		ao := c[len(c)-1]
		if ao.Sep == ";" {
			ao.Sep = ""
			return func() { ao.Sep = ";" }
		}
	case *ast.AndOrList:
		if c.Sep == ";" {
			c.Sep = ""
			return func() { c.Sep = ";" }
		}
	}
	return nil
}

func (p *printer) push() {
	p.stack = append(p.stack, nil)
}

func (p *printer) heredoc() {
	// pop
	list := p.stack[len(p.stack)-1]
	p.stack = p.stack[:len(p.stack)-1]
	for _, r := range list {
		p.newline()
		p.word(r.Heredoc)
		p.word(r.Delim)
	}
}

func (p *printer) word(w ast.Word) {
	for _, w := range w {
		p.wordPart(w)
	}
}

func (p *printer) wordPart(w ast.WordPart) {
	switch w := w.(type) {
	case *ast.Lit:
		p.lit(w)
	case *ast.Quote:
		p.quote(w)
	case *ast.ParamExp:
		p.paramExp(w)
	default:
		panic("sh/printer: unsupported node in ast.Word")
	}
}

func (p *printer) lit(w *ast.Lit) {
	p.w.WriteString(w.Value)
}

func (p *printer) quote(w *ast.Quote) {
	switch w.Tok {
	case `\`:
		p.w.Write([]byte{'\\', w.Value[0].(*ast.Lit).Value[0]})
	case `'`, `"`:
		p.w.WriteString(w.Tok)
		p.word(w.Value)
		p.w.WriteString(w.Tok)
	}
}

func (p *printer) paramExp(w *ast.ParamExp) {
	if w.Braces {
		switch {
		case w.Op == "":
			p.w.WriteString("${" + w.Name.Value + "}")
		case w.Op == "#" && w.OpPos.Before(w.Name.ValuePos):
			// string length
			p.w.WriteString("${#" + w.Name.Value + "}")
		default:
			p.w.WriteString("${" + w.Name.Value + w.Op)
			p.word(w.Word)
			p.w.WriteByte('}')
		}
	} else {
		p.w.WriteString("$" + w.Name.Value)
	}
}

func (p *printer) comment(c *ast.Comment) {
	p.w.WriteString("#" + c.Text)
}
