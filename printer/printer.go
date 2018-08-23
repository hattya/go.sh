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
	c := new(Config)
	return c.Fprint(w, n)
}

// Config controls the output of Fprint.
type Config struct {
}

// Fprint pretty-prints an AST node to w with the specified configuration.
func (c *Config) Fprint(w io.Writer, n ast.Node) error {
	p := &printer{
		cfg: *c,
		w:   bufio.NewWriter(w),
	}
	return p.print(n)
}

type printer struct {
	cfg Config
	w   *bufio.Writer
}

func (p *printer) print(n ast.Node) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = errors.New(e.(string))
		}
	}()

	switch n := n.(type) {
	case ast.Word:
		p.word(n)
	case ast.WordPart:
		p.wordPart(n)
	case *ast.Comment:
		p.comment(n)
	default:
		return fmt.Errorf("sh/printer: unsupported node: %T", n)
	}
	return p.w.Flush()
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
