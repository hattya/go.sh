//
// go.sh/pattern :: pattern_test.go
//
//   Copyright (c) 2021-2022 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
//

package pattern_test

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/hattya/go.sh/pattern"
)

var matchTests = []struct {
	patterns []string
	mode     pattern.Mode
	s        string
	e        string
}{
	{[]string{""}, 0, "", ""},
	{[]string{"", ""}, 0, "", ""},
	{[]string{"*.go"}, 0, "go.mod", ""},
	{[]string{"*.go"}, 0, "pattern.go", "pattern.go"},
	{[]string{"*.sw?"}, 0, ".pattern.go.swp", ".pattern.go.swp"},
	{[]string{"\\w"}, 0, "w", "w"},
	{[]string{"\\["}, 0, "[", "["},
	{[]string{"abc[lmn]xyz"}, 0, "abcmxyz", "abcmxyz"},
	{[]string{"abc[!lmn]xyz"}, 0, "abc-xyz", "abc-xyz"},
	{[]string{"[]\\-]"}, 0, "-", "-"},
	{[]string{"[[\\+]"}, 0, "+", "+"},
	{[]string{"[[:digit:]]"}, 0, "1", "1"},
	{[]string{"[[:digit]"}, 0, ":", ":"},

	{[]string{"/*"}, pattern.Smallest | pattern.Suffix, "foo", ""},
	{[]string{"/*"}, pattern.Smallest | pattern.Suffix, "foo/bar/baz", "/baz"},
	{[]string{"/*"}, pattern.Largest | pattern.Suffix, "foo", ""},
	{[]string{"/*"}, pattern.Largest | pattern.Suffix, "foo/bar/baz", "/bar/baz"},

	{[]string{"*/"}, pattern.Smallest | pattern.Prefix, "foo", ""},
	{[]string{"*/"}, pattern.Smallest | pattern.Prefix, "foo/bar/baz", "foo/"},
	{[]string{"*/"}, pattern.Largest | pattern.Prefix, "foo", ""},
	{[]string{"*/"}, pattern.Largest | pattern.Prefix, "foo/bar/baz", "foo/bar/"},

	{[]string{"*"}, pattern.Smallest | pattern.Suffix, "", ""},
	{[]string{"*"}, pattern.Smallest | pattern.Prefix, "", ""},
	{[]string{"*"}, pattern.Suffix | pattern.Prefix, "foo", ""},
	{[]string{"?"}, pattern.Smallest | pattern.Suffix, "\xf0\xff", "\xff"},
	{[]string{"?"}, pattern.Smallest | pattern.Prefix, "\xf0\xff", "\xf0"},
}

func TestMatch(t *testing.T) {
	for _, tt := range matchTests {
		switch g, err := pattern.Match(tt.patterns, tt.mode, tt.s); {
		case err != nil && err != pattern.NoMatch:
			t.Error("unexpected error:", err)
		case g != tt.e:
			t.Errorf("expected %q, got %q", tt.e, g)
		}
	}
}

var matchErrorTests = [][]string{
	{"\xff"},
	{"\\"},
	{"\\\xff"},
	{"["},
	{"[\xff"},
	{"[\\"},
	{"[\\\xff"},
	{"[["},
	{"[[\xff"},
	{"[[\\"},
}

func TestMatchError(t *testing.T) {
	for _, patterns := range matchErrorTests {
		if _, err := pattern.Match(patterns, 0, ""); err == nil {
			t.Error("expected error")
		}
	}
}

var globTests = []struct {
	pattern string
	paths   []string
}{
	{"*.go", []string{"a.go"}},
	{"[.a]*", []string{"a.go"}},
	{".*", []string{".", "..", ".git", ".gitignore"}},
	{"*", []string{"a.go", "bar", "baz", "foo"}},
	{"*/*", []string{"bar/a.go", "baz/a.go", "foo/a.go"}},
	{"foo/*", []string{"foo/a.go"}},
	{"foo//*", []string{"foo//a.go"}},
	{`foo\/*`, []string{"foo/a.go"}},
	{`foo/\*`, nil},
	{"_.go", nil},
	{"_/*", nil},
	{"_/_", nil},
	{".", []string{"."}},
	{"..", []string{".."}},
	{"", nil},

	{"${PATDIR}/*.go", []string{"${LITDIR}/a.go"}},
	{"${PATDIR}/.*", []string{"${LITDIR}/.", "${LITDIR}/..", "${LITDIR}/.git", "${LITDIR}/.gitignore"}},
	{"${PATDIR}/foo/*", []string{"${LITDIR}/foo/a.go"}},
	{"${PATDIR}/foo//*", []string{"${LITDIR}/foo//a.go"}},
	{`${PATDIR}/foo\/*`, []string{"${LITDIR}/foo/a.go"}},
	{`${PATDIR}/fo/\*`, nil},
	{"${PATDIR}/_.go", nil},
	{"${PATDIR}/_/*", nil},
	{"${PATDIR}/_/_", nil},
	{"${PATDIR}/", []string{"${LITDIR}/"}},
	{"${PATDIR}/.", []string{"${LITDIR}/."}},
	{"${PATDIR}/..", []string{"${LITDIR}/.."}},
}

func TestGlob(t *testing.T) {
	dir := t.TempDir()
	popd, err := pushd(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer popd()

	for _, p := range []string{
		filepath.Join(".git", "config"),
		".gitignore",
		"a.go",
		filepath.Join("foo", "a.go"),
		filepath.Join("bar", "a.go"),
		filepath.Join("baz", "a.go"),
	} {
		if dir := filepath.Dir(p); dir != "" {
			if err := mkdir(dir); err != nil {
				t.Fatal(err)
			}
		}
		if err := touch(p); err != nil {
			t.Fatal(err)
		}
	}

	mapper := func(k string) string {
		switch k {
		case "PATDIR":
			return strings.ReplaceAll(dir, `\`, `\\`)
		case "LITDIR":
			return dir
		}
		return ""
	}
	for _, tt := range globTests {
		g, err := pattern.Glob(os.Expand(tt.pattern, mapper))
		if err != nil {
			t.Error("unexpected error:", err)
		}
		var e []string
		for _, p := range tt.paths {
			e = append(e, os.Expand(p, mapper))
		}
		if !reflect.DeepEqual(g, e) {
			t.Errorf("expected %#v, got %#v", e, g)
		}
	}
}

var globErrorTests = []string{
	"*\xff",
	"_\xff",
}

func TestGlobError(t *testing.T) {
	for _, pat := range globErrorTests {
		if _, err := pattern.Glob(pat); err == nil {
			t.Error("expected error")
		}
	}
}

func mkdir(s ...string) error {
	return os.MkdirAll(filepath.Join(s...), 0o777)
}

func pushd(path string) (func() error, error) {
	wd, err := os.Getwd()
	popd := func() error {
		if err != nil {
			return err
		}
		return os.Chdir(wd)
	}
	return popd, os.Chdir(path)
}

func touch(s ...string) error {
	return os.WriteFile(filepath.Join(s...), nil, 0o666)
}
