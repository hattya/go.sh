//
// go.sh/interp :: interp_windows.go
//
//   Copyright (c) 2021 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
//

package interp

import "strings"

func (env *ExecEnv) keyFor(name string) string {
	return strings.ToUpper(name)
}
