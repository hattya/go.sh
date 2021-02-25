//
// go.sh/interp :: expand_test.go
//
//   Copyright (c) 2021 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
//

package interp_test

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"testing"

	"github.com/hattya/go.sh/ast"
	"github.com/hattya/go.sh/interp"
)

const (
	sep = string(os.PathListSeparator)
	V   = "value"
	E   = ""
)

var expandTests = []struct {
	word ast.Word
	s    string
}{
	{word(), ""},
	{word(lit("foo")), "foo"},
	{word(lit("foo"), lit("bar")), "foobar"},
}

var tildeExpTests = []struct {
	word   ast.Word
	assign bool
	s      string
}{
	{word(lit("~")), false, homeDir()},
	{word(lit("~/")), false, homeDir() + "/"},
	{word(lit("~"), lit("/")), false, homeDir() + "/"},

	{word(lit("~" + username())), false, homeDir()},
	{word(lit("~" + username() + "/")), false, homeDir() + "/"},
	{word(lit("~"), lit(username()), lit("/")), false, homeDir() + "/"},

	{word(lit("~_")), false, "~_"},
	{word(lit("~_/")), false, "~_/"},
	{word(lit("~"), lit("_"), lit("/")), false, "~_/"},

	{word(lit(sep)), true, sep},
	{word(litf("~/foo%v~/bar", sep)), true, fmt.Sprintf("%v/foo%v%[1]v/bar", homeDir(), sep)},
	{word(lit("~/foo"), lit(sep), lit("~/bar")), true, fmt.Sprintf("%v/foo%v%[1]v/bar", homeDir(), sep)},
	{word(lit("~"), lit("/foo"), lit(sep), lit("~"), lit("/bar")), true, fmt.Sprintf("%v/foo%v%[1]v/bar", homeDir(), sep)},

	{word(lit("~"), paramExp(lit("_"), "", nil), lit("/")), false, "~/"},

	{word(paramExp(lit("_"), ":-", word(litf("~/foo%v~/bar", sep)))), true, fmt.Sprintf("%v/foo%v%[1]v/bar", homeDir(), sep)},
}

var paramExpTests = []struct {
	word ast.Word
	s    string
}{
	// simplest form
	{word(paramExp(lit("V"), "", nil)), V},
	// use default values
	{word(paramExp(lit("V"), ":-", word(lit("...")))), V},
	{word(paramExp(lit("V"), "-", word(lit("...")))), V},
	{word(paramExp(lit("E"), ":-", word(lit("...")))), "..."},
	{word(paramExp(lit("E"), "-", word(lit("...")))), ""},
	{word(paramExp(lit("E"), ":-", word())), ""},
	{word(paramExp(lit("_"), ":-", word(lit("...")))), "..."},
	{word(paramExp(lit("_"), "-", word(lit("...")))), "..."},
	{word(paramExp(lit("_"), ":-", word())), ""},
	{word(paramExp(lit("_"), "-", word())), ""},
}

func TestExpand(t *testing.T) {
	env := interp.NewExecEnv()
	for _, tt := range expandTests {
		if g, e := env.Expand(tt.word, false), tt.s; g != e {
			t.Errorf("expected %q, got %q", e, g)
		}
	}
	t.Run("TildeExp", func(t *testing.T) {
		env := interp.NewExecEnv()
		env.Unset("_")
		for _, tt := range tildeExpTests {
			if g, e := env.Expand(tt.word, tt.assign), tt.s; g != e {
				t.Errorf("expected %q, got %q", e, g)
			}
		}
	})
	t.Run("ParamExp", func(t *testing.T) {
		for _, tt := range paramExpTests {
			env := interp.NewExecEnv()
			env.Set("V", V)
			env.Set("E", E)
			env.Unset("_")
			if g, e := env.Expand(tt.word, false), tt.s; g != e {
				t.Errorf("expected %q, got %q", e, g)
			}
		}
	})
}

func word(w ...ast.WordPart) ast.Word {
	if len(w) == 0 {
		return ast.Word{}
	}
	return ast.Word(w)
}

func lit(s string) *ast.Lit {
	return &ast.Lit{Value: s}
}

func litf(format string, a ...interface{}) *ast.Lit {
	return &ast.Lit{Value: fmt.Sprintf(format, a...)}
}

func paramExp(name *ast.Lit, op string, word ast.Word) *ast.ParamExp {
	return &ast.ParamExp{
		Name: name,
		Op:   op,
		Word: word,
	}
}

func username() string {
	u, err := user.Current()
	if err != nil {
		panic(err)
	}
	return u.Username
}

func homeDir() string {
	u, err := user.Current()
	if err != nil {
		panic(err)
	}
	return filepath.ToSlash(u.HomeDir)
}
