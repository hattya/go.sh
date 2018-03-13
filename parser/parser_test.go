//
// go.sh/parser :: parser_test.go
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

package parser_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"reflect"
	"testing"

	"github.com/hattya/go.sh/ast"
	"github.com/hattya/go.sh/parser"
)

var parseCommandTests = []struct {
	src      string
	cmd      ast.Command
	comments []*ast.Comment
}{
	// <blank>
	{
		src: "",
	},
	{
		src: "\n",
	},
	{
		src: "\t ",
	},
	{
		src: "echo 1\t\t22 \t 333",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(lit(1, 6, "1")),
			word(lit(1, 9, "22")),
			word(lit(1, 14, "333")),
		),
	},
	{
		src: "echo\t1  22\t \t333\n",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(lit(1, 6, "1")),
			word(lit(1, 9, "22")),
			word(lit(1, 14, "333")),
		),
	},
	// quoting
	{
		src: "\\pwd\n",
		cmd: simple_command(
			word(quote(1, 1, "\\", word(lit(1, 2, "p"))), lit(1, 3, "wd")),
		),
	},
	{
		src: "p\\wd\n",
		cmd: simple_command(
			word(lit(1, 1, "p"), quote(1, 2, "\\", word(lit(1, 3, "w"))), lit(1, 4, "d")),
		),
	},
	{
		src: "pw\\d\n",
		cmd: simple_command(
			word(lit(1, 1, "pw"), quote(1, 3, "\\", word(lit(1, 4, "d")))),
		),
	},
	{
		src: "pwd\\\n",
		cmd: simple_command(
			word(lit(1, 1, "pwd")),
		),
	},
	{
		src: "pwd\\",
		cmd: simple_command(
			word(lit(1, 1, "pwd"), quote(1, 4, "\\", nil)),
		),
	},
	{
		src: "echo 'foo bar baz'",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(quote(1, 6, "'", word(lit(1, 7, "foo bar baz")))),
		),
	},
	// <newline>
	{
		src: "echo 1\necho 2\n",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(lit(1, 6, "1")),
		),
	},
	// simple command
	{
		src: "FOO=foo",
		cmd: simple_command(
			assignment_word(1, 1, "FOO", word(lit(1, 5, "foo"))),
		),
	},
	{
		src: "FOO=foo BAR=bar env",
		cmd: simple_command(
			assignment_word(1, 1, "FOO", word(lit(1, 5, "foo"))),
			assignment_word(1, 9, "BAR", word(lit(1, 13, "bar"))),
			word(lit(1, 17, "env")),
		),
	},
	{
		src: "FOO=foo BAR=bar env -u BAR BAZ=baz",
		cmd: simple_command(
			assignment_word(1, 1, "FOO", word(lit(1, 5, "foo"))),
			assignment_word(1, 9, "BAR", word(lit(1, 13, "bar"))),
			word(lit(1, 17, "env")),
			word(lit(1, 21, "-u")),
			word(lit(1, 24, "BAR")),
			word(lit(1, 28, "BAZ=baz")),
		),
	},
	{
		src: "123=123",
		cmd: simple_command(
			word(lit(1, 1, "123=123")),
		),
	},
	{
		src: "=env",
		cmd: simple_command(
			word(lit(1, 1, "=env")),
		),
	},
	{
		src: "cat <file",
		cmd: simple_command(
			word(lit(1, 1, "cat")),
			redir(nil, 1, 5, "<", word(lit(1, 6, "file"))),
		),
	},
	{
		src: "echo foo >file",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(lit(1, 6, "foo")),
			redir(nil, 1, 10, ">", word(lit(1, 11, "file"))),
		),
	},
	{
		src: "echo baz >|file",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(lit(1, 6, "baz")),
			redir(nil, 1, 10, ">|", word(lit(1, 12, "file"))),
		),
	},
	{
		src: "echo bar >>file",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(lit(1, 6, "bar")),
			redir(nil, 1, 10, ">>", word(lit(1, 12, "file"))),
		),
	},
	{
		src: "exec 3<&-",
		cmd: simple_command(
			word(lit(1, 1, "exec")),
			redir(lit(1, 6, "3"), 1, 7, "<&", word(lit(1, 9, "-"))),
		),
	},
	{
		src: "echo foo >&1",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(lit(1, 6, "foo")),
			redir(nil, 1, 10, ">&", word(lit(1, 12, "1"))),
		),
	},
	{
		src: "exec 3<>file",
		cmd: simple_command(
			word(lit(1, 1, "exec")),
			redir(lit(1, 6, "3"), 1, 7, "<>", word(lit(1, 9, "file"))),
		),
	},
	{
		src: "echo foo>&2",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(lit(1, 6, "foo")),
			redir(nil, 1, 9, ">&", word(lit(1, 11, "2"))),
		),
	},
	{
		src: "echo \\foo>&2",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(quote(1, 6, "\\", word(lit(1, 7, "f"))), lit(1, 8, "oo")),
			redir(nil, 1, 10, ">&", word(lit(1, 12, "2"))),
		),
	},
	{
		src: ">/dev/null 2>&1 echo foo",
		cmd: simple_command(
			redir(nil, 1, 1, ">", word(lit(1, 2, "/dev/null"))),
			redir(lit(1, 12, "2"), 1, 13, ">&", word(lit(1, 15, "1"))),
			word(lit(1, 17, "echo")),
			word(lit(1, 22, "foo")),
		),
	},
	{
		src: ">/dev/null echo foo 2>&1",
		cmd: simple_command(
			redir(nil, 1, 1, ">", word(lit(1, 2, "/dev/null"))),
			word(lit(1, 12, "echo")),
			word(lit(1, 17, "foo")),
			redir(lit(1, 21, "2"), 1, 22, ">&", word(lit(1, 24, "1"))),
		),
	},
	{
		src: "echo foo >/dev/null 2>&1",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(lit(1, 6, "foo")),
			redir(nil, 1, 10, ">", word(lit(1, 11, "/dev/null"))),
			redir(lit(1, 21, "2"), 1, 22, ">&", word(lit(1, 24, "1"))),
		),
	},
	// comment
	{
		src: "# comment",
		comments: []*ast.Comment{
			comment(1, 1, " comment"),
		},
	},
	{
		src: "# comment\n\ngo version",
		cmd: simple_command(
			word(lit(3, 1, "go")),
			word(lit(3, 4, "version")),
		),
		comments: []*ast.Comment{
			comment(1, 1, " comment"),
		},
	},
	{
		src: "go version# comment\n",
		cmd: simple_command(
			word(lit(1, 1, "go")),
			word(lit(1, 4, "version")),
		),
		comments: []*ast.Comment{
			comment(1, 11, " comment"),
		},
	},
	// pipeline
	{
		src: "echo foo | grep o",
		cmd: pipeline(
			simple_command(
				word(lit(1, 1, "echo")),
				word(lit(1, 6, "foo")),
			),
			pipe(1, 10, "|", simple_command(
				word(lit(1, 12, "grep")),
				word(lit(1, 17, "o")),
			)),
		),
	},
	{
		src: "! echo foo | grep x",
		cmd: pipeline(
			pos(1, 1), // !
			simple_command(
				word(lit(1, 3, "echo")),
				word(lit(1, 8, "foo")),
			),
			pipe(1, 12, "|", simple_command(
				word(lit(1, 14, "grep")),
				word(lit(1, 19, "x")),
			)),
		),
	},
	{
		src: "! echo foo |\n\n# | comment\n\ngrep x",
		cmd: pipeline(
			pos(1, 1), // !
			simple_command(
				word(lit(1, 3, "echo")),
				word(lit(1, 8, "foo")),
			),
			pipe(1, 12, "|", simple_command(
				word(lit(5, 1, "grep")),
				word(lit(5, 6, "x")),
			)),
		),
		comments: []*ast.Comment{
			comment(3, 1, " | comment"),
		},
	},
	// list
	{
		src: "sleep 1;",
		cmd: list(
			simple_command(
				word(lit(1, 1, "sleep")),
				word(lit(1, 7, "1")),
			),
			sep(1, 8, ";"),
		),
	},
	{
		src: "cd; pwd",
		cmd: list(
			simple_command(
				word(lit(1, 1, "cd")),
			),
			sep(1, 3, ";"),
		),
	},
	{
		src: "sleep 1 &",
		cmd: list(
			simple_command(
				word(lit(1, 1, "sleep")),
				word(lit(1, 7, "1")),
			),
			sep(1, 9, "&"),
		),
	},
	{
		src: "make & fg",
		cmd: list(
			simple_command(
				word(lit(1, 1, "make")),
			),
			sep(1, 6, "&"),
		),
	},
	{
		src: "false && echo foo || echo bar",
		cmd: list(
			simple_command(
				word(lit(1, 1, "false")),
			),
			and_or(1, 7, "&&", pipeline(
				simple_command(
					word(lit(1, 10, "echo")),
					word(lit(1, 15, "foo")),
				),
			)),
			and_or(1, 19, "||", pipeline(
				simple_command(
					word(lit(1, 22, "echo")),
					word(lit(1, 27, "bar")),
				),
			)),
		),
	},
	{
		src: "true || echo foo && echo bar",
		cmd: list(
			simple_command(
				word(lit(1, 1, "true")),
			),
			and_or(1, 6, "||", pipeline(
				simple_command(
					word(lit(1, 9, "echo")),
					word(lit(1, 14, "foo")),
				),
			)),
			and_or(1, 18, "&&", pipeline(
				simple_command(
					word(lit(1, 21, "echo")),
					word(lit(1, 26, "bar")),
				),
			)),
		),
	},
	{
		src: "true ||\n\n# || comment\n\necho foo &&\n\n# && comment\n\necho bar",
		cmd: list(
			simple_command(
				word(lit(1, 1, "true")),
			),
			and_or(1, 6, "||", pipeline(
				simple_command(
					word(lit(5, 1, "echo")),
					word(lit(5, 6, "foo")),
				),
			)),
			and_or(5, 10, "&&", pipeline(
				simple_command(
					word(lit(9, 1, "echo")),
					word(lit(9, 6, "bar")),
				),
			)),
		),
		comments: []*ast.Comment{
			comment(3, 1, " || comment"),
			comment(7, 1, " && comment"),
		},
	},
}

func TestParseCommand(t *testing.T) {
	for i, tt := range parseCommandTests {
		cmd, comments, err := parser.ParseCommand(fmt.Sprintf("%v.sh", i), tt.src)
		switch {
		case err != nil:
			t.Error(err)
		case !reflect.DeepEqual(cmd, tt.cmd):
			t.Errorf("unexpected command for %q", tt.src)
		case !reflect.DeepEqual(comments, tt.comments):
			t.Errorf("unexpected comments for %q", tt.src)
		}
	}
}

var pos = ast.NewPos

func sep(line, col int, sep string) *ast.Lit {
	return &ast.Lit{
		ValuePos: ast.NewPos(line, col),
		Value:    sep,
	}
}

func list(nodes ...ast.Node) *ast.List {
	cmd := new(ast.List)
	for _, n := range nodes {
		switch n := n.(type) {
		case *ast.Cmd:
			cmd.Pipeline = &ast.Pipeline{Cmd: n}
		case *ast.AndOr:
			cmd.List = append(cmd.List, n)
		case *ast.Lit:
			cmd.SepPos = n.ValuePos
			cmd.Sep = n.Value
		}
	}
	return cmd
}

func and_or(line, col int, op string, pipeline *ast.Pipeline) *ast.AndOr {
	return &ast.AndOr{
		OpPos:    ast.NewPos(line, col),
		Op:       op,
		Pipeline: pipeline,
	}
}

func pipeline(args ...interface{}) *ast.Pipeline {
	cmd := new(ast.Pipeline)
	for _, a := range args {
		switch a := a.(type) {
		case ast.Pos:
			cmd.Bang = a
		case *ast.Cmd:
			cmd.Cmd = a
		case *ast.Pipe:
			cmd.List = append(cmd.List, a)
		}
	}
	return cmd
}

func pipe(line, col int, op string, cmd *ast.Cmd) *ast.Pipe {
	return &ast.Pipe{
		OpPos: ast.NewPos(line, col),
		Op:    op,
		Cmd:   cmd,
	}
}

func simple_command(nodes ...ast.Node) *ast.Cmd {
	x := new(ast.SimpleCmd)
	cmd := &ast.Cmd{Expr: x}
	for _, n := range nodes {
		switch n := n.(type) {
		case *ast.Assign:
			x.Assigns = append(x.Assigns, n)
		case ast.Word:
			x.Args = append(x.Args, n)
		case *ast.Redir:
			cmd.Redirs = append(cmd.Redirs, n)
		}
	}
	return cmd
}

func redir(n *ast.Lit, line, col int, op string, word ast.Word) *ast.Redir {
	return &ast.Redir{
		N:     n,
		OpPos: ast.NewPos(line, col),
		Op:    op,
		Word:  word,
	}
}

func word(w ...ast.WordPart) ast.Word {
	return ast.Word(w)
}

func lit(line, col int, v string) *ast.Lit {
	return &ast.Lit{
		ValuePos: ast.NewPos(line, col),
		Value:    v,
	}
}

func quote(line, col int, tok string, v ast.Word) *ast.Quote {
	return &ast.Quote{
		TokPos: ast.NewPos(line, col),
		Tok:    tok,
		Value:  v,
	}
}

func assignment_word(line, col int, k string, v ast.Word) *ast.Assign {
	return &ast.Assign{
		Symbol: &ast.Lit{
			ValuePos: ast.NewPos(line, col),
			Value:    k,
		},
		Op:    "=",
		Value: v,
	}
}

func comment(line, col int, text string) *ast.Comment {
	return &ast.Comment{
		Hash: ast.NewPos(line, col),
		Text: text,
	}
}

var parseErrorTests = []struct {
	src, err string
}{
	// quoting
	{
		src: "'q",
		err: ":1:1: syntax error: reached EOF while parsing single-quotes",
	},
	// simple command
	{
		src: "<",
		err: ":1:1: syntax error: unexpected EOF, expecting WORD",
	},
	{
		src: ">",
		err: ":1:1: syntax error: unexpected EOF, expecting WORD",
	},
	// pipeline
	{
		src: "!",
		err: ":1:1: syntax error: unexpected EOF",
	},
	{
		src: "echo foo | !",
		err: ":1:12: syntax error: unexpected '!'",
	},
	// list
	{
		src: ";",
		err: ":1:1: syntax error: unexpected ';'",
	},
	{
		src: "&",
		err: ":1:1: syntax error: unexpected '&'",
	},
	{
		src: "true &&",
		err: ":1:6: syntax error: unexpected EOF",
	},
	{
		src: "false ||",
		err: ":1:7: syntax error: unexpected EOF",
	},
}

func TestParseError(t *testing.T) {
	for i, tt := range parseErrorTests {
		name := fmt.Sprintf("%v.sh", i)
		switch _, _, err := parser.ParseCommand(name, tt.src); {
		case err == nil:
			t.Error("expected error")
		case err.Error()[len(name):] != tt.err:
			t.Error("unexpected error:", err)
		}
	}
}

func TestOpen(t *testing.T) {
	for _, src := range []interface{}{
		[]byte{},
		"",
		new(bytes.Buffer),
		&reader{
			data: "",
			eof:  io.EOF,
		},
	} {
		if _, _, err := parser.ParseCommand("", src); err != nil {
			t.Error(err)
		}
	}

	if _, _, err := parser.ParseCommand("", nil); err == nil {
		t.Error("expected error")
	}
}

func TestReadError(t *testing.T) {
	for _, data := range []string{
		"",
		"\\",
		"'",
	} {
		src := &reader{
			data: data,
			eof:  errors.New("read error"),
		}
		if _, _, err := parser.ParseCommand("", src); err == nil {
			t.Error("expected error")
		}
	}
}

type reader struct {
	data string
	i    int
	eof  error
}

func (r *reader) Read(p []byte) (int, error) {
	if r.i < len(r.data) {
		p[0] = r.data[r.i]
		r.i++
		return 1, nil
	}
	return 0, r.eof
}
