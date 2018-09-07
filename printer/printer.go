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
	"errors"
	"fmt"
	"io"

	"github.com/hattya/go.sh/ast"
)

// Fprint pretty-prints an AST node to w.
func Fprint(w io.Writer, n ast.Node) error {
	c := &Config{
		Redir:  After,
		Assign: Before,
	}
	return c.Fprint(w, n)
}

// Config controls the output of Fprint.
type Config struct {
	// Redir controls the output of redirections:
	//   - before commands
	//   - after commands (default)
	//   - space after operators, except for "<&" and ">&" operators
	Redir Style

	// Assign controls the output of assignments in simple commands:
	//   - before redirections (default)
	//   - after redirections
	Assign Style
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
	Space = 1 << iota
	Before
	After
)

type printer struct {
	cfg Config
	w   *bufio.Writer

	stack [][]*ast.Redir
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
	case *ast.Pipeline:
		p.pipeline(c)
	case *ast.Cmd:
		p.cmd(c)
	default:
		panic("sh/printer: unsupported ast.Command")
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
		panic("sh/printer: unsupported ast.CmdExpr")
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

func (p *printer) comment(c *ast.Comment) {
	p.w.WriteString("#" + c.Text)
}
