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

// ExecEnv represents a shell execution environment.
type ExecEnv struct {
	Aliases map[string]string
}

// NewExecEnv returns a new ExecEnv.
func NewExecEnv() *ExecEnv {
	return &ExecEnv{
		Aliases: make(map[string]string),
	}
}
