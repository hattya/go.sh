//
// go.sh/printer :: printer_test.go
//
//   Copyright (c) 2018 Akinori Hattori <hattya@gmail.com>
//
//   Permission is hereby granted, free of charge, to any person
//   obtaining a copy of this software and associated documentation files
//   (the "Software"), to deal in the Software without restriction,
//   including without limitation the rights to use, copy, modify, merge,
//   publish, distribute, sublicense, and/or sell copies of the Software,
//   and to permit persons to whom the Software is furnished to do so,
//   subject to the following conditions:
//
//   The above copyright notice and this permission notice shall be
//   included in all copies or substantial portions of the Software.
//
//   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
//   EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
//   MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
//   NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS
//   BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN
//   ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
//   CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
//   SOFTWARE.
//

package printer_test

import (
	"bytes"
	"testing"

	"github.com/hattya/go.sh/ast"
	"github.com/hattya/go.sh/parser"
	"github.com/hattya/go.sh/printer"
)

var listTests = []struct {
	n ast.Node
	e string
}{
	{
		parse("cd; pwd"),
		"cd; pwd",
	},
	{
		parse("sleep 7 & wait"),
		"sleep 7 & wait",
	},
	{
		parse("cat <<EOF; echo bar\nfoo\nEOF"),
		"cat <<EOF; echo bar\nfoo\nEOF",
	},
}

func TestList(t *testing.T) {
	var b bytes.Buffer
	for _, tt := range listTests {
		b.Reset()
		if err := printer.Fprint(&b, tt.n); err != nil {
			t.Error(err)
		}
		if g, e := b.String(), tt.e; g != e {
			t.Errorf("expected %q, got %q", e, g)
		}
	}
}

var andOrListTests = []struct {
	n ast.Node
	e string
}{
	{
		parse("true && echo foo || echo bar"),
		"true && echo foo || echo bar",
	},
	{
		parse("false && cat <<EOF1 || cat <<EOF2\nfoo\nEOF1\nbar\nEOF2"),
		"false && cat <<EOF1 || cat <<EOF2\nfoo\nEOF1\nbar\nEOF2",
	},
}

func TestAndOrList(t *testing.T) {
	var b bytes.Buffer
	for _, tt := range andOrListTests {
		b.Reset()
		if err := printer.Fprint(&b, tt.n); err != nil {
			t.Error(err)
		}
		if g, e := b.String(), tt.e; g != e {
			t.Errorf("expected %q, got %q", e, g)
		}
	}
}

var pipelineTests = []struct {
	n ast.Node
	e string
}{
	{
		parse("echo foo | grep o"),
		"echo foo | grep o",
	},
	{
		parse("cat <<EOF | grep o\nfoo\nEOF"),
		"cat <<EOF | grep o\nfoo\nEOF",
	},
	{
		parse("! echo foo | grep x"),
		"! echo foo | grep x",
	},
	{
		parse("! cat <<EOF | grep x\nfoo\nEOF"),
		"! cat <<EOF | grep x\nfoo\nEOF",
	},
}

func TestPipeline(t *testing.T) {
	var b bytes.Buffer
	for _, tt := range pipelineTests {
		b.Reset()
		if err := printer.Fprint(&b, tt.n); err != nil {
			t.Error(err)
		}
		if g, e := b.String(), tt.e; g != e {
			t.Errorf("expected %q, got %q", e, g)
		}
	}
}

var simpleCmdTests = []struct {
	n ast.Node
	e []string
}{
	{
		parse("FOO=foo BAR=bar command env >/dev/null 2>&1"),
		[]string{
			"FOO=foo BAR=bar >/dev/null 2>&1 command env",
			"FOO=foo BAR=bar > /dev/null 2>&1 command env",
			">/dev/null 2>&1 FOO=foo BAR=bar command env",
			"> /dev/null 2>&1 FOO=foo BAR=bar command env",
			"FOO=foo BAR=bar command env >/dev/null 2>&1",
			"FOO=foo BAR=bar command env > /dev/null 2>&1",
		},
	},
	{
		parse("NAME=value cat <<EOF\nheredoc\nEOF"),
		[]string{
			"NAME=value <<EOF cat\nheredoc\nEOF",
			"NAME=value << EOF cat\nheredoc\nEOF",
			"<<EOF NAME=value cat\nheredoc\nEOF",
			"<< EOF NAME=value cat\nheredoc\nEOF",
			"NAME=value cat <<EOF\nheredoc\nEOF",
			"NAME=value cat << EOF\nheredoc\nEOF",
		},
	},
}

func TestSimpleCmd(t *testing.T) {
	var b bytes.Buffer
	for _, tt := range simpleCmdTests {
		for i, cfg := range []printer.Config{
			{
				Redir:  printer.Before,
				Assign: printer.Before,
			},
			{
				Redir:  printer.Before | printer.Space,
				Assign: printer.Before,
			},
			{
				Redir:  printer.Before,
				Assign: printer.After,
			},
			{
				Redir:  printer.Before | printer.Space,
				Assign: printer.After,
			},
			{
				Redir: printer.After,
			},
			{
				Redir: printer.After | printer.Space,
			},
		} {
			b.Reset()
			if err := cfg.Fprint(&b, tt.n); err != nil {
				t.Error(err)
			}
			if g, e := b.String(), tt.e[i]; g != e {
				t.Errorf("expected %q, got %q", e, g)
			}
		}
	}
}

var subshellTests = []struct {
	n ast.Node
	e []string
}{
	{
		parse("(cd /usr/src/linux; make -j3) >/dev/null 2>&1"),
		[]string{
			"(cd /usr/src/linux; make -j3) >/dev/null 2>&1",
			"(cd /usr/src/linux; make -j3) >/dev/null 2>&1",
		},
	},
	{
		parse("(\n\tcd /usr/src/linux; make -j3\n) >/dev/null 2>&1"),
		[]string{
			"(\n\tcd /usr/src/linux; make -j3\n) >/dev/null 2>&1",
			"(\n  cd /usr/src/linux; make -j3\n) >/dev/null 2>&1",
		},
	},
	{
		parse("(\n\tcd /usr/src/linux\n\tmake -j3\n) >/dev/null 2>&1"),
		[]string{
			"(\n\tcd /usr/src/linux\n\tmake -j3\n) >/dev/null 2>&1",
			"(\n  cd /usr/src/linux\n  make -j3\n) >/dev/null 2>&1",
		},
	},
	{
		parse("(cat <<EOF; echo bar)\nfoo\nEOF"),
		[]string{
			"(cat <<EOF; echo bar)\nfoo\nEOF",
			"(cat <<EOF; echo bar)\nfoo\nEOF",
		},
	},
	{
		parse("(\n\tcat <<EOF; echo bar\nfoo\nEOF\n)"),
		[]string{
			"(\n\tcat <<EOF; echo bar\nfoo\nEOF\n)",
			"(\n  cat <<EOF; echo bar\nfoo\nEOF\n)",
		},
	},
	{
		parse("(\n\tcat <<EOF\nfoo\nEOF\n\techo bar\n)"),
		[]string{
			"(\n\tcat <<EOF\nfoo\nEOF\n\techo bar\n)",
			"(\n  cat <<EOF\nfoo\nEOF\n  echo bar\n)",
		},
	},
}

func TestSubshell(t *testing.T) {
	var b bytes.Buffer
	for _, tt := range subshellTests {
		for i, cfg := range []printer.Config{
			{
				Indent: printer.Tab,
			},
			{
				Indent: printer.Space,
				Width:  2,
			},
		} {
			b.Reset()
			if err := cfg.Fprint(&b, tt.n); err != nil {
				t.Error(err)
			}
			if g, e := b.String(), tt.e[i]; g != e {
				t.Errorf("expected %q, got %q", e, g)
			}
		}
	}
}

var groupTests = []struct {
	n ast.Node
	e []string
}{
	{
		parse("{ ./configure; make -j3; } >/dev/null 2>&1"),
		[]string{
			"{ ./configure; make -j3; } >/dev/null 2>&1",
			"{ ./configure; make -j3; } >/dev/null 2>&1",
		},
	},
	{
		parse("{\n\t./configure; make -j3;\n} >/dev/null 2>&1"),
		[]string{
			"{\n\t./configure; make -j3;\n} >/dev/null 2>&1",
			"{\n  ./configure; make -j3;\n} >/dev/null 2>&1",
		},
	},
	{
		parse("{\n\t./configure\n\tmake -j3;\n} >/dev/null 2>&1"),
		[]string{
			"{\n\t./configure\n\tmake -j3;\n} >/dev/null 2>&1",
			"{\n  ./configure\n  make -j3;\n} >/dev/null 2>&1",
		},
	},
	{
		parse("{ cat <<EOF; echo bar; }\nfoo\nEOF"),
		[]string{
			"{ cat <<EOF; echo bar; }\nfoo\nEOF",
			"{ cat <<EOF; echo bar; }\nfoo\nEOF",
		},
	},
	{
		parse("{\n\tcat <<EOF; echo bar\nfoo\nEOF\n}"),
		[]string{
			"{\n\tcat <<EOF; echo bar\nfoo\nEOF\n}",
			"{\n  cat <<EOF; echo bar\nfoo\nEOF\n}",
		},
	},
	{
		parse("{\n\tcat <<EOF\nfoo\nEOF\n\techo bar\n}"),
		[]string{
			"{\n\tcat <<EOF\nfoo\nEOF\n\techo bar\n}",
			"{\n  cat <<EOF\nfoo\nEOF\n  echo bar\n}",
		},
	},
}

func TestGroup(t *testing.T) {
	var b bytes.Buffer
	for _, tt := range groupTests {
		for i, cfg := range []printer.Config{
			{
				Indent: printer.Tab,
			},
			{
				Indent: printer.Space,
				Width:  2,
			},
		} {
			b.Reset()
			if err := cfg.Fprint(&b, tt.n); err != nil {
				t.Error(err)
			}
			if g, e := b.String(), tt.e[i]; g != e {
				t.Errorf("expected %q, got %q", e, g)
			}
		}
	}
}

var forClauseTests = []struct {
	n ast.Node
	e []string
}{
	{
		parse("for name do echo $name; done >/dev/null 2>&1"),
		[]string{
			"for name do echo $name; done >/dev/null 2>&1",
			"for name do echo $name; done >/dev/null 2>&1",
		},
	},
	{
		parse("for name; do echo $name; done >/dev/null 2>&1"),
		[]string{
			"for name do echo $name; done >/dev/null 2>&1",
			"for name do echo $name; done >/dev/null 2>&1",
		},
	},
	{
		parse("for name do\n\techo $name\ndone >/dev/null 2>&1"),
		[]string{
			"for name do\n\techo $name\ndone >/dev/null 2>&1",
			"for name\ndo\n\techo $name\ndone >/dev/null 2>&1",
		},
	},
	{
		parse("for name; do\n\techo $name\ndone >/dev/null 2>&1"),
		[]string{
			"for name do\n\techo $name\ndone >/dev/null 2>&1",
			"for name\ndo\n\techo $name\ndone >/dev/null 2>&1",
		},
	},
	{
		parse("for name\ndo\n\techo $name\ndone >/dev/null 2>&1"),
		[]string{
			"for name do\n\techo $name\ndone >/dev/null 2>&1",
			"for name\ndo\n\techo $name\ndone >/dev/null 2>&1",
		},
	},
	{
		parse("for name in; do echo $name; done >/dev/null 2>&1"),
		[]string{
			"for name in; do echo $name; done >/dev/null 2>&1",
			"for name in; do echo $name; done >/dev/null 2>&1",
		},
	},
	{
		parse("for name in; do\n\techo $name\ndone >/dev/null 2>&1"),
		[]string{
			"for name in; do\n\techo $name\ndone >/dev/null 2>&1",
			"for name in\ndo\n\techo $name\ndone >/dev/null 2>&1",
		},
	},
	{
		parse("for name in\ndo\n\techo $name\ndone >/dev/null 2>&1"),
		[]string{
			"for name in; do\n\techo $name\ndone >/dev/null 2>&1",
			"for name in\ndo\n\techo $name\ndone >/dev/null 2>&1",
		},
	},
	{
		parse("for name in foo bar baz; do echo $name; done >/dev/null 2>&1"),
		[]string{
			"for name in foo bar baz; do echo $name; done >/dev/null 2>&1",
			"for name in foo bar baz; do echo $name; done >/dev/null 2>&1",
		},
	},
	{
		parse("for name in foo bar baz; do\n\techo $name\ndone >/dev/null 2>&1"),
		[]string{
			"for name in foo bar baz; do\n\techo $name\ndone >/dev/null 2>&1",
			"for name in foo bar baz\ndo\n\techo $name\ndone >/dev/null 2>&1",
		},
	},
	{
		parse("for name in foo bar baz\ndo\n\techo $name\ndone >/dev/null 2>&1"),
		[]string{
			"for name in foo bar baz; do\n\techo $name\ndone >/dev/null 2>&1",
			"for name in foo bar baz\ndo\n\techo $name\ndone >/dev/null 2>&1",
		},
	},
	{
		parse("for name do cat <<EOF; echo $name; done\nfoo\nEOF"),
		[]string{
			"for name do cat <<EOF; echo $name; done\nfoo\nEOF",
			"for name do cat <<EOF; echo $name; done\nfoo\nEOF",
		},
	},
	{
		parse("for name do\n\tcat <<EOF\nfoo\nEOF\n\techo $name\ndone"),
		[]string{
			"for name do\n\tcat <<EOF\nfoo\nEOF\n\techo $name\ndone",
			"for name\ndo\n\tcat <<EOF\nfoo\nEOF\n\techo $name\ndone",
		},
	},
}

func TestForClause(t *testing.T) {
	var b bytes.Buffer
	for _, tt := range forClauseTests {
		for i, cfg := range []printer.Config{
			{
				Do: 0,
			},
			{
				Do: printer.Newline,
			},
		} {
			b.Reset()
			if err := cfg.Fprint(&b, tt.n); err != nil {
				t.Error(err)
			}
			if g, e := b.String(), tt.e[i]; g != e {
				t.Errorf("expected %q, got %q", e, g)
			}
		}
	}
}

var ifClauseTests = []struct {
	n ast.Node
	e []string
}{
	{
		parse("if true && false; false; then echo if; elif true; false || false; then echo elif; else echo else; fi >/dev/null 2>&1"),
		[]string{
			"if true && false; false; then echo if; elif true; false || false; then echo elif; else echo else; fi >/dev/null 2>&1",
			"if true && false; false; then echo if; elif true; false || false; then echo elif; else echo else; fi >/dev/null 2>&1",
		},
	},
	{
		parse("if true && false; false; then\n\techo if\nelif true; false || false; then\n\techo elif\nelse\n\techo else\nfi >/dev/null 2>&1"),
		[]string{
			"if true && false; false; then\n\techo if\nelif true; false || false; then\n\techo elif\nelse\n\techo else\nfi >/dev/null 2>&1",
			"if true && false; false\nthen\n\techo if\nelif true; false || false\nthen\n\techo elif\nelse\n\techo else\nfi >/dev/null 2>&1",
		},
	},
	{
		parse("if true && false\n\tfalse\nthen\n\techo if\nelif true\n\tfalse || false\nthen\n\techo elif\nelse\n\techo else\nfi >/dev/null 2>&1"),
		[]string{
			"if true && false; false; then\n\techo if\nelif true; false || false; then\n\techo elif\nelse\n\techo else\nfi >/dev/null 2>&1",
			"if true && false; false\nthen\n\techo if\nelif true; false || false\nthen\n\techo elif\nelse\n\techo else\nfi >/dev/null 2>&1",
		},
	},
	{
		parse("if false; then echo if; cat <<EOF1; elif false; then echo elif; cat <<EOF2; else echo else; cat <<EOF3; fi\nfoo\nEOF1\nbar\nEOF2\nbaz\nEOF3"),
		[]string{
			"if false; then echo if; cat <<EOF1; elif false; then echo elif; cat <<EOF2; else echo else; cat <<EOF3; fi\nfoo\nEOF1\nbar\nEOF2\nbaz\nEOF3",
			"if false; then echo if; cat <<EOF1; elif false; then echo elif; cat <<EOF2; else echo else; cat <<EOF3; fi\nfoo\nEOF1\nbar\nEOF2\nbaz\nEOF3",
		},
	},
	{
		parse("if false\nthen\n\tcat <<EOF\n1\nEOF\n\techo 2\nelif false; then\n\tcat <<EOF\n2\nEOF\n\techo 3\nelse\n\tcat <<EOF\n3\nEOF\n\techo 4\nfi"),
		[]string{
			"if false; then\n\tcat <<EOF\n1\nEOF\n\techo 2\nelif false; then\n\tcat <<EOF\n2\nEOF\n\techo 3\nelse\n\tcat <<EOF\n3\nEOF\n\techo 4\nfi",
			"if false\nthen\n\tcat <<EOF\n1\nEOF\n\techo 2\nelif false\nthen\n\tcat <<EOF\n2\nEOF\n\techo 3\nelse\n\tcat <<EOF\n3\nEOF\n\techo 4\nfi",
		},
	},
	{
		parse("if cat <<EOF1 | grep x; then echo if; elif cat <<EOF2 | grep x; then echo elif; fi\nfoo\nEOF1\nbar\nEOF2"),
		[]string{
			"if cat <<EOF1 | grep x; then echo if; elif cat <<EOF2 | grep x; then echo elif; fi\nfoo\nEOF1\nbar\nEOF2",
			"if cat <<EOF1 | grep x; then echo if; elif cat <<EOF2 | grep x; then echo elif; fi\nfoo\nEOF1\nbar\nEOF2",
		},
	},
	{
		parse("if cat <<EOF | grep x\nfoo\nEOF\nthen\n\techo if\nelif cat <<EOF | grep x\nbar\nEOF\nthen\n\techo elif\nfi"),
		[]string{
			"if cat <<EOF | grep x; then\nfoo\nEOF\n\techo if\nelif cat <<EOF | grep x; then\nbar\nEOF\n\techo elif\nfi",
			"if cat <<EOF | grep x\nfoo\nEOF\nthen\n\techo if\nelif cat <<EOF | grep x\nbar\nEOF\nthen\n\techo elif\nfi",
		},
	},
}

func TestIfClause(t *testing.T) {
	var b bytes.Buffer
	for _, tt := range ifClauseTests {
		for i, cfg := range []printer.Config{
			{
				Then: 0,
			},
			{
				Then: printer.Newline,
			},
		} {
			b.Reset()
			if err := cfg.Fprint(&b, tt.n); err != nil {
				t.Error(err)
			}
			if g, e := b.String(), tt.e[i]; g != e {
				t.Errorf("expected %q, got %q", e, g)
			}
		}
	}
}

var whileClauseTests = []struct {
	n ast.Node
	e []string
}{
	{
		parse("while false; true && true; do echo while; done >/dev/null 2>&1"),
		[]string{
			"while false; true && true; do echo while; done >/dev/null 2>&1",
			"while false; true && true; do echo while; done >/dev/null 2>&1",
		},
	},
	{
		parse("while false; true && true; do\n\techo while\ndone >/dev/null 2>&1"),
		[]string{
			"while false; true && true; do\n\techo while\ndone >/dev/null 2>&1",
			"while false; true && true\ndo\n\techo while\ndone >/dev/null 2>&1",
		},
	},
	{
		parse("while false\n\ttrue && true\ndo\n\techo while\ndone >/dev/null 2>&1"),
		[]string{
			"while false; true && true; do\n\techo while\ndone >/dev/null 2>&1",
			"while false; true && true\ndo\n\techo while\ndone >/dev/null 2>&1",
		},
	},
	{
		parse("while true; do cat <<EOF; echo bar; done\nfoo\nEOF"),
		[]string{
			"while true; do cat <<EOF; echo bar; done\nfoo\nEOF",
			"while true; do cat <<EOF; echo bar; done\nfoo\nEOF",
		},
	},
	{
		parse("while true\ndo\n\tcat <<EOF\nfoo\nEOF\n\techo bar\ndone"),
		[]string{
			"while true; do\n\tcat <<EOF\nfoo\nEOF\n\techo bar\ndone",
			"while true\ndo\n\tcat <<EOF\nfoo\nEOF\n\techo bar\ndone",
		},
	},
	{
		parse("while cat <<EOF | grep o; do echo while; done\nfoo\nEOF"),
		[]string{
			"while cat <<EOF | grep o; do echo while; done\nfoo\nEOF",
			"while cat <<EOF | grep o; do echo while; done\nfoo\nEOF",
		},
	},
	{
		parse("while cat <<EOF | grep o\nfoo\nEOF\ndo\n\techo while\ndone"),
		[]string{
			"while cat <<EOF | grep o; do\nfoo\nEOF\n\techo while\ndone",
			"while cat <<EOF | grep o\nfoo\nEOF\ndo\n\techo while\ndone",
		},
	},
}

func TestWhileClause(t *testing.T) {
	var b bytes.Buffer
	for _, tt := range whileClauseTests {
		for i, cfg := range []printer.Config{
			{
				Do: 0,
			},
			{
				Do: printer.Newline,
			},
		} {
			b.Reset()
			if err := cfg.Fprint(&b, tt.n); err != nil {
				t.Error(err)
			}
			if g, e := b.String(), tt.e[i]; g != e {
				t.Errorf("expected %q, got %q", e, g)
			}
		}
	}
}

func parse(src string) ast.Command {
	cmd, _, err := parser.ParseCommand("<stdin>", src)
	if err != nil {
		panic(err)
	}
	return cmd
}

var wordTests = []struct {
	n ast.Node
	e string
}{
	{
		&ast.Lit{
			ValuePos: ast.NewPos(1, 1),
			Value:    "lit",
		},
		"lit",
	},
	{
		ast.Word{
			&ast.Lit{
				ValuePos: ast.NewPos(1, 1),
				Value:    "1",
			},
			&ast.Lit{
				ValuePos: ast.NewPos(1, 3),
				Value:    "2",
			},
			&ast.Lit{
				ValuePos: ast.NewPos(2, 2),
				Value:    "3",
			},
		},
		"123",
	},
	// quoting
	{
		&ast.Quote{
			TokPos: ast.NewPos(1, 1),
			Tok:    `\`,
			Value: ast.Word{
				&ast.Lit{
					ValuePos: ast.NewPos(1, 2),
					Value:    "q",
				},
			},
		},
		`\q`,
	},
	{
		&ast.Quote{
			TokPos: ast.NewPos(1, 1),
			Tok:    `'`,
			Value: ast.Word{
				&ast.Lit{
					ValuePos: ast.NewPos(1, 2),
					Value:    "quote",
				},
			},
		},
		`'quote'`,
	},
	{
		&ast.Quote{
			TokPos: ast.NewPos(1, 1),
			Tok:    `"`,
			Value: ast.Word{
				&ast.Lit{
					ValuePos: ast.NewPos(1, 2),
					Value:    "quote",
				},
			},
		},
		`"quote"`,
	},
	{
		&ast.Quote{
			TokPos: ast.NewPos(1, 1),
			Tok:    `'`,
			Value: ast.Word{
				&ast.Lit{
					ValuePos: ast.NewPos(1, 2),
					Value:    "qu",
				},
				&ast.Lit{
					ValuePos: ast.NewPos(2, 3),
					Value:    "o",
				},
				&ast.Lit{
					ValuePos: ast.NewPos(3, 4),
					Value:    "te",
				},
			},
		},
		"'quote'",
	},
	{
		ast.Word{
			&ast.Lit{
				ValuePos: ast.NewPos(1, 1),
				Value:    "qu",
			},
			&ast.Quote{
				TokPos: ast.NewPos(1, 3),
				Tok:    `\`,
				Value: ast.Word{
					&ast.Lit{
						ValuePos: ast.NewPos(1, 4),
						Value:    "o",
					},
				},
			},
			&ast.Lit{
				ValuePos: ast.NewPos(1, 5),
				Value:    "te",
			},
		},
		`qu\ote`,
	},
	{
		ast.Word{
			&ast.Lit{
				ValuePos: ast.NewPos(1, 1),
				Value:    "qu",
			},
			&ast.Quote{
				TokPos: ast.NewPos(1, 3),
				Tok:    `'`,
				Value: ast.Word{
					&ast.Lit{
						ValuePos: ast.NewPos(1, 4),
						Value:    "o",
					},
				},
			},
			&ast.Lit{
				ValuePos: ast.NewPos(1, 6),
				Value:    "te",
			},
		},
		`qu'o'te`,
	},
	{
		ast.Word{
			&ast.Lit{
				ValuePos: ast.NewPos(1, 1),
				Value:    "qu",
			},
			&ast.Quote{
				TokPos: ast.NewPos(1, 3),
				Tok:    `"`,
				Value: ast.Word{
					&ast.Lit{
						ValuePos: ast.NewPos(1, 4),
						Value:    "o",
					},
				},
			},
			&ast.Lit{
				ValuePos: ast.NewPos(1, 6),
				Value:    "te",
			},
		},
		`qu"o"te`,
	},
	// parameter expansion
	{
		&ast.ParamExp{
			Dollar: ast.NewPos(1, 1),
			Name: &ast.Lit{
				ValuePos: ast.NewPos(1, 2),
				Value:    "@",
			},
		},
		"$@",
	},
	{
		&ast.ParamExp{
			Dollar: ast.NewPos(1, 1),
			Braces: true,
			Name: &ast.Lit{
				ValuePos: ast.NewPos(1, 3),
				Value:    "@",
			},
		},
		"${@}",
	},
	{
		&ast.ParamExp{
			Dollar: ast.NewPos(1, 1),
			Braces: true,
			Name: &ast.Lit{
				ValuePos: ast.NewPos(1, 4),
				Value:    "LANG",
			},
			OpPos: ast.NewPos(1, 3),
			Op:    "#",
		},
		"${#LANG}",
	},
	{
		&ast.ParamExp{
			Dollar: ast.NewPos(1, 1),
			Braces: true,
			Name: &ast.Lit{
				ValuePos: ast.NewPos(1, 3),
				Value:    "LANG",
			},
			OpPos: ast.NewPos(1, 7),
			Op:    "#",
			Word: ast.Word{
				&ast.Lit{
					ValuePos: ast.NewPos(1, 8),
					Value:    "*.",
				},
			},
		},
		"${LANG#*.}",
	},
	// command substitution
	{
		&ast.CmdSubst{
			Dollar: true,
			Left:   ast.NewPos(1, 2),
			List: []ast.Command{
				&ast.Cmd{
					Expr: &ast.SimpleCmd{
						Args: []ast.Word{
							{
								&ast.Lit{
									ValuePos: ast.NewPos(1, 3),
									Value:    "pwd",
								},
							},
						},
					},
				},
			},
			Right: ast.NewPos(1, 6),
		},
		"$(pwd)",
	},
	{
		&ast.CmdSubst{
			Dollar: true,
			Left:   ast.NewPos(1, 2),
			List: []ast.Command{
				&ast.Cmd{
					Expr: &ast.SimpleCmd{
						Args: []ast.Word{
							{
								&ast.Lit{
									ValuePos: ast.NewPos(2, 2),
									Value:    "pwd",
								},
							},
						},
					},
				},
			},
			Right: ast.NewPos(3, 1),
		},
		"$(\n\tpwd\n)",
	},
	{
		&ast.CmdSubst{
			Left: ast.NewPos(1, 1),
			List: []ast.Command{
				&ast.Cmd{
					Expr: &ast.SimpleCmd{
						Args: []ast.Word{
							{
								&ast.Lit{
									ValuePos: ast.NewPos(1, 2),
									Value:    "pwd",
								},
							},
						},
					},
				},
			},
			Right: ast.NewPos(1, 5),
		},
		"`pwd`",
	},
	{
		&ast.CmdSubst{
			Left: ast.NewPos(1, 1),
			List: []ast.Command{
				&ast.Cmd{
					Expr: &ast.SimpleCmd{
						Args: []ast.Word{
							{
								&ast.Lit{
									ValuePos: ast.NewPos(2, 2),
									Value:    "pwd",
								},
							},
						},
					},
				},
			},
			Right: ast.NewPos(3, 1),
		},
		"`\n\tpwd\n`",
	},
}

func TestWord(t *testing.T) {
	var b bytes.Buffer
	for _, tt := range wordTests {
		b.Reset()
		if err := printer.Fprint(&b, tt.n); err != nil {
			t.Error(err)
		}
		if g, e := b.String(), tt.e; g != e {
			t.Errorf("expected %q, got %q", e, g)
		}
	}
}

func TestComment(t *testing.T) {
	var b bytes.Buffer
	n := &ast.Comment{
		Hash: ast.NewPos(1, 1),
		Text: " comment",
	}
	if err := printer.Fprint(&b, n); err != nil {
		t.Error(err)
	}
	if g, e := b.String(), "# comment"; g != e {
		t.Errorf("expected %q, got %q", e, g)
	}
}

func TestPrintError(t *testing.T) {
	for _, n := range []ast.Node{nil, &ast.Cmd{Expr: &ast.Subshell{List: []ast.Command{nil}}}, new(ast.Cmd), ast.Word{nil}} {
		if err := printer.Fprint(nil, n); err == nil {
			t.Error("expected error")
		}
	}
}
