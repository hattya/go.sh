//
// go.sh/interp :: arith_test.go
//
//   Copyright (c) 2021 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
//

package interp_test

import (
	"testing"

	"github.com/hattya/go.sh/interp"
)

var evalTests = []struct {
	expr string
	n    int
}{
	{" - 1 ", -1},
	{"   0 ", 0},
	{" + 1 ", 1},
	// dec
	{"9", 9},
	{"+9", 9},
	{"-9", -9},
	{"~0", -1},
	{"!0", 1},
	{"!9", 0},
	{"!+9", 0},
	{"!-9", 0},
	{"(9)", 9},
	{"((9))", 9},
	{"+(-(+9))", -9},
	{"-(+(-9))", 9},
	// oct
	{"07", 7},
	{"+07", 7},
	{"-07", -7},
	// hex
	{"0xf", 15},
	{"0Xf", 15},
	{"+0xf", 15},
	{"+0Xf", 15},
	{"-0xf", -15},
	{"-0Xf", -15},
	// ident
	{"_", 0},
	{"E", 0},
	{"D", 4},
	{"O", 3},
	{"X", 7},
	{"+X", 7},
	{"-X", -7},
	{"~X", -8},
	{"!X", 0},
	{"((X))", 7},
	{"+(-(+X))", -7},
	{"-(+(-X))", 7},
	// inc
	{"X++", 7},
	{"+X++", 7},
	{"-X++", -7},
	{"++X", 8},
	{"+(++X)", 8},
	{"-(++X)", -8},
	// dec
	{"X--", 7},
	{"+X--", 7},
	{"-X--", -7},
	{"--X", 6},
	{"+(--X)", 6},
	{"-(--X)", -6},
}

func TestEval(t *testing.T) {
	env := interp.NewExecEnv(name)
	env.Unset("_")
	for _, tt := range evalTests {
		env.Set("E", "")
		env.Set("D", "4")
		env.Set("O", "03")
		env.Set("X", "0x7")
		switch g, err := env.Eval(tt.expr); {
		case err != nil:
			t.Error("unexpected error:", err)
		case g != tt.n:
			t.Errorf("expected %v, got %v", tt.n, g)
		}
	}
}

var evalErrorTests = []struct {
	expr string
	err  string
}{
	// empty
	{"", "unexpected EOF"},
	// number
	{"09", `invalid number "09"`},
	{"0xz", `invalid number "0xz"`},
	{"0 1 2", "unexpected NUMBER"},
	// ident
	{"A", `invalid number "alpha"`},
	{"Z", `invalid number "0z777"`},
	{"M N", "unexpected IDENT"},
	// parenthesis
	{"(", "unexpected EOF"},
	{")", "unexpected ')'"},
	// op
	{"$", "unexpected '$'"},
	{"++_--", "'++' requires lvalue"},
	{"+++_", "'++' requires lvalue"},
	{"++0--", "'++' requires lvalue"},
	{"0++", "'++' requires lvalue"},
	{"++0", "'++' requires lvalue"},
	{"+++0", "'++' requires lvalue"},
	{"--_++", "'--' requires lvalue"},
	{"---_", "'--' requires lvalue"},
	{"--0++", "'--' requires lvalue"},
	{"0--", "'--' requires lvalue"},
	{"--0", "'--' requires lvalue"},
	{"---0", "'--' requires lvalue"},
}

func TestEvalError(t *testing.T) {
	env := interp.NewExecEnv(name)
	env.Set("A", "alpha")
	env.Set("Z", "0z777")
	env.Unset("_")
	for _, tt := range evalErrorTests {
		switch _, err := env.Eval(tt.expr); {
		case err == nil:
			t.Error("expected error")
		case err.Error() != tt.err:
			t.Error("unexpected error:", err)
		}
	}
}
