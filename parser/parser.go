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
	item     *ast.CaseItem
	items    []*ast.CaseItem
	else_    []ast.ElsePart
	cmds     []ast.Command
	redir    *ast.Redir
	redirs   []*ast.Redir
	word     ast.Word
	words    []ast.Word
}

const AND = 57346
const OR = 57347
const BREAK = 57348
const CLOBBER = 57349
const APPEND = 57350
const DUPIN = 57351
const DUPOUT = 57352
const RDWR = 57353
const IO_NUMBER = 57354
const WORD = 57355
const NAME = 57356
const ASSIGNMENT_WORD = 57357
const Bang = 57358
const Lbrace = 57359
const Rbrace = 57360
const For = 57361
const Case = 57362
const Esac = 57363
const In = 57364
const If = 57365
const Elif = 57366
const Then = 57367
const Else = 57368
const Fi = 57369
const While = 57370
const Until = 57371
const Do = 57372
const Done = 57373

var yyToknames = [...]string{
	"$end",
	"error",
	"$unk",
	"AND",
	"OR",
	"'|'",
	"'('",
	"')'",
	"BREAK",
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
	"Case",
	"Esac",
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
		case "BREAK":
			s = "';;'"
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
		case "Case":
			s = "'case'"
		case "Esac":
			s = "'esac'"
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
	-1, 84,
	37, 89,
	-2, 92,
}

const yyPrivate = 57344

const yyLast = 297

var yyAct = [...]int{

	56, 55, 128, 127, 113, 20, 126, 46, 39, 57,
	81, 77, 2, 8, 47, 51, 49, 52, 6, 42,
	43, 58, 78, 151, 3, 59, 144, 22, 62, 63,
	64, 119, 32, 33, 34, 35, 36, 37, 38, 30,
	11, 19, 21, 5, 23, 83, 24, 25, 58, 117,
	26, 83, 70, 116, 52, 27, 28, 72, 135, 83,
	122, 82, 85, 69, 71, 67, 68, 105, 110, 120,
	84, 80, 94, 58, 88, 89, 87, 72, 114, 58,
	115, 112, 93, 131, 96, 91, 92, 58, 98, 99,
	100, 114, 104, 115, 143, 86, 106, 97, 111, 29,
	129, 95, 103, 102, 139, 109, 107, 79, 47, 60,
	118, 148, 129, 130, 130, 66, 132, 133, 109, 121,
	61, 136, 134, 160, 70, 130, 153, 141, 138, 149,
	65, 137, 142, 123, 141, 75, 140, 145, 74, 54,
	44, 146, 147, 40, 41, 150, 4, 1, 90, 31,
	154, 155, 45, 157, 158, 156, 76, 22, 125, 159,
	161, 162, 32, 33, 34, 35, 36, 37, 38, 30,
	11, 19, 21, 5, 23, 124, 24, 25, 108, 101,
	26, 18, 22, 17, 152, 27, 28, 32, 33, 34,
	35, 36, 37, 38, 30, 11, 19, 21, 5, 23,
	16, 24, 25, 15, 14, 26, 13, 22, 12, 10,
	27, 28, 32, 33, 34, 35, 36, 37, 38, 30,
	11, 19, 21, 22, 23, 7, 24, 25, 9, 0,
	26, 40, 41, 0, 0, 27, 28, 42, 43, 0,
	23, 0, 24, 25, 0, 0, 26, 0, 0, 0,
	0, 27, 28, 32, 33, 34, 35, 36, 37, 38,
	30, 48, 0, 50, 32, 33, 34, 35, 36, 37,
	38, 30, 73, 32, 33, 34, 35, 36, 37, 38,
	30, 53, 32, 33, 34, 35, 36, 37, 38, 30,
	32, 33, 34, 35, 36, 37, 38,
}
var yyPact = [...]int{

	20, -1000, 227, -1000, 134, 200, -1000, -1000, 270, -1000,
	241, 261, -1000, -1000, -1000, -1000, -1000, -1000, -1000, 132,
	-1000, -1000, -18, -18, 88, 100, -18, -18, -18, -1000,
	278, 95, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	20, 20, -1000, -1000, 200, 134, 270, -1000, 261, -1000,
	-1000, 252, -1000, -1000, 130, 127, 20, -17, -1000, 82,
	34, -18, 63, 39, 37, -1000, -1000, -1000, -1000, -1000,
	-1000, 252, -1000, -1000, -18, -1000, 9, 139, -1000, -1000,
	-18, 35, 72, -18, -17, 68, -18, -18, -18, 216,
	20, -18, -17, 29, -18, 48, -1000, -18, 47, 15,
	11, -1000, 270, 139, -1000, -1000, -7, 32, 40, -17,
	-1000, 105, -1000, 49, -18, -18, -1000, -1000, 270, -1000,
	-18, 21, -1000, -1000, 93, 76, -1000, -1000, 128, 94,
	-1000, -1000, 62, -1000, -12, -18, -1000, -1000, -1000, -1000,
	-18, 91, 121, -18, -1000, -15, 175, 117, -1000, -18,
	60, -1000, -18, -18, 150, 114, -1000, -1000, -1000, -18,
	-18, -1000, -1000,
}
var yyPgo = [...]int{

	0, 11, 24, 146, 18, 228, 225, 209, 15, 13,
	208, 206, 204, 203, 200, 183, 181, 179, 178, 2,
	6, 3, 175, 158, 4, 1, 156, 5, 99, 7,
	149, 148, 10, 8, 147, 0, 9,
}
var yyR1 = [...]int{

	0, 34, 34, 34, 1, 1, 1, 2, 2, 3,
	3, 4, 4, 4, 4, 6, 6, 6, 6, 6,
	7, 7, 7, 7, 8, 8, 8, 8, 9, 9,
	9, 9, 9, 9, 9, 10, 11, 12, 12, 12,
	12, 18, 18, 13, 13, 13, 22, 22, 20, 20,
	20, 20, 23, 23, 21, 21, 21, 21, 19, 19,
	14, 14, 24, 24, 24, 15, 16, 5, 17, 17,
	25, 25, 26, 26, 29, 29, 27, 27, 28, 30,
	30, 30, 30, 30, 30, 30, 31, 31, 32, 32,
	33, 33, 35, 35, 36, 36,
}
var yyR2 = [...]int{

	0, 2, 1, 0, 1, 3, 3, 1, 2, 1,
	3, 1, 1, 2, 1, 3, 2, 1, 2, 1,
	1, 2, 1, 2, 1, 2, 1, 2, 1, 1,
	1, 1, 1, 1, 1, 3, 3, 5, 6, 8,
	9, 1, 2, 6, 7, 7, 1, 2, 5, 5,
	6, 6, 1, 2, 3, 3, 4, 4, 1, 3,
	5, 6, 4, 5, 2, 5, 5, 5, 1, 2,
	3, 2, 1, 3, 1, 2, 1, 2, 2, 1,
	1, 1, 1, 1, 1, 1, 2, 1, 2, 1,
	1, 1, 1, 0, 1, 2,
}
var yyChk = [...]int{

	-1000, -34, -1, -2, -3, 23, -4, -6, -9, -5,
	-7, 20, -10, -11, -12, -13, -14, -15, -16, 21,
	-27, 22, 7, 24, 26, 27, 30, 35, 36, -28,
	19, -30, 12, 13, 14, 15, 16, 17, 18, -33,
	4, 5, 10, 11, 6, -3, -29, -27, 20, -27,
	22, -8, -27, 20, 7, -25, -35, -36, 39, -25,
	21, 20, -25, -25, -25, -28, 20, -2, -2, -4,
	-27, -8, -27, 20, 8, 8, -26, -1, 39, 25,
	37, -32, -35, 11, -36, -35, 32, 37, 37, -35,
	-31, -33, -36, -25, 37, 29, -35, 29, -25, -25,
	-25, -17, -9, -1, -35, 38, -25, -32, -18, -36,
	20, -35, 34, -24, 31, 33, 38, 38, -29, 38,
	37, -32, 20, 28, -22, -23, -20, -21, -19, 7,
	20, 34, -25, -25, -25, 37, 28, -20, -21, 28,
	8, 6, -19, 32, 38, -25, -35, -25, 20, 8,
	-25, 38, 9, 9, -35, -25, -24, -35, -35, 9,
	9, -35, -35,
}
var yyDef = [...]int{

	3, -2, 2, 4, 7, 0, 9, 11, 12, 14,
	17, 19, 28, 29, 30, 31, 32, 33, 34, 0,
	20, 22, 93, 93, 0, 0, 93, 93, 93, 76,
	0, 0, 79, 80, 81, 82, 83, 84, 85, 1,
	0, 0, 90, 91, 0, 8, 13, 74, 16, 21,
	23, 18, 24, 26, 0, 0, 0, 92, 94, 0,
	93, 93, 0, 0, 0, 77, 78, 5, 6, 10,
	75, 15, 25, 27, 93, 35, 71, 72, 95, 36,
	93, 0, 0, 93, -2, 0, 93, 93, 93, 0,
	70, 93, 87, 0, 93, 0, 88, 93, 0, 0,
	0, 67, 68, 73, 86, 37, 0, 0, 0, 89,
	41, 0, 60, 0, 93, 93, 65, 66, 69, 38,
	93, 0, 42, 43, 0, 0, 46, 52, 0, 0,
	58, 61, 0, 64, 0, 93, 44, 47, 53, 45,
	93, 0, 0, 93, 39, 0, 54, 55, 59, 93,
	62, 40, 93, 93, 56, 57, 63, 48, 49, 93,
	93, 50, 51,
}
var yyTok1 = [...]int{

	1, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	39, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 10, 3,
	7, 8, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 11,
	12, 3, 13, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 6,
}
var yyTok2 = [...]int{

	2, 3, 4, 5, 9, 14, 15, 16, 17, 18,
	19, 20, 21, 22, 23, 24, 25, 26, 27, 28,
	29, 30, 31, 32, 33, 34, 35, 36, 37, 38,
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
	case 15:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.elt = &element{
				redirs:  append(yyDollar[1].elt.redirs, yyDollar[3].elt.redirs...),
				assigns: yyDollar[1].elt.assigns,
			}
			yyVAL.elt.args = append(yyVAL.elt.args, yyDollar[2].word)
			yyVAL.elt.args = append(yyVAL.elt.args, yyDollar[3].elt.args...)
		}
	case 16:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.elt = yyDollar[1].elt
			yyVAL.elt.args = append(yyVAL.elt.args, yyDollar[2].word)
		}
	case 18:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.elt = &element{redirs: yyDollar[2].elt.redirs}
			yyVAL.elt.args = append(yyVAL.elt.args, yyDollar[1].word)
			yyVAL.elt.args = append(yyVAL.elt.args, yyDollar[2].elt.args...)
		}
	case 19:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.elt = new(element)
			yyVAL.elt.args = append(yyVAL.elt.args, yyDollar[1].word)
		}
	case 20:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.elt = new(element)
			yyVAL.elt.redirs = append(yyVAL.elt.redirs, yyDollar[1].redir)
		}
	case 21:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.elt.redirs = append(yyVAL.elt.redirs, yyDollar[2].redir)
		}
	case 22:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.elt = new(element)
			yyVAL.elt.assigns = append(yyVAL.elt.assigns, assign(yyDollar[1].word))
		}
	case 23:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.elt.assigns = append(yyVAL.elt.assigns, assign(yyDollar[2].word))
		}
	case 24:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.elt = new(element)
			yyVAL.elt.redirs = append(yyVAL.elt.redirs, yyDollar[1].redir)
		}
	case 25:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.elt.redirs = append(yyVAL.elt.redirs, yyDollar[2].redir)
		}
	case 26:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.elt = new(element)
			yyVAL.elt.args = append(yyVAL.elt.args, yyDollar[1].word)
		}
	case 27:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.elt.args = append(yyVAL.elt.args, yyDollar[2].word)
		}
	case 35:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.expr = &ast.Subshell{
				Lparen: yyDollar[1].token.pos,
				List:   yyDollar[2].cmds,
				Rparen: yyDollar[3].token.pos,
			}
		}
	case 36:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.expr = &ast.Group{
				Lbrace: yyDollar[1].token.pos,
				List:   yyDollar[2].cmds,
				Rbrace: yyDollar[3].token.pos,
			}
		}
	case 37:
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
	case 38:
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
	case 39:
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
	case 40:
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
	case 41:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.words = append(yyVAL.words, yyDollar[1].word)
		}
	case 42:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.words = append(yyVAL.words, yyDollar[2].word)
		}
	case 43:
		yyDollar = yyS[yypt-6 : yypt+1]
		{
			yyVAL.expr = &ast.CaseClause{
				Case: yyDollar[1].token.pos,
				Word: yyDollar[2].word,
				In:   yyDollar[4].token.pos,
				Esac: yyDollar[6].token.pos,
			}
		}
	case 44:
		yyDollar = yyS[yypt-7 : yypt+1]
		{
			yyVAL.expr = &ast.CaseClause{
				Case:  yyDollar[1].token.pos,
				Word:  yyDollar[2].word,
				In:    yyDollar[4].token.pos,
				Items: yyDollar[6].items,
				Esac:  yyDollar[7].token.pos,
			}
		}
	case 45:
		yyDollar = yyS[yypt-7 : yypt+1]
		{
			yyVAL.expr = &ast.CaseClause{
				Case:  yyDollar[1].token.pos,
				Word:  yyDollar[2].word,
				In:    yyDollar[4].token.pos,
				Items: yyDollar[6].items,
				Esac:  yyDollar[7].token.pos,
			}
		}
	case 46:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.items = append(yyVAL.items, yyDollar[1].item)
		}
	case 47:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.items = append(yyVAL.items, yyDollar[2].item)
		}
	case 48:
		yyDollar = yyS[yypt-5 : yypt+1]
		{
			yyVAL.item = &ast.CaseItem{
				Patterns: yyDollar[1].words,
				Rparen:   yyDollar[2].token.pos,
				Break:    yyDollar[4].token.pos,
			}
		}
	case 49:
		yyDollar = yyS[yypt-5 : yypt+1]
		{
			yyVAL.item = &ast.CaseItem{
				Patterns: yyDollar[1].words,
				Rparen:   yyDollar[2].token.pos,
				List:     yyDollar[3].cmds,
				Break:    yyDollar[4].token.pos,
			}
		}
	case 50:
		yyDollar = yyS[yypt-6 : yypt+1]
		{
			yyVAL.item = &ast.CaseItem{
				Lparen:   yyDollar[1].token.pos,
				Patterns: yyDollar[2].words,
				Rparen:   yyDollar[3].token.pos,
				Break:    yyDollar[5].token.pos,
			}
		}
	case 51:
		yyDollar = yyS[yypt-6 : yypt+1]
		{
			yyVAL.item = &ast.CaseItem{
				Lparen:   yyDollar[1].token.pos,
				Patterns: yyDollar[2].words,
				Rparen:   yyDollar[3].token.pos,
				List:     yyDollar[4].cmds,
				Break:    yyDollar[5].token.pos,
			}
		}
	case 52:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.items = append(yyVAL.items, yyDollar[1].item)
		}
	case 53:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.items = append(yyVAL.items, yyDollar[2].item)
		}
	case 54:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.item = &ast.CaseItem{
				Patterns: yyDollar[1].words,
				Rparen:   yyDollar[2].token.pos,
			}
		}
	case 55:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.item = &ast.CaseItem{
				Patterns: yyDollar[1].words,
				Rparen:   yyDollar[2].token.pos,
				List:     yyDollar[3].cmds,
			}
		}
	case 56:
		yyDollar = yyS[yypt-4 : yypt+1]
		{
			yyVAL.item = &ast.CaseItem{
				Lparen:   yyDollar[1].token.pos,
				Patterns: yyDollar[2].words,
				Rparen:   yyDollar[3].token.pos,
			}
		}
	case 57:
		yyDollar = yyS[yypt-4 : yypt+1]
		{
			yyVAL.item = &ast.CaseItem{
				Lparen:   yyDollar[1].token.pos,
				Patterns: yyDollar[2].words,
				Rparen:   yyDollar[3].token.pos,
				List:     yyDollar[4].cmds,
			}
		}
	case 58:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.words = append(yyVAL.words, yyDollar[1].word)
		}
	case 59:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.words = append(yyVAL.words, yyDollar[3].word)
		}
	case 60:
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
	case 61:
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
	case 62:
		yyDollar = yyS[yypt-4 : yypt+1]
		{
			yyVAL.else_ = append(yyVAL.else_, &ast.ElifClause{
				Elif: yyDollar[1].token.pos,
				Cond: yyDollar[2].cmds,
				Then: yyDollar[3].token.pos,
				List: yyDollar[4].cmds,
			})
		}
	case 63:
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
	case 64:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.else_ = append(yyVAL.else_, &ast.ElseClause{
				Else: yyDollar[1].token.pos,
				List: yyDollar[2].cmds,
			})
		}
	case 65:
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
	case 66:
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
	case 67:
		yyDollar = yyS[yypt-5 : yypt+1]
		{
			x := yyDollar[5].expr.(*ast.FuncDef)
			x.Name = yyDollar[1].word[0].(*ast.Lit)
			x.Lparen = yyDollar[2].token.pos
			x.Rparen = yyDollar[3].token.pos
			yyVAL.cmd = &ast.Cmd{Expr: yyDollar[5].expr}
		}
	case 68:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.expr = &ast.FuncDef{
				Body: &ast.Cmd{
					Expr: yyDollar[1].expr,
				},
			}
		}
	case 69:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.expr = &ast.FuncDef{
				Body: &ast.Cmd{
					Expr:   yyDollar[1].expr,
					Redirs: yyDollar[2].redirs,
				},
			}
		}
	case 70:
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
	case 71:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyDollar[2].cmds[len(yyDollar[2].cmds)-1] = extract(yyDollar[2].cmds[len(yyDollar[2].cmds)-1].(*ast.List))
			yyVAL.cmds = yyDollar[2].cmds
		}
	case 72:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.cmds = []ast.Command{yyDollar[1].list}
		}
	case 73:
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
	case 74:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.redirs = append(yyVAL.redirs, yyDollar[1].redir)
		}
	case 75:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.redirs = append(yyVAL.redirs, yyDollar[2].redir)
		}
	case 77:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.redir = yyDollar[2].redir
			yyVAL.redir.N = yyDollar[1].word[0].(*ast.Lit)
		}
	case 78:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.redir = &ast.Redir{
				OpPos: yyDollar[1].token.pos,
				Op:    yyDollar[1].token.val,
				Word:  yyDollar[2].word,
			}
		}
	case 87:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
		}
	case 89:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.token.pos = ast.Pos{}
		}
	}
	goto yystack /* stack new state and value */
}
