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
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/hattya/go.sh/ast"
	"github.com/hattya/go.sh/pattern"
)

// ExpMode controls the behavior of word expansions.
type ExpMode uint

const (
	// Expands multiple tilde-prefixes in a word as if it is in an
	// assignment.
	Assign ExpMode = 1 << iota

	// Expand a word into a single field, field splitting and pathname
	// expansion will not be performed.
	Literal

	// Expand a word into a single field, field splitting and pathname
	// expansion will not be performed. The result will be retained if
	// '?', '*', and '[' are quoted.
	Pattern

	// Expands a word as if it is within double-quotes, field splitting
	// and pathname expansion will not be performed.
	Quote
)

// Expand expands a word into multiple fields.
func (env *ExecEnv) Expand(word ast.Word, mode ExpMode) ([]string, error) {
	fields, err := env.expand(word, mode)
	if err != nil {
		return nil, err
	}
	var rv []string
	switch {
	case mode&Literal != 0:
		rv = []string{env.join(fields...).unquote()}
	case mode&Pattern != 0:
		rv = []string{env.join(fields...).pattern()}
	default:
		for _, f := range fields {
			switch {
			case mode&Quote != 0:
				rv = append(rv, f.unquote())
			case !f.empty():
				for _, f := range env.split(f) {
					if !f.empty() {
						if env.Opts&NoGlob != 0 {
							rv = append(rv, f.unquote())
						} else {
							rv = append(rv, env.expandPath(f)...)
						}
					}
				}
			}
		}
	}
	return rv, nil
}

func (env *ExecEnv) expand(word ast.Word, mode ExpMode) (fields []*field, err error) {
	fields = []*field{{}}
	if mode&Quote != 0 {
		fields[0].join("", true)
	}
	for i := 0; i < len(word); i++ {
		switch w := word[i].(type) {
		case *ast.Lit:
			f := fields[len(fields)-1]
			s := w.Value
			if i == 0 {
				off, col := env.expandTilde(f, s, word[i+1:], mode)
				if off != 0 {
					i += off
					s = word[i].(*ast.Lit).Value
				}
				s = s[col:]
			}
			for mode&Assign != 0 {
				// separator
				j := strings.IndexByte(s, os.PathListSeparator)
				if j == -1 {
					break
				}
				f.join(s[:j+1], mode&Quote != 0)
				s = s[j+1:]
				// expansion
				if s == "" && i+1 < len(word) {
					if w, ok := word[i+1].(*ast.Lit); ok {
						i++
						s = w.Value
					}
				}
				off, col := env.expandTilde(f, s, word[i+1:], mode)
				if off != 0 {
					i += off
					s = word[i].(*ast.Lit).Value
				}
				s = s[col:]
			}
			// remaining
			f.join(s, mode&Quote != 0)
		case *ast.Quote:
			switch w.Tok {
			case `\`, `'`:
				var s string
				if len(w.Value) != 0 {
					s = w.Value[0].(*ast.Lit).Value
				}
				fields[len(fields)-1].join(s, true)
			case `"`:
				word, err := env.expand(w.Value, Quote)
				if err != nil {
					return nil, err
				}
				fields[len(fields)-1].merge(word[0])
				fields = append(fields, word[1:]...)
			}
		case *ast.ParamExp:
			if fields, err = env.expandParam(fields, w, mode); err != nil {
				return
			}
		}
	}
	return
}

// expandTilde performs tilde expansion.
func (env *ExecEnv) expandTilde(f *field, s string, word ast.Word, mode ExpMode) (off, col int) {
	if mode&Quote != 0 || !strings.HasPrefix(s, "~") {
		return
	}
	s = s[1:]
	col = 1
	// login name
	var name, sep string
	if mode&Assign != 0 {
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
			if runtime.GOOS == "windows" {
				if w, ok := word[off].(*ast.Quote); ok && w.Tok == `\` && w.Value[0].(*ast.Lit).Value == `\` {
					break
				}
			}
			goto Fail
		}
	}
	// home directory
	if dir := env.homeDir(name); dir != "" {
		f.join(dir, true)
	} else {
		goto Fail
	}
	return
Fail:
	f.join("~"+name, false)
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
func (env *ExecEnv) expandParam(fields []*field, pe *ast.ParamExp, mode ExpMode) ([]*field, error) {
	quote := mode&Quote != 0
	var a []string
	var set, null bool
	switch pe.Name.Value {
	case "@":
		set = true
		switch len(env.Args) {
		case 1:
			null = true
		case 2:
			null = env.Args[1] == ""
			fallthrough
		default:
			a = make([]string, len(env.Args)-1)
			copy(a, env.Args[1:])
		}
	case "*":
		set = true
		switch len(env.Args) {
		case 1:
			null = true
		case 2:
			a = []string{env.Args[1]}
			null = env.Args[1] == ""
		default:
			var b strings.Builder
			sep := env.ifs()
			for i, s := range env.Args[1:] {
				if i > 0 {
					b.WriteString(sep)
				}
				b.WriteString(s)
			}
			a = []string{b.String()}
		}
	default:
		var v Var
		if v, set = env.Get(pe.Name.Value); set {
			a = []string{v.Value}
			null = v.Value == ""
		}
	}
	switch {
	case pe.Op == "":
		// simplest form
		if set && !null {
			goto Param
		}
	case pe.Word == nil:
		// string length
		if pe.Op == "#" {
			switch {
			case set:
				var n int
				if pe.Name.Value == "@" {
					n = len(a)
				} else {
					n = utf8.RuneCountInString(a[0])
				}
				fields[len(fields)-1].join(strconv.Itoa(n), quote)
			case !set && env.Opts&NoUnset != 0:
				goto Unset
			}
		}
	default:
		switch pe.Op {
		case ":-", "-":
			// use default values
			switch {
			case set && !null:
				goto Param
			case !set || pe.Op == ":-":
				word, err := env.expand(pe.Word, mode&(Assign|Quote)|Literal)
				if err != nil {
					return nil, err
				}
				fields[len(fields)-1].merge(word[0])
				fields = append(fields, word[1:]...)
			}
		case ":=", "=":
			// assign default values
			switch {
			case set && !null:
				goto Param
			case !set || pe.Op == ":=":
				if env.isSpParam(pe.Name.Value) || env.isPosParam(pe.Name.Value) {
					return nil, ParamExpError{
						ParamExp: pe,
						Msg:      "cannot assign in this way",
					}
				}
				word, err := env.expand(pe.Word, mode&Quote|Literal)
				if err != nil {
					return nil, err
				}
				env.Set(pe.Name.Value, env.join(word...).unquote())
				fields[len(fields)-1].merge(word[0])
				fields = append(fields, word[1:]...)
			}
		case ":?", "?":
			// indicate error if unset or null
			switch {
			case set && !null:
				goto Param
			case !set || pe.Op == ":?":
				var msg string
				if len(pe.Word) == 0 {
					msg = "parameter is unset or null"
				} else {
					word, err := env.expand(pe.Word, mode&Quote|Literal)
					if err != nil {
						return nil, err
					}
					msg = env.join(word...).unquote()
				}
				return nil, ParamExpError{
					ParamExp: pe,
					Msg:      msg,
				}
			}
		case ":+", "+":
			// use alternative values
			if set && (!null || pe.Op == "+") {
				word, err := env.expand(pe.Word, mode&(Assign|Quote)|Literal)
				if err != nil {
					return nil, err
				}
				fields[len(fields)-1].merge(word[0])
				fields = append(fields, word[1:]...)
			}
		case "%", "%%":
			// remove suffix pattern
			switch {
			case set && !null:
				{
					word, err := env.expand(pe.Word, Pattern)
					if err != nil {
						return nil, err
					}
					pats := []string{env.join(word...).pattern()}
					mode := pattern.Suffix
					if pe.Op == "%" {
						mode |= pattern.Smallest
					} else {
						mode |= pattern.Largest
					}
					for i, s := range a {
						m, err := pattern.Match(pats, mode, s)
						if err != nil && err != pattern.NoMatch {
							return nil, err
						}
						if i > 0 {
							fields = append(fields, &field{})
						}
						fields[len(fields)-1].join(s[:len(s)-len(m)], quote)
					}
				}
			case !set && env.Opts&NoUnset != 0:
				goto Unset
			}
		case "#", "##":
			// remove prefix pattern
			switch {
			case set && !null:
				{
					word, err := env.expand(pe.Word, Pattern)
					if err != nil {
						return nil, err
					}
					pats := []string{env.join(word...).pattern()}
					mode := pattern.Prefix
					if pe.Op == "#" {
						mode |= pattern.Smallest
					} else {
						mode |= pattern.Largest
					}
					for i, s := range a {
						m, err := pattern.Match(pats, mode, s)
						if err != nil && err != pattern.NoMatch {
							return nil, err
						}
						if i > 0 {
							fields = append(fields, &field{})
						}
						fields[len(fields)-1].join(s[len(m):], quote)
					}
				}
			case !set && env.Opts&NoUnset != 0:
				goto Unset
			}
		}
	}
	return fields, nil
Param:
	for i, s := range a {
		if i > 0 {
			fields = append(fields, &field{})
		}
		fields[len(fields)-1].join(s, quote)
	}
	return fields, nil
Unset:
	return nil, ParamExpError{
		ParamExp: pe,
		Msg:      "parameter is unset",
	}
}

// spilt performs field splitting.
func (env *ExecEnv) split(f *field) []*field {
	var ifs string
	if v, set := env.Get("IFS"); set {
		ifs = v.Value
	} else {
		ifs = IFS
	}

	if ifs == "" {
		return []*field{f}
	}
	fields := []*field{{}}
	ws := true
	for i := 0; i < len(f.b); i++ {
		s := f.b[i]
		if f.quote[i] {
			fields[len(fields)-1].join(s, true)
			ws = false
		} else {
			var i int
			for j, r := range s {
				if strings.ContainsRune(ifs, r) {
					switch {
					case unicode.IsSpace(r):
						// IFS white space
						if ws {
							break
						}
						ws = true
						fallthrough
					case !ws:
						fields[len(fields)-1].join(s[i:j], false)
						fields = append(fields, &field{})
					default:
						ws = false
					}
					i = j + utf8.RuneLen(r)
				} else {
					ws = false
				}
			}
			if i < len(s) {
				fields[len(fields)-1].join(s[i:], false)
			}
		}
	}
	if len(fields[len(fields)-1].b) == 0 && ws {
		fields = fields[:len(fields)-1]
	}
	return fields
}

// join joins the specified fields into a single field.
func (env *ExecEnv) join(fields ...*field) *field {
	dst := &field{}
	sep := env.ifs()
	for i, f := range fields {
		if i > 0 {
			dst.join(sep, false)
		}
		dst.merge(f)
	}
	return dst
}

// ifs returns a separator determined by the IFS variable.
func (env *ExecEnv) ifs() string {
	if v, set := env.Get("IFS"); set {
		if v.Value != "" {
			return v.Value[:1]
		}
		return ""
	}
	return " "
}

// expandPath performs pathname expansion.
func (env *ExecEnv) expandPath(f *field) []string {
	paths, err := pattern.Glob(f.pattern())
	if err != nil || len(paths) == 0 {
		return []string{f.unquote()}
	}
	return paths
}

// ParamExpError represents an error in parameter expansion.
type ParamExpError struct {
	ParamExp *ast.ParamExp
	Msg      string
}

func (e ParamExpError) Error() string {
	return fmt.Sprintf("$%s: %s", e.ParamExp.Name.Value, e.Msg)
}

type field struct {
	b     []string
	quote []bool
}

func (f *field) empty() bool {
	for i := 0; i < len(f.b); i++ {
		if f.quote[i] || f.b[i] != "" {
			return false
		}
	}
	return true
}

func (f *field) join(s string, quote bool) {
	f.b = append(f.b, s)
	f.quote = append(f.quote, quote)
}

func (f *field) merge(t *field) {
	f.b = append(f.b, t.b...)
	f.quote = append(f.quote, t.quote...)
}

func (f *field) pattern() string {
	var b strings.Builder
	for i := 0; i < len(f.b); i++ {
		s := f.b[i]
		if f.quote[i] {
			for {
				i := strings.IndexAny(s, `?*[\`)
				if i == -1 {
					b.WriteString(s)
					break
				}
				b.WriteString(s[:i])
				b.WriteByte('\\')
				b.WriteByte(s[i])
				s = s[i+1:]
			}
		} else {
			b.WriteString(s)
		}
	}
	return b.String()
}

func (f *field) unquote() string {
	return strings.Join(f.b, "")
}
