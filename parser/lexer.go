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
// (POSIX.1-2017).
package parser

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"unicode"

	"github.com/hattya/go.sh/ast"
	"github.com/hattya/go.sh/printer"
)

var (
	ops = map[int]string{
		AND:      "&&",
		OR:       "||",
		LAE:      "((",
		RAE:      "))",
		BREAK:    ";;",
		int('|'): "|",
		int('('): "(",
		int(')'): ")",
		int('<'): "<",
		int('>'): ">",
		CLOBBER:  ">|",
		APPEND:   ">>",
		HEREDOC:  "<<",
		HEREDOCI: "<<-",
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

	errParamExp = errors.New("syntax error: invalid parameter expansion")
)

type action func() action

type lexer struct {
	name     string
	r        io.RuneScanner
	cmd      ast.Command
	comments []*ast.Comment
	cmdSubst rune
	eof      bool
	err      error
	token    chan ast.Node
	done     chan struct{}

	stack     []int
	arithExpr bool
	paren     int
	heredoc   heredoc
	word      ast.Word
	b         bytes.Buffer
	line      int
	col       int
	prevCol   int
	pos       ast.Pos
	last      atomic.Value
}

func newLexer(name string, r io.RuneScanner) *lexer {
	l := &lexer{
		name:    name,
		r:       r,
		token:   make(chan ast.Node),
		heredoc: heredoc{c: make(chan struct{})},
		line:    1,
		col:     1,
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
	if l.done != nil {
		close(l.done)
	}
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
	case '<', '>', CLOBBER, APPEND, HEREDOC, HEREDOCI, DUPIN, DUPOUT, RDWR, IO_NUMBER:
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
	case LAE:
		return l.lexArithEval
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
	case '<', '>', CLOBBER, APPEND, HEREDOC, HEREDOCI, DUPIN, DUPOUT, RDWR:
		// redirection operator
		l.emit(tok)
		if tok = l.scanRedir(tok); tok == WORD {
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
	case '<', '>', CLOBBER, APPEND, HEREDOC, HEREDOCI, DUPIN, DUPOUT, RDWR:
		// redirection operator
		l.emit(tok)
		if tok = l.scanRedir(tok); tok == WORD {
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

func (l *lexer) lexArithEval() action {
	l.emit(LAE)
	// push
	l.stack = append(l.stack, RAE)
	for {
		if tok := l.scanArithExpr(); tok == WORD {
			l.emit(WORD)
		} else {
			return l.lexToken(tok)
		}
	}
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
	// clear
	l.paren = 0
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
		return l.lexPipeline
	case '\n':
		switch {
		case l.heredoc.exists():
			return l.lexHeredoc
		case len(l.stack) != 0:
			l.emit('\n')
			return l.lexPipeline
		}
	case ')', RAE:
		if l.cmdSubst != 0 && len(l.stack) == 1 {
			l.emit(tok)
			l.stack = nil
			break
		}
		fallthrough
	case Rbrace, Esac, Fi, Done:
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
	case '<', '>', CLOBBER, APPEND, HEREDOC, HEREDOCI, DUPIN, DUPOUT, RDWR:
		// redirection operator
		l.emit(tok)
		if tok = l.scanRedir(tok); tok == WORD {
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

func (l *lexer) lexHeredoc() action {
	var b bytes.Buffer
	find := func(r *ast.Redir, delim string) bool {
		for i := len(l.word) - 1; i >= 0; i-- {
			if l.word[i].Pos().Col() == 1 {
				b.Reset()
				printer.Fprint(&b, l.word[i:])
				if s := b.String(); strings.ContainsRune(s, '\n') {
					break
				} else if s == delim {
					r.Heredoc = l.word[:i]
					r.Delim = l.word[i:]
					l.word = nil
					return true
				}
			}
		}
		return false
	}
	for h := l.heredoc.pop(); h != nil; h = l.heredoc.pop() {
		l.mark(0)
		// unquote
		var word ast.Word
		var quoted bool
		for _, w := range h.Word {
			if q, ok := w.(*ast.Quote); ok {
				word = append(word, q.Value...)
				quoted = true
			} else {
				word = append(word, w)
			}
		}
		// token → string
		b.Reset()
		printer.Fprint(&b, word)
		delim := b.String()
	Heredoc:
		for {
			r, err := l.read()
			if err != nil {
				if !l.heredoc.exists() {
					if l.lit(); find(h, delim) {
						return nil
					}
				}
				goto Error
			}

			switch {
			case r == '\n':
				// <newline>
				if l.lit(); find(h, delim) {
					break Heredoc
				}
				// store <newline>
				if w1, ok := l.word[len(l.word)-1].(*ast.Lit); ok {
					w1.Value += "\n"
					// concatenate
					if len(l.word) > 1 {
						if w2, ok := l.word[len(l.word)-2].(*ast.Lit); ok && w2.End() == w1.Pos() {
							w2.Value += w1.Value
							l.word = l.word[:len(l.word)-1]
						}
					}
				} else {
					l.b.WriteRune('\n')
					l.lit()
				}
				l.mark(0)
			case !quoted:
				switch r {
				case '\\':
					// escape character
					r, err = l.read()
					if err != nil {
						goto Error
					}
					l.esc(r)
				case '$':
					// parameter expansion
					l.lit()
					l.mark(-1)
					if !l.scanParamExp() {
						return nil
					}
				case '`':
					// command substitution
					l.lit()
					l.mark(-1)
					if !l.scanCmdSubst('`') {
						return nil
					}
				default:
					l.b.WriteRune(r)
				}
			default:
				l.b.WriteRune(r)
			}

			continue
		Error:
			if err == io.EOF {
				l.last.Store(h.OpPos)
				l.Error("syntax error: here-document delimited by EOF")
			}
			return nil
		}
	}
	return l.lexToken('\n')
}

func (l *lexer) scanArithExpr() int {
	for {
		r, err := l.read()
		if err != nil {
			if err == io.EOF {
				l.Error("syntax error: reached EOF while looking for matching '))'")
			}
			return -1
		}

		switch r {
		case '(', ')':
			// operator
			if l.lit(); len(l.word) != 0 {
				l.unread()
				return WORD
			}
			if l.scanOp(r) == RAE {
				return RAE
			} else {
				l.b.WriteRune(r)
				l.lit()
				return WORD
			}
		case '\\', '\'', '"':
			// quoting
			l.lit()
			l.mark(-1)
			if !l.scanQuote(r) {
				return -1
			}
		case '$':
			// parameter expansion
			l.lit()
			l.mark(-1)
			if !l.scanParamExp() {
				return -1
			}
		case '`':
			// command substitution
			l.lit()
			l.mark(-1)
			if !l.scanCmdSubst('`') {
				return -1
			}
		case '\t', ' ':
			// <blank>
			fallthrough
		case '\n':
			// <newline>
			if l.lit(); len(l.word) != 0 {
				return WORD
			}
			l.mark(0)
		default:
			l.b.WriteRune(r)
		}
	}
}

func (l *lexer) scanRedir(tok int) int {
	var heredoc bool
	switch tok {
	case HEREDOC, HEREDOCI:
		heredoc = true
	}
	tok = l.scanToken()
	if tok == WORD && heredoc {
		var b bytes.Buffer
		printer.Fprint(&b, l.word)
		if strings.ContainsRune(b.String(), '\n') {
			l.last.Store(l.word.Pos())
			l.Error(`syntax error: here-document delimiter contains '\n'`)
			return -1
		}
		l.heredoc.inc()
	}
	return tok
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
		case '$':
			// parameter expansion
			l.lit()
			l.mark(-1)
			if !l.scanParamExp() {
				return -1
			}
		case '`':
			// command substitution
			if l.cmdSubst != '`' {
				l.lit()
				l.mark(-1)
				if !l.scanCmdSubst('`') {
					return -1
				}
			} else {
				if l.lit(); len(l.word) != 0 {
					l.unread()
					return WORD
				}
				if len(l.stack) != 0 {
					return ')'
				}
				return '('
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
		l.paren++
		if l.paren == 1 {
			if r, _ = l.read(); l.err == nil {
				if r == '(' {
					op = LAE
					l.paren++
					l.arithExpr = true
				} else {
					l.unread()
				}
			}
		}
	case ')':
		op = int(')')
		l.paren--
		if l.arithExpr && l.paren == 1 {
			if r, _ = l.read(); l.err == nil {
				if r == ')' {
					op = RAE
					l.paren--
					l.arithExpr = false
				} else {
					l.unread()
				}
			}
		}
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
			case '<':
				op = HEREDOC
				if r, _ = l.read(); l.err == nil {
					if r == '-' {
						op = HEREDOCI
					} else {
						l.unread()
					}
				}
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
				l.esc(r)
			case '$':
				// parameter expansion
				l.lit()
				l.mark(-1)
				if !l.scanParamExp() {
					return false
				}
			case '`':
				// command substitution
				l.lit()
				l.mark(-1)
				if !l.scanCmdSubst('`') {
					return false
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

func (l *lexer) scanParamExp() bool {
	r, err := l.read()
	if err != nil {
		if err == io.EOF {
			l.b.WriteRune('$')
			return true
		}
		return false
	}

	var pe *ast.ParamExp
	switch r {
	case '{':
		// enclosed in braces
		return l.scanParamExpInBraces()
	case '(':
		// command substitution
		l.mark(-1)
		return l.scanCmdSubst('(')
	case '@', '*', '#', '?', '-', '$', '!', '0':
		// special parameter
		pe = &ast.ParamExp{
			Dollar: l.pos,
			Name: &ast.Lit{
				ValuePos: ast.NewPos(l.line, l.col-1),
				Value:    string(r),
			},
		}
	default:
		pe = &ast.ParamExp{Dollar: l.pos}
		l.mark(-1)
		switch {
		case unicode.IsDigit(r):
			// positional parameter
			l.b.WriteRune(r)
		case r == '_' || unicode.IsLetter(r):
			// XBD Name
			for l.isNameRune(r) {
				l.b.WriteRune(r)
				if r, err = l.read(); err != nil {
					if err == io.EOF {
						break
					}
					return false
				}
			}
			if err == nil {
				l.unread()
			}
		default:
			// continue as WORD
			l.unread()
			l.b.WriteRune('$')
			l.mark(-1)
			return true
		}
		pe.Name = &ast.Lit{
			ValuePos: l.pos,
			Value:    l.b.String(),
		}
		l.b.Reset()
	}
	l.word = append(l.word, pe)
	l.mark(0)
	return true
}

func (l *lexer) scanParamExpInBraces() bool {
	pe := &ast.ParamExp{
		Dollar: l.pos,
		Braces: true,
	}
	l.mark(0)
	// inside braces
	r, err := l.read()
	if err == nil {
		if r == '#' {
			if r, err = l.read(); err == nil {
				switch r {
				case ':', '=', '+', '%', '}':
					// special parameter
					pe.Name = &ast.Lit{
						ValuePos: l.pos,
						Value:    "#",
					}
					l.mark(-1)
					goto Op
				case '#', '?', '-':
					v := r
					if r, err = l.read(); err == nil {
						l.unread()
						if r != '}' {
							// special parameter
							pe.Name = &ast.Lit{
								ValuePos: l.pos,
								Value:    "#",
							}
							l.mark(-1)
							r = v
							goto Op
						} else {
							// string length
							pe.OpPos = l.pos
							pe.Op = "#"
							l.mark(-1)
							pe.Name = &ast.Lit{
								ValuePos: l.pos,
								Value:    string(v),
							}
							goto Rbrace
						}
					}
				default:
					// string length
					l.unread()
					pe.OpPos = l.pos
					pe.Op = "#"
					l.mark(0)
				}
			}
		} else {
			l.unread()
		}
	}
	if err != nil {
		goto Error
	}
	// name
	switch r, _ = l.read(); r {
	case '@', '*', '?', '-', '$', '!', '0':
		// special parameter
		l.b.WriteRune(r)
	default:
		// XBD Name
		for l.isNameRune(r) {
			l.b.WriteRune(r)
			if r, err = l.read(); err != nil {
				goto Error
			}
		}
		l.unread()
		if l.b.Len() == 0 {
			err = errParamExp
			goto Error
		}
	}
	pe.Name = &ast.Lit{
		ValuePos: l.pos,
		Value:    l.b.String(),
	}
	l.b.Reset()
	l.mark(0)
	// op
	r, err = l.read()
Op:
	if err == nil {
		switch r {
		case ':':
			if r, err = l.read(); err == nil {
				switch r {
				case '-', '=', '?', '+':
					pe.Op = ":" + string(r)
				default:
					l.unread()
					err = errParamExp
				}
			}
		case '-', '=', '?', '+':
			pe.Op = string(r)
		case '%', '#':
			pe.Op = string(r)
			if r, err = l.read(); err == nil {
				switch r {
				case '%', '#':
					if s := string(r); pe.Op == s {
						pe.Op += s
					} else {
						l.unread()
						err = errParamExp
					}
				default:
					l.unread()
				}
			}
		default:
			l.unread()
			goto Rbrace
		}
	}
	switch {
	case err != nil:
		goto Error
	case pe.Op != "":
		pe.OpPos = l.pos
		l.mark(0)
	}
	// word
	{
		// save current word
		word := l.word
		l.word = nil
	Word:
		for {
			r, err = l.read()
			if err != nil {
				goto Error
			}

			switch r {
			case '\\', '\'', '"':
				// quoting
				l.lit()
				l.mark(-1)
				if !l.scanQuote(r) {
					return false
				}
			case '$':
				// parameter expansion
				l.lit()
				l.mark(-1)
				if !l.scanParamExp() {
					return false
				}
			case '}':
				// right brace
				l.unread()
				l.lit()
				break Word
			default:
				l.b.WriteRune(r)
			}
		}
		// restore current word
		pe.Word = l.word
		l.word = word
	}
Rbrace:
	if r, err = l.read(); err != nil || r != '}' {
		goto Error
	}
	l.word = append(l.word, pe)
	l.mark(0)
	return true
Error:
	switch err {
	case nil, io.EOF:
		l.last.Store(pe.Dollar)
		l.Error("syntax error: reached EOF while looking for matching '}'")
	case errParamExp:
		l.last.Store(pe.Dollar)
		l.Error(err.Error())
	}
	return false
}

// isNameRune reports whether r can be used in XBD Name.
func (l *lexer) isNameRune(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

func (l *lexer) scanCmdSubst(r rune) bool {
	off := 0
	switch r {
	case '(':
		r = '$'
		off = -1
		fallthrough
	case '`':
		l.unread()
		left := l.pos
		// nest
		ll := &lexer{
			name:     l.name,
			r:        l.r,
			cmdSubst: r,
			token:    make(chan ast.Node),
			done:     make(chan struct{}),
			heredoc:  heredoc{c: make(chan struct{})},
			line:     l.line,
			col:      l.col,
		}
		ll.mark(off)
		ll.last.Store(ll.pos)
		go ll.run()
		yyParse(ll)
		<-ll.done
		if ll.err != nil {
			l.err = ll.err
			if len(ll.stack) == 0 && r == '`' {
				err := l.err.(Error)
				l.err = Error{
					Name: err.Name,
					Pos:  err.Pos,
					Msg:  "syntax error: unexpected '`'",
				}
			}
			break
		}
		// apply changes
		l.comments = append(l.comments, ll.comments...)
		l.line = ll.line
		l.col = ll.col
		l.pos = ll.pos
		// append to current word
		switch x := ll.cmd.(*ast.Cmd).Expr.(type) {
		case *ast.Subshell:
			l.word = append(l.word, &ast.CmdSubst{
				Dollar: r == '$',
				Left:   left,
				List:   x.List,
				Right:  x.Rparen,
			})
		case *ast.ArithEval:
			l.word = append(l.word, &ast.ArithExp{
				Left:  ast.NewPos(left.Line(), left.Col()-1),
				Expr:  x.Expr,
				Right: x.Right,
			})
		}
		return true
	}
	return false
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

func (l *lexer) esc(r rune) {
	switch r {
	case '\n', '"', '$', '\\', '`':
		l.lit()
		if r != '\n' {
			l.word = append(l.word, &ast.Quote{
				TokPos: ast.NewPos(l.line, l.col-2),
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

type heredoc struct {
	c     chan struct{}
	n     uint32
	mu    sync.Mutex
	stack []*ast.Redir
}

func (h *heredoc) exists() bool {
	return atomic.LoadUint32(&h.n) != 0
}

func (h *heredoc) inc() {
	atomic.AddUint32(&h.n, 1)
}

func (h *heredoc) push(r *ast.Redir) {
	h.mu.Lock()
	h.stack = append(h.stack, r)
	h.mu.Unlock()
	// incoming
	select {
	case h.c <- struct{}{}:
	default:
	}
}

func (h *heredoc) pop() *ast.Redir {
	for atomic.LoadUint32(&h.n) != 0 {
		h.mu.Lock()
		if n := len(h.stack); n != 0 {
			r := h.stack[0]
			h.stack = h.stack[1:]
			h.mu.Unlock()
			atomic.AddUint32(&h.n, ^uint32(0))
			if n > 1 {
				// not empty
				select {
				case h.c <- struct{}{}:
				default:
				}
			}
			return r
		}
		h.mu.Unlock()
		// wait
		<-h.c
	}
	return nil
}
