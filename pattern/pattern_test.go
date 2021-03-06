//
// go.sh/pattern :: pattern_test.go
//
//   Copyright (c) 2021 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
//

package pattern_test

import (
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
