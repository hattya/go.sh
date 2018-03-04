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
	cmd      []ast.Word
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
		cmd: []ast.Word{
			word(lit(1, 1, "echo")),
			word(lit(1, 6, "1")),
			word(lit(1, 9, "22")),
			word(lit(1, 14, "333")),
		},
	},
	{
		src: "echo\t1  22\t \t333\n",
		cmd: []ast.Word{
			word(lit(1, 1, "echo")),
			word(lit(1, 6, "1")),
			word(lit(1, 9, "22")),
			word(lit(1, 14, "333")),
		},
	},
	// quoting
	{
		src: "\\pwd\n",
		cmd: []ast.Word{
			word(quote(1, 1, "\\", word(lit(1, 2, "p"))), lit(1, 3, "wd")),
		},
	},
	{
		src: "p\\wd\n",
		cmd: []ast.Word{
			word(lit(1, 1, "p"), quote(1, 2, "\\", word(lit(1, 3, "w"))), lit(1, 4, "d")),
		},
	},
	{
		src: "pw\\d\n",
		cmd: []ast.Word{
			word(lit(1, 1, "pw"), quote(1, 3, "\\", word(lit(1, 4, "d")))),
		},
	},
	{
		src: "pwd\\\n",
		cmd: []ast.Word{
			word(lit(1, 1, "pwd")),
		},
	},
	{
		src: "pwd\\",
		cmd: []ast.Word{
			word(lit(1, 1, "pwd"), quote(1, 4, "\\", nil)),
		},
	},
	// <newline>
	{
		src: "echo 1\necho 2\n",
		cmd: []ast.Word{
			word(lit(1, 1, "echo")),
			word(lit(1, 6, "1")),
		},
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
		cmd: []ast.Word{
			word(lit(3, 1, "go")),
			word(lit(3, 4, "version")),
		},
		comments: []*ast.Comment{
			comment(1, 1, " comment"),
		},
	},
	{
		src: "go version# comment\n",
		cmd: []ast.Word{
			word(lit(1, 1, "go")),
			word(lit(1, 4, "version")),
		},
		comments: []*ast.Comment{
			comment(1, 11, " comment"),
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

func comment(line, col int, text string) *ast.Comment {
	return &ast.Comment{
		Hash: ast.NewPos(line, col),
		Text: text,
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
