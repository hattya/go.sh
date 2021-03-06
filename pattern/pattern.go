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
	"regexp"
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
