//
// go.sh/interp :: expand_test.go
//
//   Copyright (c) 2021-2024 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
//

package interp_test

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/hattya/go.sh/ast"
	"github.com/hattya/go.sh/interp"
	"github.com/hattya/go.sh/printer"
)

const (
	sep = string(os.PathListSeparator)
	V   = "value"
	E   = ""
	P   = "foo/bar/baz"
)

var expandTests = []struct {
	word   ast.Word
	mode   interp.ExpMode
	fields []string
}{
	{word(), 0, nil},
	{word(), interp.Quote, []string{""}},
	{word(quote(`'`, word())), 0, []string{""}},
	{word(quote(`"`, word())), 0, []string{""}},

	{word(lit("foo")), 0, []string{"foo"}},
	{word(quote(`\`, word(lit("f"))), lit("oo")), 0, []string{"foo"}},
	{word(quote(`'`, word(lit("foo")))), 0, []string{"foo"}},
	{word(quote(`"`, word(lit("foo")))), 0, []string{"foo"}},
	{word(lit("foo"), lit("bar")), 0, []string{"foobar"}},

	{word(quote(`\`, word(lit("*")))), 0, []string{`*`}},
	{word(quote(`'`, word(lit("*")))), 0, []string{`*`}},
	{word(quote(`"`, word(lit("*")))), 0, []string{`*`}},

	{word(quote(`\`, word(lit("*")))), interp.Literal, []string{`*`}},
	{word(quote(`'`, word(lit("*")))), interp.Literal, []string{`*`}},
	{word(quote(`"`, word(lit("*")))), interp.Literal, []string{`*`}},

	{word(quote(`\`, word(lit("*")))), interp.Pattern, []string{`\*`}},
	{word(quote(`'`, word(lit("*")))), interp.Pattern, []string{`\*`}},
	{word(quote(`"`, word(lit("*")))), interp.Pattern, []string{`\*`}},

	{word(paramExp(lit("E"), "", nil), lit(" * 1")), interp.Arith, []string{"E * 1"}},
	{word(quote(`"`, word(paramExp(lit("E"), "", nil))), lit(" * 1")), interp.Arith, []string{"E * 1"}},
}

var tildeExpTests = []struct {
	word   ast.Word
	mode   interp.ExpMode
	fields []string
}{
	{word(lit("~")), 0, []string{homeDir()}},
	{word(lit("~/")), 0, []string{homeDir() + "/"}},
	{word(lit("~"), lit("/")), 0, []string{homeDir() + "/"}},
	{word(lit("~"), quote(`\`, word(lit(`/`)))), 0, []string{"~/"}},
	{word(lit("~"), quote(`\`, word(lit(`\`)))), 0, []string{homeDir() + `\`}},

	{word(lit("~" + username())), 0, []string{homeDir()}},
	{word(lit("~" + username() + "/")), 0, []string{homeDir() + "/"}},
	{word(lit("~"), lit(username()), lit("/")), 0, []string{homeDir() + "/"}},
	{word(lit("~"+username()), quote(`\`, word(lit(`/`)))), 0, []string{"~" + username() + "/"}},
	{word(lit("~"+username()), quote(`\`, word(lit(`\`)))), 0, []string{homeDir() + `\`}},

	{word(lit("~_")), 0, []string{"~_"}},
	{word(lit("~_/")), 0, []string{"~_/"}},
	{word(lit("~"), lit("_"), lit("/")), 0, []string{"~_/"}},
	{word(lit("~_"), quote(`\`, word(lit("/")))), 0, []string{"~_/"}},
	{word(lit("~_"), quote(`\`, word(lit(`\`)))), 0, []string{`~_\`}},

	{word(lit(sep)), interp.Assign, []string{sep}},
	{word(litf("~/foo%v~/bar", sep)), interp.Assign, []string{fmt.Sprintf("%v/foo%v%[1]v/bar", homeDir(), sep)}},
	{word(lit("~/foo"), lit(sep), lit("~/bar")), interp.Assign, []string{fmt.Sprintf("%v/foo%v%[1]v/bar", homeDir(), sep)}},
	{word(lit("~"), lit("/foo"), lit(sep), lit("~"), lit("/bar")), interp.Assign, []string{fmt.Sprintf("%v/foo%v%[1]v/bar", homeDir(), sep)}},
	{word(lit("~"), quote(`\`, word(lit(`/`))), litf("foo%v~", sep), quote(`\`, word(lit(`/`))), lit("bar")), interp.Assign, []string{fmt.Sprintf("~/foo%v~/bar", sep)}},
	{word(lit("~"), quote(`\`, word(lit(`\`))), litf("foo%v~", sep), quote(`\`, word(lit(`/`))), lit("bar")), interp.Assign, []string{fmt.Sprintf(`%v\foo%v~/bar`, homeDir(), sep)}},
	{word(lit("~"), quote(`\`, word(lit(`/`))), litf("foo%v~", sep), quote(`\`, word(lit(`\`))), lit("bar")), interp.Assign, []string{fmt.Sprintf(`~/foo%v%v\bar`, sep, homeDir())}},
	{word(lit("~"), quote(`\`, word(lit(`\`))), litf("foo%v~", sep), quote(`\`, word(lit(`\`))), lit("bar")), interp.Assign, []string{fmt.Sprintf(`%v\foo%v%[1]v\bar`, homeDir(), sep)}},

	{word(lit("~"), paramExp(lit("_"), "", nil), lit("/")), 0, []string{"~/"}},
	{word(lit("~/")), interp.Quote, []string{"~/"}},
	{word(quote(`"`, word(lit("~/")))), 0, []string{"~/"}},

	{word(paramExp(lit("_"), ":-", word(litf("~/foo%v~/bar", sep)))), interp.Assign, []string{fmt.Sprintf("%v/foo%v%[1]v/bar", homeDir(), sep)}},
	{word(paramExp(lit("E"), "+", word(litf("~/foo%v~/bar", sep)))), interp.Assign, []string{fmt.Sprintf("%v/foo%v%[1]v/bar", homeDir(), sep)}},
}

var paramExpTests = []struct {
	word   ast.Word
	fields []string
	err    string
	assign bool
}{
	// simplest form
	{word(paramExp(lit("V"), "", nil)), []string{V}, "", false},
	// use default values
	{word(paramExp(lit("V"), ":-", word(lit("...")))), []string{V}, "", false},
	{word(paramExp(lit("V"), "-", word(lit("...")))), []string{V}, "", false},
	{word(paramExp(lit("E"), ":-", word(lit("...")))), []string{"..."}, "", false},
	{word(paramExp(lit("E"), "-", word(lit("...")))), []string{""}, "", false},
	{word(paramExp(lit("E"), ":-", word())), []string{""}, "", false},
	{word(paramExp(lit("_"), ":-", word(quote(`'`, word(lit("...")))))), []string{"..."}, "", false},
	{word(paramExp(lit("_"), "-", word(quote(`'`, word(lit("...")))))), []string{"..."}, "", false},
	{word(paramExp(lit("_"), ":-", word())), []string{""}, "", false},
	{word(paramExp(lit("_"), "-", word())), []string{""}, "", false},

	{word(paramExp(lit("_"), ":-", word(quote(`"`, word(paramExp(lit("1"), ":=", word(lit("...")))))))), nil, "$1: cannot assign ", false},
	{word(paramExp(lit("_"), "-", word(quote(`"`, word(paramExp(lit("1"), "=", word(lit("...")))))))), nil, "$1: cannot assign ", false},
	// assign default values
	{word(paramExp(lit("V"), ":=", word(lit("...")))), []string{V}, "", false},
	{word(paramExp(lit("V"), "=", word(lit("...")))), []string{V}, "", false},
	{word(paramExp(lit("E"), ":=", word(lit("...")))), []string{"..."}, "", true},
	{word(paramExp(lit("E"), "=", word(lit("...")))), []string{""}, "", false},
	{word(paramExp(lit("E"), ":=", word())), []string{""}, "", true},
	{word(paramExp(lit("_"), ":=", word(lit("...")))), []string{"..."}, "", true},
	{word(paramExp(lit("_"), "=", word(lit("...")))), []string{"..."}, "", true},
	{word(paramExp(lit("_"), ":=", word())), []string{""}, "", true},
	{word(paramExp(lit("_"), "=", word())), []string{""}, "", true},

	{word(paramExp(lit("1"), ":=", word(lit("...")))), nil, "$1: cannot assign ", false},
	{word(paramExp(lit("1"), "=", word(lit("...")))), nil, "$1: cannot assign ", false},
	{word(paramExp(lit("@"), ":=", word(lit("...")))), nil, "$@: cannot assign ", false},
	{word(paramExp(lit("_"), ":=", word(quote(`"`, word(paramExp(lit("1"), ":=", word(lit("...")))))))), nil, "$1: cannot assign ", false},
	{word(paramExp(lit("_"), "=", word(quote(`"`, word(paramExp(lit("1"), "=", word(lit("...")))))))), nil, "$1: cannot assign ", false},
	// indicate error if unset or null
	{word(paramExp(lit("V"), ":?", word(lit("...")))), []string{V}, "", false},
	{word(paramExp(lit("V"), "?", word(lit("...")))), []string{V}, "", false},
	{word(paramExp(lit("E"), ":?", word(lit("...")))), nil, "$E: ...", false},
	{word(paramExp(lit("E"), "?", word(lit("...")))), []string{""}, "", false},
	{word(paramExp(lit("_"), ":?", word(quote(`'`, word(lit("...")))))), nil, "$_: ...", false},
	{word(paramExp(lit("_"), "?", word(quote(`'`, word(lit("...")))))), nil, "$_: ...", false},
	{word(paramExp(lit("_"), ":?", word())), nil, "$_: parameter is unset or null", false},
	{word(paramExp(lit("_"), "?", word())), nil, "$_: parameter is unset or null", false},

	{word(paramExp(lit("_"), ":?", word(quote(`"`, word(paramExp(lit("1"), ":=", word(lit("...")))))))), nil, "$1: cannot assign ", false},
	{word(paramExp(lit("_"), "?", word(quote(`"`, word(paramExp(lit("1"), "=", word(lit("...")))))))), nil, "$1: cannot assign ", false},
	// use alternative values
	{word(paramExp(lit("V"), ":+", word(lit("...")))), []string{"..."}, "", false},
	{word(paramExp(lit("V"), "+", word(lit("...")))), []string{"..."}, "", false},
	{word(paramExp(lit("E"), ":+", word(lit("...")))), []string{""}, "", false},
	{word(paramExp(lit("E"), "+", word(lit("...")))), []string{"..."}, "", false},
	{word(paramExp(lit("E"), "+", word())), []string{""}, "", false},
	{word(paramExp(lit("_"), ":+", word(lit("...")))), []string{""}, "", false},
	{word(paramExp(lit("_"), "+", word(lit("...")))), []string{""}, "", false},

	{word(paramExp(lit("V"), ":+", word(quote(`"`, word(paramExp(lit("1"), ":=", word(lit("...")))))))), nil, "$1: cannot assign ", false},
	{word(paramExp(lit("V"), "+", word(quote(`"`, word(paramExp(lit("1"), "=", word(lit("...")))))))), nil, "$1: cannot assign ", false},
	// string length
	{word(paramExp(lit("V"), "#", nil)), []string{strconv.Itoa(len(V))}, "", false},
	{word(paramExp(lit("E"), "#", nil)), []string{"0"}, "", false},
	{word(paramExp(lit("_"), "#", nil)), nil, "$_: parameter is unset", false},
	// remove suffix pattern
	{word(paramExp(lit("P"), "%", word(lit("/*")))), []string{"foo/bar"}, "", false},
	{word(paramExp(lit("P"), "%%", word(lit("/*")))), []string{"foo"}, "", false},
	{word(paramExp(lit("P"), "%", word(quote(`'`, word(lit("/*")))))), []string{"foo/bar/baz"}, "", false},
	{word(paramExp(lit("P"), "%%", word(quote(`'`, word(lit("/*")))))), []string{"foo/bar/baz"}, "", false},
	{word(paramExp(lit("P"), "%", word())), []string{"foo/bar/baz"}, "", false},
	{word(paramExp(lit("P"), "%%", word())), []string{"foo/bar/baz"}, "", false},
	{word(paramExp(lit("_"), "%", word())), nil, "$_: parameter is unset", false},
	{word(paramExp(lit("_"), "%%", word())), nil, "$_: parameter is unset", false},

	{word(paramExp(lit("V"), "%", word(quote(`"`, word(paramExp(lit("1"), "=", word(lit("...")))))))), nil, "$1: cannot assign ", false},
	{word(paramExp(lit("V"), "%%", word(quote(`"`, word(paramExp(lit("1"), ":=", word(lit("...")))))))), nil, "$1: cannot assign ", false},
	{word(paramExp(lit("V"), "%", word(lit("\xff")))), nil, "regexp: invalid UTF-8", false},
	{word(paramExp(lit("V"), "%%", word(lit("\xff")))), nil, "regexp: invalid UTF-8", false},
	// remove prefix pattern
	{word(paramExp(lit("P"), "#", word(lit("*/")))), []string{"bar/baz"}, "", false},
	{word(paramExp(lit("P"), "##", word(lit("*/")))), []string{"baz"}, "", false},
	{word(paramExp(lit("P"), "#", word(quote(`'`, word(lit("*/")))))), []string{"foo/bar/baz"}, "", false},
	{word(paramExp(lit("P"), "##", word(quote(`'`, word(lit("*/")))))), []string{"foo/bar/baz"}, "", false},
	{word(paramExp(lit("P"), "#", word())), []string{"foo/bar/baz"}, "", false},
	{word(paramExp(lit("P"), "##", word())), []string{"foo/bar/baz"}, "", false},
	{word(paramExp(lit("_"), "#", word())), nil, "$_: parameter is unset", false},
	{word(paramExp(lit("_"), "##", word())), nil, "$_: parameter is unset", false},

	{word(paramExp(lit("V"), "#", word(quote(`"`, word(paramExp(lit("1"), "=", word(lit("...")))))))), nil, "$1: cannot assign ", false},
	{word(paramExp(lit("V"), "##", word(quote(`"`, word(paramExp(lit("1"), ":=", word(lit("...")))))))), nil, "$1: cannot assign ", false},
	{word(paramExp(lit("V"), "#", word(lit("\xff")))), nil, "regexp: invalid UTF-8", false},
	{word(paramExp(lit("V"), "##", word(lit("\xff")))), nil, "regexp: invalid UTF-8", false},
}

var spParamTests = []struct {
	word   ast.Word
	mode   interp.ExpMode
	args   []string
	ifs    any
	fields []string
	err    string
	assign string
}{
	// simplest form
	{word(paramExp(lit("@"), "", nil)), 0, nil, nil, nil, "", ""},
	{word(paramExp(lit("@"), "", nil)), 0, []string{""}, nil, nil, "", ""},
	{word(quote(`"`, word(paramExp(lit("@"), "", nil)))), 0, []string{""}, nil, []string{""}, "", ""},
	{word(paramExp(lit("@"), "", nil)), 0, []string{"1"}, nil, []string{"1"}, "", ""},
	{word(paramExp(lit("@"), "", nil)), 0, []string{"1", "2"}, nil, []string{"1", "2"}, "", ""},
	{word(paramExp(lit("@"), "", nil)), 0, []string{"", "2", ""}, nil, []string{"2"}, "", ""},
	{word(quote(`"`, word(paramExp(lit("@"), "", nil)))), 0, []string{"", "2", ""}, nil, []string{"", "2", ""}, "", ""},

	{word(paramExp(lit("*"), "", nil)), 0, nil, nil, nil, "", ""},
	{word(paramExp(lit("*"), "", nil)), 0, []string{""}, nil, nil, "", ""},
	{word(quote(`"`, word(paramExp(lit("*"), "", nil)))), 0, []string{""}, nil, []string{""}, "", ""},
	{word(paramExp(lit("*"), "", nil)), 0, []string{"1"}, nil, []string{"1"}, "", ""},
	{word(paramExp(lit("*"), "", nil)), 0, []string{"1", "2"}, nil, []string{"1", "2"}, "", ""},
	{word(paramExp(lit("*"), "", nil)), 0, []string{"", "2", ""}, nil, []string{"2"}, "", ""},
	{word(quote(`"`, word(paramExp(lit("*"), "", nil)))), 0, []string{"", "2", ""}, nil, []string{" 2 "}, "", ""},
	{word(quote(`"`, word(paramExp(lit("*"), "", nil)))), 0, []string{"", "2", ""}, ", \t\n", []string{",2,"}, "", ""},
	{word(quote(`"`, word(paramExp(lit("*"), "", nil)))), 0, []string{"", "2", ""}, "", []string{"2"}, "", ""},

	{word(paramExp(lit("@"), "", nil)), interp.Literal, []string{"*", "?"}, nil, []string{"* ?"}, "", ""},
	{word(paramExp(lit("@"), "", nil)), interp.Literal, []string{"*", "?"}, ", \t\n", []string{"*,?"}, "", ""},
	{word(paramExp(lit("@"), "", nil)), interp.Literal, []string{"*", "?"}, "", []string{"*?"}, "", ""},
	{word(quote(`"`, word(paramExp(lit("@"), "", nil)))), interp.Literal, []string{"*", "?"}, nil, []string{"* ?"}, "", ""},
	{word(quote(`"`, word(paramExp(lit("@"), "", nil)))), interp.Literal, []string{"*", "?"}, ", \t\n", []string{"*,?"}, "", ""},
	{word(quote(`"`, word(paramExp(lit("@"), "", nil)))), interp.Literal, []string{"*", "?"}, "", []string{"*?"}, "", ""},

	{word(paramExp(lit("@"), "", nil)), interp.Pattern, []string{"*", "?"}, nil, []string{"* ?"}, "", ""},
	{word(paramExp(lit("@"), "", nil)), interp.Pattern, []string{"*", "?"}, ", \t\n", []string{"*,?"}, "", ""},
	{word(paramExp(lit("@"), "", nil)), interp.Pattern, []string{"*", "?"}, "", []string{"*?"}, "", ""},
	{word(quote(`"`, word(paramExp(lit("@"), "", nil)))), interp.Pattern, []string{"*", "?"}, nil, []string{`\* \?`}, "", ""},
	{word(quote(`"`, word(paramExp(lit("@"), "", nil)))), interp.Pattern, []string{"*", "?"}, ", \t\n", []string{`\*,\?`}, "", ""},
	{word(quote(`"`, word(paramExp(lit("@"), "", nil)))), interp.Pattern, []string{"*", "?"}, "", []string{`\*\?`}, "", ""},
	// use default values
	{word(paramExp(lit("@"), ":-", word(lit("...")))), 0, []string{"", "2", ""}, nil, []string{"2"}, "", ""},
	{word(paramExp(lit("_"), ":-", word(paramExp(lit("@"), "", nil)))), 0, []string{"", "2", ""}, nil, []string{"2"}, "", ""},
	{word(paramExp(lit("_"), ":-", word(quote(`"`, word(paramExp(lit("@"), "", nil)))))), 0, []string{"", "2", ""}, nil, []string{"", "2", ""}, "", ""},
	{word(quote(`"`, word(paramExp(lit("_"), ":-", word(paramExp(lit("@"), "", nil)))))), 0, []string{"", "2", ""}, nil, []string{"", "2", ""}, "", ""},
	{word(quote(`"`, word(paramExp(lit("_"), ":-", word(quote(`"`, word(paramExp(lit("@"), "", nil)))))))), 0, []string{"", "2", ""}, nil, []string{"", "2", ""}, "", ""},
	// assign default values
	{word(paramExp(lit("@"), ":=", word(lit("...")))), 0, []string{"", "2", ""}, nil, []string{"2"}, "", ""},
	{word(paramExp(lit("_"), ":=", word(paramExp(lit("@"), "", nil)))), 0, []string{"", "2", ""}, nil, []string{"2"}, "", " 2 "},
	{word(paramExp(lit("_"), ":=", word(paramExp(lit("@"), "", nil)))), 0, []string{"", "2", ""}, ", \n\t", []string{"2"}, "", ",2,"},
	{word(paramExp(lit("_"), ":=", word(paramExp(lit("@"), "", nil)))), 0, []string{"", "2", ""}, "", []string{"2"}, "", "2"},
	{word(paramExp(lit("_"), ":=", word(quote(`"`, word(paramExp(lit("@"), "", nil)))))), 0, []string{"", "2", ""}, nil, []string{"", "2", ""}, "", " 2 "},
	{word(quote(`"`, word(paramExp(lit("_"), ":=", word(paramExp(lit("@"), "", nil)))))), 0, []string{"", "2", ""}, nil, []string{"", "2", ""}, "", " 2 "},
	{word(quote(`"`, word(paramExp(lit("_"), ":=", word(quote(`"`, word(paramExp(lit("@"), "", nil)))))))), 0, []string{"", "2", ""}, nil, []string{"", "2", ""}, "", " 2 "},
	// indicate error if unset or null
	{word(paramExp(lit("@"), ":?", word(lit("...")))), 0, []string{"", "2", ""}, nil, []string{"2"}, "", ""},
	{word(paramExp(lit("_"), ":?", word(paramExp(lit("@"), "", nil)))), 0, []string{"", "2", ""}, nil, nil, "$_:  2 ", ""},
	{word(paramExp(lit("_"), ":?", word(paramExp(lit("@"), "", nil)))), 0, []string{"", "2", ""}, ", \n\t", nil, "$_: ,2,", ""},
	{word(paramExp(lit("_"), ":?", word(paramExp(lit("@"), "", nil)))), 0, []string{"", "2", ""}, "", nil, "$_: 2", ""},
	{word(paramExp(lit("_"), ":?", word(quote(`"`, word(paramExp(lit("@"), "", nil)))))), 0, []string{"", "2", ""}, nil, nil, "$_:  2 ", ""},
	{word(quote(`"`, word(paramExp(lit("_"), ":?", word(paramExp(lit("@"), "", nil)))))), 0, []string{"", "2", ""}, nil, nil, "$_:  2 ", ""},
	{word(quote(`"`, word(paramExp(lit("_"), ":?", word(quote(`"`, word(paramExp(lit("@"), "", nil)))))))), 0, []string{"", "2", ""}, nil, nil, "$_:  2 ", ""},
	// use alternative values
	{word(paramExp(lit("@"), ":+", word(lit("...")))), 0, []string{"", "2", ""}, nil, []string{"..."}, "", ""},
	{word(paramExp(lit("V"), ":+", word(paramExp(lit("@"), "", nil)))), 0, []string{"", "2", ""}, nil, []string{"2"}, "", ""},
	{word(paramExp(lit("V"), ":+", word(quote(`"`, word(paramExp(lit("@"), "", nil)))))), 0, []string{"", "2", ""}, nil, []string{"", "2", ""}, "", ""},
	{word(quote(`"`, word(paramExp(lit("V"), ":+", word(paramExp(lit("@"), "", nil)))))), 0, []string{"", "2", ""}, nil, []string{"", "2", ""}, "", ""},
	{word(quote(`"`, word(paramExp(lit("V"), ":+", word(quote(`"`, word(paramExp(lit("@"), "", nil)))))))), 0, []string{"", "2", ""}, nil, []string{"", "2", ""}, "", ""},
	// string length
	{word(paramExp(lit("@"), "#", nil)), 0, nil, nil, []string{"0"}, "", ""},
	{word(paramExp(lit("@"), "#", nil)), 0, []string{""}, nil, []string{"1"}, "", ""},
	{word(paramExp(lit("@"), "#", nil)), 0, []string{"1"}, nil, []string{"1"}, "", ""},
	{word(paramExp(lit("@"), "#", nil)), 0, []string{"1", "2"}, nil, []string{"2"}, "", ""},
	// remove suffix pattern
	{word(paramExp(lit("@"), "%", word(lit("/*")))), 0, []string{"foo/bar/baz", "qux/quux"}, nil, []string{"foo/bar", "qux"}, "", ""},
	{word(paramExp(lit("@"), "%%", word(lit("/*")))), 0, []string{"foo/bar/baz", "qux/quux"}, nil, []string{"foo", "qux"}, "", ""},
	{word(paramExp(lit("@"), "%", word())), 0, []string{"foo/bar/baz", "qux/quux"}, nil, []string{"foo/bar/baz", "qux/quux"}, "", ""},
	{word(paramExp(lit("@"), "%%", word())), 0, []string{"foo/bar/baz", "qux/quux"}, nil, []string{"foo/bar/baz", "qux/quux"}, "", ""},

	{word(paramExp(lit("P"), "%", word(paramExp(lit("@"), "", nil)))), 0, []string{"", "*"}, "/", []string{"foo", "bar"}, "", ""},
	{word(paramExp(lit("P"), "%%", word(paramExp(lit("@"), "", nil)))), 0, []string{"", "*"}, "/", []string{"foo"}, "", ""},
	{word(quote(`"`, word(paramExp(lit("P"), "%", word(paramExp(lit("@"), "", nil)))))), 0, []string{"", "*"}, "/", []string{"foo/bar"}, "", ""},
	{word(quote(`"`, word(paramExp(lit("P"), "%%", word(paramExp(lit("@"), "", nil)))))), 0, []string{"", "*"}, "/", []string{"foo"}, "", ""},
	// remove prefix pattern
	{word(paramExp(lit("@"), "#", word(lit("*/")))), 0, []string{"foo/bar/baz", "qux/quux"}, nil, []string{"bar/baz", "quux"}, "", ""},
	{word(paramExp(lit("@"), "##", word(lit("*/")))), 0, []string{"foo/bar/baz", "qux/quux"}, nil, []string{"baz", "quux"}, "", ""},
	{word(paramExp(lit("@"), "#", word())), 0, []string{"foo/bar/baz", "qux/quux"}, nil, []string{"foo/bar/baz", "qux/quux"}, "", ""},
	{word(paramExp(lit("@"), "##", word())), 0, []string{"foo/bar/baz", "qux/quux"}, nil, []string{"foo/bar/baz", "qux/quux"}, "", ""},

	{word(paramExp(lit("P"), "#", word(paramExp(lit("@"), "", nil)))), 0, []string{"*", ""}, "/", []string{"bar", "baz"}, "", ""},
	{word(paramExp(lit("P"), "##", word(paramExp(lit("@"), "", nil)))), 0, []string{"*", ""}, "/", []string{"baz"}, "", ""},
	{word(quote(`"`, word(paramExp(lit("P"), "#", word(paramExp(lit("@"), "", nil)))))), 0, []string{"*", ""}, "/", []string{"bar/baz"}, "", ""},
	{word(quote(`"`, word(paramExp(lit("P"), "##", word(paramExp(lit("@"), "", nil)))))), 0, []string{"*", ""}, "/", []string{"baz"}, "", ""},
}

var posParamTests = []struct {
	word   ast.Word
	args   []string
	fields []string
}{
	{word(paramExp(lit("1"), "", nil)), []string{}, nil},
	{word(paramExp(lit("1"), "", nil)), []string{""}, nil},
	{word(quote(`"`, word(paramExp(lit("1"), "", nil)))), []string{""}, []string{""}},
	{word(paramExp(lit("1"), "", nil)), []string{"1"}, []string{"1"}},
	{word(paramExp(lit("1"), "", nil)), []string{"1", "2"}, []string{"1"}},

	{word(paramExp(lit("01"), "", nil)), []string{}, nil},
	{word(paramExp(lit("01"), "", nil)), []string{""}, nil},
	{word(quote(`"`, word(paramExp(lit("01"), "", nil)))), []string{""}, []string{""}},
	{word(paramExp(lit("01"), "", nil)), []string{"1"}, []string{"1"}},
	{word(paramExp(lit("01"), "", nil)), []string{"1", "2"}, []string{"1"}},

	{word(paramExp(lit("2"), "", nil)), []string{}, nil},
	{word(paramExp(lit("2"), "", nil)), []string{"1"}, nil},
	{word(paramExp(lit("2"), "", nil)), []string{"1", "2"}, []string{"2"}},
	{word(paramExp(lit("2"), "", nil)), []string{"1", "2", "3"}, []string{"2"}},

	{word(paramExp(lit("02"), "", nil)), []string{}, nil},
	{word(paramExp(lit("02"), "", nil)), []string{"1"}, nil},
	{word(paramExp(lit("02"), "", nil)), []string{"1", "2"}, []string{"2"}},
	{word(paramExp(lit("02"), "", nil)), []string{"1", "2", "3"}, []string{"2"}},
}

var arithExpTests = []struct {
	word   ast.Word
	fields []string
	err    string
}{
	{word(arithExp(word(lit("2 - 1")))), []string{"1"}, ""},
	{word(quote(`"`, word(arithExp(word(lit("1 + 1")))))), []string{"2"}, ""},
	{word(arithExp(word(quote(`"`, word(lit("6 / 2")))))), []string{"3"}, ""},
	{word(arithExp(word(lit("1 << "), quote(`'`, word(lit("2")))))), []string{"4"}, ""},

	{word(arithExp(word(lit("E")))), []string{"0"}, ""},
	{word(arithExp(word(quote(`'`, word(lit("E")))))), []string{"0"}, ""},
	{word(arithExp(word(lit("1 + E")))), []string{"1"}, ""},
	{word(arithExp(word(lit("1 + "), quote(`'`, word(lit("E")))))), []string{"1"}, ""},
	{word(arithExp(word(lit("++E + 1")))), []string{"2"}, ""},
	{word(arithExp(word(lit("++"), quote(`'`, word(lit("E"))), lit(" + 1")))), []string{"2"}, ""},

	{word(arithExp(word(paramExp(lit("E"), "", nil)))), []string{"0"}, ""},
	{word(arithExp(word(quote(`"`, word(paramExp(lit("E"), "", nil)))))), []string{"0"}, ""},
	{word(arithExp(word(lit("1 + "), paramExp(lit("E"), "", nil)))), []string{"1"}, ""},
	{word(arithExp(word(lit("1 + "), quote(`"`, word(paramExp(lit("E"), "", nil)))))), []string{"1"}, ""},
	{word(arithExp(word(lit("++"), paramExp(lit("E"), "", nil), lit(" + 1")))), []string{"2"}, ""},
	{word(arithExp(word(lit("++"), quote(`"`, word(paramExp(lit("E"), "", nil))), lit(" + 1")))), []string{"2"}, ""},
	{word(arithExp(word(paramExp(lit("E"), ":-", word(arithExp(word(lit("7 % 4")))))))), []string{"3"}, ""},

	{word(arithExp(word(lit("1 * "), arithExp(word(lit("2 + 3")))))), []string{"5"}, ""},
	{word(arithExp(word(lit("1 * "), quote(`"`, word(arithExp(word(lit("2 + 3")))))))), []string{"5"}, ""},

	{word(arithExp(word())), nil, "arithmetic expression is missing"},
	{word(arithExp(word(paramExp(lit("9"), ":=", word(lit("...")))))), nil, "$9: cannot assign "},
	{word(arithExp(word(paramExp(lit("@"), "", nil)))), nil, "1 2 3: unexpected NUMBER"},
	{word(arithExp(word(paramExp(lit("#"), "", nil), lit("++")))), nil, "3++: '++' requires lvalue"},
	{word(arithExp(word(paramExp(lit("1"), "", nil), lit("--")))), nil, "1--: '--' requires lvalue"},
}

var fieldSplitTests = []struct {
	word   ast.Word
	ifs    any
	fields []string
}{
	{word(lit(" \t abc \t xyz \t ")), nil, []string{"abc", "xyz"}},
	{word(lit(" \t abc \t, \t ,\t xyz \t ")), " \t\n,", []string{"abc", "xyz"}},
	{word(lit(" \t abc \t, "), quote(`"`, word()), lit(" ,\t xyz \t ")), " \t\n,", []string{"abc", "", "xyz"}},
	{word(lit(" \t,abc \t xyz \t,")), " \t\n,", []string{"abc", "xyz"}},
	{word(quote(`"`, word()), lit("\t,abc \t xyz \t,")), " \t\n,", []string{"", "abc", "xyz"}},
	{word(lit(" \t,abc \t xyz \t,"), quote(`"`, word())), " \t\n,", []string{"abc", "xyz", ""}},
	{word(lit("abc \xff xyz")), nil, []string{"abc", "\xff", "xyz"}},
	{word(lit("abc \t xyz")), "", []string{"abc \t xyz"}},
	{word(quote(`'`, word(lit("abc \t xyz")))), "", []string{"abc \t xyz"}},
	{word(quote(`"`, word(lit("abc \t xyz")))), "", []string{"abc \t xyz"}},
}

var pathExpTests = []struct {
	word   ast.Word
	opts   interp.Option
	fields []string
}{
	{word(), 0, nil},
	{word(lit("foo")), 0, []string{"foo"}},
	{word(lit("qux")), 0, []string{"qux"}},

	{word(lit("b*")), 0, []string{"bar", "baz"}},
	{word(lit("b*")), interp.NoGlob, []string{"b*"}},
	{word(lit("b"), quote(`\`, word(lit("*")))), 0, []string{"b*"}},
	{word(quote(`'`, word(lit("b*")))), 0, []string{"b*"}},
	{word(quote(`"`, word(lit("b*")))), 0, []string{"b*"}},

	{word(lit("q*")), 0, []string{"q*"}},
	{word(lit("q*")), interp.NoGlob, []string{"q*"}},
	{word(lit("q"), quote(`\`, word(lit("*")))), 0, []string{"q*"}},
	{word(quote(`'`, word(lit("q*")))), 0, []string{"q*"}},
	{word(quote(`"`, word(lit("q*")))), 0, []string{"q*"}},

	{word(lit("\xff*")), 0, []string{"\xff*"}},
	{word(lit("\xff*")), interp.NoGlob, []string{"\xff*"}},
	{word(lit("\xff"), quote(`\`, word(lit("*")))), 0, []string{"\xff*"}},
	{word(quote(`'`, word(lit("\xff*")))), 0, []string{"\xff*"}},
	{word(quote(`"`, word(lit("\xff*")))), 0, []string{"\xff*"}},
}

func TestExpand(t *testing.T) {
	env := interp.NewExecEnv(name)
	env.Set("E", E)
	for _, tt := range expandTests {
		g, _ := env.Expand(tt.word, tt.mode)
		if e := tt.fields; !reflect.DeepEqual(g, e) {
			t.Errorf("expected %#v, got %#v", e, g)
		}
	}
	t.Run("TildeExp", func(t *testing.T) {
		env := interp.NewExecEnv(name)
		env.Set("E", E)
		env.Unset("_")
		for _, tt := range tildeExpTests {
			if runtime.GOOS != "windows" && strings.ContainsRune(tt.fields[0], '\\') {
				continue
			}
			g, _ := env.Expand(tt.word, tt.mode)
			if e := tt.fields; !reflect.DeepEqual(g, e) {
				t.Errorf("expected %#v, got %#v", e, g)
			}
		}
	})
	t.Run("ParamExp", func(t *testing.T) {
		for _, tt := range paramExpTests {
			env := interp.NewExecEnv(name)
			env.Opts |= interp.NoUnset
			env.Set("V", V)
			env.Set("E", E)
			env.Set("P", P)
			env.Unset("_")
			g, err := env.Expand(tt.word, interp.Quote)
			switch {
			case err == nil && tt.err != "":
				t.Error("expected error")
			case err != nil && (tt.err == "" || !strings.Contains(err.Error(), tt.err)):
				t.Error("unexpected error:", err)
			default:
				if e := tt.fields; !reflect.DeepEqual(g, e) {
					t.Errorf("expected %#v, got %#v", e, g)
				}
				if tt.assign {
					pe := tt.word[0].(*ast.ParamExp)
					if v, set := env.Get(pe.Name.Value); !set {
						t.Errorf("%v is unset", pe.Name.Value)
					} else {
						var b strings.Builder
						printer.Fprint(&b, pe.Word)
						if g, e := v.Value, b.String(); g != e {
							t.Errorf("expected %q, got %q", e, g)
						}
					}
				}
			}
		}
	})
	t.Run("SpParam", func(t *testing.T) {
		for _, tt := range spParamTests {
			env := interp.NewExecEnv(name, tt.args...)
			env.Set("V", V)
			env.Set("P", P)
			env.Unset("_")
			if tt.ifs != nil {
				env.Set("IFS", tt.ifs.(string))
			} else {
				env.Unset("IFS")
			}
			g, err := env.Expand(tt.word, tt.mode)
			switch {
			case err == nil && tt.err != "":
				t.Error("expected error")
			case err != nil && (tt.err == "" || !strings.Contains(err.Error(), tt.err)):
				t.Error("unexpected error:", err)
			default:
				if e := tt.fields; !reflect.DeepEqual(g, e) {
					t.Errorf("expected %#v, got %#v", e, g)
				}
				if tt.assign != "" {
					pe, ok := tt.word[0].(*ast.ParamExp)
					if !ok {
						pe = tt.word[0].(*ast.Quote).Value[0].(*ast.ParamExp)
					}
					if v, set := env.Get(pe.Name.Value); !set {
						t.Errorf("%v is unset", pe.Name.Value)
					} else {
						if g, e := v.Value, tt.assign; g != e {
							t.Errorf("expected %q, got %q", e, g)
						}
					}
				}
			}
		}
	})
	t.Run("PosParam", func(t *testing.T) {
		for _, tt := range posParamTests {
			env := interp.NewExecEnv(name, tt.args...)
			g, _ := env.Expand(tt.word, 0)
			if e := tt.fields; !reflect.DeepEqual(g, e) {
				t.Errorf("expected %#v, got %#v", e, g)
			}
		}
	})
	t.Run("ArithExp", func(t *testing.T) {
		for _, tt := range arithExpTests {
			env := interp.NewExecEnv(name, "1", "2", "3")
			env.Set("E", E)
			g, err := env.Expand(tt.word, 0)
			switch {
			case err == nil && tt.err != "":
				t.Error("expected error")
			case err != nil && (tt.err == "" || !strings.Contains(err.Error(), tt.err)):
				t.Error("unexpected error:", err)
			default:
				if e := tt.fields; !reflect.DeepEqual(g, e) {
					t.Errorf("expected %#v, got %#v", e, g)
				}
			}
		}
	})
	t.Run("FieldSplit", func(t *testing.T) {
		for _, tt := range fieldSplitTests {
			env := interp.NewExecEnv(name)
			if tt.ifs != nil {
				env.Set("IFS", tt.ifs.(string))
			} else {
				env.Unset("IFS")
			}
			g, _ := env.Expand(tt.word, 0)
			if e := tt.fields; !reflect.DeepEqual(g, e) {
				t.Errorf("expected %#v, got %#v", e, g)
			}
		}
	})
	t.Run("PathExp", func(t *testing.T) {
		popd, err := pushd(t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		defer popd()

		for _, name := range []string{"foo", "bar", "baz"} {
			if err := touch(name); err != nil {
				t.Fatal(err)
			}
		}

		for _, tt := range pathExpTests {
			env := interp.NewExecEnv(name)
			env.Opts = tt.opts
			g, _ := env.Expand(tt.word, 0)
			if e := tt.fields; !reflect.DeepEqual(g, e) {
				t.Errorf("expected %#v, got %#v", e, g)
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

func litf(format string, a ...any) *ast.Lit {
	return &ast.Lit{Value: fmt.Sprintf(format, a...)}
}

func quote(tok string, word ast.Word) *ast.Quote {
	return &ast.Quote{
		Tok:   tok,
		Value: word,
	}
}

func paramExp(name *ast.Lit, op string, word ast.Word) *ast.ParamExp {
	return &ast.ParamExp{
		Name: name,
		Op:   op,
		Word: word,
	}
}

func arithExp(expr ast.Word) *ast.ArithExp {
	return &ast.ArithExp{Expr: expr}
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
