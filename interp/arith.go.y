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
%token<op>   '*' '/' '%'

%type<expr> primary_expr
%type<expr> postfix_expr unary_expr
%type<op>   unary_op
%type<expr> mul_expr
%type<expr> expr

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
			$$ = calculate(yylex, $1, $2, $3)
		}
	|	mul_expr '/' unary_expr
		{
			$$ = calculate(yylex, $1, $2, $3)
		}
	|	mul_expr '%' unary_expr
		{
			$$ = calculate(yylex, $1, $2, $3)
		}

expr:
		mul_expr

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
