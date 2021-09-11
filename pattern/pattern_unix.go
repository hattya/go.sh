//
// go.sh/pattern :: pattern_unix.go
//
//   Copyright (c) 2021 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
//

//go:build !plan9 && !windows
// +build !plan9,!windows

package pattern

import "strings"

func indexSep(pat string) (int, int) {
	n := len(pat)
	for {
		switch i := strings.IndexAny(pat, `/\`); {
		case i == -1:
			return -1, 0
		case pat[i] == '\\' && i+1 < len(pat):
			if pat[i+1] == '/' {
				return n - len(pat[i:]), 2
			}
			pat = pat[i+1:]
		default:
			return n - len(pat[i:]), 1
		}
	}
}

func split(pat string) (string, string) {
	if i, w := indexSep(pat); i == 0 {
		return "/", pat[w:]
	}
	return ".", pat
}
