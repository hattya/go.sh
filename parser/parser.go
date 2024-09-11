// Code generated by goyacc -l -o parser.go parser.go.y. DO NOT EDIT.
//
// go.sh/parser :: parser.go
//
//   Copyright (c) 2018-2024 Akinori Hattori <hattya@gmail.com>
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
	list  any
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
const FALLTHROUGH = 57351
const CLOBBER = 57352
const APPEND = 57353
const HEREDOC = 57354
const HEREDOCI = 57355
const DUPIN = 57356
const DUPOUT = 57357
const RDWR = 57358
const IO_NUMBER = 57359
const IO_LOCATION = 57360
const WORD = 57361
const NAME = 57362
const ASSIGNMENT_WORD = 57363
const Bang = 57364
const Lbrace = 57365
const Rbrace = 57366
const For = 57367
const Case = 57368
const Esac = 57369
const In = 57370
const If = 57371
const Elif = 57372
const Then = 57373
const Else = 57374
const Fi = 57375
const While = 57376
const Until = 57377
const Do = 57378
const Done = 57379

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
	"FALLTHROUGH",
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
	"IO_LOCATION",
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
		case "FALLTHROUGH":
			s = "';&'"
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

func location(n *ast.Lit) *ast.Lit {
	n.ValuePos = ast.NewPos(n.ValuePos.Line(), n.ValuePos.Col()+1)
	n.Value = n.Value[1 : len(n.Value)-1]
	return n
}

// ParseCommands parses src, including alias substitution, and returns
// commands.
func ParseCommands(env *interp.ExecEnv, name string, src any) ([]ast.Command, []*ast.Comment, error) {
	r, err := open(src)
	if err != nil {
		return nil, nil, err
	}

	l := newLexer(env, name, r)
	yyParse(l)
	return l.cmds, l.comments, l.err
}

// ParseCommand parses src and returns a command.
func ParseCommand(name string, src any) (ast.Command, []*ast.Comment, error) {
	cmds, comments, err := ParseCommands(nil, name, src)
	if len(cmds) == 0 {
		return nil, comments, err
	}
	return cmds[0], comments, err
}

func open(src any) (r io.RuneScanner, err error) {
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

var yyExca = [...]int8{
	-1, 1,
	1, -1,
	-2, 0,
	-1, 104,
	43, 102,
	-2, 105,
}

const yyPrivate = 57344

const yyLast = 336

var yyAct = [...]uint8{
	69, 68, 5, 49, 24, 147, 133, 146, 59, 70,
	101, 11, 50, 9, 52, 64, 60, 3, 62, 65,
	53, 54, 6, 51, 85, 170, 164, 139, 26, 71,
	28, 137, 136, 75, 76, 77, 40, 41, 42, 43,
	47, 48, 44, 45, 46, 35, 36, 14, 23, 25,
	8, 27, 51, 29, 30, 86, 125, 31, 103, 26,
	155, 28, 32, 33, 90, 140, 65, 114, 84, 92,
	142, 89, 97, 108, 102, 105, 103, 91, 87, 88,
	107, 151, 27, 104, 29, 30, 103, 163, 31, 51,
	134, 117, 135, 32, 33, 109, 92, 134, 130, 135,
	132, 106, 113, 115, 116, 100, 112, 51, 118, 119,
	120, 111, 124, 123, 150, 159, 126, 51, 131, 37,
	98, 122, 34, 73, 168, 129, 127, 60, 162, 83,
	82, 138, 74, 149, 72, 99, 152, 153, 129, 141,
	67, 156, 154, 90, 150, 53, 54, 173, 174, 161,
	158, 160, 157, 95, 94, 79, 81, 165, 78, 80,
	57, 166, 167, 149, 2, 169, 55, 56, 1, 97,
	110, 143, 176, 177, 178, 179, 175, 26, 39, 28,
	38, 96, 148, 145, 144, 40, 41, 42, 43, 47,
	48, 44, 45, 46, 35, 36, 14, 23, 25, 8,
	27, 7, 29, 30, 128, 121, 31, 22, 21, 20,
	58, 32, 33, 19, 26, 85, 28, 18, 171, 172,
	17, 16, 40, 41, 42, 43, 47, 48, 44, 45,
	46, 35, 36, 14, 23, 25, 8, 27, 15, 29,
	30, 13, 26, 31, 28, 10, 12, 4, 32, 33,
	40, 41, 42, 43, 47, 48, 44, 45, 46, 35,
	36, 14, 23, 25, 0, 27, 0, 29, 30, 0,
	0, 31, 0, 0, 0, 0, 32, 33, 40, 41,
	42, 43, 47, 48, 44, 45, 46, 35, 36, 61,
	0, 63, 40, 41, 42, 43, 47, 48, 44, 45,
	46, 35, 36, 93, 40, 41, 42, 43, 47, 48,
	44, 45, 46, 35, 36, 66, 40, 41, 42, 43,
	47, 48, 44, 45, 46, 35, 36, 40, 41, 42,
	43, 47, 48, 44, 45, 46,
}

var yyPact = [...]int16{
	21, -1000, -22, -1000, 132, 162, -1000, 154, 235, -1000,
	-1000, 301, -1000, 263, 289, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, 133, -1000, -1000, -22, -22, 108, 96,
	106, -22, -22, -22, -1000, 312, 312, -1000, 104, 103,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	170, -1000, 21, -1000, -1000, 21, 21, 235, 154, 301,
	-1000, 289, -1000, -1000, 277, -1000, -1000, 146, 145, 21,
	-21, 89, 125, 62, -22, 63, 37, 30, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, 162, -1000, -1000, -1000,
	-1000, 277, -1000, -1000, -22, -1000, 7, 162, -1000, -1000,
	-22, 24, 68, -22, -21, 56, -22, -22, -22, 52,
	21, -22, -21, 12, -22, 72, -1000, -22, 60, -12,
	-13, -1000, 301, 162, -1000, -1000, -17, 22, 44, -21,
	-1000, 137, -1000, 41, -22, -22, -1000, -1000, 301, -1000,
	-22, 17, -1000, -1000, 107, 81, -1000, -1000, 143, -1000,
	102, -1000, 49, -1000, -18, -22, -1000, -1000, -1000, -1000,
	-22, 98, -1000, -22, -1000, -19, 207, 136, -1000, 53,
	-1000, -22, -22, -22, -22, -1000, -1000, -1000, -1000, -1000,
}

var yyPgo = [...]uint8{
	0, 17, 247, 2, 22, 201, 13, 246, 245, 241,
	15, 11, 238, 221, 220, 217, 213, 209, 208, 207,
	205, 204, 184, 183, 7, 5, 182, 6, 1, 181,
	8, 4, 122, 119, 180, 178, 170, 10, 14, 168,
	164, 0, 9,
}

var yyR1 = [...]int8{
	0, 39, 39, 40, 40, 1, 1, 2, 2, 3,
	3, 3, 4, 4, 5, 5, 6, 6, 6, 6,
	8, 8, 8, 8, 8, 9, 9, 9, 9, 10,
	10, 10, 10, 11, 11, 11, 11, 11, 11, 11,
	11, 12, 13, 14, 15, 15, 15, 15, 21, 21,
	16, 16, 16, 22, 22, 24, 24, 24, 24, 23,
	23, 25, 25, 26, 26, 26, 17, 17, 27, 27,
	27, 18, 19, 7, 20, 20, 28, 28, 29, 29,
	30, 30, 31, 31, 31, 31, 31, 31, 32, 34,
	34, 34, 34, 34, 34, 34, 33, 35, 35, 36,
	36, 37, 37, 38, 38, 41, 41, 42, 42,
}

var yyR2 = [...]int8{
	0, 2, 0, 1, 3, 2, 1, 1, 3, 1,
	3, 3, 1, 2, 1, 3, 1, 1, 2, 1,
	3, 2, 1, 2, 1, 1, 2, 1, 2, 1,
	2, 1, 2, 1, 1, 1, 1, 1, 1, 1,
	1, 3, 3, 3, 5, 6, 8, 9, 1, 2,
	6, 7, 7, 1, 2, 5, 5, 5, 5, 1,
	2, 3, 3, 1, 2, 3, 5, 6, 4, 5,
	2, 5, 5, 5, 1, 2, 3, 2, 1, 3,
	1, 2, 1, 2, 2, 1, 2, 2, 2, 1,
	1, 1, 1, 1, 1, 1, 2, 1, 1, 2,
	1, 2, 1, 1, 1, 1, 0, 1, 2,
}

var yyChk = [...]int16{
	-1000, -39, -40, -1, -2, -3, -4, -5, 29, -6,
	-8, -11, -7, -9, 26, -12, -13, -14, -15, -16,
	-17, -18, -19, 27, -31, 28, 7, 30, 9, 32,
	33, 36, 41, 42, -32, 24, 25, -33, -34, -35,
	15, 16, 17, 18, 21, 22, 23, 19, 20, -41,
	-42, 45, -38, 13, 14, 4, 5, 6, -5, -30,
	-31, 26, -31, 28, -10, -31, 26, 7, -28, -41,
	-42, -28, 26, 27, 26, -28, -28, -28, -32, -33,
	-32, -33, 26, 26, -1, 45, -3, -4, -4, -6,
	-31, -10, -31, 26, 8, 8, -29, -3, 31, 10,
	43, -37, -41, 14, -42, -41, 38, 43, 43, -41,
	-36, -38, -42, -28, 43, 35, -41, 35, -28, -28,
	-28, -20, -11, -3, -41, 44, -28, -37, -21, -42,
	26, -41, 40, -27, 37, 39, 44, 44, -30, 44,
	43, -37, 26, 34, -22, -23, -24, -25, -26, 26,
	7, 40, -28, -28, -28, 43, 34, -24, -25, 34,
	8, 6, 26, 38, 44, -28, -41, -28, 26, -28,
	44, 11, 12, 11, 12, -27, -41, -41, -41, -41,
}

var yyDef = [...]int8{
	2, -2, 106, 3, 6, 7, 9, 12, 0, 14,
	16, 17, 19, 22, 24, 33, 34, 35, 36, 37,
	38, 39, 40, 0, 25, 27, 106, 106, 0, 0,
	0, 106, 106, 106, 82, 0, 0, 85, 0, 0,
	89, 90, 91, 92, 93, 94, 95, 97, 98, 1,
	105, 107, 5, 103, 104, 0, 0, 0, 13, 18,
	80, 21, 26, 28, 23, 29, 31, 0, 0, 0,
	105, 0, 0, 106, 106, 0, 0, 0, 83, 86,
	84, 87, 88, 96, 4, 108, 8, 10, 11, 15,
	81, 20, 30, 32, 106, 41, 77, 78, 42, 43,
	106, 0, 0, 106, -2, 0, 106, 106, 106, 0,
	76, 106, 100, 0, 106, 0, 101, 106, 0, 0,
	0, 73, 74, 79, 99, 44, 0, 0, 0, 102,
	48, 0, 66, 0, 106, 106, 71, 72, 75, 45,
	106, 0, 49, 50, 0, 0, 53, 59, 0, 63,
	0, 67, 0, 70, 0, 106, 51, 54, 60, 52,
	106, 0, 64, 106, 46, 0, 61, 62, 65, 68,
	47, 106, 106, 106, 106, 69, 55, 57, 56, 58,
}

var yyTok1 = [...]int8{
	1, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	45, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 13, 3,
	7, 8, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 14,
	15, 3, 16, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 6,
}

var yyTok2 = [...]int8{
	2, 3, 4, 5, 9, 10, 11, 12, 17, 18,
	19, 20, 21, 22, 23, 24, 25, 26, 27, 28,
	29, 30, 31, 32, 33, 34, 35, 36, 37, 38,
	39, 40, 41, 42, 43, 44,
}

var yyTok3 = [...]int8{
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
	base := int(yyPact[state])
	for tok := TOKSTART; tok-1 < len(yyToknames); tok++ {
		if n := base + tok; n >= 0 && n < yyLast && int(yyChk[int(yyAct[n])]) == tok {
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}
	}

	if yyDef[state] == -2 {
		i := 0
		for yyExca[i] != -1 || int(yyExca[i+1]) != state {
			i += 2
		}

		// Look for tokens that we accept or reduce.
		for i += 2; yyExca[i] >= 0; i += 2 {
			tok := int(yyExca[i])
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
		token = int(yyTok1[0])
		goto out
	}
	if char < len(yyTok1) {
		token = int(yyTok1[char])
		goto out
	}
	if char >= yyPrivate {
		if char < yyPrivate+len(yyTok2) {
			token = int(yyTok2[char-yyPrivate])
			goto out
		}
	}
	for i := 0; i < len(yyTok3); i += 2 {
		token = int(yyTok3[i+0])
		if token == char {
			token = int(yyTok3[i+1])
			goto out
		}
	}

out:
	if token == 0 {
		token = int(yyTok2[1]) /* unknown char */
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
	yyn = int(yyPact[yystate])
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
	yyn = int(yyAct[yyn])
	if int(yyChk[yyn]) == yytoken { /* valid shift */
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
	yyn = int(yyDef[yystate])
	if yyn == -2 {
		if yyrcvr.char < 0 {
			yyrcvr.char, yytoken = yylex1(yylex, &yyrcvr.lval)
		}

		/* look through exception table */
		xi := 0
		for {
			if yyExca[xi+0] == -1 && int(yyExca[xi+1]) == yystate {
				break
			}
			xi += 2
		}
		for xi += 2; ; xi += 2 {
			yyn = int(yyExca[xi+0])
			if yyn < 0 || yyn == yytoken {
				break
			}
		}
		yyn = int(yyExca[xi+1])
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
				yyn = int(yyPact[yyS[yyp].yys]) + yyErrCode
				if yyn >= 0 && yyn < yyLast {
					yystate = int(yyAct[yyn]) /* simulate a shift of "error" */
					if int(yyChk[yystate]) == yyErrCode {
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

	yyp -= int(yyR2[yyn])
	// yyp is now the index of $0. Perform the default action. Iff the
	// reduced production is ε, $1 is possibly out of range.
	if yyp+1 >= len(yyS) {
		nyys := make([]yySymType, len(yyS)*2)
		copy(nyys, yyS)
		yyS = nyys
	}
	yyVAL = yyS[yyp+1]

	/* consult goto table to find next state */
	yyn = int(yyR1[yyn])
	yyg := int(yyPgo[yyn])
	yyj := yyg + yyS[yyp].yys + 1

	if yyj >= yyLast {
		yystate = int(yyAct[yyg])
	} else {
		yystate = int(yyAct[yyj])
		if int(yyChk[yystate]) != -yyn {
			yystate = int(yyAct[yyg])
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
			ci := yyDollar[1].node.(*ast.CaseItem)
			ci.Rparen = yyDollar[2].token.pos
			ci.Break = yyDollar[4].token.pos
			yyVAL.node = yyDollar[1].node
		}
	case 56:
		yyDollar = yyS[yypt-5 : yypt+1]
		{
			ci := yyDollar[1].node.(*ast.CaseItem)
			ci.Rparen = yyDollar[2].token.pos
			ci.List = yyDollar[3].list.([]ast.Command)
			ci.Break = yyDollar[4].token.pos
			yyVAL.node = yyDollar[1].node
		}
	case 57:
		yyDollar = yyS[yypt-5 : yypt+1]
		{
			ci := yyDollar[1].node.(*ast.CaseItem)
			ci.Rparen = yyDollar[2].token.pos
			ci.Fallthrough = yyDollar[4].token.pos
			yyVAL.node = yyDollar[1].node
		}
	case 58:
		yyDollar = yyS[yypt-5 : yypt+1]
		{
			ci := yyDollar[1].node.(*ast.CaseItem)
			ci.Rparen = yyDollar[2].token.pos
			ci.List = yyDollar[3].list.([]ast.Command)
			ci.Fallthrough = yyDollar[4].token.pos
			yyVAL.node = yyDollar[1].node
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
			yyDollar[1].node.(*ast.CaseItem).Rparen = yyDollar[2].token.pos
			yyVAL.node = yyDollar[1].node
		}
	case 62:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			ci := yyDollar[1].node.(*ast.CaseItem)
			ci.Rparen = yyDollar[2].token.pos
			ci.List = yyDollar[3].list.([]ast.Command)
			yyVAL.node = yyDollar[1].node
		}
	case 63:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.node = &ast.CaseItem{Patterns: []ast.Word{yyDollar[1].word}}
		}
	case 64:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.node = &ast.CaseItem{
				Lparen:   yyDollar[1].token.pos,
				Patterns: []ast.Word{yyDollar[2].word},
			}
		}
	case 65:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.node.(*ast.CaseItem).Patterns = append(yyVAL.node.(*ast.CaseItem).Patterns, yyDollar[3].word)
		}
	case 66:
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
	case 67:
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
	case 68:
		yyDollar = yyS[yypt-4 : yypt+1]
		{
			yyVAL.list = []ast.ElsePart{&ast.ElifClause{
				Elif: yyDollar[1].token.pos,
				Cond: yyDollar[2].list.([]ast.Command),
				Then: yyDollar[3].token.pos,
				List: yyDollar[4].list.([]ast.Command),
			}}
		}
	case 69:
		yyDollar = yyS[yypt-5 : yypt+1]
		{
			yyVAL.list = append([]ast.ElsePart{&ast.ElifClause{
				Elif: yyDollar[1].token.pos,
				Cond: yyDollar[2].list.([]ast.Command),
				Then: yyDollar[3].token.pos,
				List: yyDollar[4].list.([]ast.Command),
			}}, yyDollar[5].list.([]ast.ElsePart)...)
		}
	case 70:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.list = []ast.ElsePart{&ast.ElseClause{
				Else: yyDollar[1].token.pos,
				List: yyDollar[2].list.([]ast.Command),
			}}
		}
	case 71:
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
	case 72:
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
	case 73:
		yyDollar = yyS[yypt-5 : yypt+1]
		{
			x := yyDollar[5].node.(*ast.FuncDef)
			x.Name = yyDollar[1].word[0].(*ast.Lit)
			x.Lparen = yyDollar[2].token.pos
			x.Rparen = yyDollar[3].token.pos
			yyVAL.node = &ast.Cmd{Expr: yyDollar[5].node.(ast.CmdExpr)}
		}
	case 74:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.node = &ast.FuncDef{
				Body: &ast.Cmd{
					Expr: yyDollar[1].node.(ast.CmdExpr),
				},
			}
		}
	case 75:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.node = &ast.FuncDef{
				Body: &ast.Cmd{
					Expr:   yyDollar[1].node.(ast.CmdExpr),
					Redirs: yyDollar[2].list.([]*ast.Redir),
				},
			}
		}
	case 76:
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
	case 77:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			cmds := yyDollar[2].list.([]ast.Command)
			l := cmds[len(cmds)-1].(ast.List)
			if len(l) == 1 {
				cmds[len(cmds)-1] = extract(l[0])
			}
			yyVAL.list = yyDollar[2].list
		}
	case 78:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.list = []ast.Command{ast.List{yyDollar[1].node.(*ast.AndOrList)}}
		}
	case 79:
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
	case 80:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.list = []*ast.Redir{yyDollar[1].node.(*ast.Redir)}
		}
	case 81:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.list = append(yyVAL.list.([]*ast.Redir), yyDollar[2].node.(*ast.Redir))
		}
	case 83:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.node = yyDollar[2].node
			yyVAL.node.(*ast.Redir).N = yyDollar[1].word[0].(*ast.Lit)
		}
	case 84:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.node = yyDollar[2].node
			yyVAL.node.(*ast.Redir).N = location(yyDollar[1].word[0].(*ast.Lit))
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
			yyVAL.node = yyDollar[2].node
			yyVAL.node.(*ast.Redir).N = location(yyDollar[1].word[0].(*ast.Lit))
		}
	case 88:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.node = &ast.Redir{
				OpPos: yyDollar[1].token.pos,
				Op:    yyDollar[1].token.val,
				Word:  yyDollar[2].word,
			}
		}
	case 96:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.node = &ast.Redir{
				OpPos: yyDollar[1].token.pos,
				Op:    yyDollar[1].token.val,
				Word:  yyDollar[2].word,
			}
			yylex.(*lexer).heredoc.push(yyVAL.node.(*ast.Redir))
		}
	case 100:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
		}
	case 102:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.token.pos = ast.Pos{}
		}
	}
	goto yystack /* stack new state and value */
}
