//
// go.sh/interp :: expand.go
//
//   Copyright (c) 2021 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
//

package interp

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/hattya/go.sh/ast"
)

// Expand expands a word and returns a string.
//
// If assign is true, it expands multiple tilde-prefixes in a word as if
// it is in an assignment.
func (env *ExecEnv) Expand(word ast.Word, assign bool) (string, error) {
	var b strings.Builder
	for i := 0; i < len(word); i++ {
		switch w := word[i].(type) {
		case *ast.Lit:
			s := w.Value
			if i == 0 {
				off, col := env.expandTilde(&b, s, word[i+1:], assign)
				if off != 0 {
					i += off
					s = word[i].(*ast.Lit).Value
				}
				s = s[col:]
			}
			for assign {
				// separator
				j := strings.IndexByte(s, os.PathListSeparator)
				if j == -1 {
					break
				}
				b.WriteString(s[:j+1])
				s = s[j+1:]
				// expansion
				if s == "" && i+1 < len(word) {
					if w, ok := word[i+1].(*ast.Lit); ok {
						i++
						s = w.Value
					}
				}
				off, col := env.expandTilde(&b, s, word[i+1:], assign)
				if off != 0 {
					i += off
					s = word[i].(*ast.Lit).Value
				}
				s = s[col:]
			}
			// remaining
			b.WriteString(s)
		case *ast.ParamExp:
			if err := env.expandParam(&b, w); err != nil {
				return "", err
			}
		}
	}
	return b.String(), nil
}

// expandTilde performs tilde expansion.
func (env *ExecEnv) expandTilde(b *strings.Builder, s string, word ast.Word, assign bool) (off, col int) {
	if !strings.HasPrefix(s, "~") {
		return
	}
	s = s[1:]
	col = 1
	// login name
	var name, sep string
	if assign {
		sep = string(os.PathListSeparator) + "/"
	} else {
		sep = "/"
	}
	for {
		if i := strings.IndexAny(s, sep); i != -1 {
			name += s[:i]
			col += i
			break
		}
		name += s
		col += len(s)
		if off >= len(word) {
			break
		} else if w, ok := word[off].(*ast.Lit); ok {
			s = w.Value
			off += 1
			col = 0
		} else {
			goto Fail
		}
	}
	// home directory
	if dir := env.homeDir(name); dir != "" {
		b.WriteString(dir)
	} else {
		goto Fail
	}
	return
Fail:
	b.WriteByte('~')
	b.WriteString(name)
	return
}

func (env *ExecEnv) homeDir(name string) string {
	var dir string
	if name == "" {
		if v, set := env.Get("HOME"); set {
			dir = v.Value
		} else if runtime.GOOS == "windows" {
			if v, set := env.Get("USERPROFILE"); set {
				dir = v.Value
			}
		}
	} else if u, err := user.Lookup(name); err == nil {
		dir = u.HomeDir
	}
	return filepath.ToSlash(dir)
}

// expandParam performs parameter expansion.
func (env *ExecEnv) expandParam(b *strings.Builder, pe *ast.ParamExp) error {
	switch v, set := env.Get(pe.Name.Value); {
	case pe.Op == "":
		// simplest form
		if set && v.Value != "" {
			b.WriteString(v.Value)
		}
	case pe.Word != nil:
		switch pe.Op {
		case ":-", "-":
			// use default values
			switch {
			case set && v.Value != "":
				b.WriteString(v.Value)
			case !set || pe.Op == ":-":
				s, err := env.Expand(pe.Word, true)
				if err != nil {
					return err
				}
				b.WriteString(s)
			}
		case ":=", "=":
			// assign default values
			switch {
			case set && v.Value != "":
				b.WriteString(v.Value)
			case !set || pe.Op == ":=":
				if env.isSpParam(pe.Name.Value) || env.isPosParam(pe.Name.Value) {
					return ParamExpError{
						ParamExp: pe,
						Msg:      "cannot assign in this way",
					}
				}
				s, err := env.Expand(pe.Word, false)
				if err != nil {
					return err
				}
				env.Set(pe.Name.Value, s)
				b.WriteString(s)
			}
		}
	}
	return nil
}

// ParamExpError represents an error in parameter expansion.
type ParamExpError struct {
	ParamExp *ast.ParamExp
	Msg      string
}

func (e ParamExpError) Error() string {
	return fmt.Sprintf("$%s: %s", e.ParamExp.Name.Value, e.Msg)
}
