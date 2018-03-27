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
	elt      *element
	else_    []ast.ElsePart
	cmds     []ast.Command
	redir    *ast.Redir
	redirs   []*ast.Redir
	word     ast.Word
	words    []ast.Word
}

const AND = 57346
const OR = 57347
const CLOBBER = 57348
const APPEND = 57349
const DUPIN = 57350
const DUPOUT = 57351
const RDWR = 57352
const IO_NUMBER = 57353
const WORD = 57354
const NAME = 57355
const ASSIGNMENT_WORD = 57356
const Bang = 57357
const Lbrace = 57358
const Rbrace = 57359
const For = 57360
const In = 57361
const If = 57362
const Elif = 57363
const Then = 57364
const Else = 57365
const Fi = 57366
const While = 57367
const Until = 57368
const Do = 57369
const Done = 57370

var yyToknames = [...]string{
	"$end",
	"error",
	"$unk",
	"AND",
	"OR",
	"'|'",
	"'('",
	"')'",
	"'&'",
	"';'",
	"'<'",
	"'>'",
	"CLOBBER",
	"APPEND",
	"DUPIN",
	"DUPOUT",
	"RDWR",
	"IO_NUMBER",
	"WORD",
	"NAME",
	"ASSIGNMENT_WORD",
	"Bang",
	"Lbrace",
	"Rbrace",
	"For",
	"In",
	"If",
	"Elif",
	"Then",
	"Else",
	"Fi",
	"While",
	"Until",
	"Do",
	"Done",
	"'\\n'",
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
		case "CLOBBER":
			s = "'>|'"
		case "APPEND":
			s = "'>>'"
		case "DUPIN":
			s = "'<&'"
		case "DUPOUT":
			s = "'>&'"
		case "RDWR":
			s = "'<>'"
		case "Bang":
			s = "'!'"
		case "Lbrace":
			s = "'{'"
		case "Rbrace":
			s = "'}'"
		case "For":
			s = "'for'"
		case "In":
			s = "'in'"
		case "If":
			s = "'if'"
		case "Elif":
			s = "'elif'"
		case "Then":
			s = "'then'"
		case "Else":
			s = "'else'"
		case "Fi":
			s = "'fi'"
		case "While":
			s = "'while'"
		case "Until":
			s = "'until'"
		case "Do":
			s = "'do'"
		case "Done":
			s = "'done'"
		}
		yyToknames[i] = s
	}
}

type element struct {
	redirs  []*ast.Redir
	assigns []*ast.Assign
	args    []ast.Word
}

func extract(cmd *ast.List) ast.Command {
	switch {
	case len(cmd.List) != 0 || !cmd.SepPos.IsZero():
		return cmd
	case !cmd.Pipeline.Bang.IsZero() || len(cmd.Pipeline.List) != 0:
		return cmd.Pipeline
	}
	return cmd.Pipeline.Cmd
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
	-1, 77,
	34, 67,
	-2, 70,
}

const yyPrivate = 57344

const yyLast = 189

var yyAct = [...]int{

	50, 100, 52, 74, 17, 51, 35, 53, 71, 6,
	2, 47, 118, 43, 45, 48, 115, 105, 38, 39,
	104, 54, 3, 56, 57, 58, 103, 93, 19, 113,
	106, 85, 28, 29, 30, 31, 32, 33, 34, 26,
	10, 76, 18, 5, 20, 53, 21, 64, 22, 48,
	63, 80, 66, 23, 24, 109, 65, 76, 77, 61,
	62, 75, 70, 79, 114, 73, 108, 53, 76, 101,
	66, 102, 83, 78, 84, 86, 82, 98, 25, 88,
	89, 90, 87, 53, 72, 60, 94, 55, 92, 97,
	95, 68, 91, 101, 53, 102, 99, 36, 37, 97,
	107, 40, 110, 111, 4, 59, 1, 112, 81, 27,
	41, 42, 36, 37, 116, 117, 19, 38, 39, 119,
	28, 29, 30, 31, 32, 33, 34, 26, 10, 69,
	18, 96, 20, 16, 21, 15, 22, 14, 13, 12,
	11, 23, 24, 28, 29, 30, 31, 32, 33, 34,
	26, 44, 8, 46, 28, 29, 30, 31, 32, 33,
	34, 26, 67, 28, 29, 30, 31, 32, 33, 34,
	26, 49, 28, 29, 30, 31, 32, 33, 34, 26,
	28, 29, 30, 31, 32, 33, 34, 9, 7,
}
var yyPact = [...]int{

	21, -1000, 108, -1000, 95, 109, -1000, -1000, 161, 132,
	152, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -29,
	-29, 67, -29, -29, -29, -1000, 169, 66, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, 21, 21, -1000, -1000,
	109, 95, 161, -1000, 152, -1000, -1000, 143, -1000, -1000,
	83, 21, -28, -1000, 60, 31, 44, 29, 17, -1000,
	-1000, -1000, -1000, -1000, -1000, 143, -1000, -1000, -1000, 9,
	93, -1000, -1000, -29, -3, 49, -29, -28, -29, -29,
	-29, 21, -29, -28, -8, -29, 58, -1000, 65, -9,
	-15, 93, -1000, -1000, -18, -4, 47, -28, -1000, -1000,
	24, -29, -29, -1000, -1000, -1000, -29, -5, -1000, -1000,
	35, -1000, -19, -29, -29, -1000, -23, 41, -1000, -1000,
}
var yyPgo = [...]int{

	0, 10, 22, 104, 9, 188, 187, 11, 152, 140,
	139, 138, 137, 135, 133, 131, 1, 0, 129, 4,
	78, 111, 109, 108, 3, 6, 106, 5, 2,
}
var yyR1 = [...]int{

	0, 26, 26, 26, 1, 1, 1, 2, 2, 3,
	3, 4, 4, 4, 5, 5, 5, 5, 5, 6,
	6, 6, 6, 7, 7, 7, 7, 8, 8, 8,
	8, 8, 8, 9, 10, 11, 11, 11, 11, 15,
	15, 12, 12, 16, 16, 16, 13, 14, 17, 17,
	18, 18, 21, 21, 19, 19, 20, 22, 22, 22,
	22, 22, 22, 22, 23, 23, 24, 24, 25, 25,
	27, 27, 28, 28,
}
var yyR2 = [...]int{

	0, 2, 1, 0, 1, 3, 3, 1, 2, 1,
	3, 1, 1, 2, 3, 2, 1, 2, 1, 1,
	2, 1, 2, 1, 2, 1, 2, 1, 1, 1,
	1, 1, 1, 3, 3, 5, 6, 8, 9, 1,
	2, 5, 6, 4, 5, 2, 5, 5, 3, 2,
	1, 3, 1, 2, 1, 2, 2, 1, 1, 1,
	1, 1, 1, 1, 2, 1, 2, 1, 1, 1,
	1, 0, 1, 2,
}
var yyChk = [...]int{

	-1000, -26, -1, -2, -3, 22, -4, -5, -8, -6,
	19, -9, -10, -11, -12, -13, -14, -19, 21, 7,
	23, 25, 27, 32, 33, -20, 18, -22, 11, 12,
	13, 14, 15, 16, 17, -25, 4, 5, 9, 10,
	6, -3, -21, -19, 19, -19, 21, -7, -19, 19,
	-17, -27, -28, 36, -17, 20, -17, -17, -17, -20,
	19, -2, -2, -4, -19, -7, -19, 19, 8, -18,
	-1, 36, 24, 34, -24, -27, 10, -28, 29, 34,
	34, -23, -25, -28, -17, 34, 26, -27, -17, -17,
	-17, -1, -27, 35, -17, -24, -15, -28, 19, 31,
	-16, 28, 30, 35, 35, 35, 34, -24, 19, 31,
	-17, -17, -17, 34, 29, 35, -17, -17, 35, -16,
}
var yyDef = [...]int{

	3, -2, 2, 4, 7, 0, 9, 11, 12, 16,
	18, 27, 28, 29, 30, 31, 32, 19, 21, 71,
	71, 0, 71, 71, 71, 54, 0, 0, 57, 58,
	59, 60, 61, 62, 63, 1, 0, 0, 68, 69,
	0, 8, 13, 52, 15, 20, 22, 17, 23, 25,
	0, 0, 70, 72, 0, 71, 0, 0, 0, 55,
	56, 5, 6, 10, 53, 14, 24, 26, 33, 49,
	50, 73, 34, 71, 0, 0, 71, -2, 71, 71,
	71, 48, 71, 65, 0, 71, 0, 66, 0, 0,
	0, 51, 64, 35, 0, 0, 0, 67, 39, 41,
	0, 71, 71, 46, 47, 36, 71, 0, 40, 42,
	0, 45, 0, 71, 71, 37, 0, 43, 38, 44,
}
var yyTok1 = [...]int{

	1, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	36, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 9, 3,
	7, 8, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 10,
	11, 3, 12, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 6,
}
var yyTok2 = [...]int{

	2, 3, 4, 5, 13, 14, 15, 16, 17, 18,
	19, 20, 21, 22, 23, 24, 25, 26, 27, 28,
	29, 30, 31, 32, 33, 34, 35,
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
			yylex.(*lexer).cmd = extract(yyDollar[1].list)
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
			yyVAL.cmd = &ast.Cmd{
				Expr: &ast.SimpleCmd{
					Assigns: yyDollar[1].elt.assigns,
					Args:    yyDollar[1].elt.args,
				},
				Redirs: yyDollar[1].elt.redirs,
			}
		}
	case 12:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.cmd = &ast.Cmd{Expr: yyDollar[1].expr}
		}
	case 13:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.cmd = &ast.Cmd{
				Expr:   yyDollar[1].expr,
				Redirs: yyDollar[2].redirs,
			}
		}
	case 14:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.elt = &element{
				redirs:  append(yyDollar[1].elt.redirs, yyDollar[3].elt.redirs...),
				assigns: yyDollar[1].elt.assigns,
			}
			yyVAL.elt.args = append(yyVAL.elt.args, yyDollar[2].word)
			yyVAL.elt.args = append(yyVAL.elt.args, yyDollar[3].elt.args...)
		}
	case 15:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.elt = yyDollar[1].elt
			yyVAL.elt.args = append(yyVAL.elt.args, yyDollar[2].word)
		}
	case 17:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.elt = &element{redirs: yyDollar[2].elt.redirs}
			yyVAL.elt.args = append(yyVAL.elt.args, yyDollar[1].word)
			yyVAL.elt.args = append(yyVAL.elt.args, yyDollar[2].elt.args...)
		}
	case 18:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.elt = new(element)
			yyVAL.elt.args = append(yyVAL.elt.args, yyDollar[1].word)
		}
	case 19:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.elt = new(element)
			yyVAL.elt.redirs = append(yyVAL.elt.redirs, yyDollar[1].redir)
		}
	case 20:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.elt.redirs = append(yyVAL.elt.redirs, yyDollar[2].redir)
		}
	case 21:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.elt = new(element)
			yyVAL.elt.assigns = append(yyVAL.elt.assigns, assign(yyDollar[1].word))
		}
	case 22:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.elt.assigns = append(yyVAL.elt.assigns, assign(yyDollar[2].word))
		}
	case 23:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.elt = new(element)
			yyVAL.elt.redirs = append(yyVAL.elt.redirs, yyDollar[1].redir)
		}
	case 24:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.elt.redirs = append(yyVAL.elt.redirs, yyDollar[2].redir)
		}
	case 25:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.elt = new(element)
			yyVAL.elt.args = append(yyVAL.elt.args, yyDollar[1].word)
		}
	case 26:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.elt.args = append(yyVAL.elt.args, yyDollar[2].word)
		}
	case 33:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.expr = &ast.Subshell{
				Lparen: yyDollar[1].token.pos,
				List:   yyDollar[2].cmds,
				Rparen: yyDollar[3].token.pos,
			}
		}
	case 34:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.expr = &ast.Group{
				Lbrace: yyDollar[1].token.pos,
				List:   yyDollar[2].cmds,
				Rbrace: yyDollar[3].token.pos,
			}
		}
	case 35:
		yyDollar = yyS[yypt-5 : yypt+1]
		{
			yyVAL.expr = &ast.ForClause{
				For:  yyDollar[1].token.pos,
				Name: yyDollar[2].word[0].(*ast.Lit),
				Do:   yyDollar[3].token.pos,
				List: yyDollar[4].cmds,
				Done: yyDollar[5].token.pos,
			}
		}
	case 36:
		yyDollar = yyS[yypt-6 : yypt+1]
		{
			yyVAL.expr = &ast.ForClause{
				For:       yyDollar[1].token.pos,
				Name:      yyDollar[2].word[0].(*ast.Lit),
				Semicolon: yyDollar[3].token.pos,
				Do:        yyDollar[4].token.pos,
				List:      yyDollar[5].cmds,
				Done:      yyDollar[6].token.pos,
			}
		}
	case 37:
		yyDollar = yyS[yypt-8 : yypt+1]
		{
			yyVAL.expr = &ast.ForClause{
				For:       yyDollar[1].token.pos,
				Name:      yyDollar[2].word[0].(*ast.Lit),
				In:        yyDollar[4].token.pos,
				Semicolon: yyDollar[5].token.pos,
				Do:        yyDollar[6].token.pos,
				List:      yyDollar[7].cmds,
				Done:      yyDollar[8].token.pos,
			}
		}
	case 38:
		yyDollar = yyS[yypt-9 : yypt+1]
		{
			yyVAL.expr = &ast.ForClause{
				For:       yyDollar[1].token.pos,
				Name:      yyDollar[2].word[0].(*ast.Lit),
				In:        yyDollar[4].token.pos,
				Items:     yyDollar[5].words,
				Semicolon: yyDollar[6].token.pos,
				Do:        yyDollar[7].token.pos,
				List:      yyDollar[8].cmds,
				Done:      yyDollar[9].token.pos,
			}
		}
	case 39:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.words = append(yyVAL.words, yyDollar[1].word)
		}
	case 40:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.words = append(yyVAL.words, yyDollar[2].word)
		}
	case 41:
		yyDollar = yyS[yypt-5 : yypt+1]
		{
			yyVAL.expr = &ast.IfClause{
				If:   yyDollar[1].token.pos,
				Cond: yyDollar[2].cmds,
				Then: yyDollar[3].token.pos,
				List: yyDollar[4].cmds,
				Fi:   yyDollar[5].token.pos,
			}
		}
	case 42:
		yyDollar = yyS[yypt-6 : yypt+1]
		{
			yyVAL.expr = &ast.IfClause{
				If:   yyDollar[1].token.pos,
				Cond: yyDollar[2].cmds,
				Then: yyDollar[3].token.pos,
				List: yyDollar[4].cmds,
				Else: yyDollar[5].else_,
				Fi:   yyDollar[6].token.pos,
			}
		}
	case 43:
		yyDollar = yyS[yypt-4 : yypt+1]
		{
			yyVAL.else_ = append(yyVAL.else_, &ast.ElifClause{
				Elif: yyDollar[1].token.pos,
				Cond: yyDollar[2].cmds,
				Then: yyDollar[3].token.pos,
				List: yyDollar[4].cmds,
			})
		}
	case 44:
		yyDollar = yyS[yypt-5 : yypt+1]
		{
			yyVAL.else_ = append(yyVAL.else_, &ast.ElifClause{
				Elif: yyDollar[1].token.pos,
				Cond: yyDollar[2].cmds,
				Then: yyDollar[3].token.pos,
				List: yyDollar[4].cmds,
			})
			yyVAL.else_ = append(yyVAL.else_, yyDollar[5].else_...)
		}
	case 45:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.else_ = append(yyVAL.else_, &ast.ElseClause{
				Else: yyDollar[1].token.pos,
				List: yyDollar[2].cmds,
			})
		}
	case 46:
		yyDollar = yyS[yypt-5 : yypt+1]
		{
			yyVAL.expr = &ast.WhileClause{
				While: yyDollar[1].token.pos,
				Cond:  yyDollar[2].cmds,
				Do:    yyDollar[3].token.pos,
				List:  yyDollar[4].cmds,
				Done:  yyDollar[5].token.pos,
			}
		}
	case 47:
		yyDollar = yyS[yypt-5 : yypt+1]
		{
			yyVAL.expr = &ast.UntilClause{
				Until: yyDollar[1].token.pos,
				Cond:  yyDollar[2].cmds,
				Do:    yyDollar[3].token.pos,
				List:  yyDollar[4].cmds,
				Done:  yyDollar[5].token.pos,
			}
		}
	case 48:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			cmd := yyDollar[2].cmds[len(yyDollar[2].cmds)-1].(*ast.List)
			if yyDollar[3].token.typ != '\n' {
				cmd.SepPos = yyDollar[3].token.pos
				cmd.Sep = yyDollar[3].token.val
			} else {
				yyDollar[2].cmds[len(yyDollar[2].cmds)-1] = extract(cmd)
			}
			yyVAL.cmds = yyDollar[2].cmds
		}
	case 49:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyDollar[2].cmds[len(yyDollar[2].cmds)-1] = extract(yyDollar[2].cmds[len(yyDollar[2].cmds)-1].(*ast.List))
			yyVAL.cmds = yyDollar[2].cmds
		}
	case 50:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.cmds = []ast.Command{yyDollar[1].list}
		}
	case 51:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			cmd := yyVAL.cmds[len(yyVAL.cmds)-1].(*ast.List)
			if yyDollar[2].token.typ != '\n' {
				cmd.SepPos = yyDollar[2].token.pos
				cmd.Sep = yyDollar[2].token.val
			} else {
				yyVAL.cmds[len(yyVAL.cmds)-1] = extract(cmd)
			}
			yyVAL.cmds = append(yyVAL.cmds, yyDollar[3].list)
		}
	case 52:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.redirs = append(yyVAL.redirs, yyDollar[1].redir)
		}
	case 53:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.redirs = append(yyVAL.redirs, yyDollar[2].redir)
		}
	case 55:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.redir = yyDollar[2].redir
			yyVAL.redir.N = yyDollar[1].word[0].(*ast.Lit)
		}
	case 56:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.redir = &ast.Redir{
				OpPos: yyDollar[1].token.pos,
				Op:    yyDollar[1].token.val,
				Word:  yyDollar[2].word,
			}
		}
	case 65:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
		}
	case 67:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.token.pos = ast.Pos{}
		}
	}
	goto yystack /* stack new state and value */
}
