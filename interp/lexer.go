//
// go.sh/interp :: lexer.go
//
//   Copyright (c) 2021 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
//

//go:generate goyacc -l -o arith.go arith.go.y

package interp

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"unicode"
)

var ops = map[int]string{
	'(': "(",
	')': ")",
	INC: "++",
	DEC: "--",
	'+': "+",
	'-': "-",
	'~': "~",
	'!': "!",
	'*': "*",
	'/': "/",
	'%': "%",
	LSH: "<<",
	RSH: ">>",
	'<': "<",
	'>': ">",
	LE:  "<=",
	GE:  ">=",
}

type lexer struct {
	env   *ExecEnv
	r     io.RuneScanner
	n     int
	token chan interface{}

	mu     sync.Mutex
	err    error
	cancel chan struct{}

	b strings.Builder
}

func newLexer(env *ExecEnv, r io.RuneScanner) *lexer {
	l := &lexer{
		env:    env,
		r:      r,
		token:  make(chan interface{}),
		cancel: make(chan struct{}),
	}
	go l.run()
	return l
}

func (l *lexer) Lex(lval *yySymType) int {
	switch tok := (<-l.token).(type) {
	case token:
		lval.expr.s = tok.val
		return tok.typ
	case int:
		lval.op = ops[tok]
		return tok
	}
	return 0
}

func (l *lexer) run() {
	defer func() {
		close(l.token)

		if e := recover(); e != nil {
			// re-panic
			panic(e)
		}
	}()

	for action := l.lexToken; action != nil; {
		action = action()
	}
}

func (l *lexer) lexToken() action {
Read:
	r, err := l.read()
	if err != nil {
		return nil
	}
	switch r {
	case ' ', '\t', '\n':
		goto Read
	}
	l.unread()

	switch {
	case '0' <= r && r <= '9':
		return l.lexNumber
	case r == '_' || unicode.IsLetter(r):
		return l.lexIdent
	}
	return l.lexOp
}

func (l *lexer) lexNumber() action {
	r, _ := l.read()
	l.b.WriteRune(r)
	var hex bool
	if r == '0' {
		r, err := l.read()
		switch {
		case err != nil:
			goto Number
		case r == 'X' || r == 'x':
			hex = true
		case r < '0' || '9' < r:
			l.unread()
			goto Number
		}
		l.b.WriteRune(r)
	}

	for {
		r, err := l.read()
		switch {
		case err != nil:
			goto Number
		case '0' <= r && r <= '9' || hex && ('A' <= r && r <= 'Z' || 'a' <= r && r <= 'z'):
			l.b.WriteRune(r)
		default:
			l.unread()
			goto Number
		}
	}
Number:
	l.emit(NUMBER)
	return l.lexToken
}

func (l *lexer) lexIdent() action {
	for {
		r, err := l.read()
		switch {
		case err != nil:
			goto Ident
		case r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r):
			l.b.WriteRune(r)
		default:
			l.unread()
			goto Ident
		}
	}
Ident:
	l.emit(IDENT)
	return l.lexToken
}

func (l *lexer) lexOp() action {
	var op int
	switch r, _ := l.read(); r {
	case '(', ')', '~', '!', '*', '/', '%':
		op = int(r)
	case '+':
		op = '+'
		if r, err := l.read(); err == nil {
			if r == '+' {
				op = INC
			} else {
				l.unread()
			}
		}
	case '-':
		op = '-'
		if r, err := l.read(); err == nil {
			if r == '-' {
				op = DEC
			} else {
				l.unread()
			}
		}
	case '<':
		op = '<'
		if r, err := l.read(); err == nil {
			switch r {
			case '<':
				op = LSH
			case '=':
				op = LE
			default:
				l.unread()
			}
		}
	case '>':
		op = '>'
		if r, err := l.read(); err == nil {
			switch r {
			case '>':
				op = RSH
			case '=':
				op = GE
			default:
				l.unread()
			}
		}
	default:
		l.Error(fmt.Sprintf("unexpected %q", r))
		return nil
	}
	l.emit(op)
	return l.lexToken
}

func (l *lexer) emit(typ int) {
	var tok interface{}
	switch typ {
	case NUMBER, IDENT:
		tok = token{
			typ: typ,
			val: l.b.String(),
		}
		l.b.Reset()
	default:
		tok = typ
	}
	select {
	case l.token <- tok:
	case <-l.cancel:
		// bailout
		panic(nil)
	}
}

func (l *lexer) read() (rune, error) {
	r, _, err := l.r.ReadRune()
	return r, err
}

func (l *lexer) unread() {
	l.r.UnreadRune()
}

func (l *lexer) Error(s string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	switch {
	case strings.HasPrefix(s, "syntax error: "):
		s = s[14:]
		if l.err != nil && s == "unexpected EOF" {
			return // lexing was interrupted
		}
	case strings.HasPrefix(s, "runtime error: "):
		s = s[15:]
	}
	l.err = ArithExprError(s)

	select {
	case <-l.cancel:
	default:
		close(l.cancel)
	}
}

type action func() action

type token struct {
	typ int
	val string
}

// ArithExprError represents an arithmetic expression error.
type ArithExprError string

func (e ArithExprError) Error() string {
	return string(e)
}
