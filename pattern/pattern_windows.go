//
// go.sh/pattern :: pattern_windows.go
//
//   Copyright (c) 2021 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
//

package pattern

import (
	"strings"
	"unicode/utf8"
)

func indexSep(pat string) (int, int) {
	n := len(pat)
	for {
		switch i := strings.IndexAny(pat, `\/`); {
		case i == -1:
			return -1, 0
		case pat[i] == '\\' && i+1 < len(pat):
			if isSep(rune(pat[i+1])) {
				return n - len(pat[i:]), 2
			}
			pat = pat[i+1:]
		default:
			return n - len(pat[i:]), 1
		}
	}
}

func split(pat string) (string, string) {
	// drive or volume
	var b strings.Builder
	i := 0
	for j := 0; j < 2; j++ {
		r, w := unquoteRune(pat[i:])
		b.WriteRune(r)
		i += w
	}
	switch s := b.String(); {
	case s[1] == ':' && ('A' <= s[0] && s[0] <= 'Z' || 'a' <= s[0] && s[0] <= 'z'):
		// drive letter
	case isSep(rune(s[0])) && isSep(rune(s[1])):
		sep := 0
		for i < len(pat) {
			r, w := unquoteRune(pat[i:])
			if isSep(r) {
				sep++
				if sep == 2 {
					break
				}
			}
			b.WriteRune(r)
			i += w
		}
		switch s := b.String(); {
		case len(s) >= 4 && isSep(rune(s[3])) && (s[2] == '.' || s[2] == '?'):
			// DOS device
			if len(s) == 7 && strings.HasSuffix(s, "UNC") {
				for i < len(pat) {
					r, w := unquoteRune(pat[i:])
					if isSep(r) {
						sep++
						if sep == 5 {
							break
						}
					}
					b.WriteRune(r)
					i += w
				}
			}
		case sep > 0:
			// UNC
		default:
			b.Reset()
			i = 0
		}
	default:
		b.Reset()
		i = 0
	}
	// path
	if r, w := unquoteRune(pat[i:]); isSep(r) {
		b.WriteRune(r)
		i += w
	}

	if b.Len() == 0 {
		b.WriteByte('.')
	}
	return b.String(), pat[i:]
}

func isSep(r rune) bool {
	return r == '\\' || r == '/'
}

func unquoteRune(s string) (rune, int) {
	r, w := utf8.DecodeRuneInString(s)
	if r == '\\' {
		r, w = utf8.DecodeRuneInString(s[w:])
		w++
	}
	return r, w
}
