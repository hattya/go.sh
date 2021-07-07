%{
//
// go.sh/interp :: arith.go
//
//   Copyright (c) 2021 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
//

package interp

import (
	"fmt"
	"strconv"
	"strings"
)
%}

%union {
	op   string
	expr expr
}

%token<expr> NUMBER IDENT
%token<op>   '(' ')'
%token<op>   INC DEC '+' '-' '~' '!'
%token<op>   '*' '/' '%' LSH RSH '<' '>' LE GE EQ NE '&' '^' '|' LAND LOR
%token<op>   '?' ':'
%token<op>   '=' MUL_ASSIGN DIV_ASSIGN MOD_ASSIGN ADD_ASSIGN SUB_ASSIGN LSH_ASSIGN RSH_ASSIGN AND_ASSIGN XOR_ASSIGN OR_ASSIGN

%type<expr> primary_expr
%type<expr> postfix_expr unary_expr
%type<op>   unary_op
%type<expr> mul_expr add_expr shift_expr rel_expr eq_expr and_expr xor_expr or_expr land_expr lor_expr
%type<expr> cond_expr
%type<expr> expr
%type<op>   assign_op

%right '=' MUL_ASSIGN DIV_ASSIGN MOD_ASSIGN ADD_ASSIGN SUB_ASSIGN LSH_ASSIGN RSH_ASSIGN AND_ASSIGN XOR_ASSIGN OR_ASSIGN
%right '?'
%left  LOR
%left  LAND
%left  '|'
%left  '^'
%left  '&'
%left  EQ  NE
%left  '<' '>' LE  GE
%left  LSH RSH
%left  '+' '-'
%left  '*' '/' '%'
%right INC DEC

%%

arith:
		expr
		{
			if n, ok := expand(yylex, $1); ok {
				yylex.(*lexer).n = n
			}
		}

primary_expr:
		NUMBER
		{
			$$.s = ""
			if n, err := strconv.ParseInt($1.s, 0, 0); err != nil {
				yylex.Error(fmt.Sprintf("invalid number %q", $1.s))
			} else {
				$$.n = int(n)
			}
		}
	|	IDENT
	|	'(' expr ')'
		{
			$$ = $2
		}

postfix_expr:
		primary_expr
	|	postfix_expr INC
		{
			$$.s = ""
			if $1.s == "" {
				yylex.Error(errLValue($2))
			} else if n, ok := expand(yylex, $1); ok {
				$$.n = n
				yylex.(*lexer).env.Set($1.s, strconv.Itoa($$.n + 1))
			}
		}
	|	postfix_expr DEC
		{
			$$.s = ""
			if $1.s == "" {
				yylex.Error(errLValue($2))
			} else if n, ok := expand(yylex, $1); ok {
				$$.n = n
				yylex.(*lexer).env.Set($1.s, strconv.Itoa($$.n - 1))
			}
		}

unary_expr:
		         postfix_expr
	|	INC      unary_expr
		{
			$$.s = ""
			if $2.s == "" {
				yylex.Error(errLValue($1))
			} else if n, ok := expand(yylex, $2); ok {
				$$.n = n + 1
				yylex.(*lexer).env.Set($2.s, strconv.Itoa($$.n))
			}
		}
	|	DEC      unary_expr
		{
			$$.s = ""
			if $2.s == "" {
				yylex.Error(errLValue($1))
			} else if n, ok := expand(yylex, $2); ok {
				$$.n = n - 1
				yylex.(*lexer).env.Set($2.s, strconv.Itoa($$.n))
			}
		}
	|	unary_op unary_expr
		{
			$$.s = ""
			if n, ok := expand(yylex, $2); ok {
				switch $1 {
				case "+":
					$$.n = +n
				case "-":
					$$.n = -n
				case "~":
					$$.n = ^n
				case "!":
					if n == 0 {
						$$.n = 1
					} else {
						$$.n = 0
					}
				}
			}
		}

unary_op:
		'+'
	|	'-'
	|	'~'
	|	'!'

mul_expr:
		             unary_expr
	|	mul_expr '*' unary_expr
		{
			$$, _ = calculate(yylex, $1, $2, $3)
		}
	|	mul_expr '/' unary_expr
		{
			$$, _ = calculate(yylex, $1, $2, $3)
		}
	|	mul_expr '%' unary_expr
		{
			$$, _ = calculate(yylex, $1, $2, $3)
		}

add_expr:
		             mul_expr
	|	add_expr '+' mul_expr
		{
			$$, _ = calculate(yylex, $1, $2, $3)
		}
	|	add_expr '-' mul_expr
		{
			$$, _ = calculate(yylex, $1, $2, $3)
		}

shift_expr:
		               add_expr
	|	shift_expr LSH add_expr
		{
			$$, _ = calculate(yylex, $1, $2, $3)
		}
	|	shift_expr RSH add_expr
		{
			$$, _ = calculate(yylex, $1, $2, $3)
		}

rel_expr:
		             shift_expr
	|	rel_expr '<' shift_expr
		{
			$$ = compare(yylex, $1, $2, $3)
		}
	|	rel_expr '>' shift_expr
		{
			$$ = compare(yylex, $1, $2, $3)
		}
	|	rel_expr LE  shift_expr
		{
			$$ = compare(yylex, $1, $2, $3)
		}
	|	rel_expr GE  shift_expr
		{
			$$ = compare(yylex, $1, $2, $3)
		}

eq_expr:
		           rel_expr
	|	eq_expr EQ rel_expr
		{
			$$ = compare(yylex, $1, $2, $3)
		}
	|	eq_expr NE rel_expr
		{
			$$ = compare(yylex, $1, $2, $3)
		}

and_expr:
		             eq_expr
	|	and_expr '&' eq_expr
		{
			$$, _ = calculate(yylex, $1, $2, $3)
		}

xor_expr:
		             and_expr
	|	xor_expr '^' and_expr
		{
			$$, _ = calculate(yylex, $1, $2, $3)
		}

or_expr:
		            xor_expr
	|	or_expr '|' xor_expr
		{
			$$, _ = calculate(yylex, $1, $2, $3)
		}

land_expr:
		               or_expr
	|	land_expr LAND or_expr
		{
			$$.n = 0
			$$.s = ""
			if l, ok := expand(yylex, $1); ok && l != 0 {
				if r, ok := expand(yylex, $3); ok && r != 0 {
					$$.n = 1
				}
			}
		}

lor_expr:
		             land_expr
	|	lor_expr LOR land_expr
		{
			$$.n = 0
			$$.s = ""
			if l, ok := expand(yylex, $1); ok && l != 0 {
				$$.n = 1
			} else if r, ok := expand(yylex, $3); ok && r != 0 {
				$$.n = 1
			}
		}

cond_expr:
		lor_expr
	|	lor_expr '?' expr ':' cond_expr
		{
			$$.s = ""
			if l, ok := expand(yylex, $1); ok {
				if l != 0 {
					$$.n, _ = expand(yylex, $3)
				} else {
					$$.n, _ = expand(yylex, $5)
				}
			}
		}

expr:
		cond_expr
	|	unary_expr assign_op expr
		{
			$$.s = ""
			if $1.s == "" {
				yylex.Error(errLValue($2))
			} else {
				var ok bool
				if $2 == "=" {
					$$.n, ok = expand(yylex, $3)
				} else {
					$$, ok = calculate(yylex, $1, $2[:len($2)-1], $3)
				}
				if ok {
					yylex.(*lexer).env.Set($1.s, strconv.Itoa($$.n))
				}
			}
		}

assign_op:
		'='
	|	MUL_ASSIGN
	|	DIV_ASSIGN
	|	MOD_ASSIGN
	|	ADD_ASSIGN
	|	SUB_ASSIGN
	|	LSH_ASSIGN
	|	RSH_ASSIGN
	|	AND_ASSIGN
	|	XOR_ASSIGN
	|	 OR_ASSIGN

%%

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
		case "LAND":
			s = "'&&'"
		case "LOR":
			s = "'||'"
		case "MUL_ASSIGN":
			s = "'*='"
		case "DIV_ASSIGN":
			s = "'/='"
		case "MOD_ASSIGN":
			s = "'%='"
		case "ADD_ASSIGN":
			s = "'+='"
		case "SUB_ASSIGN":
			s = "'-='"
		case "LSH_ASSIGN":
			s = "'<<='"
		case "RSH_ASSIGN":
			s = "'>>='"
		case "AND_ASSIGN":
			s = "'&='"
		case "XOR_ASSIGN":
			s = "'^='"
		case "OR_ASSIGN":
			s = "'|='"
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

func calculate(yylex yyLexer, l expr, op string, r expr) (x expr, ok bool) {
	if l, ok1 := expand(yylex, l); ok1 {
		if r, ok2 := expand(yylex, r); ok2 {
			ok = true
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
			case "&":
				x.n = l & r
			case "^":
				x.n = l ^ r
			case "|":
				x.n = l | r
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
