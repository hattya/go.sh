//
// go.sh/parser :: parser.go
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

package parser

import __yyfmt__ "fmt"

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"

	"github.com/hattya/go.sh/ast"
)

type yySymType struct {
	yys      int
	list     *ast.List
	pipeline *ast.Pipeline
	cmd      *ast.Cmd
	expr     ast.CmdExpr
	token    token
	assigns  []*ast.Assign
	word     ast.Word
	words    []ast.Word
}

const AND = 57346
const OR = 57347
const WORD = 57348
const ASSIGNMENT_WORD = 57349
const Bang = 57350

var yyToknames = [...]string{
	"$end",
	"error",
	"$unk",
	"AND",
	"OR",
	"'|'",
	"'&'",
	"';'",
	"WORD",
	"ASSIGNMENT_WORD",
	"Bang",
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
		case "AND":
			s = "'&&'"
		case "OR":
			s = "'||'"
		case "Bang":
			s = "'!'"
		}
		yyToknames[i] = s
	}
}

func assign(w ast.Word) *ast.Assign {
	k := w[0].(*ast.Lit)
	i := strings.IndexRune(k.Value, '=')
	v := w[1:]
	if i < len(k.Value)-1 {
		v = append(ast.Word{
			&ast.Lit{
				ValuePos: ast.NewPos(k.ValuePos.Line(), k.ValuePos.Col()+i+1),
				Value:    k.Value[i+1:],
			},
		}, v...)
		k.Value = k.Value[:i]
	}
	return &ast.Assign{
		Symbol: k,
		Op:     "=",
		Value:  v,
	}
}

// ParseCommand parses src and returns a command.
func ParseCommand(name string, src interface{}) (ast.Command, []*ast.Comment, error) {
	r, err := open(src)
	if err != nil {
		return nil, nil, err
	}

	l := newLexer(name, r)
	yyParse(l)
	return l.cmd, l.comments, l.err
}

func open(src interface{}) (r io.RuneScanner, err error) {
	switch src := src.(type) {
	case []byte:
		r = bytes.NewReader(src)
	case string:
		r = strings.NewReader(src)
	case io.RuneScanner:
		r = src
	case io.Reader:
		r = bufio.NewReader(src)
	default:
		err = errors.New("invalid source")
	}
	return
}

var yyExca = [...]int{
	-1, 1,
	1, -1,
	-2, 0,
}

const yyPrivate = 57344

const yyLast = 28

var yyAct = [...]int{

	20, 6, 3, 9, 10, 5, 9, 10, 18, 19,
	12, 13, 26, 14, 15, 22, 23, 21, 24, 25,
	16, 4, 1, 11, 8, 7, 2, 17,
}
var yyPact = [...]int{

	-6, -1000, 6, -1000, 14, -3, -1000, -1000, -1, 8,
	-1000, -1000, -6, -6, -1000, -1000, -3, 14, 8, -1000,
	3, -1000, -1000, -1000, -1000, 3, -1000,
}
var yyPgo = [...]int{

	0, 26, 2, 21, 1, 25, 24, 0, 23, 22,
}
var yyR1 = [...]int{

	0, 9, 9, 9, 1, 1, 1, 2, 2, 3,
	3, 4, 5, 5, 5, 5, 5, 6, 6, 7,
	7, 8, 8,
}
var yyR2 = [...]int{

	0, 2, 1, 0, 1, 3, 3, 1, 2, 1,
	3, 1, 3, 2, 1, 2, 1, 1, 2, 1,
	2, 1, 1,
}
var yyChk = [...]int{

	-1000, -9, -1, -2, -3, 11, -4, -5, -6, 9,
	10, -8, 4, 5, 7, 8, 6, -3, 9, 10,
	-7, 9, -2, -2, -4, -7, 9,
}
var yyDef = [...]int{

	3, -2, 2, 4, 7, 0, 9, 11, 14, 16,
	17, 1, 0, 0, 21, 22, 0, 8, 13, 18,
	15, 19, 5, 6, 10, 12, 20,
}
var yyTok1 = [...]int{

	1, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 7, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 8,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 6,
}
var yyTok2 = [...]int{

	2, 3, 4, 5, 9, 10, 11,
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
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyDollar[1].list.SepPos = yyDollar[2].token.pos
			yyDollar[1].list.Sep = yyDollar[2].token.val
			yylex.(*lexer).cmd = yyDollar[1].list
		}
	case 2:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			l := yylex.(*lexer)
			switch {
			case len(yyDollar[1].list.List) != 0:
				l.cmd = yyDollar[1].list
			case !yyDollar[1].list.Pipeline.Bang.IsZero() || len(yyDollar[1].list.Pipeline.List) != 0:
				l.cmd = yyDollar[1].list.Pipeline
			default:
				l.cmd = yyDollar[1].list.Pipeline.Cmd
			}
		}
	case 4:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.list = &ast.List{Pipeline: yyDollar[1].pipeline}
		}
	case 5:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.list.List = append(yyVAL.list.List, &ast.AndOr{
				OpPos:    yyDollar[2].token.pos,
				Op:       yyDollar[2].token.val,
				Pipeline: yyDollar[3].pipeline,
			})
		}
	case 6:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.list.List = append(yyVAL.list.List, &ast.AndOr{
				OpPos:    yyDollar[2].token.pos,
				Op:       yyDollar[2].token.val,
				Pipeline: yyDollar[3].pipeline,
			})
		}
	case 8:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.pipeline = yyDollar[2].pipeline
			yyVAL.pipeline.Bang = yyDollar[1].token.pos
		}
	case 9:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.pipeline = &ast.Pipeline{Cmd: yyDollar[1].cmd}
		}
	case 10:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.pipeline.List = append(yyVAL.pipeline.List, &ast.Pipe{
				OpPos: yyDollar[2].token.pos,
				Op:    yyDollar[2].token.val,
				Cmd:   yyDollar[3].cmd,
			})
		}
	case 11:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.cmd = &ast.Cmd{Expr: yyDollar[1].expr}
		}
	case 12:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.expr = &ast.SimpleCmd{
				Assigns: yyDollar[1].assigns,
				Args:    append([]ast.Word{yyDollar[2].word}, yyDollar[3].words...),
			}
		}
	case 13:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.expr = &ast.SimpleCmd{
				Assigns: yyDollar[1].assigns,
				Args:    []ast.Word{yyDollar[2].word},
			}
		}
	case 14:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.expr = &ast.SimpleCmd{Assigns: yyDollar[1].assigns}
		}
	case 15:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.expr = &ast.SimpleCmd{Args: append([]ast.Word{yyDollar[1].word}, yyDollar[2].words...)}
		}
	case 16:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.expr = &ast.SimpleCmd{Args: []ast.Word{yyDollar[1].word}}
		}
	case 17:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.assigns = append(yyVAL.assigns, assign(yyDollar[1].word))
		}
	case 18:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.assigns = append(yyVAL.assigns, assign(yyDollar[2].word))
		}
	case 19:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.words = append(yyVAL.words, yyDollar[1].word)
		}
	case 20:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.words = append(yyVAL.words, yyDollar[2].word)
		}
	}
	goto yystack /* stack new state and value */
}
