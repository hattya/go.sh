//
// go.sh/interp :: interp_unix.go
//
//   Copyright (c) 2021-2023 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
//

//go:build unix

package interp

func (env *ExecEnv) keyFor(name string) string {
	return name
}
