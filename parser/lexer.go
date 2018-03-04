//
// go.sh/parser :: lexer.go
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

//go:generate goyacc -l -o parser.go parser.go.y

// Package parser implemnets a parser for the Shell Command Language
// (POSIX.1-2008).
package parser

import (
	"bytes"
	"fmt"
	"io"
	"sync/atomic"

	"github.com/hattya/go.sh/ast"
)

var ops = map[int]string{
	int('&'): "&",
	int(';'): ";",
}

type action func() action

type lexer struct {
	name     string
	r        io.RuneScanner
	cmd      ast.Command
	comments []*ast.Comment
	eof      bool
	err      error
	token    chan ast.Node

	word    ast.Word
	b       bytes.Buffer
	line    int
	col     int
	prevCol int
	pos     ast.Pos
	last    atomic.Value
}

func newLexer(name string, r io.RuneScanner) *lexer {
	l := &lexer{
		name:  name,
		r:     r,
		token: make(chan ast.Node),
		line:  1,
		col:   1,
	}
	l.mark(0)
	go l.run()
	return l
}

func (l *lexer) Lex(lval *yySymType) int {
	tok := <-l.token
	switch tok := tok.(type) {
	case token:
		lval.token = tok
		return tok.typ
	case ast.Word:
		lval.word = tok
		return WORD
	}
	return 0
}

func (l *lexer) run() {
	for action := l.lexWord; action != nil; {
		action = action()
	}
	close(l.token)
}

func (l *lexer) lexWord() action {
	tok := l.scanToken()
	if tok == WORD {
		l.emit(WORD)
		return l.lexWord
	}
	return l.lexToken(tok)
}

func (l *lexer) lexToken(tok int) action {
	switch tok {
	case '\n':
	default:
		if tok > 0 {
			l.emit(tok)
		}
	}
	return nil
}

func (l *lexer) scanToken() int {
	for {
		r, err := l.read()
		if err != nil {
			if err == io.EOF {
				if l.lit(); len(l.word) != 0 {
					return WORD
				}
				return 0
			}
			return -1
		}

		switch r {
		case '&', ';':
			// operator
			if l.lit(); len(l.word) != 0 {
				l.unread()
				return WORD
			}
			return l.scanOp(r)
		case '\\':
			// quoting
			l.lit()
			l.mark(-1)
			if !l.scanQuote(r) {
				return -1
			}
		case '\t', ' ':
			// <blank>
			if l.lit(); len(l.word) != 0 {
				return WORD
			}
			l.mark(0)
		case '\n':
			// <newline>
			if l.lit(); len(l.word) != 0 {
				l.unread()
				return WORD
			}
			return int(r)
		case '#':
			// comment
			l.unread()
			if l.lit(); len(l.word) != 0 {
				return WORD
			}
			if !l.linebreak() {
				return -1
			}
		default:
			l.b.WriteRune(r)
		}
	}
}

func (l *lexer) scanOp(r rune) int {
	op := -1
	switch r {
	case '&':
		op = int('&')
	case ';':
		op = int(';')
	}
	return op
}

func (l *lexer) scanQuote(r rune) bool {
	q := &ast.Quote{
		TokPos: l.pos,
		Tok:    string(r),
	}
	l.mark(0)
	switch r {
	case '\\':
		r, err := l.read()
		if err != nil {
			if err == io.EOF {
				l.word = append(l.word, q)
				break
			}
			return false
		}

		if r != '\n' {
			q.Value = ast.Word{
				&ast.Lit{
					ValuePos: l.pos,
					Value:    string(r),
				},
			}
			l.word = append(l.word, q)
		}
	}
	l.mark(0)
	return true
}

func (l *lexer) linebreak() bool {
	for hash := false; ; {
		r, err := l.read()
		if err != nil {
			l.comment()
			return false
		}

		switch r {
		case '\n':
			// <newline>
			hash = false
			l.comment()
			l.mark(0)
		case '#':
			// comment
			hash = true
			l.mark(-1)
		default:
			if !hash {
				l.unread()
				return true
			}
			l.b.WriteRune(r)
		}
	}
}

func (l *lexer) comment() {
	if l.b.Len() != 0 {
		l.comments = append(l.comments, &ast.Comment{
			Hash: l.pos,
			Text: l.b.String(),
		})
		l.b.Reset()
	}
}

func (l *lexer) lit() {
	if l.b.Len() != 0 {
		l.word = append(l.word, &ast.Lit{
			ValuePos: l.pos,
			Value:    l.b.String(),
		})
		l.b.Reset()
	}
}

func (l *lexer) emit(tok int) {
	l.last.Store(l.pos)
	switch tok {
	case WORD:
		l.token <- l.word
		l.word = nil
	default:
		l.token <- token{
			typ: tok,
			pos: l.pos,
			val: ops[tok],
		}
	}
	l.mark(0)
}

func (l *lexer) mark(off int) {
	l.pos = ast.NewPos(l.line, l.col+off)
}

func (l *lexer) read() (rune, error) {
	r, _, err := l.r.ReadRune()
	switch {
	case err != nil:
		if err == io.EOF {
			l.eof = true
		} else {
			l.err = err
		}
	case r == '\n':
		l.prevCol = l.col
		l.line++
		l.col = 1
	default:
		l.col++
	}
	return r, err
}

func (l *lexer) unread() {
	l.r.UnreadRune()
	if l.col == 1 {
		l.line--
		l.col = l.prevCol
	} else {
		l.col--
	}
}

func (l *lexer) Error(e string) {
	l.err = Error{
		Name: l.name,
		Pos:  l.last.Load().(ast.Pos),
		Msg:  e,
	}
}

// Error represents a syntax error
type Error struct {
	Name string
	Pos  ast.Pos
	Msg  string
}

func (e Error) Error() string {
	return fmt.Sprintf("%v:%v:%v: %s", e.Name, e.Pos.Line(), e.Pos.Col(), e.Msg)
}

type token struct {
	typ int
	pos ast.Pos
	val string
}

func (t token) Pos() ast.Pos { return t.pos }
func (t token) End() ast.Pos { return ast.NewPos(t.pos.Line(), t.pos.Col()+len(t.val)) }
