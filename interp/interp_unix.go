//
// go.sh/interp :: interp_unix.go
//
//   Copyright (c) 2021 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
//

// +build !plan9,!windows

package interp

func (env *ExecEnv) keyFor(name string) string {
	return name
}
