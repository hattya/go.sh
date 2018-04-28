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
const LAE = 57348
const RAE = 57349
const BREAK = 57350
const CLOBBER = 57351
const APPEND = 57352
const DUPIN = 57353
const DUPOUT = 57354
const RDWR = 57355
const IO_NUMBER = 57356
const WORD = 57357
const NAME = 57358
const ASSIGNMENT_WORD = 57359
const Bang = 57360
const Lbrace = 57361
const Rbrace = 57362
const For = 57363
const Case = 57364
const Esac = 57365
const In = 57366
const If = 57367
const Elif = 57368
const Then = 57369
const Else = 57370
const Fi = 57371
const While = 57372
const Until = 57373
const Do = 57374
const Done = 57375

var yyToknames = [...]string{
	"$end",
	"error",
	"$unk",
	"AND",
	"OR",
	"'|'",
	"'('",
	"')'",
	"LAE",
	"RAE",
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
		case "LAE":
			s = "(("
		case "RAE":
			s = "))"
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
	-1, 90,
	39, 91,
	-2, 94,
}

const yyPrivate = 57344

const yyLast = 310

var yyAct = [...]int{

	58, 57, 132, 131, 118, 21, 130, 48, 59, 87,
	62, 81, 2, 8, 49, 6, 51, 54, 41, 53,
	3, 44, 45, 60, 82, 155, 61, 23, 148, 25,
	66, 67, 68, 124, 34, 35, 36, 37, 38, 39,
	40, 32, 11, 20, 22, 5, 24, 122, 26, 27,
	60, 121, 28, 23, 74, 25, 54, 29, 30, 76,
	135, 111, 73, 71, 72, 88, 91, 139, 125, 100,
	75, 94, 24, 90, 26, 27, 93, 89, 28, 95,
	119, 76, 120, 29, 30, 147, 85, 89, 99, 98,
	102, 89, 92, 31, 104, 105, 106, 103, 110, 97,
	63, 101, 112, 143, 116, 60, 83, 64, 109, 108,
	115, 113, 114, 86, 49, 60, 123, 164, 152, 60,
	134, 136, 137, 115, 126, 133, 69, 138, 119, 74,
	120, 117, 142, 70, 65, 141, 146, 133, 84, 63,
	134, 149, 157, 79, 78, 150, 151, 56, 140, 154,
	85, 145, 134, 153, 158, 159, 46, 161, 162, 160,
	127, 145, 1, 144, 165, 166, 23, 96, 25, 33,
	163, 42, 43, 34, 35, 36, 37, 38, 39, 40,
	32, 11, 20, 22, 5, 24, 4, 26, 27, 80,
	129, 28, 47, 128, 107, 19, 29, 30, 23, 18,
	25, 17, 156, 16, 15, 34, 35, 36, 37, 38,
	39, 40, 32, 11, 20, 22, 5, 24, 14, 26,
	27, 13, 12, 28, 23, 10, 25, 7, 29, 30,
	9, 34, 35, 36, 37, 38, 39, 40, 32, 11,
	20, 22, 0, 24, 0, 26, 27, 0, 0, 28,
	0, 0, 0, 0, 29, 30, 34, 35, 36, 37,
	38, 39, 40, 32, 50, 0, 52, 34, 35, 36,
	37, 38, 39, 40, 32, 77, 34, 35, 36, 37,
	38, 39, 40, 32, 55, 34, 35, 36, 37, 38,
	39, 40, 32, 34, 35, 36, 37, 38, 39, 40,
	42, 43, 0, 0, 0, 0, 0, 0, 44, 45,
}
var yyPact = [...]int{

	20, -1000, 296, -1000, 150, 217, -1000, -1000, 271, -1000,
	242, 262, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	140, -1000, -1000, -18, -18, 117, 84, 112, -18, -18,
	-18, -1000, 279, 111, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, 20, 20, -1000, -1000, 217, 150, 271, -1000,
	262, -1000, -1000, 253, -1000, -1000, 136, 135, 20, -17,
	-1000, 79, 128, -1000, 74, -18, 58, 37, 32, -1000,
	-1000, -1000, -1000, -1000, -1000, 253, -1000, -1000, -18, -1000,
	9, 167, -1000, -1000, -1000, -1000, -18, 30, 70, -18,
	-17, 66, -18, -18, -18, 46, 20, -18, -17, 21,
	-18, 78, -1000, -18, 95, 11, 7, -1000, 271, 167,
	-1000, -1000, -7, 29, 64, -17, 130, -1000, 24, -18,
	-18, -1000, -1000, 271, -1000, -18, 28, -1000, 118, 73,
	-1000, -1000, 155, 98, -1000, -1000, 51, -1000, -12, -18,
	-1000, -1000, -1000, -1000, -18, 96, 145, -18, -1000, -15,
	191, 131, -1000, -18, 47, -1000, -18, -18, 159, 106,
	-1000, -1000, -1000, -18, -18, -1000, -1000,
}
var yyPgo = [...]int{

	0, 11, 20, 186, 15, 230, 227, 225, 19, 13,
	222, 221, 218, 204, 203, 201, 199, 195, 194, 10,
	2, 6, 3, 193, 190, 4, 1, 189, 5, 93,
	7, 169, 167, 9, 18, 162, 0, 8,
}
var yyR1 = [...]int{

	0, 35, 35, 35, 1, 1, 1, 2, 2, 3,
	3, 4, 4, 4, 4, 6, 6, 6, 6, 6,
	7, 7, 7, 7, 8, 8, 8, 8, 9, 9,
	9, 9, 9, 9, 9, 9, 10, 11, 12, 13,
	13, 13, 13, 19, 19, 14, 14, 14, 23, 23,
	21, 21, 21, 21, 24, 24, 22, 22, 22, 22,
	20, 20, 15, 15, 25, 25, 25, 16, 17, 5,
	18, 18, 26, 26, 27, 27, 30, 30, 28, 28,
	29, 31, 31, 31, 31, 31, 31, 31, 32, 32,
	33, 33, 34, 34, 36, 36, 37, 37,
}
var yyR2 = [...]int{

	0, 2, 1, 0, 1, 3, 3, 1, 2, 1,
	3, 1, 1, 2, 1, 3, 2, 1, 2, 1,
	1, 2, 1, 2, 1, 2, 1, 2, 1, 1,
	1, 1, 1, 1, 1, 1, 3, 3, 3, 5,
	6, 8, 9, 1, 2, 6, 7, 7, 1, 2,
	5, 5, 6, 6, 1, 2, 3, 3, 4, 4,
	1, 3, 5, 6, 4, 5, 2, 5, 5, 5,
	1, 2, 3, 2, 1, 3, 1, 2, 1, 2,
	2, 1, 1, 1, 1, 1, 1, 1, 2, 1,
	2, 1, 1, 1, 1, 0, 1, 2,
}
var yyChk = [...]int{

	-1000, -35, -1, -2, -3, 25, -4, -6, -9, -5,
	-7, 22, -10, -11, -12, -13, -14, -15, -16, -17,
	23, -28, 24, 7, 26, 9, 28, 29, 32, 37,
	38, -29, 21, -31, 14, 15, 16, 17, 18, 19,
	20, -34, 4, 5, 12, 13, 6, -3, -30, -28,
	22, -28, 24, -8, -28, 22, 7, -26, -36, -37,
	41, -26, -19, 22, 23, 22, -26, -26, -26, -29,
	22, -2, -2, -4, -28, -8, -28, 22, 8, 8,
	-27, -1, 41, 27, 10, 22, 39, -33, -36, 13,
	-37, -36, 34, 39, 39, -36, -32, -34, -37, -26,
	39, 31, -36, 31, -26, -26, -26, -18, -9, -1,
	-36, 40, -26, -33, -19, -37, -36, 36, -25, 33,
	35, 40, 40, -30, 40, 39, -33, 30, -23, -24,
	-21, -22, -20, 7, 22, 36, -26, -26, -26, 39,
	30, -21, -22, 30, 8, 6, -20, 34, 40, -26,
	-36, -26, 22, 8, -26, 40, 11, 11, -36, -26,
	-25, -36, -36, 11, 11, -36, -36,
}
var yyDef = [...]int{

	3, -2, 2, 4, 7, 0, 9, 11, 12, 14,
	17, 19, 28, 29, 30, 31, 32, 33, 34, 35,
	0, 20, 22, 95, 95, 0, 0, 0, 95, 95,
	95, 78, 0, 0, 81, 82, 83, 84, 85, 86,
	87, 1, 0, 0, 92, 93, 0, 8, 13, 76,
	16, 21, 23, 18, 24, 26, 0, 0, 0, 94,
	96, 0, 0, 43, 95, 95, 0, 0, 0, 79,
	80, 5, 6, 10, 77, 15, 25, 27, 95, 36,
	73, 74, 97, 37, 38, 44, 95, 0, 0, 95,
	-2, 0, 95, 95, 95, 0, 72, 95, 89, 0,
	95, 0, 90, 95, 0, 0, 0, 69, 70, 75,
	88, 39, 0, 0, 0, 91, 0, 62, 0, 95,
	95, 67, 68, 71, 40, 95, 0, 45, 0, 0,
	48, 54, 0, 0, 60, 63, 0, 66, 0, 95,
	46, 49, 55, 47, 95, 0, 0, 95, 41, 0,
	56, 57, 61, 95, 64, 42, 95, 95, 58, 59,
	65, 50, 51, 95, 95, 52, 53,
}
var yyTok1 = [...]int{

	1, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	41, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 12, 3,
	7, 8, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 13,
	14, 3, 15, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 6,
}
var yyTok2 = [...]int{

	2, 3, 4, 5, 9, 10, 11, 16, 17, 18,
	19, 20, 21, 22, 23, 24, 25, 26, 27, 28,
	29, 30, 31, 32, 33, 34, 35, 36, 37, 38,
	39, 40,
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
	case 36:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.expr = &ast.Subshell{
				Lparen: yyDollar[1].token.pos,
				List:   yyDollar[2].cmds,
				Rparen: yyDollar[3].token.pos,
			}
		}
	case 37:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.expr = &ast.Group{
				Lbrace: yyDollar[1].token.pos,
				List:   yyDollar[2].cmds,
				Rbrace: yyDollar[3].token.pos,
			}
		}
	case 38:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.expr = &ast.ArithEval{
				Left:  yyDollar[1].token.pos,
				Expr:  yyDollar[2].words,
				Right: yyDollar[3].token.pos,
			}
		}
	case 39:
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
	case 40:
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
	case 41:
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
	case 42:
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
	case 43:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.words = append(yyVAL.words, yyDollar[1].word)
		}
	case 44:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.words = append(yyVAL.words, yyDollar[2].word)
		}
	case 45:
		yyDollar = yyS[yypt-6 : yypt+1]
		{
			yyVAL.expr = &ast.CaseClause{
				Case: yyDollar[1].token.pos,
				Word: yyDollar[2].word,
				In:   yyDollar[4].token.pos,
				Esac: yyDollar[6].token.pos,
			}
		}
	case 46:
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
	case 47:
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
	case 48:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.items = append(yyVAL.items, yyDollar[1].item)
		}
	case 49:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.items = append(yyVAL.items, yyDollar[2].item)
		}
	case 50:
		yyDollar = yyS[yypt-5 : yypt+1]
		{
			yyVAL.item = &ast.CaseItem{
				Patterns: yyDollar[1].words,
				Rparen:   yyDollar[2].token.pos,
				Break:    yyDollar[4].token.pos,
			}
		}
	case 51:
		yyDollar = yyS[yypt-5 : yypt+1]
		{
			yyVAL.item = &ast.CaseItem{
				Patterns: yyDollar[1].words,
				Rparen:   yyDollar[2].token.pos,
				List:     yyDollar[3].cmds,
				Break:    yyDollar[4].token.pos,
			}
		}
	case 52:
		yyDollar = yyS[yypt-6 : yypt+1]
		{
			yyVAL.item = &ast.CaseItem{
				Lparen:   yyDollar[1].token.pos,
				Patterns: yyDollar[2].words,
				Rparen:   yyDollar[3].token.pos,
				Break:    yyDollar[5].token.pos,
			}
		}
	case 53:
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
	case 54:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.items = append(yyVAL.items, yyDollar[1].item)
		}
	case 55:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.items = append(yyVAL.items, yyDollar[2].item)
		}
	case 56:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.item = &ast.CaseItem{
				Patterns: yyDollar[1].words,
				Rparen:   yyDollar[2].token.pos,
			}
		}
	case 57:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.item = &ast.CaseItem{
				Patterns: yyDollar[1].words,
				Rparen:   yyDollar[2].token.pos,
				List:     yyDollar[3].cmds,
			}
		}
	case 58:
		yyDollar = yyS[yypt-4 : yypt+1]
		{
			yyVAL.item = &ast.CaseItem{
				Lparen:   yyDollar[1].token.pos,
				Patterns: yyDollar[2].words,
				Rparen:   yyDollar[3].token.pos,
			}
		}
	case 59:
		yyDollar = yyS[yypt-4 : yypt+1]
		{
			yyVAL.item = &ast.CaseItem{
				Lparen:   yyDollar[1].token.pos,
				Patterns: yyDollar[2].words,
				Rparen:   yyDollar[3].token.pos,
				List:     yyDollar[4].cmds,
			}
		}
	case 60:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.words = append(yyVAL.words, yyDollar[1].word)
		}
	case 61:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.words = append(yyVAL.words, yyDollar[3].word)
		}
	case 62:
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
	case 63:
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
	case 64:
		yyDollar = yyS[yypt-4 : yypt+1]
		{
			yyVAL.else_ = append(yyVAL.else_, &ast.ElifClause{
				Elif: yyDollar[1].token.pos,
				Cond: yyDollar[2].cmds,
				Then: yyDollar[3].token.pos,
				List: yyDollar[4].cmds,
			})
		}
	case 65:
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
	case 66:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.else_ = append(yyVAL.else_, &ast.ElseClause{
				Else: yyDollar[1].token.pos,
				List: yyDollar[2].cmds,
			})
		}
	case 67:
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
	case 68:
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
	case 69:
		yyDollar = yyS[yypt-5 : yypt+1]
		{
			x := yyDollar[5].expr.(*ast.FuncDef)
			x.Name = yyDollar[1].word[0].(*ast.Lit)
			x.Lparen = yyDollar[2].token.pos
			x.Rparen = yyDollar[3].token.pos
			yyVAL.cmd = &ast.Cmd{Expr: yyDollar[5].expr}
		}
	case 70:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.expr = &ast.FuncDef{
				Body: &ast.Cmd{
					Expr: yyDollar[1].expr,
				},
			}
		}
	case 71:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.expr = &ast.FuncDef{
				Body: &ast.Cmd{
					Expr:   yyDollar[1].expr,
					Redirs: yyDollar[2].redirs,
				},
			}
		}
	case 72:
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
	case 73:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyDollar[2].cmds[len(yyDollar[2].cmds)-1] = extract(yyDollar[2].cmds[len(yyDollar[2].cmds)-1].(*ast.List))
			yyVAL.cmds = yyDollar[2].cmds
		}
	case 74:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.cmds = []ast.Command{yyDollar[1].list}
		}
	case 75:
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
	case 76:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.redirs = append(yyVAL.redirs, yyDollar[1].redir)
		}
	case 77:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.redirs = append(yyVAL.redirs, yyDollar[2].redir)
		}
	case 79:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.redir = yyDollar[2].redir
			yyVAL.redir.N = yyDollar[1].word[0].(*ast.Lit)
		}
	case 80:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.redir = &ast.Redir{
				OpPos: yyDollar[1].token.pos,
				Op:    yyDollar[1].token.val,
				Word:  yyDollar[2].word,
			}
		}
	case 89:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
		}
	case 91:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.token.pos = ast.Pos{}
		}
	}
	goto yystack /* stack new state and value */
}
