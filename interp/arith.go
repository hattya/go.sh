// Code generated by goyacc -l -o arith.go arith.go.y. DO NOT EDIT.
//
// go.sh/interp :: arith.go
//
//   Copyright (c) 2021 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
//

package interp

import __yyfmt__ "fmt"

import (
	"fmt"
	"strconv"
	"strings"
)

type yySymType struct {
	yys  int
	op   string
	expr expr
}

const NUMBER = 57346
const IDENT = 57347
const INC = 57348
const DEC = 57349
const LSH = 57350
const RSH = 57351
const LE = 57352
const GE = 57353
const EQ = 57354
const NE = 57355

var yyToknames = [...]string{
	"$end",
	"error",
	"$unk",
	"NUMBER",
	"IDENT",
	"'('",
	"')'",
	"INC",
	"DEC",
	"'+'",
	"'-'",
	"'~'",
	"'!'",
	"'*'",
	"'/'",
	"'%'",
	"LSH",
	"RSH",
	"'<'",
	"'>'",
	"LE",
	"GE",
	"EQ",
	"NE",
}

var yyStatenames = [...]string{}

const yyEofCode = 1
const yyErrCode = 2
const yyInitialStackSize = 16

func init() {
	yyErrorVerbose = true

	for i, s := range yyToknames {
		switch s {
		case "$end":
			s = "EOF"
		case "INC":
			s = "'++'"
		case "DEC":
			s = "'--'"
		case "LSH":
			s = "'<<'"
		case "RSH":
			s = "'>>'"
		case "LE":
			s = "'<='"
		case "GE":
			s = "'>='"
		case "EQ":
			s = "'=='"
		case "NE":
			s = "'!='"
		}
		yyToknames[i] = s
	}
}

type expr struct {
	n int
	s string
}

func errLValue(op string) string {
	return fmt.Sprintf("'%v' requires lvalue", op)
}

func expand(yylex yyLexer, x expr) (int, bool) {
	if x.s == "" {
		return x.n, true
	} else if v, set := yylex.(*lexer).env.Get(x.s); !set || v.Value == "" {
		return 0, true
	} else if n, err := strconv.ParseInt(v.Value, 0, 0); err != nil {
		yylex.Error(fmt.Sprintf("invalid number %q", v.Value))
		return 0, false
	} else {
		return int(n), true
	}
}

func calculate(yylex yyLexer, l expr, op string, r expr) (x expr) {
	if l, ok := expand(yylex, l); ok {
		if r, ok := expand(yylex, r); ok {
			switch op {
			case "*":
				x.n = l * r
			case "/":
				x.n = l / r
			case "%":
				x.n = l % r
			case "+":
				x.n = l + r
			case "-":
				x.n = l - r
			case "<<":
				x.n = l << r
			case ">>":
				x.n = l >> r
			}
		}
	}
	return
}

func compare(yylex yyLexer, l expr, op string, r expr) (x expr) {
	if l, ok := expand(yylex, l); ok {
		if r, ok := expand(yylex, r); ok {
			var b bool
			switch op {
			case "<":
				b = l < r
			case ">":
				b = l > r
			case "<=":
				b = l <= r
			case ">=":
				b = l >= r
			case "==":
				b = l == r
			case "!=":
				b = l != r
			}
			if b {
				x.n = 1
			}
		}
	}
	return
}

// Eval evaluates an arithmetic expression.
func (env *ExecEnv) Eval(expr string) (n int, err error) {
	l := newLexer(env, strings.NewReader(expr))
	defer func() {
		if e := recover(); e != nil {
			l.Error(e.(error).Error())
			err = l.err
		}
	}()

	yyParse(l)
	return l.n, l.err
}

var yyExca = [...]int{
	-1, 1,
	1, -1,
	-2, 0,
}

const yyPrivate = 57344

const yyLast = 53

var yyAct = [...]int{
	8, 5, 6, 21, 22, 7, 23, 24, 25, 26,
	53, 36, 37, 38, 1, 4, 27, 28, 31, 32,
	33, 2, 29, 30, 3, 42, 43, 44, 45, 12,
	46, 47, 50, 51, 52, 48, 49, 40, 41, 18,
	19, 20, 39, 10, 11, 14, 15, 16, 17, 34,
	35, 9, 13,
}

var yyPact = [...]int{
	35, -1000, -1000, -20, -13, -1, 12, 4, -1000, 41,
	35, 35, 35, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	35, 35, 35, 35, 35, 35, 35, 35, 35, 35,
	35, 35, 35, 35, -1000, -1000, -1000, -1000, -1000, 3,
	-13, -13, -1, -1, -1, -1, 12, 12, 4, 4,
	-1000, -1000, -1000, -1000,
}

var yyPgo = [...]int{
	0, 52, 51, 0, 29, 5, 2, 1, 15, 24,
	21, 14,
}

var yyR1 = [...]int{
	0, 11, 1, 1, 1, 2, 2, 2, 3, 3,
	3, 3, 4, 4, 4, 4, 5, 5, 5, 5,
	6, 6, 6, 7, 7, 7, 8, 8, 8, 8,
	8, 9, 9, 9, 10,
}

var yyR2 = [...]int{
	0, 1, 1, 1, 3, 1, 2, 2, 1, 2,
	2, 2, 1, 1, 1, 1, 1, 3, 3, 3,
	1, 3, 3, 1, 3, 3, 1, 3, 3, 3,
	3, 1, 3, 3, 1,
}

var yyChk = [...]int{
	-1000, -11, -10, -9, -8, -7, -6, -5, -3, -2,
	8, 9, -4, -1, 10, 11, 12, 13, 4, 5,
	6, 23, 24, 19, 20, 21, 22, 17, 18, 10,
	11, 14, 15, 16, 8, 9, -3, -3, -3, -10,
	-8, -8, -7, -7, -7, -7, -6, -6, -5, -5,
	-3, -3, -3, 7,
}

var yyDef = [...]int{
	0, -2, 1, 34, 31, 26, 23, 20, 16, 8,
	0, 0, 0, 5, 12, 13, 14, 15, 2, 3,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 6, 7, 9, 10, 11, 0,
	32, 33, 27, 28, 29, 30, 24, 25, 21, 22,
	17, 18, 19, 4,
}

var yyTok1 = [...]int{
	1, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 13, 3, 3, 3, 16, 3, 3,
	6, 7, 14, 10, 3, 11, 3, 15, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	19, 3, 20, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 12,
}

var yyTok2 = [...]int{
	2, 3, 4, 5, 8, 9, 17, 18, 21, 22,
	23, 24,
}

var yyTok3 = [...]int{
	0,
}

var yyErrorMessages = [...]struct {
	state int
	token int
	msg   string
}{}

/*	parser for yacc output	*/

var (
	yyDebug        = 0
	yyErrorVerbose = false
)

type yyLexer interface {
	Lex(lval *yySymType) int
	Error(s string)
}

type yyParser interface {
	Parse(yyLexer) int
	Lookahead() int
}

type yyParserImpl struct {
	lval  yySymType
	stack [yyInitialStackSize]yySymType
	char  int
}

func (p *yyParserImpl) Lookahead() int {
	return p.char
}

func yyNewParser() yyParser {
	return &yyParserImpl{}
}

const yyFlag = -1000

func yyTokname(c int) string {
	if c >= 1 && c-1 < len(yyToknames) {
		if yyToknames[c-1] != "" {
			return yyToknames[c-1]
		}
	}
	return __yyfmt__.Sprintf("tok-%v", c)
}

func yyStatname(s int) string {
	if s >= 0 && s < len(yyStatenames) {
		if yyStatenames[s] != "" {
			return yyStatenames[s]
		}
	}
	return __yyfmt__.Sprintf("state-%v", s)
}

func yyErrorMessage(state, lookAhead int) string {
	const TOKSTART = 4

	if !yyErrorVerbose {
		return "syntax error"
	}

	for _, e := range yyErrorMessages {
		if e.state == state && e.token == lookAhead {
			return "syntax error: " + e.msg
		}
	}

	res := "syntax error: unexpected " + yyTokname(lookAhead)

	// To match Bison, suggest at most four expected tokens.
	expected := make([]int, 0, 4)

	// Look for shiftable tokens.
	base := yyPact[state]
	for tok := TOKSTART; tok-1 < len(yyToknames); tok++ {
		if n := base + tok; n >= 0 && n < yyLast && yyChk[yyAct[n]] == tok {
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}
	}

	if yyDef[state] == -2 {
		i := 0
		for yyExca[i] != -1 || yyExca[i+1] != state {
			i += 2
		}

		// Look for tokens that we accept or reduce.
		for i += 2; yyExca[i] >= 0; i += 2 {
			tok := yyExca[i]
			if tok < TOKSTART || yyExca[i+1] == 0 {
				continue
			}
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}

		// If the default action is to accept or reduce, give up.
		if yyExca[i+1] != 0 {
			return res
		}
	}

	for i, tok := range expected {
		if i == 0 {
			res += ", expecting "
		} else {
			res += " or "
		}
		res += yyTokname(tok)
	}
	return res
}

func yylex1(lex yyLexer, lval *yySymType) (char, token int) {
	token = 0
	char = lex.Lex(lval)
	if char <= 0 {
		token = yyTok1[0]
		goto out
	}
	if char < len(yyTok1) {
		token = yyTok1[char]
		goto out
	}
	if char >= yyPrivate {
		if char < yyPrivate+len(yyTok2) {
			token = yyTok2[char-yyPrivate]
			goto out
		}
	}
	for i := 0; i < len(yyTok3); i += 2 {
		token = yyTok3[i+0]
		if token == char {
			token = yyTok3[i+1]
			goto out
		}
	}

out:
	if token == 0 {
		token = yyTok2[1] /* unknown char */
	}
	if yyDebug >= 3 {
		__yyfmt__.Printf("lex %s(%d)\n", yyTokname(token), uint(char))
	}
	return char, token
}

func yyParse(yylex yyLexer) int {
	return yyNewParser().Parse(yylex)
}

func (yyrcvr *yyParserImpl) Parse(yylex yyLexer) int {
	var yyn int
	var yyVAL yySymType
	var yyDollar []yySymType
	_ = yyDollar // silence set and not used
	yyS := yyrcvr.stack[:]

	Nerrs := 0   /* number of errors */
	Errflag := 0 /* error recovery flag */
	yystate := 0
	yyrcvr.char = -1
	yytoken := -1 // yyrcvr.char translated into internal numbering
	defer func() {
		// Make sure we report no lookahead when not parsing.
		yystate = -1
		yyrcvr.char = -1
		yytoken = -1
	}()
	yyp := -1
	goto yystack

ret0:
	return 0

ret1:
	return 1

yystack:
	/* put a state and value onto the stack */
	if yyDebug >= 4 {
		__yyfmt__.Printf("char %v in %v\n", yyTokname(yytoken), yyStatname(yystate))
	}

	yyp++
	if yyp >= len(yyS) {
		nyys := make([]yySymType, len(yyS)*2)
		copy(nyys, yyS)
		yyS = nyys
	}
	yyS[yyp] = yyVAL
	yyS[yyp].yys = yystate

yynewstate:
	yyn = yyPact[yystate]
	if yyn <= yyFlag {
		goto yydefault /* simple state */
	}
	if yyrcvr.char < 0 {
		yyrcvr.char, yytoken = yylex1(yylex, &yyrcvr.lval)
	}
	yyn += yytoken
	if yyn < 0 || yyn >= yyLast {
		goto yydefault
	}
	yyn = yyAct[yyn]
	if yyChk[yyn] == yytoken { /* valid shift */
		yyrcvr.char = -1
		yytoken = -1
		yyVAL = yyrcvr.lval
		yystate = yyn
		if Errflag > 0 {
			Errflag--
		}
		goto yystack
	}

yydefault:
	/* default state action */
	yyn = yyDef[yystate]
	if yyn == -2 {
		if yyrcvr.char < 0 {
			yyrcvr.char, yytoken = yylex1(yylex, &yyrcvr.lval)
		}

		/* look through exception table */
		xi := 0
		for {
			if yyExca[xi+0] == -1 && yyExca[xi+1] == yystate {
				break
			}
			xi += 2
		}
		for xi += 2; ; xi += 2 {
			yyn = yyExca[xi+0]
			if yyn < 0 || yyn == yytoken {
				break
			}
		}
		yyn = yyExca[xi+1]
		if yyn < 0 {
			goto ret0
		}
	}
	if yyn == 0 {
		/* error ... attempt to resume parsing */
		switch Errflag {
		case 0: /* brand new error */
			yylex.Error(yyErrorMessage(yystate, yytoken))
			Nerrs++
			if yyDebug >= 1 {
				__yyfmt__.Printf("%s", yyStatname(yystate))
				__yyfmt__.Printf(" saw %s\n", yyTokname(yytoken))
			}
			fallthrough

		case 1, 2: /* incompletely recovered error ... try again */
			Errflag = 3

			/* find a state where "error" is a legal shift action */
			for yyp >= 0 {
				yyn = yyPact[yyS[yyp].yys] + yyErrCode
				if yyn >= 0 && yyn < yyLast {
					yystate = yyAct[yyn] /* simulate a shift of "error" */
					if yyChk[yystate] == yyErrCode {
						goto yystack
					}
				}

				/* the current p has no shift on "error", pop stack */
				if yyDebug >= 2 {
					__yyfmt__.Printf("error recovery pops state %d\n", yyS[yyp].yys)
				}
				yyp--
			}
			/* there is no state on the stack with an error shift ... abort */
			goto ret1

		case 3: /* no shift yet; clobber input char */
			if yyDebug >= 2 {
				__yyfmt__.Printf("error recovery discards %s\n", yyTokname(yytoken))
			}
			if yytoken == yyEofCode {
				goto ret1
			}
			yyrcvr.char = -1
			yytoken = -1
			goto yynewstate /* try again in the same state */
		}
	}

	/* reduction by production yyn */
	if yyDebug >= 2 {
		__yyfmt__.Printf("reduce %v in:\n\t%v\n", yyn, yyStatname(yystate))
	}

	yynt := yyn
	yypt := yyp
	_ = yypt // guard against "declared and not used"

	yyp -= yyR2[yyn]
	// yyp is now the index of $0. Perform the default action. Iff the
	// reduced production is ε, $1 is possibly out of range.
	if yyp+1 >= len(yyS) {
		nyys := make([]yySymType, len(yyS)*2)
		copy(nyys, yyS)
		yyS = nyys
	}
	yyVAL = yyS[yyp+1]

	/* consult goto table to find next state */
	yyn = yyR1[yyn]
	yyg := yyPgo[yyn]
	yyj := yyg + yyS[yyp].yys + 1

	if yyj >= yyLast {
		yystate = yyAct[yyg]
	} else {
		yystate = yyAct[yyj]
		if yyChk[yystate] != -yyn {
			yystate = yyAct[yyg]
		}
	}
	// dummy call; replaced with literal code
	switch yynt {

	case 1:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			if n, ok := expand(yylex, yyDollar[1].expr); ok {
				yylex.(*lexer).n = n
			}
		}
	case 2:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.expr.s = ""
			if n, err := strconv.ParseInt(yyDollar[1].expr.s, 0, 0); err != nil {
				yylex.Error(fmt.Sprintf("invalid number %q", yyDollar[1].expr.s))
			} else {
				yyVAL.expr.n = int(n)
			}
		}
	case 4:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 6:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.expr.s = ""
			if yyDollar[1].expr.s == "" {
				yylex.Error(errLValue(yyDollar[2].op))
			} else if n, ok := expand(yylex, yyDollar[1].expr); ok {
				yyVAL.expr.n = n
				yylex.(*lexer).env.Set(yyDollar[1].expr.s, strconv.Itoa(yyVAL.expr.n+1))
			}
		}
	case 7:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.expr.s = ""
			if yyDollar[1].expr.s == "" {
				yylex.Error(errLValue(yyDollar[2].op))
			} else if n, ok := expand(yylex, yyDollar[1].expr); ok {
				yyVAL.expr.n = n
				yylex.(*lexer).env.Set(yyDollar[1].expr.s, strconv.Itoa(yyVAL.expr.n-1))
			}
		}
	case 9:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.expr.s = ""
			if yyDollar[2].expr.s == "" {
				yylex.Error(errLValue(yyDollar[1].op))
			} else if n, ok := expand(yylex, yyDollar[2].expr); ok {
				yyVAL.expr.n = n + 1
				yylex.(*lexer).env.Set(yyDollar[2].expr.s, strconv.Itoa(yyVAL.expr.n))
			}
		}
	case 10:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.expr.s = ""
			if yyDollar[2].expr.s == "" {
				yylex.Error(errLValue(yyDollar[1].op))
			} else if n, ok := expand(yylex, yyDollar[2].expr); ok {
				yyVAL.expr.n = n - 1
				yylex.(*lexer).env.Set(yyDollar[2].expr.s, strconv.Itoa(yyVAL.expr.n))
			}
		}
	case 11:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.expr.s = ""
			if n, ok := expand(yylex, yyDollar[2].expr); ok {
				switch yyDollar[1].op {
				case "+":
					yyVAL.expr.n = +n
				case "-":
					yyVAL.expr.n = -n
				case "~":
					yyVAL.expr.n = ^n
				case "!":
					if n == 0 {
						yyVAL.expr.n = 1
					} else {
						yyVAL.expr.n = 0
					}
				}
			}
		}
	case 17:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.expr = calculate(yylex, yyDollar[1].expr, yyDollar[2].op, yyDollar[3].expr)
		}
	case 18:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.expr = calculate(yylex, yyDollar[1].expr, yyDollar[2].op, yyDollar[3].expr)
		}
	case 19:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.expr = calculate(yylex, yyDollar[1].expr, yyDollar[2].op, yyDollar[3].expr)
		}
	case 21:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.expr = calculate(yylex, yyDollar[1].expr, yyDollar[2].op, yyDollar[3].expr)
		}
	case 22:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.expr = calculate(yylex, yyDollar[1].expr, yyDollar[2].op, yyDollar[3].expr)
		}
	case 24:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.expr = calculate(yylex, yyDollar[1].expr, yyDollar[2].op, yyDollar[3].expr)
		}
	case 25:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.expr = calculate(yylex, yyDollar[1].expr, yyDollar[2].op, yyDollar[3].expr)
		}
	case 27:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.expr = compare(yylex, yyDollar[1].expr, yyDollar[2].op, yyDollar[3].expr)
		}
	case 28:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.expr = compare(yylex, yyDollar[1].expr, yyDollar[2].op, yyDollar[3].expr)
		}
	case 29:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.expr = compare(yylex, yyDollar[1].expr, yyDollar[2].op, yyDollar[3].expr)
		}
	case 30:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.expr = compare(yylex, yyDollar[1].expr, yyDollar[2].op, yyDollar[3].expr)
		}
	case 32:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.expr = compare(yylex, yyDollar[1].expr, yyDollar[2].op, yyDollar[3].expr)
		}
	case 33:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.expr = compare(yylex, yyDollar[1].expr, yyDollar[2].op, yyDollar[3].expr)
		}
	}
	goto yystack /* stack new state and value */
}
