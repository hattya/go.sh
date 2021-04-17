//
// go.sh/pattern :: pattern.go
//
//   Copyright (c) 2021 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
//

// Package pattern implements the pattern matching notation.
package pattern

import (
	"errors"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"
)

// NoMatch indicates that the pattern does not match anything.
var NoMatch = errors.New("no match")

// Mode controls the behavior of Match.
type Mode uint

const (
	Smallest Mode = 1 << iota // smallest match
	Largest                   // largest match
	Suffix                    // pattern matching with suffix
	Prefix                    // pattern matching with prefix
)

// Match returns a string holding the portion of the match in s of the
// patterns. The patterns will be joined by "|", and translated into a
// regular expression.
// If no match is found, the error returned is NoMatch.
//
// Longest is default and has priority. Suffix and Prefix are mutually
// exclusive.
func Match(patterns []string, mode Mode, s string) (string, error) {
	if mode&Suffix != 0 && mode&Prefix != 0 {
		return "", NoMatch
	}
	rx, err := compile(patterns, mode)
	if err != nil {
		return "", err
	}
	if m := rx.FindStringSubmatch(s); m != nil {
		for mode&Smallest != 0 && mode&Suffix != 0 {
			s = s[len(s)-len(m[0]):]
			r, w := utf8.DecodeRuneInString(s)
			if r == utf8.RuneError {
				if w == 0 {
					break
				} else {
					m[0] = m[0][w:]
					continue
				}
			}
			sm := rx.FindStringSubmatch(s[w:])
			if sm == nil {
				break
			}
			m = sm
		}
		return m[1], nil
	}
	return "", NoMatch
}

// Glob returns paths that matches pattern.
func Glob(pattern string) ([]string, error) {
	if pattern == "" {
		return nil, nil
	}
	base, pattern := split(pattern)
	paths := []string{base}
	for pattern != "" {
		i, w := indexSep(pattern)
		var sep string
		if i == -1 {
			i = len(pattern)
		} else {
			sep = pattern[i+w-1 : i+w]
		}

		switch {
		case i > 0:
			var matches []string
			if name, lit := unquote(pattern[:i]); lit {
				// literal
				for _, p := range paths {
					if p == "." {
						p = name
					} else {
						p += name
					}
					if _, err := os.Lstat(p); err == nil {
						matches = append(matches, p+sep)
					}
				}
			} else {
				// pattern
				rx, err := compile([]string{pattern[:i]}, Prefix|Suffix)
				if err != nil {
					return nil, err
				}
				for _, p := range paths {
					err := glob(p, rx, func(name string) {
						if p != "." {
							name = p + name
						}
						matches = append(matches, name+sep)
					})
					if err != nil {
						return nil, err
					}
				}
			}
			if len(matches) == 0 {
				// no match
				return nil, nil
			}
			paths = matches
			sort.Strings(paths)
		case w > 0:
			// sep
			for i := range paths {
				paths[i] += sep
			}
		}
		pattern = pattern[i+w:]
	}
	return paths, nil
}

func glob(path string, rx *regexp.Regexp, fn func(string)) error {
	d, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer d.Close()

	var dot bool
	if strings.HasPrefix(rx.String(), `^(\.`) {
		dot = true
		for _, n := range []string{".", ".."} {
			if rx.MatchString(n) {
				fn(n)
			}
		}
	}
	for {
		switch n, err := d.Readdirnames(1); {
		case err != nil:
			if err == io.EOF {
				return nil
			}
			return err
		case rx.MatchString(n[0]):
			if dot || !strings.HasPrefix(n[0], ".") {
				fn(n[0])
			}
		}
	}
}

func unquote(s string) (string, bool) {
	var b strings.Builder
	var esc bool
	for _, r := range s {
		switch r {
		case utf8.RuneError:
			return "", false
		case '\\':
			if !esc {
				esc = true
				continue
			}
		case '?', '*', '[':
			if !esc {
				return "", false
			}
		}
		b.WriteRune(r)
		esc = false
	}
	return b.String(), true
}

func compile(patterns []string, mode Mode) (*regexp.Regexp, error) {
	var b strings.Builder
	if mode&Prefix != 0 {
		b.WriteByte('^')
	}
	b.WriteByte('(')
	for i, pat := range patterns {
		if i > 0 {
			b.WriteByte('|')
		}
	Pattern:
		for pat != "" {
			r, w := utf8.DecodeRuneInString(pat)
			switch r {
			case utf8.RuneError:
				b.WriteString(pat[:w])
			case '?':
				b.WriteByte('.')
			case '*':
				if mode&Smallest == 0 || mode&Largest != 0 {
					b.WriteString(".*")
				} else {
					b.WriteString(".*?")
				}
			case '[':
				b.WriteByte('[')
				pat = pat[w:]
				r, w = utf8.DecodeRuneInString(pat)
				if r == '^' || r == '!' {
					b.WriteByte('^')
					pat = pat[w:]
					r, w = utf8.DecodeRuneInString(pat)
				}
				if r == ']' {
					b.WriteByte(']')
					pat = pat[w:]
					r, w = utf8.DecodeRuneInString(pat)
				}
			Bracket:
				for {
					switch r {
					case utf8.RuneError:
						if w == 0 {
							break Pattern
						}
						b.WriteString(pat[:w])
					case '[':
						b.WriteByte('[')
						pat = pat[w:]
						r, w = utf8.DecodeRuneInString(pat)
						switch r {
						case utf8.RuneError:
							if w == 0 {
								break Pattern
							}
							b.WriteString(pat[:w])
						case '.', '=', ':':
							b.WriteRune(r)
							pat = pat[w:]
							j := strings.Index(pat, string(r)+"]")
							if j == -1 {
								break Bracket
							}
							w = j + 2
							b.WriteString(pat[:w])
						default:
							b.WriteRune(r)
							break Bracket
						}
					case ']':
						b.WriteByte(']')
						break Bracket
					case '\\':
						pat = pat[w:]
						r, w = utf8.DecodeRuneInString(pat)
						switch r {
						case utf8.RuneError:
							b.WriteByte('\\')
							if w == 0 {
								break Pattern
							}
							b.WriteString(pat[:w])
						case '!', '-', '[', ']', '^':
							b.WriteByte('\\')
						}
						b.WriteRune(r)
					default:
						b.WriteRune(r)
					}
					pat = pat[w:]
					r, w = utf8.DecodeRuneInString(pat)
				}
			case '\\':
				pat = pat[w:]
				r, w = utf8.DecodeRuneInString(pat)
				switch r {
				case utf8.RuneError:
					b.WriteByte('\\')
					if w == 0 {
						break Pattern
					}
					b.WriteString(pat[:w])
				case '\\', '.', '+', '*', '?', '(', ')', '|', '[', ']', '{', '}', '^', '$':
					b.WriteByte('\\')
				}
				b.WriteRune(r)
			case '.', '+', '(', ')', '|', '{', '}', '^', '$':
				b.WriteByte('\\')
				b.WriteRune(r)
			default:
				b.WriteRune(r)
			}
			pat = pat[w:]
		}
	}
	b.WriteByte(')')
	if mode&Suffix != 0 {
		b.WriteByte('$')
	}
	return regexp.Compile(b.String())
}
