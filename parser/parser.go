// Code generated by goyacc -l -o parser.go parser.go.y. DO NOT EDIT.
//
// go.sh/parser :: parser.go
//
//   Copyright (c) 2018-2021 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
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
	"github.com/hattya/go.sh/interp"
)

type yySymType struct {
	yys   int
	list  interface{}
	node  ast.Node
	elt   *element
	word  ast.Word
	token token
}

const AND = 57346
const OR = 57347
const LAE = 57348
const RAE = 57349
const BREAK = 57350
const CLOBBER = 57351
const APPEND = 57352
const HEREDOC = 57353
const HEREDOCI = 57354
const DUPIN = 57355
const DUPOUT = 57356
const RDWR = 57357
const IO_NUMBER = 57358
const WORD = 57359
const NAME = 57360
const ASSIGNMENT_WORD = 57361
const Bang = 57362
const Lbrace = 57363
const Rbrace = 57364
const For = 57365
const Case = 57366
const Esac = 57367
const In = 57368
const If = 57369
const Elif = 57370
const Then = 57371
const Else = 57372
const Fi = 57373
const While = 57374
const Until = 57375
const Do = 57376
const Done = 57377

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
	"HEREDOC",
	"HEREDOCI",
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
		case "HEREDOC":
			s = "'<<'"
		case "HEREDOCI":
			s = "'<<-'"
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

func extract(cmd *ast.AndOrList) ast.Command {
	switch {
	case len(cmd.List) != 0 || !cmd.SepPos.IsZero():
		return cmd
	case !cmd.Pipeline.Bang.IsZero() || len(cmd.Pipeline.List) != 0:
		return cmd.Pipeline
	}
	return cmd.Pipeline.Cmd
}

func assign(w ast.Word) *ast.Assign {
	n := w[0].(*ast.Lit)
	if i := strings.IndexRune(n.Value, '='); 0 < i && i < len(n.Value)-1 {
		w[0] = &ast.Lit{
			ValuePos: ast.NewPos(n.ValuePos.Line(), n.ValuePos.Col()+i+1),
			Value:    n.Value[i+1:],
		}
		n.Value = n.Value[:i]
	} else {
		w = w[1:]
		n.Value = n.Value[:len(n.Value)-1]
	}
	return &ast.Assign{
		Name:  n,
		Op:    "=",
		Value: w,
	}
}

// ParseCommands parses src, including alias substitution, and returns
// commands.
func ParseCommands(env *interp.ExecEnv, name string, src interface{}) ([]ast.Command, []*ast.Comment, error) {
	r, err := open(src)
	if err != nil {
		return nil, nil, err
	}

	l := newLexer(env, name, r)
	yyParse(l)
	return l.cmds, l.comments, l.err
}

// ParseCommand parses src and returns a command.
func ParseCommand(name string, src interface{}) (ast.Command, []*ast.Comment, error) {
	cmds, comments, err := ParseCommands(nil, name, src)
	if len(cmds) == 0 {
		return nil, comments, err
	}
	return cmds[0], comments, err
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
	-1, 101,
	41, 101,
	-2, 104,
}

const yyPrivate = 57344

const yyLast = 386

var yyAct = [...]int{
	68, 67, 145, 48, 130, 24, 144, 58, 143, 69,
	98, 11, 49, 94, 5, 63, 51, 59, 9, 61,
	64, 3, 6, 26, 100, 28, 50, 100, 100, 70,
	52, 53, 82, 74, 75, 76, 168, 161, 139, 127,
	136, 134, 133, 122, 27, 152, 29, 30, 137, 111,
	31, 105, 97, 104, 50, 32, 33, 50, 50, 148,
	131, 50, 132, 5, 87, 83, 64, 160, 103, 89,
	146, 81, 36, 99, 102, 86, 88, 84, 85, 146,
	34, 131, 101, 132, 129, 114, 112, 147, 156, 95,
	72, 165, 106, 147, 89, 153, 147, 80, 79, 110,
	73, 113, 71, 109, 140, 115, 116, 117, 78, 121,
	108, 52, 53, 123, 177, 128, 77, 170, 119, 96,
	7, 120, 126, 124, 158, 59, 166, 135, 158, 57,
	157, 92, 91, 149, 150, 126, 138, 66, 56, 151,
	2, 87, 54, 55, 1, 107, 38, 37, 155, 159,
	154, 93, 142, 141, 162, 125, 118, 22, 163, 164,
	21, 20, 167, 19, 18, 17, 16, 171, 172, 15,
	174, 175, 173, 26, 13, 28, 10, 178, 179, 12,
	39, 40, 41, 42, 46, 47, 43, 44, 45, 35,
	14, 23, 25, 8, 27, 4, 29, 30, 0, 0,
	31, 0, 0, 0, 0, 32, 33, 0, 26, 82,
	28, 0, 176, 0, 0, 39, 40, 41, 42, 46,
	47, 43, 44, 45, 35, 14, 23, 25, 8, 27,
	0, 29, 30, 0, 0, 31, 0, 0, 0, 0,
	32, 33, 26, 0, 28, 0, 169, 0, 0, 39,
	40, 41, 42, 46, 47, 43, 44, 45, 35, 14,
	23, 25, 8, 27, 0, 29, 30, 0, 0, 31,
	26, 0, 28, 0, 32, 33, 0, 39, 40, 41,
	42, 46, 47, 43, 44, 45, 35, 14, 23, 25,
	8, 27, 0, 29, 30, 0, 0, 31, 26, 0,
	28, 0, 32, 33, 0, 39, 40, 41, 42, 46,
	47, 43, 44, 45, 35, 14, 23, 25, 0, 27,
	0, 29, 30, 0, 0, 31, 0, 0, 0, 0,
	32, 33, 39, 40, 41, 42, 46, 47, 43, 44,
	45, 35, 60, 0, 62, 39, 40, 41, 42, 46,
	47, 43, 44, 45, 35, 90, 39, 40, 41, 42,
	46, 47, 43, 44, 45, 35, 65, 39, 40, 41,
	42, 46, 47, 43, 44, 45, 35, 39, 40, 41,
	42, 46, 47, 43, 44, 45,
}

var yyPact = [...]int{
	263, -1000, -17, -1000, 99, 138, -1000, 132, 291, -1000,
	-1000, 353, -1000, 318, 342, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, 130, -1000, -1000, -17, -17, 78, 65,
	76, -17, -17, -17, -1000, 363, -1000, 74, 73, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, 166,
	-1000, 263, -1000, -1000, 263, 263, 291, 132, 353, -1000,
	342, -1000, -1000, 331, -1000, -1000, 124, 123, 263, -11,
	60, 109, 11, -17, 32, 12, 10, -1000, -1000, -1000,
	-1000, -1000, -1000, 138, -1000, -1000, -1000, -1000, 331, -1000,
	-1000, -17, -1000, 18, 138, -1000, -1000, -17, 8, 53,
	-17, -11, 52, -17, -17, -17, 16, 263, -17, -11,
	1, -17, 15, -1000, -17, 46, 0, -1, -1000, 353,
	138, -1000, -1000, -2, 7, 14, -11, -1000, 72, -1000,
	21, -17, -17, -1000, -1000, 353, -1000, -17, 4, -1000,
	-1000, 63, 56, -1000, -1000, 122, 69, -1000, -1000, 31,
	-1000, -5, -17, -1000, -1000, -1000, -1000, -17, 67, 118,
	-17, -1000, -6, 235, 106, -1000, -17, 25, -1000, -17,
	-17, 201, 103, -1000, -1000, -1000, -17, -17, -1000, -1000,
}

var yyPgo = [...]int{
	0, 21, 195, 13, 22, 120, 18, 179, 176, 174,
	15, 11, 169, 166, 165, 164, 163, 161, 160, 157,
	156, 155, 2, 153, 152, 8, 6, 4, 1, 151,
	7, 5, 80, 72, 147, 146, 145, 10, 16, 144,
	140, 0, 9,
}

var yyR1 = [...]int{
	0, 39, 39, 40, 40, 1, 1, 2, 2, 3,
	3, 3, 4, 4, 5, 5, 6, 6, 6, 6,
	8, 8, 8, 8, 8, 9, 9, 9, 9, 10,
	10, 10, 10, 11, 11, 11, 11, 11, 11, 11,
	11, 12, 13, 14, 15, 15, 15, 15, 21, 21,
	16, 16, 16, 23, 23, 25, 25, 25, 25, 24,
	24, 26, 26, 26, 26, 22, 22, 17, 17, 27,
	27, 27, 18, 19, 7, 20, 20, 28, 28, 29,
	29, 30, 30, 31, 31, 31, 31, 32, 34, 34,
	34, 34, 34, 34, 34, 33, 35, 35, 36, 36,
	37, 37, 38, 38, 41, 41, 42, 42,
}

var yyR2 = [...]int{
	0, 2, 0, 1, 3, 2, 1, 1, 3, 1,
	3, 3, 1, 2, 1, 3, 1, 1, 2, 1,
	3, 2, 1, 2, 1, 1, 2, 1, 2, 1,
	2, 1, 2, 1, 1, 1, 1, 1, 1, 1,
	1, 3, 3, 3, 5, 6, 8, 9, 1, 2,
	6, 7, 7, 1, 2, 5, 5, 6, 6, 1,
	2, 3, 3, 4, 4, 1, 3, 5, 6, 4,
	5, 2, 5, 5, 5, 1, 2, 3, 2, 1,
	3, 1, 2, 1, 2, 1, 2, 2, 1, 1,
	1, 1, 1, 1, 1, 2, 1, 1, 2, 1,
	2, 1, 1, 1, 1, 0, 1, 2,
}

var yyChk = [...]int{
	-1000, -39, -40, -1, -2, -3, -4, -5, 27, -6,
	-8, -11, -7, -9, 24, -12, -13, -14, -15, -16,
	-17, -18, -19, 25, -31, 26, 7, 28, 9, 30,
	31, 34, 39, 40, -32, 23, -33, -34, -35, 14,
	15, 16, 17, 20, 21, 22, 18, 19, -41, -42,
	43, -38, 12, 13, 4, 5, 6, -5, -30, -31,
	24, -31, 26, -10, -31, 24, 7, -28, -41, -42,
	-28, 24, 25, 24, -28, -28, -28, -32, -33, 24,
	24, -1, 43, -3, -4, -4, -6, -31, -10, -31,
	24, 8, 8, -29, -3, 29, 10, 41, -37, -41,
	13, -42, -41, 36, 41, 41, -41, -36, -38, -42,
	-28, 41, 33, -41, 33, -28, -28, -28, -20, -11,
	-3, -41, 42, -28, -37, -21, -42, 24, -41, 38,
	-27, 35, 37, 42, 42, -30, 42, 41, -37, 24,
	32, -23, -24, -25, -26, -22, 7, 24, 38, -28,
	-28, -28, 41, 32, -25, -26, 32, 8, 6, -22,
	36, 42, -28, -41, -28, 24, 8, -28, 42, 11,
	11, -41, -28, -27, -41, -41, 11, 11, -41, -41,
}

var yyDef = [...]int{
	2, -2, 105, 3, 6, 7, 9, 12, 0, 14,
	16, 17, 19, 22, 24, 33, 34, 35, 36, 37,
	38, 39, 40, 0, 25, 27, 105, 105, 0, 0,
	0, 105, 105, 105, 83, 0, 85, 0, 0, 88,
	89, 90, 91, 92, 93, 94, 96, 97, 1, 104,
	106, 5, 102, 103, 0, 0, 0, 13, 18, 81,
	21, 26, 28, 23, 29, 31, 0, 0, 0, 104,
	0, 0, 105, 105, 0, 0, 0, 84, 86, 87,
	95, 4, 107, 8, 10, 11, 15, 82, 20, 30,
	32, 105, 41, 78, 79, 42, 43, 105, 0, 0,
	105, -2, 0, 105, 105, 105, 0, 77, 105, 99,
	0, 105, 0, 100, 105, 0, 0, 0, 74, 75,
	80, 98, 44, 0, 0, 0, 101, 48, 0, 67,
	0, 105, 105, 72, 73, 76, 45, 105, 0, 49,
	50, 0, 0, 53, 59, 0, 0, 65, 68, 0,
	71, 0, 105, 51, 54, 60, 52, 105, 0, 0,
	105, 46, 0, 61, 62, 66, 105, 69, 47, 105,
	105, 63, 64, 70, 55, 56, 105, 105, 57, 58,
}

var yyTok1 = [...]int{
	1, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	43, 3, 3, 3, 3, 3, 3, 3, 3, 3,
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
	39, 40, 41, 42,
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

	case 3:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			l := yyDollar[1].list.(ast.List)
			if len(l) > 1 {
				yylex.(*lexer).cmds = append(yylex.(*lexer).cmds, l)
			} else {
				yylex.(*lexer).cmds = append(yylex.(*lexer).cmds, extract(l[0]))
			}
		}
	case 4:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			l := yyDollar[3].list.(ast.List)
			if len(l) > 1 {
				yylex.(*lexer).cmds = append(yylex.(*lexer).cmds, l)
			} else {
				yylex.(*lexer).cmds = append(yylex.(*lexer).cmds, extract(l[0]))
			}
		}
	case 5:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			l := yyDollar[1].list.(ast.List)
			ao := l[len(l)-1]
			ao.SepPos = yyDollar[2].token.pos
			ao.Sep = yyDollar[2].token.val
		}
	case 7:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.list = ast.List{yyDollar[1].node.(*ast.AndOrList)}
		}
	case 8:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			l := yyVAL.list.(ast.List)
			ao := l[len(l)-1]
			ao.SepPos = yyDollar[2].token.pos
			ao.Sep = yyDollar[2].token.val
			yyVAL.list = append(l, yyDollar[3].node.(*ast.AndOrList))
		}
	case 9:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.node = &ast.AndOrList{Pipeline: yyDollar[1].node.(*ast.Pipeline)}
		}
	case 10:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.node.(*ast.AndOrList).List = append(yyVAL.node.(*ast.AndOrList).List, &ast.AndOr{
				OpPos:    yyDollar[2].token.pos,
				Op:       yyDollar[2].token.val,
				Pipeline: yyDollar[3].node.(*ast.Pipeline),
			})
		}
	case 11:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.node.(*ast.AndOrList).List = append(yyVAL.node.(*ast.AndOrList).List, &ast.AndOr{
				OpPos:    yyDollar[2].token.pos,
				Op:       yyDollar[2].token.val,
				Pipeline: yyDollar[3].node.(*ast.Pipeline),
			})
		}
	case 13:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.node = yyDollar[2].node
			yyVAL.node.(*ast.Pipeline).Bang = yyDollar[1].token.pos
		}
	case 14:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.node = &ast.Pipeline{Cmd: yyDollar[1].node.(*ast.Cmd)}
		}
	case 15:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.node.(*ast.Pipeline).List = append(yyVAL.node.(*ast.Pipeline).List, &ast.Pipe{
				OpPos: yyDollar[2].token.pos,
				Op:    yyDollar[2].token.val,
				Cmd:   yyDollar[3].node.(*ast.Cmd),
			})
		}
	case 16:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.node = &ast.Cmd{
				Expr: &ast.SimpleCmd{
					Assigns: yyDollar[1].elt.assigns,
					Args:    yyDollar[1].elt.args,
				},
				Redirs: yyDollar[1].elt.redirs,
			}
		}
	case 17:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.node = &ast.Cmd{Expr: yyDollar[1].node.(ast.CmdExpr)}
		}
	case 18:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.node = &ast.Cmd{
				Expr:   yyDollar[1].node.(ast.CmdExpr),
				Redirs: yyDollar[2].list.([]*ast.Redir),
			}
		}
	case 20:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.elt = &element{
				redirs:  append(yyDollar[1].elt.redirs, yyDollar[3].elt.redirs...),
				assigns: yyDollar[1].elt.assigns,
				args:    append([]ast.Word{yyDollar[2].word}, yyDollar[3].elt.args...),
			}
		}
	case 21:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.elt = yyDollar[1].elt
			yyVAL.elt.args = append(yyVAL.elt.args, yyDollar[2].word)
		}
	case 23:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.elt = &element{
				redirs: yyDollar[2].elt.redirs,
				args:   append([]ast.Word{yyDollar[1].word}, yyDollar[2].elt.args...),
			}
		}
	case 24:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.elt = &element{args: []ast.Word{yyDollar[1].word}}
		}
	case 25:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.elt = &element{redirs: []*ast.Redir{yyDollar[1].node.(*ast.Redir)}}
		}
	case 26:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.elt.redirs = append(yyVAL.elt.redirs, yyDollar[2].node.(*ast.Redir))
		}
	case 27:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.elt = &element{assigns: []*ast.Assign{assign(yyDollar[1].word)}}
		}
	case 28:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.elt.assigns = append(yyVAL.elt.assigns, assign(yyDollar[2].word))
		}
	case 29:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.elt = &element{redirs: []*ast.Redir{yyDollar[1].node.(*ast.Redir)}}
		}
	case 30:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.elt.redirs = append(yyVAL.elt.redirs, yyDollar[2].node.(*ast.Redir))
		}
	case 31:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.elt = &element{args: []ast.Word{yyDollar[1].word}}
		}
	case 32:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.elt.args = append(yyVAL.elt.args, yyDollar[2].word)
		}
	case 41:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.node = &ast.Subshell{
				Lparen: yyDollar[1].token.pos,
				List:   yyDollar[2].list.([]ast.Command),
				Rparen: yyDollar[3].token.pos,
			}
		}
	case 42:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.node = &ast.Group{
				Lbrace: yyDollar[1].token.pos,
				List:   yyDollar[2].list.([]ast.Command),
				Rbrace: yyDollar[3].token.pos,
			}
		}
	case 43:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.node = &ast.ArithEval{
				Left:  yyDollar[1].token.pos,
				Expr:  yyDollar[2].word,
				Right: yyDollar[3].token.pos,
			}
		}
	case 44:
		yyDollar = yyS[yypt-5 : yypt+1]
		{
			yyVAL.node = &ast.ForClause{
				For:  yyDollar[1].token.pos,
				Name: yyDollar[2].word[0].(*ast.Lit),
				Do:   yyDollar[3].token.pos,
				List: yyDollar[4].list.([]ast.Command),
				Done: yyDollar[5].token.pos,
			}
		}
	case 45:
		yyDollar = yyS[yypt-6 : yypt+1]
		{
			yyVAL.node = &ast.ForClause{
				For:       yyDollar[1].token.pos,
				Name:      yyDollar[2].word[0].(*ast.Lit),
				Semicolon: yyDollar[3].token.pos,
				Do:        yyDollar[4].token.pos,
				List:      yyDollar[5].list.([]ast.Command),
				Done:      yyDollar[6].token.pos,
			}
		}
	case 46:
		yyDollar = yyS[yypt-8 : yypt+1]
		{
			yyVAL.node = &ast.ForClause{
				For:       yyDollar[1].token.pos,
				Name:      yyDollar[2].word[0].(*ast.Lit),
				In:        yyDollar[4].token.pos,
				Semicolon: yyDollar[5].token.pos,
				Do:        yyDollar[6].token.pos,
				List:      yyDollar[7].list.([]ast.Command),
				Done:      yyDollar[8].token.pos,
			}
		}
	case 47:
		yyDollar = yyS[yypt-9 : yypt+1]
		{
			yyVAL.node = &ast.ForClause{
				For:       yyDollar[1].token.pos,
				Name:      yyDollar[2].word[0].(*ast.Lit),
				In:        yyDollar[4].token.pos,
				Items:     yyDollar[5].list.([]ast.Word),
				Semicolon: yyDollar[6].token.pos,
				Do:        yyDollar[7].token.pos,
				List:      yyDollar[8].list.([]ast.Command),
				Done:      yyDollar[9].token.pos,
			}
		}
	case 48:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.list = []ast.Word{yyDollar[1].word}
		}
	case 49:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.list = append(yyVAL.list.([]ast.Word), yyDollar[2].word)
		}
	case 50:
		yyDollar = yyS[yypt-6 : yypt+1]
		{
			yyVAL.node = &ast.CaseClause{
				Case: yyDollar[1].token.pos,
				Word: yyDollar[2].word,
				In:   yyDollar[4].token.pos,
				Esac: yyDollar[6].token.pos,
			}
		}
	case 51:
		yyDollar = yyS[yypt-7 : yypt+1]
		{
			yyVAL.node = &ast.CaseClause{
				Case:  yyDollar[1].token.pos,
				Word:  yyDollar[2].word,
				In:    yyDollar[4].token.pos,
				Items: yyDollar[6].list.([]*ast.CaseItem),
				Esac:  yyDollar[7].token.pos,
			}
		}
	case 52:
		yyDollar = yyS[yypt-7 : yypt+1]
		{
			yyVAL.node = &ast.CaseClause{
				Case:  yyDollar[1].token.pos,
				Word:  yyDollar[2].word,
				In:    yyDollar[4].token.pos,
				Items: yyDollar[6].list.([]*ast.CaseItem),
				Esac:  yyDollar[7].token.pos,
			}
		}
	case 53:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.list = []*ast.CaseItem{yyDollar[1].node.(*ast.CaseItem)}
		}
	case 54:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.list = append(yyVAL.list.([]*ast.CaseItem), yyDollar[2].node.(*ast.CaseItem))
		}
	case 55:
		yyDollar = yyS[yypt-5 : yypt+1]
		{
			yyVAL.node = &ast.CaseItem{
				Patterns: yyDollar[1].list.([]ast.Word),
				Rparen:   yyDollar[2].token.pos,
				Break:    yyDollar[4].token.pos,
			}
		}
	case 56:
		yyDollar = yyS[yypt-5 : yypt+1]
		{
			yyVAL.node = &ast.CaseItem{
				Patterns: yyDollar[1].list.([]ast.Word),
				Rparen:   yyDollar[2].token.pos,
				List:     yyDollar[3].list.([]ast.Command),
				Break:    yyDollar[4].token.pos,
			}
		}
	case 57:
		yyDollar = yyS[yypt-6 : yypt+1]
		{
			yyVAL.node = &ast.CaseItem{
				Lparen:   yyDollar[1].token.pos,
				Patterns: yyDollar[2].list.([]ast.Word),
				Rparen:   yyDollar[3].token.pos,
				Break:    yyDollar[5].token.pos,
			}
		}
	case 58:
		yyDollar = yyS[yypt-6 : yypt+1]
		{
			yyVAL.node = &ast.CaseItem{
				Lparen:   yyDollar[1].token.pos,
				Patterns: yyDollar[2].list.([]ast.Word),
				Rparen:   yyDollar[3].token.pos,
				List:     yyDollar[4].list.([]ast.Command),
				Break:    yyDollar[5].token.pos,
			}
		}
	case 59:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.list = []*ast.CaseItem{yyDollar[1].node.(*ast.CaseItem)}
		}
	case 60:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.list = append(yyVAL.list.([]*ast.CaseItem), yyDollar[2].node.(*ast.CaseItem))
		}
	case 61:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.node = &ast.CaseItem{
				Patterns: yyDollar[1].list.([]ast.Word),
				Rparen:   yyDollar[2].token.pos,
			}
		}
	case 62:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.node = &ast.CaseItem{
				Patterns: yyDollar[1].list.([]ast.Word),
				Rparen:   yyDollar[2].token.pos,
				List:     yyDollar[3].list.([]ast.Command),
			}
		}
	case 63:
		yyDollar = yyS[yypt-4 : yypt+1]
		{
			yyVAL.node = &ast.CaseItem{
				Lparen:   yyDollar[1].token.pos,
				Patterns: yyDollar[2].list.([]ast.Word),
				Rparen:   yyDollar[3].token.pos,
			}
		}
	case 64:
		yyDollar = yyS[yypt-4 : yypt+1]
		{
			yyVAL.node = &ast.CaseItem{
				Lparen:   yyDollar[1].token.pos,
				Patterns: yyDollar[2].list.([]ast.Word),
				Rparen:   yyDollar[3].token.pos,
				List:     yyDollar[4].list.([]ast.Command),
			}
		}
	case 65:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.list = []ast.Word{yyDollar[1].word}
		}
	case 66:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.list = append(yyVAL.list.([]ast.Word), yyDollar[3].word)
		}
	case 67:
		yyDollar = yyS[yypt-5 : yypt+1]
		{
			yyVAL.node = &ast.IfClause{
				If:   yyDollar[1].token.pos,
				Cond: yyDollar[2].list.([]ast.Command),
				Then: yyDollar[3].token.pos,
				List: yyDollar[4].list.([]ast.Command),
				Fi:   yyDollar[5].token.pos,
			}
		}
	case 68:
		yyDollar = yyS[yypt-6 : yypt+1]
		{
			yyVAL.node = &ast.IfClause{
				If:   yyDollar[1].token.pos,
				Cond: yyDollar[2].list.([]ast.Command),
				Then: yyDollar[3].token.pos,
				List: yyDollar[4].list.([]ast.Command),
				Else: yyDollar[5].list.([]ast.ElsePart),
				Fi:   yyDollar[6].token.pos,
			}
		}
	case 69:
		yyDollar = yyS[yypt-4 : yypt+1]
		{
			yyVAL.list = []ast.ElsePart{&ast.ElifClause{
				Elif: yyDollar[1].token.pos,
				Cond: yyDollar[2].list.([]ast.Command),
				Then: yyDollar[3].token.pos,
				List: yyDollar[4].list.([]ast.Command),
			}}
		}
	case 70:
		yyDollar = yyS[yypt-5 : yypt+1]
		{
			yyVAL.list = append([]ast.ElsePart{&ast.ElifClause{
				Elif: yyDollar[1].token.pos,
				Cond: yyDollar[2].list.([]ast.Command),
				Then: yyDollar[3].token.pos,
				List: yyDollar[4].list.([]ast.Command),
			}}, yyDollar[5].list.([]ast.ElsePart)...)
		}
	case 71:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.list = []ast.ElsePart{&ast.ElseClause{
				Else: yyDollar[1].token.pos,
				List: yyDollar[2].list.([]ast.Command),
			}}
		}
	case 72:
		yyDollar = yyS[yypt-5 : yypt+1]
		{
			yyVAL.node = &ast.WhileClause{
				While: yyDollar[1].token.pos,
				Cond:  yyDollar[2].list.([]ast.Command),
				Do:    yyDollar[3].token.pos,
				List:  yyDollar[4].list.([]ast.Command),
				Done:  yyDollar[5].token.pos,
			}
		}
	case 73:
		yyDollar = yyS[yypt-5 : yypt+1]
		{
			yyVAL.node = &ast.UntilClause{
				Until: yyDollar[1].token.pos,
				Cond:  yyDollar[2].list.([]ast.Command),
				Do:    yyDollar[3].token.pos,
				List:  yyDollar[4].list.([]ast.Command),
				Done:  yyDollar[5].token.pos,
			}
		}
	case 74:
		yyDollar = yyS[yypt-5 : yypt+1]
		{
			x := yyDollar[5].node.(*ast.FuncDef)
			x.Name = yyDollar[1].word[0].(*ast.Lit)
			x.Lparen = yyDollar[2].token.pos
			x.Rparen = yyDollar[3].token.pos
			yyVAL.node = &ast.Cmd{Expr: yyDollar[5].node.(ast.CmdExpr)}
		}
	case 75:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.node = &ast.FuncDef{
				Body: &ast.Cmd{
					Expr: yyDollar[1].node.(ast.CmdExpr),
				},
			}
		}
	case 76:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.node = &ast.FuncDef{
				Body: &ast.Cmd{
					Expr:   yyDollar[1].node.(ast.CmdExpr),
					Redirs: yyDollar[2].list.([]*ast.Redir),
				},
			}
		}
	case 77:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			cmds := yyDollar[2].list.([]ast.Command)
			l := cmds[len(cmds)-1].(ast.List)
			if yyDollar[3].token.typ != '\n' {
				ao := l[len(l)-1]
				ao.SepPos = yyDollar[3].token.pos
				ao.Sep = yyDollar[3].token.val
			}
			if len(l) == 1 {
				cmds[len(cmds)-1] = extract(l[0])
			}
			yyVAL.list = yyDollar[2].list
		}
	case 78:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			cmds := yyDollar[2].list.([]ast.Command)
			l := cmds[len(cmds)-1].(ast.List)
			if len(l) == 1 {
				cmds[len(cmds)-1] = extract(l[0])
			}
			yyVAL.list = yyDollar[2].list
		}
	case 79:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.list = []ast.Command{ast.List{yyDollar[1].node.(*ast.AndOrList)}}
		}
	case 80:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			cmds := yyVAL.list.([]ast.Command)
			l := cmds[len(cmds)-1].(ast.List)
			if yyDollar[2].token.typ != '\n' {
				ao := l[len(l)-1]
				ao.SepPos = yyDollar[2].token.pos
				ao.Sep = yyDollar[2].token.val
				cmds[len(cmds)-1] = append(l, yyDollar[3].node.(*ast.AndOrList))
			} else {
				if len(l) == 1 {
					cmds[len(cmds)-1] = extract(l[0])
				}
				yyVAL.list = append(cmds, ast.List{yyDollar[3].node.(*ast.AndOrList)})
			}
		}
	case 81:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.list = []*ast.Redir{yyDollar[1].node.(*ast.Redir)}
		}
	case 82:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.list = append(yyVAL.list.([]*ast.Redir), yyDollar[2].node.(*ast.Redir))
		}
	case 84:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.node = yyDollar[2].node
			yyVAL.node.(*ast.Redir).N = yyDollar[1].word[0].(*ast.Lit)
		}
	case 86:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.node = yyDollar[2].node
			yyVAL.node.(*ast.Redir).N = yyDollar[1].word[0].(*ast.Lit)
		}
	case 87:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.node = &ast.Redir{
				OpPos: yyDollar[1].token.pos,
				Op:    yyDollar[1].token.val,
				Word:  yyDollar[2].word,
			}
		}
	case 95:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.node = &ast.Redir{
				OpPos: yyDollar[1].token.pos,
				Op:    yyDollar[1].token.val,
				Word:  yyDollar[2].word,
			}
			yylex.(*lexer).heredoc.push(yyVAL.node.(*ast.Redir))
		}
	case 99:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
		}
	case 101:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.token.pos = ast.Pos{}
		}
	}
	goto yystack /* stack new state and value */
}
