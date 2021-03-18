//
// go.sh/interp :: interp.go
//
//   Copyright (c) 2021 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
//

// Package interp implements an interpreter for the Shell Command Language
// (POSIX.1-2017).
package interp

import (
	"os"
	"strconv"
	"strings"
)

const IFS = " \t\n"

// ExecEnv represents a shell execution environment.
type ExecEnv struct {
	Args    []string
	Opts    Option
	Aliases map[string]string

	vars map[string]Var
}

// NewExecEnv returns a new ExecEnv.
func NewExecEnv(name string, args ...string) *ExecEnv {
	env := &ExecEnv{
		Args:    append([]string{name}, args...),
		Aliases: make(map[string]string),
		vars:    make(map[string]Var),
	}
	for _, s := range os.Environ() {
		if i := strings.IndexByte(s[1:], '='); i != -1 {
			env.vars[env.keyFor(s[:i+1])] = Var{
				Name:   s[:i+1],
				Value:  s[i+2:],
				Export: true,
			}
		}
	}
	// shell variables
	env.vars[env.keyFor("IFS")] = Var{
		Name:   "IFS",
		Value:  IFS,
		Export: true,
	}
	return env
}

// Get retrieves the variable named by the name.
func (env *ExecEnv) Get(name string) (v Var, set bool) {
	if env.isPosParam(name) {
		if i, _ := strconv.Atoi(name); i < len(env.Args) {
			v = Var{
				Name:  strconv.Itoa(i),
				Value: env.Args[i],
			}
			set = true
		}
	} else {
		v, set = env.vars[env.keyFor(name)]
	}
	return
}

// Set sets the value of the variable named by the name.
func (env *ExecEnv) Set(name, value string) {
	if env.isPosParam(name) {
		return
	}
	env.vars[env.keyFor(name)] = Var{
		Name:  name,
		Value: value,
	}
}

// Unset unsets the variable named by the name.
func (env *ExecEnv) Unset(name string) {
	delete(env.vars, env.keyFor(name))
}

// Walk walks the variables, calling fn for each.
func (env *ExecEnv) Walk(fn func(Var)) {
	for _, v := range env.vars {
		fn(v)
	}
}

// isSpParam reports whether s matches the name of a special parameter.
func (env *ExecEnv) isSpParam(s string) bool {
	switch s {
	case "@", "*", "#", "?", "-", "$", "!", "0":
		return true
	}
	return false
}

// isPosParam reports whether s matches the name of a positional
// parameter.
func (env *ExecEnv) isPosParam(s string) bool {
	for _, r := range s {
		if r < '0' || '9' < r {
			return false
		}
	}
	return s != "" && s != "0"
}

// Option represents a shell option.
type Option uint

const (
	AllExport Option = 1 << iota
	ErrExit
	IgnoreEOF
	Monitor
	NoClobber
	NoGlob
	NoExec
	NoLog
	Notify
	NoUnset
	Verbose
	Vi
	XTrace
)

// Var represents a variable.
type Var struct {
	Name  string
	Value string

	Export   bool
	ReadOnly bool
}
