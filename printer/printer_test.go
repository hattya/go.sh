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
