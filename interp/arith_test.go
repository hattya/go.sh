//
// go.sh/interp :: arith_test.go
//
//   Copyright (c) 2021-2024 Akinori Hattori <hattya@gmail.com>
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
	// mul
	{" 0 *  0 *  0 ", 0},
	{" 2 *  4 *  8 ", 64},
	{"+2 *  4 *  8 ", 64},
	{" 2 * -4 *  8 ", -64},
	{" 2 *  4 * +8 ", 64},
	{"-2 * +4 * -8 ", 64},
	{"-2 * (4 *  8)", -64},
	// div
	{" 64 /  8 /  4 ", 2},
	{"+64 /  8 /  4 ", 2},
	{" 64 / -8 /  4 ", -2},
	{" 64 /  8 / +4 ", 2},
	{"-64 / +8 / -4 ", 2},
	{"-64 / (8 /  4)", -32},
	// mod
	{" 140 %  71 %  37 ", 32},
	{"+140 %  71 %  37 ", 32},
	{" 140 % -71 %  37 ", 32},
	{" 140 %  71 % +37 ", 32},
	{"-140 % +71 % -37 ", -32},
	{"-140 % (71 %  37)", -4},
	// add
	{" 0 +  0  + 0 ", 0},
	{" 8 +  4  + 2 ", 14},
	{" 8 *  4  + 2 ", 34},
	{" 8 * (4  + 2)", 48},
	{" 8 +  4  / 2 ", 10},
	{"(8 +  4) / 2 ", 6},
	{" 8 %  4  + 2 ", 2},
	{" 8 % (4  + 2)", 2},
	// sub
	{" 0 -  0  - 0", 0},
	{" 8 -  4  - 2 ", 2},
	{" 8 *  4  - 2 ", 30},
	{" 8 * (4  - 2)", 16},
	{" 8 -  4  / 2 ", 6},
	{"(8 -  4) / 2 ", 2},
	{" 8 %  4  - 2 ", -2},
	{" 8 % (4  - 2)", 0},
	// lsh
	{" 4 <<  2  << 1 ", 32},
	{" 4 +   2  << 1 ", 12},
	{" 4 +  (2  << 1)", 8},
	{" 4 <<  2  -  1 ", 8},
	{"(4 <<  2) -  1 ", 15},
	// rsh
	{" 8 >>  2  >> 1 ", 1},
	{" 8 +   2  >> 1 ", 5},
	{" 8 +  (2  >> 1)", 9},
	{" 8 >>  2  -  1 ", 4},
	{"(8 >>  2) -  1 ", 1},
	// lt
	{" 1 <   2  ", 1},
	{" 2 <   2  ", 0},
	{" 2 <   1  ", 0},
	{" 1 <<  2  <  4 ", 0},
	{" 1 << (2  <  4)", 2},
	{" 1 <   4  >> 1 ", 1},
	{"(1 <   4) >> 1 ", 0},
	// le
	{" 1 <=  2  ", 1},
	{" 2 <=  2  ", 1},
	{" 2 <=  1  ", 0},
	{" 1 <<  2  <= 4 ", 1},
	{" 1 << (2  <= 4)", 2},
	{" 1 <=  4  >> 1 ", 1},
	{"(1 <=  4) >> 1 ", 0},
	// gt
	{" 1 >   2  ", 0},
	{" 2 >   2  ", 0},
	{" 2 >   1  ", 1},
	{" 4 <<  2  >  2 ", 1},
	{" 4 << (2  >  2)", 4},
	{" 4 >   2  >> 1 ", 1},
	{"(4 >   2) >> 1 ", 0},
	// ge
	{" 1 >=  2  ", 0},
	{" 2 >=  2  ", 1},
	{" 2 >=  1  ", 1},
	{" 4 <<  2  >= 2 ", 1},
	{" 4 << (2  >= 2)", 8},
	{" 4 >=  2  >> 1 ", 1},
	{"(4 >=  2) >> 1 ", 0},
	// eq
	{" 1 ==  1  ", 1},
	{" 1 ==  0  ", 0},
	{" 1 <   1  == 0 ", 1},
	{" 1 <  (1  == 0)", 0},
	{" 0 ==  1  <= 1 ", 0},
	{"(0 ==  1) <= 1 ", 1},
	{" 0 >   1  == 0 ", 1},
	{" 0 >  (1  == 0)", 0},
	{" 0 ==  1  >= 0 ", 0},
	{"(0 ==  1) >= 0 ", 1},
	// ne
	{" 1 !=  1  ", 0},
	{" 1 !=  2  ", 1},
	{" 1 <   0  != 1 ", 1},
	{" 1 <  (0  != 1)", 0},
	{" 1 !=  0  <= 1 ", 0},
	{"(1 !=  0) <= 1 ", 1},
	{" 0 >   1  != 1 ", 1},
	{" 0 >  (1  != 1)", 0},
	{" 1 !=  1  >= 0 ", 0},
	{"(1 !=  1) >= 0 ", 1},
	// and
	{" 0 &   0  ", 0},
	{" 0 &   1  ", 0},
	{" 1 &   0  ", 0},
	{" 1 &   1  ", 1},
	{" 0 ==  1  &  0 ", 0},
	{" 0 == (1  &  0)", 1},
	{" 0 &   1  != 1 ", 0},
	{"(0 &   1) != 1 ", 1},
	// xor
	{" 0 ^  0 ", 0},
	{" 0 ^  1 ", 1},
	{" 1 ^  0 ", 1},
	{" 1 ^  1 ", 0},
	{" 0 &  0  ^ 1 ", 1},
	{" 0 & (0  ^ 1)", 0},
	{" 1 ^  0  & 0 ", 1},
	{"(1 ^  0) & 0 ", 0},
	// or
	{" 0 |  0 ", 0},
	{" 0 |  1 ", 1},
	{" 1 |  0 ", 1},
	{" 1 |  1 ", 1},
	{" 1 ^  0  | 1 ", 1},
	{" 1 ^ (0  | 1)", 0},
	{" 1 |  0  ^ 1 ", 1},
	{"(1 |  0) ^ 1 ", 0},
	// logical and
	{" 0 &&  0  && 0 ", 0},
	{" 0 &&  0  && 1 ", 0},
	{" 0 &&  1  && 0 ", 0},
	{" 1 &&  0  && 0 ", 0},
	{" 1 &&  1  && 1 ", 1},
	{" 1 |   1  && 0 ", 0},
	{" 1 |  (1  && 0)", 1},
	{" 0 &&  1  |  1 ", 0},
	{"(0 &&  1) |  1 ", 1},
	// logical or
	{" 0 ||  0  || 0 ", 0},
	{" 0 ||  0  || 1 ", 1},
	{" 0 ||  1  || 0 ", 1},
	{" 1 ||  0  || 0 ", 1},
	{" 1 ||  1  || 1 ", 1},
	{" 0 &&  0  || 1 ", 1},
	{" 0 && (0  || 1)", 0},
	{" 1 ||  0  && 0 ", 1},
	{"(1 ||  0) && 0 ", 0},
	// conditional
	{"-1 == -1 ? -1 : 0 == 0 ? 0 : 1", -1},
	{" 0 == -1 ? -1 : 0 == 0 ? 0 : 1", 0},
	{" 1 == -1 ? -1 : 1 == 0 ? 0 : 1", 1},
	// assignment
	{"X   = 2", 2},
	{"X  *= 2", 14},
	{"X  /= 2", 3},
	{"X  %= 2", 1},
	{"X  += 2", 9},
	{"X  -= 2", 5},
	{"X <<= 2", 28},
	{"X >>= 2", 1},
	{"X  &= 2", 2},
	{"X  ^= 2", 5},
	{"X  |= 2", 7},
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
	{"<<", "unexpected '<<'"},
	{">>", "unexpected '>>'"},
	{"<", "unexpected '<'"},
	{">", "unexpected '>'"},
	{"<=", "unexpected '<='"},
	{">=", "unexpected '>='"},
	{"==", "unexpected '=='"},
	{"!=", "unexpected '!='"},
	{"&&", "unexpected '&&'"},
	{"||", "unexpected '||'"},
	{"=", "unexpected '='"},
	{"*=", "unexpected '*='"},
	{"/=", "unexpected '/='"},
	{"%=", "unexpected '%='"},
	{"+=", "unexpected '+='"},
	{"-=", "unexpected '-='"},
	{"<<=", "unexpected '<<='"},
	{">>=", "unexpected '>>='"},
	{"&=", "unexpected '&='"},
	{"^=", "unexpected '^='"},
	{"|=", "unexpected '|='"},
	{"0   = 1", "'=' requires lvalue"},
	{"0  *= 1", "'*=' requires lvalue"},
	{"0  /= 1", "'/=' requires lvalue"},
	{"0  %= 1", "'%=' requires lvalue"},
	{"0  += 1", "'+=' requires lvalue"},
	{"0  -= 1", "'-=' requires lvalue"},
	{"0 <<= 1", "'<<=' requires lvalue"},
	{"0 >>= 1", "'>>=' requires lvalue"},
	{"0  &= 1", "'&=' requires lvalue"},
	{"0  ^= 1", "'^=' requires lvalue"},
	{"0  |= 1", "'|=' requires lvalue"},
	// divide by zero
	{"0 /  0", "integer divide by zero"},
	{"0 %  0", "integer divide by zero"},
	{"M /= 0", "integer divide by zero"},
	{"M %= 0", "integer divide by zero"},
	// negative shift
	{"1 <<  -1", "negative shift amount"},
	{"1 >>  -1", "negative shift amount"},
	{"N <<= -1", "negative shift amount"},
	{"N >>= -1", "negative shift amount"},
}

func TestEvalError(t *testing.T) {
	env := interp.NewExecEnv(name)
	env.Set("A", "alpha")
	env.Set("M", "0")
	env.Set("N", "1")
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
