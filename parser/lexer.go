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
	"strings"
	"sync/atomic"
	"unicode"

	"github.com/hattya/go.sh/ast"
)

var (
	ops = map[int]string{
		AND:      "&&",
		OR:       "||",
		BREAK:    ";;",
		int('|'): "|",
		int('('): "(",
		int(')'): ")",
		int('<'): "<",
		int('>'): ">",
		CLOBBER:  ">|",
		APPEND:   ">>",
		DUPIN:    "<&",
		DUPOUT:   ">&",
		RDWR:     "<>",
		int('&'): "&",
		int(';'): ";",
	}
	words = map[string]int{
		"!":     Bang,
		"{":     Lbrace,
		"}":     Rbrace,
		"for":   For,
		"case":  Case,
		"esac":  Esac,
		"in":    In,
		"if":    If,
		"elif":  Elif,
		"then":  Then,
		"else":  Else,
		"fi":    Fi,
		"while": While,
		"until": Until,
		"do":    Do,
		"done":  Done,
	}
	builtins = map[string]struct{}{
		"break":    {},
		":":        {},
		"continue": {},
		".":        {},
		"eval":     {},
		"exec":     {},
		"exit":     {},
		"export":   {},
		"readonly": {},
		"return":   {},
		"set":      {},
		"shift":    {},
		"times":    {},
		"trap":     {},
		"unset":    {},
	}
)

type action func() action

type lexer struct {
	name     string
	r        io.RuneScanner
	cmd      ast.Command
	comments []*ast.Comment
	eof      bool
	err      error
	token    chan ast.Node

	stack   []int
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
	case word:
		lval.word = tok.val
		return tok.typ
	}
	return 0
}

func (l *lexer) run() {
	for action := l.lexPipeline; action != nil; {
		action = action()
	}
	close(l.token)
}

func (l *lexer) lexPipeline() action {
	tok := l.translate(l.scanToken())
	if tok == Bang {
		l.emit(Bang)
		tok = l.scanToken()
	}
	return l.lexCmd(tok)
}

func (l *lexer) lexCmd(tok int) action {
	tok = l.translate(tok)
	switch tok {
	case '<', '>', CLOBBER, APPEND, DUPIN, DUPOUT, RDWR, IO_NUMBER:
		return l.lexSimpleCmd(tok)
	case WORD:
		lex := l.lexSimpleCmd
		if len(l.word) == 1 && !l.isAssign() {
			if w, ok := l.word[0].(*ast.Lit); ok && l.isName(w.Value) {
				// lookahead
				l.word = nil
				l.mark(0)
				la := l.scanToken()
				if la == '(' {
					if _, ok := builtins[w.Value]; ok {
						l.last.Store(w.ValuePos)
						l.Error("syntax error: invalid function name")
						return nil
					}
					tok = NAME
					lex = l.lexFuncDef
				}
				// save lookahead token
				word := l.word
				pos := l.pos
				// emit current token
				l.word = ast.Word{w}
				l.pos = w.ValuePos
				l.emit(tok)
				// restore lookahead token
				l.word = word
				l.pos = pos
				tok = la
			}
		}
		return lex(tok)
	case '(':
		return l.lexSubshell
	case Lbrace:
		return l.lexGroup
	case For:
		return l.lexFor
	case Case:
		return l.lexCase
	case BREAK:
		return l.lexCaseBreak
	case If:
		return l.lexIf
	case Elif:
		return l.lexElif
	case Then:
		return l.lexThen
	case Else:
		return l.lexElse
	case While:
		return l.lexWhile
	case Until:
		return l.lexUntil
	case Do:
		return l.lexDo
	}
	return l.lexToken(tok)
}

// translate translates a WORD token to a reserved word token if it is.
func (l *lexer) translate(tok int) int {
	if tok == WORD && len(l.word) == 1 {
		if w, ok := l.word[0].(*ast.Lit); ok {
			if tok, ok := words[w.Value]; ok {
				return tok
			}
		}
	}
	return tok
}

func (l *lexer) lexSimpleCmd(tok int) action {
	if tok == WORD && !l.isAssign() {
		l.emit(WORD)
		return l.lexCmdSuffix
	}
	return l.lexCmdPrefix(tok)
}

func (l *lexer) lexCmdPrefix(tok int) action {
	switch tok {
	case '<', '>', CLOBBER, APPEND, DUPIN, DUPOUT, RDWR:
		// redirection operator
		l.emit(tok)
		if tok = l.scanToken(); tok == WORD {
			goto Prefix
		}
	case IO_NUMBER:
		goto Prefix
	case WORD:
		if !l.isAssign() {
			l.emit(WORD)
			return l.lexCmdSuffix
		}
		tok = ASSIGNMENT_WORD
		goto Prefix
	}
	return l.lexToken(tok)
Prefix:
	l.emit(tok)
	return l.lexCmdPrefix(l.scanToken())
}

func (l *lexer) lexCmdSuffix() action {
	tok := l.scanToken()
	switch tok {
	case '<', '>', CLOBBER, APPEND, DUPIN, DUPOUT, RDWR:
		// redirection operator
		l.emit(tok)
		if tok = l.scanToken(); tok == WORD {
			goto Suffix
		}
	case IO_NUMBER, WORD:
		goto Suffix
	}
	return l.lexToken(tok)
Suffix:
	l.emit(tok)
	return l.lexCmdSuffix
}

// isAssign reports whether the current word is ASSIGNMENT_WORD.
func (l *lexer) isAssign() bool {
	w, ok := l.word[0].(*ast.Lit)
	if !ok {
		return false
	}
	i := strings.IndexRune(w.Value, '=')
	if i < 1 {
		return false
	}
	return l.isName(w.Value[:i])
}

func (l *lexer) lexSubshell() action {
	l.emit('(')
	// push
	l.stack = append(l.stack, ')')
	return l.lexPipeline
}

func (l *lexer) lexGroup() action {
	l.emit(Lbrace)
	// push
	l.stack = append(l.stack, Rbrace)
	return l.lexPipeline
}

func (l *lexer) lexFor() action {
	l.emit(For)
	// name
	switch tok := l.scanToken(); tok {
	case WORD:
		if len(l.word) == 1 {
			if w, ok := l.word[0].(*ast.Lit); ok && l.isName(w.Value) {
				l.emit(NAME)
				break
			}
		}
		l.last.Store(l.word.Pos())
		l.Error("syntax error: invalid for loop variable")
		return nil
	default:
		return l.lexToken(tok)
	}

	switch tok := l.scanToken(); tok {
	case ';':
		l.emit(';')
		if !l.linebreak() {
			return nil
		}
		switch tok = l.translate(l.scanToken()); tok {
		case Do:
			goto Do
		default:
			return l.lexToken(tok)
		}
	case '\n':
		l.emit('\n')
		if !l.linebreak() {
			return nil
		}
		tok = l.scanToken()
		fallthrough
	default:
		switch tok = l.translate(tok); tok {
		case In:
			goto In
		case Do:
			goto Do
		default:
			return l.lexToken(tok)
		}
	}
In:
	l.emit(In)
	for {
		switch tok := l.scanToken(); tok {
		case WORD:
			l.emit(WORD)
		case ';', '\n':
			l.emit(tok)
			if !l.linebreak() {
				return nil
			}
			if tok = l.translate(l.scanToken()); tok == Do {
				goto Do
			}
			fallthrough
		default:
			return l.lexToken(tok)
		}
	}
Do:
	l.emit(Do)
	// push
	l.stack = append(l.stack, Done)
	return l.lexPipeline
}

// isName reports whether s satisfies XBD Name.
func (l *lexer) isName(s string) bool {
	for i, r := range s {
		if !(r == '_' || unicode.IsLetter(r) || (i > 0 && unicode.IsDigit(r))) {
			return false
		}
	}
	return true
}

func (l *lexer) lexCase() action {
	l.emit(Case)
	// word
	if tok := l.scanToken(); tok == WORD {
		l.emit(WORD)
	} else {
		return l.lexToken(tok)
	}
	if !l.linebreak() {
		return nil
	}
	// in
	if tok := l.scanToken(); l.translate(tok) == In {
		l.emit(In)
	} else {
		return l.lexToken(tok)
	}
	if !l.linebreak() {
		return nil
	}
	// push
	l.stack = append(l.stack, Esac)
	return l.lexCaseItem
}

func (l *lexer) lexCaseItem() action {
	tok := l.scanToken()
	// check for esac
	if l.translate(tok) == Esac {
		return l.lexToken(Esac)
	}
	// patterns
	if tok == '(' {
		l.emit('(')
		tok = l.scanToken()
	}
Pattern:
	for {
		switch tok {
		case '|', WORD:
			l.emit(tok)
		case ')':
			l.emit(')')
			if !l.linebreak() {
				return nil
			}
			break Pattern
		default:
			return l.lexToken(tok)
		}
		tok = l.scanToken()
	}
	return l.lexPipeline
}

func (l *lexer) lexCaseBreak() action {
	l.emit(BREAK)
	if !l.linebreak() {
		return nil
	}
	return l.lexCaseItem
}

func (l *lexer) lexIf() action {
	l.emit(If)
	// push
	l.stack = append(l.stack, Then)
	return l.lexPipeline
}

func (l *lexer) lexElif() action {
	l.emit(Elif)
	// pop & push
	if len(l.stack) != 0 && l.stack[len(l.stack)-1] == Fi {
		l.stack[len(l.stack)-1] = Then
		return l.lexPipeline
	}
	return nil
}

func (l *lexer) lexThen() action {
	l.emit(Then)
	// pop & push
	if len(l.stack) != 0 && l.stack[len(l.stack)-1] == Then {
		l.stack[len(l.stack)-1] = Fi
		return l.lexPipeline
	}
	return nil
}

func (l *lexer) lexElse() action {
	l.emit(Else)
	// pop & push
	if len(l.stack) != 0 && l.stack[len(l.stack)-1] == Fi {
		l.stack[len(l.stack)-1] = Fi
		return l.lexPipeline
	}
	return nil
}

func (l *lexer) lexWhile() action {
	l.emit(While)
	// push
	l.stack = append(l.stack, Do)
	return l.lexPipeline
}

func (l *lexer) lexUntil() action {
	l.emit(Until)
	// push
	l.stack = append(l.stack, Do)
	return l.lexPipeline
}

func (l *lexer) lexDo() action {
	l.emit(Do)
	// pop & push
	if len(l.stack) != 0 && l.stack[len(l.stack)-1] == Do {
		l.stack[len(l.stack)-1] = Done
		return l.lexPipeline
	}
	return nil
}

func (l *lexer) lexFuncDef(_ int) action {
	l.emit('(')
	if tok := l.scanToken(); tok == ')' {
		l.emit(')')
		if !l.linebreak() {
			return nil
		}
	} else {
		return l.lexToken(tok)
	}
	return l.lexCmd(l.scanToken())
}

func (l *lexer) lexToken(tok int) action {
	switch tok {
	case AND, OR:
		l.emit(tok)
		if l.linebreak() {
			return l.lexPipeline
		}
	case BREAK:
		return l.lexCaseBreak
	case '|':
		l.emit('|')
		if l.linebreak() {
			return l.lexCmd(l.scanToken())
		}
	case '&', ';':
		l.emit(tok)
		if len(l.stack) != 0 {
			return l.lexPipeline
		}
	case '\n':
		if len(l.stack) != 0 {
			l.emit('\n')
			return l.lexPipeline
		}
	case ')', Rbrace, Esac, Fi, Done:
		l.emit(tok)
		// pop
		if len(l.stack) != 0 && l.stack[len(l.stack)-1] == tok {
			l.stack = l.stack[:len(l.stack)-1]
			return l.lexRedir
		}
	default:
		if tok > 0 {
			l.emit(tok)
		}
	}
	return nil
}

func (l *lexer) lexRedir() action {
	tok := l.scanToken()
	switch tok {
	case '<', '>', CLOBBER, APPEND, DUPIN, DUPOUT, RDWR:
		// redirection operator
		l.emit(tok)
		if tok = l.scanToken(); tok == WORD {
			goto Redir
		}
	case IO_NUMBER:
		goto Redir
	}
	return l.lexToken(tok)
Redir:
	l.emit(tok)
	return l.lexRedir
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
		case '&', '(', ')', ';', '|':
			// operator
			if l.lit(); len(l.word) != 0 {
				l.unread()
				return WORD
			}
			return l.scanOp(r)
		case '<', '>':
			// redirection operator
			if l.lit(); len(l.word) != 0 {
				l.unread()
				if len(l.word) == 1 {
					if w, ok := l.word[0].(*ast.Lit); ok {
						for _, r := range w.Value {
							if !('0' <= r && r <= '9') {
								return WORD
							}
						}
						return IO_NUMBER
					}
				}
				return WORD
			}
			return l.scanOp(r)
		case '\\', '\'', '"':
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
		if r, _ = l.read(); l.err == nil {
			if r == '&' {
				op = AND
			} else {
				l.unread()
			}
		}
	case '(':
		op = int('(')
	case ')':
		op = int(')')
	case ';':
		op = int(';')
		if r, _ = l.read(); l.err == nil {
			if r == ';' {
				op = BREAK
			} else {
				l.unread()
			}
		}
	case '<':
		op = int('<')
		if r, _ = l.read(); l.err == nil {
			switch r {
			case '&':
				op = DUPIN
			case '>':
				op = RDWR
			default:
				l.unread()
			}
		}
	case '>':
		op = int('>')
		if r, _ = l.read(); l.err == nil {
			switch r {
			case '&':
				op = DUPOUT
			case '>':
				op = APPEND
			case '|':
				op = CLOBBER
			default:
				l.unread()
			}
		}
	case '|':
		op = int('|')
		if r, _ = l.read(); l.err == nil {
			if r == '|' {
				op = OR
			} else {
				l.unread()
			}
		}
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
		// escape character
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
	case '\'':
		// single-quotes
		for {
			r, err := l.read()
			if err != nil {
				if err == io.EOF {
					l.last.Store(q.TokPos)
					l.Error("syntax error: reached EOF while parsing single-quotes")
				}
				return false
			}

			if r == '\'' {
				break
			}
			l.b.WriteRune(r)
		}
		q.Value = ast.Word{
			&ast.Lit{
				ValuePos: l.pos,
				Value:    l.b.String(),
			},
		}
		l.b.Reset()
		l.word = append(l.word, q)
	case '"':
		// double-quotes
		var err error
		// save current word
		word := l.word
		l.word = nil
	QQ:
		for {
			r, err = l.read()
			if err != nil {
				break QQ
			}

			switch r {
			case '\\':
				// escape character
				r, err = l.read()
				if err != nil {
					break QQ
				}

				switch r {
				case '\n', '"', '$', '\\', '`':
					l.lit()
					if r != '\n' {
						l.word = append(l.word, &ast.Quote{
							TokPos: l.pos,
							Tok:    "\\",
							Value: ast.Word{
								&ast.Lit{
									ValuePos: ast.NewPos(l.line, l.col-1),
									Value:    string(r),
								},
							},
						})
					}
					l.mark(0)
				default:
					l.b.WriteRune('\\')
					l.b.WriteRune(r)
				}
			case '"':
				// right double-quote
				l.lit()
				break QQ
			default:
				l.b.WriteRune(r)
			}
		}
		if err != nil {
			if err == io.EOF {
				l.last.Store(q.TokPos)
				l.Error("syntax error: reached EOF while parsing double-quotes")
			}
			return false
		}
		// append to current word
		q.Value = l.word
		l.word = append(word, q)
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
	case IO_NUMBER, WORD, NAME, ASSIGNMENT_WORD:
		l.token <- word{
			typ: tok,
			val: l.word,
		}
		l.word = nil
	default:
		if len(l.word) != 0 {
			w := l.word[0].(*ast.Lit)
			l.token <- token{
				typ: tok,
				pos: w.ValuePos,
				val: w.Value,
			}
			l.word = nil
		} else {
			l.token <- token{
				typ: tok,
				pos: l.pos,
				val: ops[tok],
			}
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
	if l.err != nil && strings.Contains(e, ": unexpected EOF") {
		return // lexing was interrupted
	}
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

type word struct {
	typ int
	val ast.Word
}

func (w word) Pos() ast.Pos { return w.val.Pos() }
func (w word) End() ast.Pos { return w.val.End() }
