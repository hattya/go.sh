//
// go.sh/parser :: parser_test.go
//
//   Copyright (c) 2018-2019 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
//

package parser_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"reflect"
	"testing"

	"github.com/hattya/go.sh/ast"
	"github.com/hattya/go.sh/parser"
)

var parseCommandTests = []struct {
	src      string
	cmd      ast.Command
	comments []*ast.Comment
}{
	// <blank>
	{
		src: "",
	},
	{
		src: "\t ",
	},
	{
		src: "echo 1\t\t2 \t 3",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(lit(1, 6, "1")),
			word(lit(1, 9, "2")),
			word(lit(1, 13, "3")),
		),
	},
	{
		src: "echo\t1  2\t \t3",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(lit(1, 6, "1")),
			word(lit(1, 9, "2")),
			word(lit(1, 13, "3")),
		),
	},
	// <newline>
	{
		src: "\n",
	},
	{
		src: "echo 1\necho 2\n",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(lit(1, 6, "1")),
		),
	},
	// quoting
	{
		src: "\\pwd\n",
		cmd: simple_command(
			word(quote(1, 1, `\`, word(lit(1, 2, "p"))), lit(1, 3, "wd")),
		),
	},
	{
		src: "p\\wd\n",
		cmd: simple_command(
			word(lit(1, 1, "p"), quote(1, 2, `\`, word(lit(1, 3, "w"))), lit(1, 4, "d")),
		),
	},
	{
		src: "pw\\d\n",
		cmd: simple_command(
			word(lit(1, 1, "pw"), quote(1, 3, `\`, word(lit(1, 4, "d")))),
		),
	},
	{
		src: "pwd\\\n",
		cmd: simple_command(
			word(lit(1, 1, "pwd")),
		),
	},
	{
		src: "pwd\\",
		cmd: simple_command(
			word(lit(1, 1, "pwd"), quote(1, 4, `\`, nil)),
		),
	},
	{
		src: "echo 'foo bar baz'",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(quote(1, 6, "'", word(lit(1, 7, "foo bar baz")))),
		),
	},
	{
		src: `echo "foo bar baz"`,
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(quote(1, 6, `"`, word(lit(1, 7, "foo bar baz")))),
		),
	},
	{
		src: "echo \"foo\\tbar\\tbaz\"",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(quote(1, 6, `"`, word(lit(1, 7, "foo\\tbar\\tbaz")))),
		),
	},
	{
		src: "echo \"foo\\\n bar\\\n baz\"",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(quote(1, 6, `"`, word(lit(1, 7, "foo"), lit(2, 1, " bar"), lit(3, 1, " baz")))),
		),
	},
	{
		src: `echo "\$USER"`,
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(quote(1, 6, `"`, word(quote(1, 7, `\`, word(lit(1, 8, "$"))), lit(1, 9, "USER")))),
		),
	},
	{
		src: `echo "foo ${BAR:-"bar"} baz"`,
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(quote(1, 6, `"`, word(
				lit(1, 7, "foo "),
				param_exp(1, 11, true, lit(1, 13, "BAR"), lit(1, 16, ":-"), word(
					quote(1, 18, `"`, word(lit(1, 19, "bar"))),
				)),
				lit(1, 24, " baz"),
			))),
		),
	},
	{
		src: `echo "foo $(echo bar) baz"`,
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(quote(1, 6, `"`, word(
				lit(1, 7, "foo "),
				cmd_subst(
					true,       // dollar
					pos(1, 12), // left
					simple_command(
						word(lit(1, 13, "echo")),
						word(lit(1, 18, "bar")),
					),
					pos(1, 21), // right
				),
				lit(1, 22, " baz"),
			))),
		),
	},
	{
		src: "echo \"foo `echo bar` baz\"",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(quote(1, 6, `"`, word(
				lit(1, 7, "foo "),
				cmd_subst(
					false,      // dollar
					pos(1, 11), // left
					simple_command(
						word(lit(1, 12, "echo")),
						word(lit(1, 17, "bar")),
					),
					pos(1, 20), // right
				),
				lit(1, 21, " baz"),
			))),
		),
	},
	{
		src: `echo "1 + 2 = $((1 + 2))"`,
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(quote(1, 6, `"`, word(
				lit(1, 7, "1 + 2 = "),
				arith_exp(
					pos(1, 15), // left
					word(lit(1, 18, "1"), lit(1, 20, "+"), lit(1, 22, "2")),
					pos(1, 23), // right
				),
			))),
		),
	},
	// parameter expansion
	{
		src: "echo $",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(lit(1, 6, "$")),
		),
	},
	{
		src: "echo $@_",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(param_exp(1, 6, false, lit(1, 7, "@"), nil, nil), lit(1, 8, "_")),
		),
	},
	{
		src: "echo $*_",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(param_exp(1, 6, false, lit(1, 7, "*"), nil, nil), lit(1, 8, "_")),
		),
	},
	{
		src: "echo $#_",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(param_exp(1, 6, false, lit(1, 7, "#"), nil, nil), lit(1, 8, "_")),
		),
	},
	{
		src: "echo $?_",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(param_exp(1, 6, false, lit(1, 7, "?"), nil, nil), lit(1, 8, "_")),
		),
	},
	{
		src: "echo $-_",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(param_exp(1, 6, false, lit(1, 7, "-"), nil, nil), lit(1, 8, "_")),
		),
	},
	{
		src: "echo $$_",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(param_exp(1, 6, false, lit(1, 7, "$"), nil, nil), lit(1, 8, "_")),
		),
	},
	{
		src: "echo $!_",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(param_exp(1, 6, false, lit(1, 7, "!"), nil, nil), lit(1, 8, "_")),
		),
	},
	{
		src: "echo $0_",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(param_exp(1, 6, false, lit(1, 7, "0"), nil, nil), lit(1, 8, "_")),
		),
	},
	{
		src: "echo $1_",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(param_exp(1, 6, false, lit(1, 7, "1"), nil, nil), lit(1, 8, "_")),
		),
	},
	{
		src: "echo $11",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(param_exp(1, 6, false, lit(1, 7, "1"), nil, nil), lit(1, 8, "1")),
		),
	},
	{
		src: "echo $HOME",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(param_exp(1, 6, false, lit(1, 7, "HOME"), nil, nil)),
		),
	},
	{
		src: "echo $HOME.",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(param_exp(1, 6, false, lit(1, 7, "HOME"), nil, nil), lit(1, 11, ".")),
		),
	},
	{
		src: "echo $/",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(lit(1, 6, "$/")),
		),
	},
	{
		src: "echo ${@}",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(param_exp(1, 6, true, lit(1, 8, "@"), nil, nil)),
		),
	},
	{
		src: "echo ${*}",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(param_exp(1, 6, true, lit(1, 8, "*"), nil, nil)),
		),
	},
	{
		src: "echo ${#}",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(param_exp(1, 6, true, lit(1, 8, "#"), nil, nil)),
		),
	},
	{
		src: "echo ${?}",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(param_exp(1, 6, true, lit(1, 8, "?"), nil, nil)),
		),
	},
	{
		src: "echo ${-}",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(param_exp(1, 6, true, lit(1, 8, "-"), nil, nil)),
		),
	},
	{
		src: "echo ${$}",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(param_exp(1, 6, true, lit(1, 8, "$"), nil, nil)),
		),
	},
	{
		src: "echo ${!}",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(param_exp(1, 6, true, lit(1, 8, "!"), nil, nil)),
		),
	},
	{
		src: "echo ${0}",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(param_exp(1, 6, true, lit(1, 8, "0"), nil, nil)),
		),
	},
	{
		src: "echo ${1}",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(param_exp(1, 6, true, lit(1, 8, "1"), nil, nil)),
		),
	},
	{
		src: "echo ${11}",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(param_exp(1, 6, true, lit(1, 8, "11"), nil, nil)),
		),
	},
	{
		src: "echo ${HOME}",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(param_exp(1, 6, true, lit(1, 8, "HOME"), nil, nil)),
		),
	},
	{
		src: "echo ${LANG:-}",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(param_exp(1, 6, true, lit(1, 8, "LANG"), lit(1, 12, ":-"), nil)),
		),
	},
	{
		src: "echo ${LANG:-C\\.${ENC:-UTF-8}}",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(param_exp(1, 6, true, lit(1, 8, "LANG"), lit(1, 12, ":-"), word(
				lit(1, 14, "C"),
				quote(1, 15, `\`, word(lit(1, 16, "."))),
				param_exp(1, 17, true, lit(1, 19, "ENC"), lit(1, 22, ":-"), word(
					lit(1, 24, "UTF-8"),
				)),
			))),
		),
	},
	{
		src: "echo ${LANG-}",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(param_exp(1, 6, true, lit(1, 8, "LANG"), lit(1, 12, "-"), nil)),
		),
	},
	{
		src: "echo ${LANG-C\\.${ENC-UTF-8}}",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(param_exp(1, 6, true, lit(1, 8, "LANG"), lit(1, 12, "-"), word(
				lit(1, 13, "C"),
				quote(1, 14, `\`, word(lit(1, 15, "."))),
				param_exp(1, 16, true, lit(1, 18, "ENC"), lit(1, 21, "-"), word(
					lit(1, 22, "UTF-8"),
				)),
			))),
		),
	},
	{
		src: "echo ${#=}",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(param_exp(1, 6, true, lit(1, 8, "#"), lit(1, 9, "="), nil)),
		),
	},
	{
		src: "echo ${#-}",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(param_exp(1, 6, true, lit(1, 9, "-"), lit(1, 8, "#"), nil)),
		),
	},
	{
		src: "echo ${#-9}",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(param_exp(1, 6, true, lit(1, 8, "#"), lit(1, 9, "-"), word(
				lit(1, 10, "9"),
			))),
		),
	},
	{
		src: "echo ${#LANG}",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(param_exp(1, 6, true, lit(1, 9, "LANG"), lit(1, 8, "#"), nil)),
		),
	},
	{
		src: "echo ${LANG#*.}",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(param_exp(1, 6, true, lit(1, 8, "LANG"), lit(1, 12, "#"), word(
				lit(1, 13, "*."),
			))),
		),
	},
	{
		src: "echo ${LANG##*.}",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(param_exp(1, 6, true, lit(1, 8, "LANG"), lit(1, 12, "##"), word(
				lit(1, 14, "*."),
			))),
		),
	},
	// command substitution
	{
		src: "echo $(basename $0).",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(
				cmd_subst(
					true,      // dollar
					pos(1, 7), // left,
					simple_command(
						word(lit(1, 8, "basename")),
						word(param_exp(1, 17, false, lit(1, 18, "0"), nil, nil)),
					),
					pos(1, 19), // right
				),
				lit(1, 20, "."),
			),
		),
	},
	{
		src: "echo $(basename $(dirname $0)).",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(
				cmd_subst(
					true,      // dollar
					pos(1, 7), // left,
					simple_command(
						word(lit(1, 8, "basename")),
						word(
							cmd_subst(
								true,       // dollar
								pos(1, 18), // left
								simple_command(
									word(lit(1, 19, "dirname")),
									word(param_exp(1, 27, false, lit(1, 28, "0"), nil, nil)),
								),
								pos(1, 29), // right
							),
						),
					),
					pos(1, 30), // right
				),
				lit(1, 31, "."),
			),
		),
	},
	{
		src: "echo $(basename `dirname $0`).",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(
				cmd_subst(
					true,      // dollar
					pos(1, 7), // left,
					simple_command(
						word(lit(1, 8, "basename")),
						word(
							cmd_subst(
								false,      // dollar
								pos(1, 17), // left
								simple_command(
									word(lit(1, 18, "dirname")),
									word(param_exp(1, 26, false, lit(1, 27, "0"), nil, nil)),
								),
								pos(1, 28), // right
							),
						),
					),
					pos(1, 29), // right
				),
				lit(1, 30, "."),
			),
		),
	},
	{
		src: "echo `basename $0`.",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(
				cmd_subst(
					false,     // dollar
					pos(1, 6), // left,
					simple_command(
						word(lit(1, 7, "basename")),
						word(param_exp(1, 16, false, lit(1, 17, "0"), nil, nil)),
					),
					pos(1, 18), // right
				),
				lit(1, 19, "."),
			),
		),
	},
	{
		src: "echo `basename $(dirname $0)`.",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(
				cmd_subst(
					false,     // dollar
					pos(1, 6), // left,
					simple_command(
						word(lit(1, 7, "basename")),
						word(
							cmd_subst(
								true,       // dollar
								pos(1, 17), // left
								simple_command(
									word(lit(1, 18, "dirname")),
									word(param_exp(1, 26, false, lit(1, 27, "0"), nil, nil)),
								),
								pos(1, 28), // right
							),
						),
					),
					pos(1, 29), // right
				),
				lit(1, 30, "."),
			),
		),
	},
	// arithmetic expansion
	{
		src: "echo $((x))",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(arith_exp(
				pos(1, 6), // left
				word(lit(1, 9, "x")),
				pos(1, 10), // right
			)),
		),
	},
	{
		src: "echo $(($x))",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(arith_exp(
				pos(1, 6), // left
				word(param_exp(1, 9, false, lit(1, 10, "x"), nil, nil)),
				pos(1, 11), // right
			)),
		),
	},
	{
		src: "echo $((\\x = ((1 + $(echo 2))) * `echo 3`))",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(arith_exp(
				pos(1, 6), // left
				word(
					quote(1, 9, `\`, word(lit(1, 10, "x"))),
					lit(1, 12, "="),
					lit(1, 14, "((1"),
					lit(1, 18, "+"),
					cmd_subst(
						true,       // dollar
						pos(1, 21), // left
						simple_command(
							word(lit(1, 22, "echo")),
							word(lit(1, 27, "2")),
						),
						pos(1, 28), // right
					),
					lit(1, 29, "))"),
					lit(1, 32, "*"),
					cmd_subst(
						false,      // dollar
						pos(1, 34), // left
						simple_command(
							word(lit(1, 35, "echo")),
							word(lit(1, 40, "3")),
						),
						pos(1, 41), // right
					),
				),
				pos(1, 42), // right
			)),
		),
	},
	// simple command
	{
		src: "FOO=foo",
		cmd: simple_command(
			assignment_word(1, 1, "FOO", word(lit(1, 5, "foo"))),
		),
	},
	{
		src: "FOO=foo BAR=bar env",
		cmd: simple_command(
			assignment_word(1, 1, "FOO", word(lit(1, 5, "foo"))),
			assignment_word(1, 9, "BAR", word(lit(1, 13, "bar"))),
			word(lit(1, 17, "env")),
		),
	},
	{
		src: "FOO=foo BAR=bar env -u BAR BAZ=baz",
		cmd: simple_command(
			assignment_word(1, 1, "FOO", word(lit(1, 5, "foo"))),
			assignment_word(1, 9, "BAR", word(lit(1, 13, "bar"))),
			word(lit(1, 17, "env")),
			word(lit(1, 21, "-u")),
			word(lit(1, 24, "BAR")),
			word(lit(1, 28, "BAZ=baz")),
		),
	},
	{
		src: "123=123",
		cmd: simple_command(
			word(lit(1, 1, "123=123")),
		),
	},
	{
		src: "=env",
		cmd: simple_command(
			word(lit(1, 1, "=env")),
		),
	},
	{
		src: "cat <file",
		cmd: simple_command(
			word(lit(1, 1, "cat")),
			redir(nil, 1, 5, "<", word(lit(1, 6, "file"))),
		),
	},
	{
		src: "echo foo >file",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(lit(1, 6, "foo")),
			redir(nil, 1, 10, ">", word(lit(1, 11, "file"))),
		),
	},
	{
		src: "echo foo >|file",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(lit(1, 6, "foo")),
			redir(nil, 1, 10, ">|", word(lit(1, 12, "file"))),
		),
	},
	{
		src: "echo foo >>file",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(lit(1, 6, "foo")),
			redir(nil, 1, 10, ">>", word(lit(1, 12, "file"))),
		),
	},
	{
		src: "cat <<EOF\nEOF\n",
		cmd: simple_command(
			word(lit(1, 1, "cat")),
			heredoc(
				nil, 1, 5, "<<", word(lit(1, 7, "EOF")),
				word(),
				word(lit(2, 1, "EOF")),
			),
		),
	},
	{
		src: "cat <<EOF\nfoo\nbar\nbaz\nEOF\n",
		cmd: simple_command(
			word(lit(1, 1, "cat")),
			heredoc(
				nil, 1, 5, "<<", word(lit(1, 7, "EOF")),
				word(lit(2, 1, "foo\nbar\nbaz\n")),
				word(lit(5, 1, "EOF")),
			),
		),
	},
	{
		src: "cat <<EOF\nfoo\nbar\nbaz\nEOF",
		cmd: simple_command(
			word(lit(1, 1, "cat")),
			heredoc(
				nil, 1, 5, "<<", word(lit(1, 7, "EOF")),
				word(lit(2, 1, "foo\nbar\nbaz\n")),
				word(lit(5, 1, "EOF")),
			),
		),
	},
	{
		src: "cat <<E\\\nO\\\nF\nfoo\nbar\nbaz\nE\\\nO\\\nF\n",
		cmd: simple_command(
			word(lit(1, 1, "cat")),
			heredoc(
				nil, 1, 5, "<<", word(lit(1, 7, "E"), lit(2, 1, "O"), lit(3, 1, "F")),
				word(lit(4, 1, "foo\nbar\nbaz\n")),
				word(lit(7, 1, "E"), lit(8, 1, "O"), lit(9, 1, "F")),
			),
		),
	},
	{
		src: "cat <<E\\\n'O'\\\nF\nfoo\nbar\nbaz\nEOF\n",
		cmd: simple_command(
			word(lit(1, 1, "cat")),
			heredoc(
				nil, 1, 5, "<<", word(lit(1, 7, "E"), quote(2, 1, "'", word(lit(2, 2, "O"))), lit(3, 1, "F")),
				word(lit(4, 1, "foo\nbar\nbaz\n")),
				word(lit(7, 1, "EOF")),
			),
		),
	},
	{
		src: "cat <<EOF\n\\foo\n$bar\n`baz`\nEOF\n",
		cmd: simple_command(
			word(lit(1, 1, "cat")),
			heredoc(
				nil, 1, 5, "<<", word(lit(1, 7, "EOF")),
				word(
					lit(2, 1, "\\foo\n"),
					param_exp(3, 1, false, lit(3, 2, "bar"), nil, nil),
					lit(3, 5, "\n"),
					cmd_subst(
						false,     // dollar
						pos(4, 1), // left
						simple_command(
							word(lit(4, 2, "baz")),
						),
						pos(4, 5), // right
					),
					lit(4, 6, "\n"),
				),
				word(lit(5, 1, "EOF")),
			),
		),
	},
	{
		src: "cat <<-EOF\n\\foo\n$bar\n`baz`\nEOF\n",
		cmd: simple_command(
			word(lit(1, 1, "cat")),
			heredoc(
				nil, 1, 5, "<<-", word(lit(1, 8, "EOF")),
				word(
					lit(2, 1, "\\foo\n"),
					param_exp(3, 1, false, lit(3, 2, "bar"), nil, nil),
					lit(3, 5, "\n"),
					cmd_subst(
						false,     // dollar
						pos(4, 1), // left
						simple_command(
							word(lit(4, 2, "baz")),
						),
						pos(4, 5), // right
					),
					lit(4, 6, "\n"),
				),
				word(lit(5, 1, "EOF")),
			),
		),
	},
	{
		src: "cat <<'EOF'\n\\foo\n$bar\n`baz`\nEOF\n",
		cmd: simple_command(
			word(lit(1, 1, "cat")),
			heredoc(
				nil, 1, 5, "<<", word(quote(1, 7, "'", word(lit(1, 8, "EOF")))),
				word(lit(2, 1, "\\foo\n$bar\n`baz`\n")),
				word(lit(5, 1, "EOF")),
			),
		),
	},
	{
		src: "cat <<EOF1; echo 2; cat <<EOF3\n1\nEOF1\n3\nEOF3\n",
		cmd: list(
			and_or_list(
				simple_command(
					word(lit(1, 1, "cat")),
					heredoc(
						nil, 1, 5, "<<", word(lit(1, 7, "EOF1")),
						word(lit(2, 1, "1\n")),
						word(lit(3, 1, "EOF1")),
					),
				),
				sep(1, 11, ";"),
			),
			and_or_list(
				simple_command(
					word(lit(1, 13, "echo")),
					word(lit(1, 18, "2")),
				),
				sep(1, 19, ";"),
			),
			and_or_list(
				simple_command(
					word(lit(1, 21, "cat")),
					heredoc(
						nil, 1, 25, "<<", word(lit(1, 27, "EOF3")),
						word(lit(4, 1, "3\n")),
						word(lit(5, 1, "EOF3")),
					),
				),
			),
		),
	},
	{
		src: "echo $(cat <<EOF\n1\nEOF\necho 2\ncat <<EOF\n3\nEOF\n)",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(
				cmd_subst(
					true,      //dollar
					pos(1, 7), // left
					simple_command(
						word(lit(1, 8, "cat")),
						heredoc(
							nil, 1, 12, "<<", word(lit(1, 14, "EOF")),
							word(lit(2, 1, "1\n")),
							word(lit(3, 1, "EOF")),
						),
					),
					simple_command(
						word(lit(4, 1, "echo")),
						word(lit(4, 6, "2")),
					),
					simple_command(
						word(lit(5, 1, "cat")),
						heredoc(
							nil, 5, 5, "<<", word(lit(5, 7, "EOF")),
							word(lit(6, 1, "3\n")),
							word(lit(7, 1, "EOF")),
						),
					),
					pos(8, 1), // right
				),
			),
		),
	},
	{
		src: "exec 3<&-",
		cmd: simple_command(
			word(lit(1, 1, "exec")),
			redir(lit(1, 6, "3"), 1, 7, "<&", word(lit(1, 9, "-"))),
		),
	},
	{
		src: "echo foo >&1",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(lit(1, 6, "foo")),
			redir(nil, 1, 10, ">&", word(lit(1, 12, "1"))),
		),
	},
	{
		src: "exec 3<>file",
		cmd: simple_command(
			word(lit(1, 1, "exec")),
			redir(lit(1, 6, "3"), 1, 7, "<>", word(lit(1, 9, "file"))),
		),
	},
	{
		src: "echo foo>&2",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(lit(1, 6, "foo")),
			redir(nil, 1, 9, ">&", word(lit(1, 11, "2"))),
		),
	},
	{
		src: `echo \2>file`,
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(quote(1, 6, `\`, word(lit(1, 7, "2")))),
			redir(nil, 1, 8, ">", word(lit(1, 9, "file"))),
		),
	},
	{
		src: `echo 2\>file`,
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(lit(1, 6, "2"), quote(1, 7, `\`, word(lit(1, 8, ">"))), lit(1, 9, "file")),
		),
	},
	{
		src: ">/dev/null 2>&1 echo foo",
		cmd: simple_command(
			redir(nil, 1, 1, ">", word(lit(1, 2, "/dev/null"))),
			redir(lit(1, 12, "2"), 1, 13, ">&", word(lit(1, 15, "1"))),
			word(lit(1, 17, "echo")),
			word(lit(1, 22, "foo")),
		),
	},
	{
		src: ">/dev/null echo foo 2>&1",
		cmd: simple_command(
			redir(nil, 1, 1, ">", word(lit(1, 2, "/dev/null"))),
			word(lit(1, 12, "echo")),
			word(lit(1, 17, "foo")),
			redir(lit(1, 21, "2"), 1, 22, ">&", word(lit(1, 24, "1"))),
		),
	},
	{
		src: "echo foo >/dev/null 2>&1",
		cmd: simple_command(
			word(lit(1, 1, "echo")),
			word(lit(1, 6, "foo")),
			redir(nil, 1, 10, ">", word(lit(1, 11, "/dev/null"))),
			redir(lit(1, 21, "2"), 1, 22, ">&", word(lit(1, 24, "1"))),
		),
	},
	// comment
	{
		src: "# comment",
		comments: []*ast.Comment{
			comment(1, 1, " comment"),
		},
	},
	{
		src: "# comment\n\ngo version",
		cmd: simple_command(
			word(lit(3, 1, "go")),
			word(lit(3, 4, "version")),
		),
		comments: []*ast.Comment{
			comment(1, 1, " comment"),
		},
	},
	{
		src: "go version# comment\n",
		cmd: simple_command(
			word(lit(1, 1, "go")),
			word(lit(1, 4, "version")),
		),
		comments: []*ast.Comment{
			comment(1, 11, " comment"),
		},
	},
	// pipeline
	{
		src: "echo foo | grep o",
		cmd: pipeline(
			simple_command(
				word(lit(1, 1, "echo")),
				word(lit(1, 6, "foo")),
			),
			pipe(1, 10, "|", simple_command(
				word(lit(1, 12, "grep")),
				word(lit(1, 17, "o")),
			)),
		),
	},
	{
		src: "! echo foo | grep x",
		cmd: pipeline(
			pos(1, 1), // !
			simple_command(
				word(lit(1, 3, "echo")),
				word(lit(1, 8, "foo")),
			),
			pipe(1, 12, "|", simple_command(
				word(lit(1, 14, "grep")),
				word(lit(1, 19, "x")),
			)),
		),
	},
	{
		src: "! echo foo |\n\n# | comment\n\ngrep x",
		cmd: pipeline(
			pos(1, 1), // !
			simple_command(
				word(lit(1, 3, "echo")),
				word(lit(1, 8, "foo")),
			),
			pipe(1, 12, "|", simple_command(
				word(lit(5, 1, "grep")),
				word(lit(5, 6, "x")),
			)),
		),
		comments: []*ast.Comment{
			comment(3, 1, " | comment"),
		},
	},
	// list
	{
		src: "sleep 7;",
		cmd: and_or_list(
			simple_command(
				word(lit(1, 1, "sleep")),
				word(lit(1, 7, "7")),
			),
			sep(1, 8, ";"),
		),
	},
	{
		src: "cd; pwd",
		cmd: list(
			and_or_list(
				simple_command(
					word(lit(1, 1, "cd")),
				),
				sep(1, 3, ";"),
			),
			and_or_list(
				simple_command(
					word(lit(1, 5, "pwd")),
				),
			),
		),
	},
	{
		src: "cd;\npwd",
		cmd: and_or_list(
			simple_command(
				word(lit(1, 1, "cd")),
			),
			sep(1, 3, ";"),
		),
	},
	{
		src: "sleep 7 &",
		cmd: and_or_list(
			simple_command(
				word(lit(1, 1, "sleep")),
				word(lit(1, 7, "7")),
			),
			sep(1, 9, "&"),
		),
	},
	{
		src: "make & wait",
		cmd: list(
			and_or_list(
				simple_command(
					word(lit(1, 1, "make")),
				),
				sep(1, 6, "&"),
			),
			and_or_list(
				simple_command(
					word(lit(1, 8, "wait")),
				),
			),
		),
	},
	{
		src: "make &\nwait",
		cmd: and_or_list(
			simple_command(
				word(lit(1, 1, "make")),
			),
			sep(1, 6, "&"),
		),
	},
	{
		src: "false && echo foo || echo bar",
		cmd: and_or_list(
			simple_command(
				word(lit(1, 1, "false")),
			),
			and_or(1, 7, "&&", pipeline(
				simple_command(
					word(lit(1, 10, "echo")),
					word(lit(1, 15, "foo")),
				),
			)),
			and_or(1, 19, "||", pipeline(
				simple_command(
					word(lit(1, 22, "echo")),
					word(lit(1, 27, "bar")),
				),
			)),
		),
	},
	{
		src: "true || echo foo && echo bar",
		cmd: and_or_list(
			simple_command(
				word(lit(1, 1, "true")),
			),
			and_or(1, 6, "||", pipeline(
				simple_command(
					word(lit(1, 9, "echo")),
					word(lit(1, 14, "foo")),
				),
			)),
			and_or(1, 18, "&&", pipeline(
				simple_command(
					word(lit(1, 21, "echo")),
					word(lit(1, 26, "bar")),
				),
			)),
		),
	},
	{
		src: "true ||\n\n# || comment\n\necho foo &&\n\n# && comment\n\necho bar",
		cmd: and_or_list(
			simple_command(
				word(lit(1, 1, "true")),
			),
			and_or(1, 6, "||", pipeline(
				simple_command(
					word(lit(5, 1, "echo")),
					word(lit(5, 6, "foo")),
				),
			)),
			and_or(5, 10, "&&", pipeline(
				simple_command(
					word(lit(9, 1, "echo")),
					word(lit(9, 6, "bar")),
				),
			)),
		),
		comments: []*ast.Comment{
			comment(3, 1, " || comment"),
			comment(7, 1, " && comment"),
		},
	},
	// grouping command
	{
		src: "(cd /usr/src/linux; make -j3)",
		cmd: subshell(
			pos(1, 1), // (
			list(
				and_or_list(
					simple_command(
						word(lit(1, 2, "cd")),
						word(lit(1, 5, "/usr/src/linux")),
					),
					sep(1, 19, ";"),
				),
				and_or_list(
					simple_command(
						word(lit(1, 21, "make")),
						word(lit(1, 26, "-j3")),
					),
				),
			),
			pos(1, 29), // )
		),
	},
	{
		src: "(\n\tcd /usr/src/linux; make -j3\n)",
		cmd: subshell(
			pos(1, 1), // (
			list(
				and_or_list(
					simple_command(
						word(lit(2, 2, "cd")),
						word(lit(2, 5, "/usr/src/linux")),
					),
					sep(2, 19, ";"),
				),
				and_or_list(
					simple_command(
						word(lit(2, 21, "make")),
						word(lit(2, 26, "-j3")),
					),
				),
			),
			pos(3, 1), // )
		),
	},
	{
		src: "(\n\tcd /usr/src/linux\n\tmake -j3\n)",
		cmd: subshell(
			pos(1, 1), // (
			simple_command(
				word(lit(2, 2, "cd")),
				word(lit(2, 5, "/usr/src/linux")),
			),
			simple_command(
				word(lit(3, 2, "make")),
				word(lit(3, 7, "-j3")),
			),
			pos(4, 1), // )
		),
	},
	{
		src: "(cd /usr/src/linux; make -j3) >/dev/null 2>&1",
		cmd: subshell(
			pos(1, 1), // (
			list(
				and_or_list(
					simple_command(
						word(lit(1, 2, "cd")),
						word(lit(1, 5, "/usr/src/linux")),
					),
					sep(1, 19, ";"),
				),
				and_or_list(
					simple_command(
						word(lit(1, 21, "make")),
						word(lit(1, 26, "-j3")),
					),
				),
			),
			pos(1, 29), // )
			redir(nil, 1, 31, ">", word(lit(1, 32, "/dev/null"))),
			redir(lit(1, 42, "2"), 1, 43, ">&", word(lit(1, 45, "1"))),
		),
	},
	{
		src: "{ ./configure; make -j3; }",
		cmd: group(
			pos(1, 1), // {
			list(
				and_or_list(
					simple_command(
						word(lit(1, 3, "./configure")),
					),
					sep(1, 14, ";"),
				),
				and_or_list(
					simple_command(
						word(lit(1, 16, "make")),
						word(lit(1, 21, "-j3")),
					),
					sep(1, 24, ";"),
				),
			),
			pos(1, 26), // }
		),
	},
	{
		src: "{\n\t./configure; make -j3\n}",
		cmd: group(
			pos(1, 1), // {
			list(
				and_or_list(
					simple_command(
						word(lit(2, 2, "./configure")),
					),
					sep(2, 13, ";"),
				),
				and_or_list(
					simple_command(
						word(lit(2, 15, "make")),
						word(lit(2, 20, "-j3")),
					),
				),
			),
			pos(3, 1), // }
		),
	},
	{
		src: "{\n\t./configure\n\tmake -j3\n}",
		cmd: group(
			pos(1, 1), // {
			simple_command(
				word(lit(2, 2, "./configure")),
			),
			simple_command(
				word(lit(3, 2, "make")),
				word(lit(3, 7, "-j3")),
			),
			pos(4, 1), // }
		),
	},
	{
		src: "{ ./configure; make; } >/dev/null 2>&1",
		cmd: group(
			pos(1, 1), // {
			list(
				and_or_list(
					simple_command(
						word(lit(1, 3, "./configure")),
					),
					sep(1, 14, ";"),
				),
				and_or_list(
					simple_command(
						word(lit(1, 16, "make")),
					),
					sep(1, 20, ";"),
				),
			),
			pos(1, 22), // }
			redir(nil, 1, 24, ">", word(lit(1, 25, "/dev/null"))),
			redir(lit(1, 35, "2"), 1, 36, ">&", word(lit(1, 38, "1"))),
		),
	},
	// arithmetic evaluation
	{
		src: "((x -= 1))",
		cmd: arith_eval(
			pos(1, 1), // left
			word(lit(1, 3, "x"), lit(1, 5, "-="), lit(1, 8, "1")),
			pos(1, 9), // right
		),
	},
	{
		src: "((\n\tx += 1\n))",
		cmd: arith_eval(
			pos(1, 1), // left
			word(lit(2, 2, "x"), lit(2, 4, "+="), lit(2, 7, "1")),
			pos(3, 1), // right
		),
	},
	{
		src: "((x <<= 1)) >/dev/null 2>&1",
		cmd: arith_eval(
			pos(1, 1), // left
			word(lit(1, 3, "x"), lit(1, 5, "<<="), lit(1, 9, "1")),
			pos(1, 10), // right
			redir(nil, 1, 13, ">", word(lit(1, 14, "/dev/null"))),
			redir(lit(1, 24, "2"), 1, 25, ">&", word(lit(1, 27, "1"))),
		),
	},
	// for loop
	{
		src: "for name do echo $name; done",
		cmd: for_clause(
			pos(1, 1), // for
			lit(1, 5, "name"),
			sep(0, 0, ""),
			pos(1, 10), // do
			and_or_list(
				simple_command(
					word(lit(1, 13, "echo")),
					word(param_exp(1, 18, false, lit(1, 19, "name"), nil, nil)),
				),
				sep(1, 23, ";"),
			),
			pos(1, 25), // done
		),
	},
	{
		src: "for name; do echo $name; done",
		cmd: for_clause(
			pos(1, 1), // for
			lit(1, 5, "name"),
			sep(1, 9, ";"),
			pos(1, 11), // do
			and_or_list(
				simple_command(
					word(lit(1, 14, "echo")),
					word(param_exp(1, 19, false, lit(1, 20, "name"), nil, nil)),
				),
				sep(1, 24, ";"),
			),
			pos(1, 26), // done
		),
	},
	{
		src: "for name do\n\techo $name\ndone",
		cmd: for_clause(
			pos(1, 1), // for
			lit(1, 5, "name"),
			sep(0, 0, ""),
			pos(1, 10), // do
			simple_command(
				word(lit(2, 2, "echo")),
				word(param_exp(2, 7, false, lit(2, 8, "name"), nil, nil)),
			),
			pos(3, 1), // done
		),
	},
	{
		src: "for name; do\n\techo $name\ndone",
		cmd: for_clause(
			pos(1, 1), // for
			lit(1, 5, "name"),
			sep(1, 9, ";"),
			pos(1, 11), // do
			simple_command(
				word(lit(2, 2, "echo")),
				word(param_exp(2, 7, false, lit(2, 8, "name"), nil, nil)),
			),
			pos(3, 1), // done
		),
	},
	{
		src: "for name\ndo\n\techo $name\ndone",
		cmd: for_clause(
			pos(1, 1), // for
			lit(1, 5, "name"),
			sep(0, 0, ""),
			pos(2, 1), // do
			simple_command(
				word(lit(3, 2, "echo")),
				word(param_exp(3, 7, false, lit(3, 8, "name"), nil, nil)),
			),
			pos(4, 1), // done
		),
	},
	{
		src: "for name in; do echo $name; done",
		cmd: for_clause(
			pos(1, 1), // for
			lit(1, 5, "name"),
			pos(1, 10), // in
			sep(1, 12, ";"),
			pos(1, 14), // do
			and_or_list(
				simple_command(
					word(lit(1, 17, "echo")),
					word(param_exp(1, 22, false, lit(1, 23, "name"), nil, nil)),
				),
				sep(1, 27, ";"),
			),
			pos(1, 29), // done
		),
	},
	{
		src: "for name in; do\n\techo $name\ndone",
		cmd: for_clause(
			pos(1, 1), // for
			lit(1, 5, "name"),
			pos(1, 10), // in
			sep(1, 12, ";"),
			pos(1, 14), // do
			simple_command(
				word(lit(2, 2, "echo")),
				word(param_exp(2, 7, false, lit(2, 8, "name"), nil, nil)),
			),
			pos(3, 1), // done
		),
	},
	{
		src: "for name in\ndo\n\techo $name\ndone",
		cmd: for_clause(
			pos(1, 1), // for
			lit(1, 5, "name"),
			pos(1, 10), // in
			sep(0, 0, ""),
			pos(2, 1), // do
			simple_command(
				word(lit(3, 2, "echo")),
				word(param_exp(3, 7, false, lit(3, 8, "name"), nil, nil)),
			),
			pos(4, 1), // done
		),
	},
	{
		src: "for name\nin\ndo\n\techo $name\ndone",
		cmd: for_clause(
			pos(1, 1), // for
			lit(1, 5, "name"),
			pos(2, 1), // in
			sep(0, 0, ""),
			pos(3, 1), // do
			simple_command(
				word(lit(4, 2, "echo")),
				word(param_exp(4, 7, false, lit(4, 8, "name"), nil, nil)),
			),
			pos(5, 1), // done
		),
	},
	{
		src: "for name in foo bar baz; do echo $name; done >/dev/null 2>&1",
		cmd: for_clause(
			pos(1, 1), // for
			lit(1, 5, "name"),
			pos(1, 10), // in
			word(lit(1, 13, "foo")),
			word(lit(1, 17, "bar")),
			word(lit(1, 21, "baz")),
			sep(1, 24, ";"),
			pos(1, 26), // do
			and_or_list(
				simple_command(
					word(lit(1, 29, "echo")),
					word(param_exp(1, 34, false, lit(1, 35, "name"), nil, nil)),
				),
				sep(1, 39, ";"),
			),
			pos(1, 41), // done
			redir(nil, 1, 46, ">", word(lit(1, 47, "/dev/null"))),
			redir(lit(1, 57, "2"), 1, 58, ">&", word(lit(1, 60, "1"))),
		),
	},
	{
		src: "for name in foo bar baz; do\n\techo $name\ndone >/dev/null 2>&1",
		cmd: for_clause(
			pos(1, 1), // for
			lit(1, 5, "name"),
			pos(1, 10), // in
			word(lit(1, 13, "foo")),
			word(lit(1, 17, "bar")),
			word(lit(1, 21, "baz")),
			sep(1, 24, ";"),
			pos(1, 26), // do
			simple_command(
				word(lit(2, 2, "echo")),
				word(param_exp(2, 7, false, lit(2, 8, "name"), nil, nil)),
			),
			pos(3, 1), // done
			redir(nil, 3, 6, ">", word(lit(3, 7, "/dev/null"))),
			redir(lit(3, 17, "2"), 3, 18, ">&", word(lit(3, 20, "1"))),
		),
	},
	{
		src: "for name in foo bar baz\ndo\n\techo $name\ndone >/dev/null 2>&1",
		cmd: for_clause(
			pos(1, 1), // for
			lit(1, 5, "name"),
			pos(1, 10), // in
			word(lit(1, 13, "foo")),
			word(lit(1, 17, "bar")),
			word(lit(1, 21, "baz")),
			sep(0, 0, ""),
			pos(2, 1), // do
			simple_command(
				word(lit(3, 2, "echo")),
				word(param_exp(3, 7, false, lit(3, 8, "name"), nil, nil)),
			),
			pos(4, 1), // done
			redir(nil, 4, 6, ">", word(lit(4, 7, "/dev/null"))),
			redir(lit(4, 17, "2"), 4, 18, ">&", word(lit(4, 20, "1"))),
		),
	},
	{
		src: "for name\nin foo bar baz\ndo\n\techo $name\ndone >/dev/null 2>&1",
		cmd: for_clause(
			pos(1, 1), // for
			lit(1, 5, "name"),
			pos(2, 1), // in
			word(lit(2, 4, "foo")),
			word(lit(2, 8, "bar")),
			word(lit(2, 12, "baz")),
			sep(0, 0, ""),
			pos(3, 1), // do
			simple_command(
				word(lit(4, 2, "echo")),
				word(param_exp(4, 7, false, lit(4, 8, "name"), nil, nil)),
			),
			pos(5, 1), // done
			redir(nil, 5, 6, ">", word(lit(5, 7, "/dev/null"))),
			redir(lit(5, 17, "2"), 5, 18, ">&", word(lit(5, 20, "1"))),
		),
	},
	// case conditional construct
	{
		src: "case word in esac",
		cmd: case_clause(
			pos(1, 1), // case
			word(lit(1, 6, "word")),
			pos(1, 11), // in
			pos(1, 14), // esac
		),
	},
	{
		src: "case word in\nesac",
		cmd: case_clause(
			pos(1, 1), // case
			word(lit(1, 6, "word")),
			pos(1, 11), // in
			pos(2, 1),  // esac
		),
	},
	{
		src: "case word\nin\nesac",
		cmd: case_clause(
			pos(1, 1), // case
			word(lit(1, 6, "word")),
			pos(2, 1), // in
			pos(3, 1), // esac
		),
	},
	{
		src: "case word in 0|foo) ;; 1|bar) echo bar ;; (2|baz) ;; (3|qux) echo qux ;; esac",
		cmd: case_clause(
			pos(1, 1), // case
			word(lit(1, 6, "word")),
			pos(1, 11), // in
			case_item(
				word(lit(1, 14, "0")),
				word(lit(1, 16, "foo")),
				pos(1, 19), // )
				pos(1, 21), // ;;
			),
			case_item(
				word(lit(1, 24, "1")),
				word(lit(1, 26, "bar")),
				pos(1, 29), // )
				simple_command(
					word(lit(1, 31, "echo")),
					word(lit(1, 36, "bar")),
				),
				pos(1, 40), // ;;
			),
			case_item(
				pos(1, 43), // (
				word(lit(1, 44, "2")),
				word(lit(1, 46, "baz")),
				pos(1, 49), // )
				pos(1, 51), // ;;
			),
			case_item(
				pos(1, 54), // (
				word(lit(1, 55, "3")),
				word(lit(1, 57, "qux")),
				pos(1, 60), // )
				simple_command(
					word(lit(1, 62, "echo")),
					word(lit(1, 67, "qux")),
				),
				pos(1, 71), // ;;
			),
			pos(1, 74), // esac
		),
	},
	{
		src: "case word in\n0|foo)\n\t;;\n1|bar)\n\techo bar\n\t;;\n(2|baz)\n\t;;\n(3|qux)\n\techo qux\n\t;;\nesac",
		cmd: case_clause(
			pos(1, 1), // case
			word(lit(1, 6, "word")),
			pos(1, 11), // in
			case_item(
				word(lit(2, 1, "0")),
				word(lit(2, 3, "foo")),
				pos(2, 6), // )
				pos(3, 2), // ;;
			),
			case_item(
				word(lit(4, 1, "1")),
				word(lit(4, 3, "bar")),
				pos(4, 6), // )
				simple_command(
					word(lit(5, 2, "echo")),
					word(lit(5, 7, "bar")),
				),
				pos(6, 2), // ;;
			),
			case_item(
				pos(7, 1), // (
				word(lit(7, 2, "2")),
				word(lit(7, 4, "baz")),
				pos(7, 7), // )
				pos(8, 2), // ;;
			),
			case_item(
				pos(9, 1), // (
				word(lit(9, 2, "3")),
				word(lit(9, 4, "qux")),
				pos(9, 7), // )
				simple_command(
					word(lit(10, 2, "echo")),
					word(lit(10, 7, "qux")),
				),
				pos(11, 2), // ;;
			),
			pos(12, 1), // esac
		),
	},
	{
		src: "case word in 0|foo) ;; 1|bar) echo bar ;; (2|baz) ;; (3|qux) echo qux; esac",
		cmd: case_clause(
			pos(1, 1), // case
			word(lit(1, 6, "word")),
			pos(1, 11), // in
			case_item(
				word(lit(1, 14, "0")),
				word(lit(1, 16, "foo")),
				pos(1, 19), // )
				pos(1, 21), // ;;
			),
			case_item(
				word(lit(1, 24, "1")),
				word(lit(1, 26, "bar")),
				pos(1, 29), // )
				simple_command(
					word(lit(1, 31, "echo")),
					word(lit(1, 36, "bar")),
				),
				pos(1, 40), // ;;
			),
			case_item(
				pos(1, 43), // (
				word(lit(1, 44, "2")),
				word(lit(1, 46, "baz")),
				pos(1, 49), // )
				pos(1, 51), // ;;
			),
			case_item(
				pos(1, 54), // (
				word(lit(1, 55, "3")),
				word(lit(1, 57, "qux")),
				pos(1, 60), // )
				and_or_list(
					simple_command(
						word(lit(1, 62, "echo")),
						word(lit(1, 67, "qux")),
					),
					sep(1, 70, ";"),
				),
			),
			pos(1, 72), // esac
		),
	},
	{
		src: "case word in\n0|foo)\n\t;;\n1|bar)\n\techo bar\n\t;;\n(2|baz)\n\t;;\n(3|qux)\n\techo qux\nesac",
		cmd: case_clause(
			pos(1, 1), // case
			word(lit(1, 6, "word")),
			pos(1, 11), // in
			case_item(
				word(lit(2, 1, "0")),
				word(lit(2, 3, "foo")),
				pos(2, 6), // )
				pos(3, 2), // ;;
			),
			case_item(
				word(lit(4, 1, "1")),
				word(lit(4, 3, "bar")),
				pos(4, 6), // )
				simple_command(
					word(lit(5, 2, "echo")),
					word(lit(5, 7, "bar")),
				),
				pos(6, 2), // ;;
			),
			case_item(
				pos(7, 1), // (
				word(lit(7, 2, "2")),
				word(lit(7, 4, "baz")),
				pos(7, 7), // )
				pos(8, 2), // ;;
			),
			case_item(
				pos(9, 1), // (
				word(lit(9, 2, "3")),
				word(lit(9, 4, "qux")),
				pos(9, 7), // )
				simple_command(
					word(lit(10, 2, "echo")),
					word(lit(10, 7, "qux")),
				),
			),
			pos(11, 1), // esac
		),
	},
	{
		src: "case word in\nfoo) ;;\nbar)\n\techo bar ;;\nbaz)\n\techo baz\n\t;;\nesac >/dev/null 2>&1",
		cmd: case_clause(
			pos(1, 1), // case
			word(lit(1, 6, "word")),
			pos(1, 11), // in
			case_item(
				word(lit(2, 1, "foo")),
				pos(2, 4), // )
				pos(2, 6), // ;;
			),
			case_item(
				word(lit(3, 1, "bar")),
				pos(3, 4), // )
				simple_command(
					word(lit(4, 2, "echo")),
					word(lit(4, 7, "bar")),
				),
				pos(4, 11), // ;;
			),
			case_item(
				word(lit(5, 1, "baz")),
				pos(5, 4), // )
				simple_command(
					word(lit(6, 2, "echo")),
					word(lit(6, 7, "baz")),
				),
				pos(7, 2), // ;;
			),
			pos(8, 1), // esac
			redir(nil, 8, 6, ">", word(lit(8, 7, "/dev/null"))),
			redir(lit(8, 17, "2"), 8, 18, ">&", word(lit(8, 20, "1"))),
		),
	},
	// if conditional construct
	{
		src: "if true; then echo if; fi",
		cmd: if_clause(
			pos(1, 1), // if
			and_or_list(
				simple_command(
					word(lit(1, 4, "true")),
				),
				sep(1, 8, ";"),
			),
			pos(1, 10), // then
			and_or_list(
				simple_command(
					word(lit(1, 15, "echo")),
					word(lit(1, 20, "if")),
				),
				sep(1, 22, ";"),
			),
			pos(1, 24), // fi
		),
	},
	{
		src: "if true; then\n\techo if\nfi",
		cmd: if_clause(
			pos(1, 1), // if
			and_or_list(
				simple_command(
					word(lit(1, 4, "true")),
				),
				sep(1, 8, ";"),
			),
			pos(1, 10), // then
			simple_command(
				word(lit(2, 2, "echo")),
				word(lit(2, 7, "if")),
			),
			pos(3, 1), // fi
		),
	},
	{
		src: "if true\nthen\n\techo if\nfi",
		cmd: if_clause(
			pos(1, 1), // if
			simple_command(
				word(lit(1, 4, "true")),
			),
			pos(2, 1), // then
			simple_command(
				word(lit(3, 2, "echo")),
				word(lit(3, 7, "if")),
			),
			pos(4, 1), // fi
		),
	},
	{
		src: "if false; then\n\techo if\nelif true; then\n\techo elif\nfi",
		cmd: if_clause(
			pos(1, 1), // if
			and_or_list(
				simple_command(
					word(lit(1, 4, "false")),
				),
				sep(1, 9, ";"),
			),
			pos(1, 11), // then
			simple_command(
				word(lit(2, 2, "echo")),
				word(lit(2, 7, "if")),
			),
			elif_clause(
				pos(3, 1), // elif
				and_or_list(
					simple_command(
						word(lit(3, 6, "true")),
					),
					sep(3, 10, ";"),
				),
				pos(3, 12), // then
				simple_command(
					word(lit(4, 2, "echo")),
					word(lit(4, 7, "elif")),
				),
			),
			pos(5, 1), // fi
		),
	},
	{
		src: "if false; then\n\techo if\nelse\n\techo else\nfi",
		cmd: if_clause(
			pos(1, 1), // if
			and_or_list(
				simple_command(
					word(lit(1, 4, "false")),
				),
				sep(1, 9, ";"),
			),
			pos(1, 11), // then
			simple_command(
				word(lit(2, 2, "echo")),
				word(lit(2, 7, "if")),
			),
			else_clause(
				pos(3, 1), // else
				simple_command(
					word(lit(4, 2, "echo")),
					word(lit(4, 7, "else")),
				),
			),
			pos(5, 1), // fi
		),
	},
	{
		src: "if false; then\n\techo if\nelif false; then\n\techo elif\nelse\n\techo else\nfi >/dev/null 2>&1",
		cmd: if_clause(
			pos(1, 1), // if
			and_or_list(
				simple_command(
					word(lit(1, 4, "false")),
				),
				sep(1, 9, ";"),
			),
			pos(1, 11), // then
			simple_command(
				word(lit(2, 2, "echo")),
				word(lit(2, 7, "if")),
			),
			elif_clause(
				pos(3, 1), // elif
				and_or_list(
					simple_command(
						word(lit(3, 6, "false")),
					),
					sep(3, 11, ";"),
				),
				pos(3, 13), // then
				simple_command(
					word(lit(4, 2, "echo")),
					word(lit(4, 7, "elif")),
				),
			),
			else_clause(
				pos(5, 1), // else
				simple_command(
					word(lit(6, 2, "echo")),
					word(lit(6, 7, "else")),
				),
			),
			pos(7, 1), // fi
			redir(nil, 7, 4, ">", word(lit(7, 5, "/dev/null"))),
			redir(lit(7, 15, "2"), 7, 16, ">&", word(lit(7, 18, "1"))),
		),
	},
	// while loop
	{
		src: "while true; do echo while; done",
		cmd: while_clause(
			pos(1, 1), // while
			and_or_list(
				simple_command(
					word(lit(1, 7, "true")),
				),
				sep(1, 11, ";"),
			),
			pos(1, 13), // do
			and_or_list(
				simple_command(
					word(lit(1, 16, "echo")),
					word(lit(1, 21, "while")),
				),
				sep(1, 26, ";"),
			),
			pos(1, 28), // done
		),
	},
	{
		src: "while true\ndo\n\techo while\ndone",
		cmd: while_clause(
			pos(1, 1), // while
			simple_command(
				word(lit(1, 7, "true")),
			),
			pos(2, 1), // do
			simple_command(
				word(lit(3, 2, "echo")),
				word(lit(3, 7, "while")),
			),
			pos(4, 1), // done
		),
	},
	{
		src: "while true; do\n\techo while\ndone >/dev/null 2>&1",
		cmd: while_clause(
			pos(1, 1), // while
			and_or_list(
				simple_command(
					word(lit(1, 7, "true")),
				),
				sep(1, 11, ";"),
			),
			pos(1, 13), // do
			simple_command(
				word(lit(2, 2, "echo")),
				word(lit(2, 7, "while")),
			),
			pos(3, 1), // done
			redir(nil, 3, 6, ">", word(lit(3, 7, "/dev/null"))),
			redir(lit(3, 17, "2"), 3, 18, ">&", word(lit(3, 20, "1"))),
		),
	},
	// until loop
	{
		src: "until false; do echo until; done",
		cmd: until_clause(
			pos(1, 1), // until
			and_or_list(
				simple_command(
					word(lit(1, 7, "false")),
				),
				sep(1, 12, ";"),
			),
			pos(1, 14), // do
			and_or_list(
				simple_command(
					word(lit(1, 17, "echo")),
					word(lit(1, 22, "until")),
				),
				sep(1, 27, ";"),
			),
			pos(1, 29), // done
		),
	},
	{
		src: "until false\ndo\n\techo until\ndone",
		cmd: until_clause(
			pos(1, 1), // until
			simple_command(
				word(lit(1, 7, "false")),
			),
			pos(2, 1), // do
			simple_command(
				word(lit(3, 2, "echo")),
				word(lit(3, 7, "until")),
			),
			pos(4, 1), // done
		),
	},
	{
		src: "until false; do\n\techo until\ndone >/dev/null 2>&1",
		cmd: until_clause(
			pos(1, 1), // until
			and_or_list(
				simple_command(
					word(lit(1, 7, "false")),
				),
				sep(1, 12, ";"),
			),
			pos(1, 14), // do
			simple_command(
				word(lit(2, 2, "echo")),
				word(lit(2, 7, "until")),
			),
			pos(3, 1), // done
			redir(nil, 3, 6, ">", word(lit(3, 7, "/dev/null"))),
			redir(lit(3, 17, "2"), 3, 18, ">&", word(lit(3, 20, "1"))),
		),
	},
	// function definition command
	{
		src: "foo() { echo foo; }",
		cmd: func_def(
			lit(1, 1, "foo"),
			pos(1, 4), // (
			pos(1, 5), // )
			group(
				pos(1, 7), // {
				and_or_list(
					simple_command(
						word(lit(1, 9, "echo")),
						word(lit(1, 14, "foo")),
					),
					sep(1, 17, ";"),
				),
				pos(1, 19), // }
			),
		),
	},
	{
		src: "foo()\n{\n\techo foo\n}",
		cmd: func_def(
			lit(1, 1, "foo"),
			pos(1, 4), // (
			pos(1, 5), // )
			group(
				pos(2, 1), // {
				simple_command(
					word(lit(3, 2, "echo")),
					word(lit(3, 7, "foo")),
				),
				pos(4, 1), // }
			),
		),
	},
	{
		src: "foo() {\n\techo foo\n} >/dev/null 2>&1",
		cmd: func_def(
			lit(1, 1, "foo"),
			pos(1, 4), // (
			pos(1, 5), // )
			group(
				pos(1, 7), // {
				simple_command(
					word(lit(2, 2, "echo")),
					word(lit(2, 7, "foo")),
				),
				pos(3, 1), // }
				redir(nil, 3, 3, ">", word(lit(3, 4, "/dev/null"))),
				redir(lit(3, 14, "2"), 3, 15, ">&", word(lit(3, 17, "1"))),
			),
		),
	},
}

func TestParseCommand(t *testing.T) {
	for i, tt := range parseCommandTests {
		cmd, comments, err := parser.ParseCommand(fmt.Sprintf("%v.sh", i), tt.src)
		switch {
		case err != nil:
			t.Error(err)
		case !reflect.DeepEqual(cmd, tt.cmd):
			t.Errorf("unexpected command for %q", tt.src)
		case !reflect.DeepEqual(comments, tt.comments):
			t.Errorf("unexpected comments for %q", tt.src)
		}
	}
}

var pos = ast.NewPos

func list(list ...*ast.AndOrList) ast.List {
	return ast.List(list)
}

func sep(line, col int, sep string) *ast.Lit {
	return &ast.Lit{
		ValuePos: ast.NewPos(line, col),
		Value:    sep,
	}
}

func and_or_list(nodes ...ast.Node) *ast.AndOrList {
	cmd := new(ast.AndOrList)
	for _, n := range nodes {
		switch n := n.(type) {
		case *ast.Cmd:
			cmd.Pipeline = &ast.Pipeline{Cmd: n}
		case *ast.AndOr:
			cmd.List = append(cmd.List, n)
		case *ast.Lit:
			cmd.SepPos = n.ValuePos
			cmd.Sep = n.Value
		}
	}
	return cmd
}

func and_or(line, col int, op string, pipeline *ast.Pipeline) *ast.AndOr {
	return &ast.AndOr{
		OpPos:    ast.NewPos(line, col),
		Op:       op,
		Pipeline: pipeline,
	}
}

func pipeline(args ...interface{}) *ast.Pipeline {
	cmd := new(ast.Pipeline)
	for _, a := range args {
		switch a := a.(type) {
		case ast.Pos:
			cmd.Bang = a
		case *ast.Cmd:
			cmd.Cmd = a
		case *ast.Pipe:
			cmd.List = append(cmd.List, a)
		}
	}
	return cmd
}

func pipe(line, col int, op string, cmd *ast.Cmd) *ast.Pipe {
	return &ast.Pipe{
		OpPos: ast.NewPos(line, col),
		Op:    op,
		Cmd:   cmd,
	}
}

func simple_command(nodes ...ast.Node) *ast.Cmd {
	x := new(ast.SimpleCmd)
	cmd := &ast.Cmd{Expr: x}
	for _, n := range nodes {
		switch n := n.(type) {
		case *ast.Assign:
			x.Assigns = append(x.Assigns, n)
		case ast.Word:
			x.Args = append(x.Args, n)
		case *ast.Redir:
			cmd.Redirs = append(cmd.Redirs, n)
		}
	}
	return cmd
}

func subshell(args ...interface{}) *ast.Cmd {
	x := new(ast.Subshell)
	cmd := &ast.Cmd{Expr: x}
	pos := 0
	for _, a := range args {
		switch a := a.(type) {
		case ast.Pos:
			switch pos {
			case 0:
				x.Lparen = a
			case 1:
				x.Rparen = a
			}
			pos++
		case ast.Command:
			x.List = append(x.List, a)
		case *ast.Redir:
			cmd.Redirs = append(cmd.Redirs, a)
		}
	}
	return cmd
}

func group(args ...interface{}) *ast.Cmd {
	x := new(ast.Group)
	cmd := &ast.Cmd{Expr: x}
	pos := 0
	for _, a := range args {
		switch a := a.(type) {
		case ast.Pos:
			switch pos {
			case 0:
				x.Lbrace = a
			case 1:
				x.Rbrace = a
			}
			pos++
		case ast.Command:
			x.List = append(x.List, a)
		case *ast.Redir:
			cmd.Redirs = append(cmd.Redirs, a)
		}
	}
	return cmd
}

func arith_eval(args ...interface{}) *ast.Cmd {
	x := new(ast.ArithEval)
	cmd := &ast.Cmd{Expr: x}
	pos := 0
	for _, a := range args {
		switch a := a.(type) {
		case ast.Pos:
			switch pos {
			case 0:
				x.Left = a
			case 1:
				x.Right = a
			}
			pos++
		case ast.Word:
			x.Expr = a
		case *ast.Redir:
			cmd.Redirs = append(cmd.Redirs, a)
		}
	}
	return cmd
}

func for_clause(args ...interface{}) *ast.Cmd {
	x := new(ast.ForClause)
	cmd := &ast.Cmd{Expr: x}
	pos := 0
	lit := 0
	for _, a := range args {
		switch a := a.(type) {
		case ast.Pos:
			switch pos {
			case 0:
				x.For = a
			case 1:
				x.In = a
			case 2:
				x.Do = a
			case 3:
				x.Done = a
			}
			pos++
		case *ast.Lit:
			switch lit {
			case 0:
				x.Name = a
			case 1:
				x.Semicolon = a.ValuePos
				pos = 2
			}
			lit++
		case ast.Word:
			x.Items = append(x.Items, a)
		case ast.Command:
			x.List = append(x.List, a)
		case *ast.Redir:
			cmd.Redirs = append(cmd.Redirs, a)
		}
	}
	return cmd
}

func case_clause(args ...interface{}) *ast.Cmd {
	x := new(ast.CaseClause)
	cmd := &ast.Cmd{Expr: x}
	pos := 0
	for _, a := range args {
		switch a := a.(type) {
		case ast.Pos:
			switch pos {
			case 0:
				x.Case = a
			case 1:
				x.In = a
			case 2:
				x.Esac = a
			}
			pos++
		case ast.Word:
			x.Word = a
		case *ast.CaseItem:
			x.Items = append(x.Items, a)
		case *ast.Redir:
			cmd.Redirs = append(cmd.Redirs, a)
		}
	}
	return cmd
}

func case_item(args ...interface{}) *ast.CaseItem {
	ci := new(ast.CaseItem)
	pos := 0
	for _, a := range args {
		switch a := a.(type) {
		case ast.Pos:
			switch pos {
			case 0:
				ci.Lparen = a
			case 1:
				ci.Rparen = a
			case 2:
				ci.Break = a
			}
			pos++
		case ast.Word:
			ci.Patterns = append(ci.Patterns, a)
			if pos == 0 {
				pos = 1
			}
		case ast.Command:
			ci.List = append(ci.List, a)
		}
	}
	return ci
}

func if_clause(args ...interface{}) *ast.Cmd {
	x := new(ast.IfClause)
	cmd := &ast.Cmd{Expr: x}
	pos := 0
	for _, a := range args {
		switch a := a.(type) {
		case ast.Pos:
			switch pos {
			case 0:
				x.If = a
			case 1:
				x.Then = a
			case 2:
				x.Fi = a
			}
			pos++
		case ast.Command:
			switch pos {
			case 1:
				x.Cond = append(x.Cond, a)
			case 2:
				x.List = append(x.List, a)
			}
		case ast.ElsePart:
			x.Else = append(x.Else, a)
		case *ast.Redir:
			cmd.Redirs = append(cmd.Redirs, a)
		}
	}
	return cmd
}

func elif_clause(args ...interface{}) *ast.ElifClause {
	e := new(ast.ElifClause)
	pos := 0
	for _, a := range args {
		switch a := a.(type) {
		case ast.Pos:
			switch pos {
			case 0:
				e.Elif = a
			case 1:
				e.Then = a
			}
			pos++
		case ast.Command:
			switch pos {
			case 1:
				e.Cond = append(e.Cond, a)
			case 2:
				e.List = append(e.List, a)
			}
		}
	}
	return e
}

func else_clause(pos ast.Pos, list ...ast.Command) *ast.ElseClause {
	return &ast.ElseClause{
		Else: pos,
		List: list,
	}
}

func while_clause(args ...interface{}) *ast.Cmd {
	x := new(ast.WhileClause)
	cmd := &ast.Cmd{Expr: x}
	pos := 0
	for _, a := range args {
		switch a := a.(type) {
		case ast.Pos:
			switch pos {
			case 0:
				x.While = a
			case 1:
				x.Do = a
			case 2:
				x.Done = a
			}
			pos++
		case ast.Command:
			switch pos {
			case 1:
				x.Cond = append(x.Cond, a)
			case 2:
				x.List = append(x.List, a)
			}
		case *ast.Redir:
			cmd.Redirs = append(cmd.Redirs, a)
		}
	}
	return cmd
}

func until_clause(args ...interface{}) *ast.Cmd {
	x := new(ast.UntilClause)
	cmd := &ast.Cmd{Expr: x}
	pos := 0
	for _, a := range args {
		switch a := a.(type) {
		case ast.Pos:
			switch pos {
			case 0:
				x.Until = a
			case 1:
				x.Do = a
			case 2:
				x.Done = a
			}
			pos++
		case ast.Command:
			switch pos {
			case 1:
				x.Cond = append(x.Cond, a)
			case 2:
				x.List = append(x.List, a)
			}
		case *ast.Redir:
			cmd.Redirs = append(cmd.Redirs, a)
		}
	}
	return cmd
}

func func_def(name *ast.Lit, lparen, rparen ast.Pos, body ast.Command) *ast.Cmd {
	return &ast.Cmd{
		Expr: &ast.FuncDef{
			Name:   name,
			Lparen: lparen,
			Rparen: rparen,
			Body:   body,
		},
	}
}

func redir(n *ast.Lit, line, col int, op string, word ast.Word) *ast.Redir {
	return &ast.Redir{
		N:     n,
		OpPos: ast.NewPos(line, col),
		Op:    op,
		Word:  word,
	}
}

func heredoc(n *ast.Lit, line, col int, op string, word, heredoc, delim ast.Word) *ast.Redir {
	return &ast.Redir{
		N:       n,
		OpPos:   ast.NewPos(line, col),
		Op:      op,
		Word:    word,
		Heredoc: heredoc,
		Delim:   delim,
	}
}

func word(w ...ast.WordPart) ast.Word {
	if len(w) == 0 {
		return ast.Word{}
	}
	return ast.Word(w)
}

func lit(line, col int, v string) *ast.Lit {
	return &ast.Lit{
		ValuePos: ast.NewPos(line, col),
		Value:    v,
	}
}

func quote(line, col int, tok string, v ast.Word) *ast.Quote {
	return &ast.Quote{
		TokPos: ast.NewPos(line, col),
		Tok:    tok,
		Value:  v,
	}
}

func param_exp(line, col int, braces bool, name, op *ast.Lit, word ast.Word) *ast.ParamExp {
	pe := &ast.ParamExp{
		Dollar: ast.NewPos(line, col),
		Braces: braces,
		Name:   name,
		Word:   word,
	}
	if op != nil {
		pe.OpPos = op.ValuePos
		pe.Op = op.Value
	}
	return pe
}

func cmd_subst(args ...interface{}) *ast.CmdSubst {
	cs := new(ast.CmdSubst)
	pos := 0
	for _, a := range args {
		switch a := a.(type) {
		case bool:
			cs.Dollar = a
		case ast.Pos:
			switch pos {
			case 0:
				cs.Left = a
			case 1:
				cs.Right = a
			}
			pos++
		case ast.Command:
			cs.List = append(cs.List, a)
		}
	}
	return cs
}

func arith_exp(args ...interface{}) *ast.ArithExp {
	x := new(ast.ArithExp)
	pos := 0
	for _, a := range args {
		switch a := a.(type) {
		case ast.Pos:
			switch pos {
			case 0:
				x.Left = a
			case 1:
				x.Right = a
			}
			pos++
		case ast.Word:
			x.Expr = a
		}
	}
	return x
}

func assignment_word(line, col int, k string, v ast.Word) *ast.Assign {
	return &ast.Assign{
		Symbol: &ast.Lit{
			ValuePos: ast.NewPos(line, col),
			Value:    k,
		},
		Op:    "=",
		Value: v,
	}
}

func comment(line, col int, text string) *ast.Comment {
	return &ast.Comment{
		Hash: ast.NewPos(line, col),
		Text: text,
	}
}

var parseErrorTests = []struct {
	src, err string
}{
	// quoting
	{
		src: "'q",
		err: ":1:1: syntax error: reached EOF while parsing single-quotes",
	},
	{
		src: `"qq`,
		err: ":1:1: syntax error: reached EOF while parsing double-quotes",
	},
	{
		src: `"\`,
		err: ":1:1: syntax error: reached EOF while parsing double-quotes",
	},
	{
		src: `"${}`,
		err: ":1:2: syntax error: invalid parameter expansion",
	},
	{
		src: `"$(!)"`,
		err: ":1:5: syntax error: unexpected ')'",
	},
	{
		src: "\"`!`\"",
		err: ":1:4: syntax error: unexpected '`'",
	},
	// parameter expansion
	{
		src: "${}",
		err: ":1:1: syntax error: invalid parameter expansion",
	},
	{
		src: "${LANG:}",
		err: ":1:1: syntax error: invalid parameter expansion",
	},
	{
		src: "${LANG%#}",
		err: ":1:1: syntax error: invalid parameter expansion",
	},
	{
		src: "${LANG-'}",
		err: ":1:8: syntax error: reached EOF while parsing single-quotes",
	},
	{
		src: "${LANG-${}}",
		err: ":1:8: syntax error: invalid parameter expansion",
	},
	{
		src: "${LANG",
		err: ":1:1: syntax error: reached EOF while looking for matching '}'",
	},
	{
		src: "${LANG ",
		err: ":1:1: syntax error: reached EOF while looking for matching '}'",
	},
	// command substitution
	{
		src: "$(!",
		err: ":1:3: syntax error: unexpected EOF",
	},
	{
		src: "$(!)",
		err: ":1:4: syntax error: unexpected ')'",
	},
	{
		src: "$(echo $(!))",
		err: ":1:11: syntax error: unexpected ')'",
	},
	{
		src: "$(echo `!`)",
		err: ":1:10: syntax error: unexpected '`'",
	},
	{
		src: "`!",
		err: ":1:2: syntax error: unexpected EOF",
	},
	{
		src: "`!`",
		err: ":1:3: syntax error: unexpected '`'",
	},
	{
		src: "`echo $(!)`",
		err: ":1:10: syntax error: unexpected ')'",
	},
	// arithmetic expansion
	{
		src: "$((",
		err: ":1:1: syntax error: reached EOF while looking for matching '))'",
	},
	{
		src: "$((x)",
		err: ":1:1: syntax error: reached EOF while looking for matching '))'",
	},
	{
		src: "$((x) ",
		err: ":1:1: syntax error: reached EOF while looking for matching '))'",
	},
	{
		src: "$(('q",
		err: ":1:4: syntax error: reached EOF while parsing single-quotes",
	},
	{
		src: `$(("qq`,
		err: ":1:4: syntax error: reached EOF while parsing double-quotes",
	},
	{
		src: "$(($(!)))",
		err: ":1:7: syntax error: unexpected ')'",
	},
	{
		src: "$((`!`))",
		err: ":1:6: syntax error: unexpected '`'",
	},
	// simple command
	{
		src: "<",
		err: ":1:1: syntax error: unexpected EOF, expecting WORD",
	},
	{
		src: ">",
		err: ":1:1: syntax error: unexpected EOF, expecting WORD",
	},
	{
		src: "cat <<'\n'",
		err: `:1:7: syntax error: here-document delimiter contains '\n'`,
	},
	{
		src: "cat <<EOF\n",
		err: ":1:5: syntax error: here-document delimited by EOF",
	},
	{
		src: "cat <<EOF\n$(!",
		err: ":2:3: syntax error: unexpected EOF",
	},
	{
		src: "cat <<EOF\n`!",
		err: ":2:2: syntax error: unexpected EOF",
	},
	// pipeline
	{
		src: "!",
		err: ":1:1: syntax error: unexpected EOF",
	},
	{
		src: "echo foo | !",
		err: ":1:12: syntax error: unexpected '!'",
	},
	// list
	{
		src: ";",
		err: ":1:1: syntax error: unexpected ';'",
	},
	{
		src: "&",
		err: ":1:1: syntax error: unexpected '&'",
	},
	{
		src: "true &&",
		err: ":1:6: syntax error: unexpected EOF",
	},
	{
		src: "false ||",
		err: ":1:7: syntax error: unexpected EOF",
	},
	// grouping command
	{
		src: "(",
		err: ":1:1: syntax error: unexpected EOF",
	},
	{
		src: ")",
		err: ":1:1: syntax error: unexpected ')'",
	},
	{
		src: "{",
		err: ":1:1: syntax error: unexpected EOF",
	},
	{
		src: "}",
		err: ":1:1: syntax error: unexpected '}'",
	},
	// for loop
	{
		src: "for 1",
		err: ":1:5: syntax error: invalid for loop variable",
	},
	{
		src: "for >/dev/null",
		err: ":1:5: syntax error: unexpected '>', expecting NAME",
	},
	{
		src: "in",
		err: ":1:1: syntax error: unexpected 'in'",
	},
	{
		src: "for name\ndone",
		err: ":2:1: syntax error: unexpected 'done', expecting 'in'",
	},
	{
		src: "do",
		err: ":1:1: syntax error: unexpected 'do'",
	},
	{
		src: "for name; done",
		err: ":1:11: syntax error: unexpected 'done', expecting 'do'",
	},
	{
		src: "for name in; done",
		err: ":1:14: syntax error: unexpected 'done', expecting 'do'",
	},
	{
		src: "for name in\ndone",
		err: ":2:1: syntax error: unexpected 'done', expecting 'do'",
	},
	{
		src: "done",
		err: ":1:1: syntax error: unexpected 'done'",
	},
	{
		src: "for name; do",
		err: ":1:11: syntax error: unexpected EOF",
	},
	{
		src: "for name\ndo",
		err: ":2:1: syntax error: unexpected EOF",
	},
	// case conditional construct
	{
		src: "case",
		err: ":1:1: syntax error: unexpected EOF, expecting WORD",
	},
	{
		src: "case word",
		err: ":1:6: syntax error: unexpected EOF, expecting 'in'",
	},
	{
		src: "case word >/dev/null",
		err: ":1:11: syntax error: unexpected '>', expecting 'in'",
	},
	{
		src: "case word in",
		err: ":1:11: syntax error: unexpected EOF, expecting '(' or WORD or 'esac'",
	},
	{
		src: "case word in >)",
		err: ":1:14: syntax error: unexpected '>', expecting '(' or WORD or 'esac'",
	},
	{
		src: "case word in *)",
		err: ":1:15: syntax error: unexpected EOF, expecting 'esac'",
	},
	{
		src: "case word in *) ;;",
		err: ":1:17: syntax error: unexpected EOF, expecting '(' or WORD or 'esac'",
	},
	// if conditional construct
	{
		src: "if",
		err: ":1:1: syntax error: unexpected EOF",
	},
	{
		src: "elif",
		err: ":1:1: syntax error: unexpected 'elif'",
	},
	{
		src: "if false; then :;elif",
		err: ":1:18: syntax error: unexpected EOF",
	},
	{
		src: "then",
		err: ":1:1: syntax error: unexpected 'then'",
	},
	{
		src: "if true; fi",
		err: ":1:10: syntax error: unexpected 'fi', expecting 'then'",
	},
	{
		src: "if false; then :;elif true; fi",
		err: ":1:29: syntax error: unexpected 'fi', expecting 'then'",
	},
	{
		src: "else",
		err: ":1:1: syntax error: unexpected 'else'",
	},
	{
		src: "if false; then :;else",
		err: ":1:18: syntax error: unexpected EOF",
	},
	{
		src: "fi",
		err: ":1:1: syntax error: unexpected 'fi'",
	},
	// while loop
	{
		src: "while",
		err: ":1:1: syntax error: unexpected EOF",
	},
	{
		src: "do",
		err: ":1:1: syntax error: unexpected 'do'",
	},
	{
		src: "while true; done",
		err: ":1:13: syntax error: unexpected 'done', expecting 'do'",
	},
	{
		src: "while true\ndone",
		err: ":2:1: syntax error: unexpected 'done', expecting 'do'",
	},
	{
		src: "done",
		err: ":1:1: syntax error: unexpected 'done'",
	},
	{
		src: "while true; do",
		err: ":1:13: syntax error: unexpected EOF",
	},
	{
		src: "while true\ndo",
		err: ":2:1: syntax error: unexpected EOF",
	},
	// until loop
	{
		src: "until",
		err: ":1:1: syntax error: unexpected EOF",
	},
	{
		src: "do",
		err: ":1:1: syntax error: unexpected 'do'",
	},
	{
		src: "until true; done",
		err: ":1:13: syntax error: unexpected 'done', expecting 'do'",
	},
	{
		src: "until true\ndone",
		err: ":2:1: syntax error: unexpected 'done', expecting 'do'",
	},
	{
		src: "done",
		err: ":1:1: syntax error: unexpected 'done'",
	},
	{
		src: "until true; do",
		err: ":1:13: syntax error: unexpected EOF",
	},
	{
		src: "until true\ndo",
		err: ":2:1: syntax error: unexpected EOF",
	},
	// function definition command
	{
		src: "exit()",
		err: ":1:1: syntax error: invalid function name",
	},
	{
		src: "fname(",
		err: ":1:6: syntax error: unexpected EOF, expecting ')'",
	},
	{
		src: "fname()",
		err: ":1:7: syntax error: unexpected EOF",
	},
}

func TestParseError(t *testing.T) {
	for i, tt := range parseErrorTests {
		name := fmt.Sprintf("%v.sh", i)
		switch _, _, err := parser.ParseCommand(name, tt.src); {
		case err == nil:
			t.Error("expected error")
		case err.Error()[len(name):] != tt.err:
			t.Error("unexpected error:", err)
		}
	}
}

func TestOpen(t *testing.T) {
	for _, src := range []interface{}{
		[]byte{},
		"",
		new(bytes.Buffer),
		&reader{
			data: "",
			eof:  io.EOF,
		},
	} {
		if _, _, err := parser.ParseCommand("", src); err != nil {
			t.Error(err)
		}
	}

	if _, _, err := parser.ParseCommand("", nil); err == nil {
		t.Error("expected error")
	}
}

func TestReadError(t *testing.T) {
	for _, data := range []string{
		"",
		`\`,
		"'",
		`"`,
		`"\`,
		"$",
		"$_",
		"${",
		"${#",
		"${#-",
		"${_",
		"${@",
		"${_-",
		"_ <<_\n\\",
		"for _;",
		"for _\n",
		"for _ in;",
		"for _ in\n",
	} {
		src := &reader{
			data: data,
			eof:  errors.New("read error"),
		}
		if _, _, err := parser.ParseCommand("", src); err == nil {
			t.Error("expected error")
		}
	}
}

type reader struct {
	data string
	i    int
	eof  error
}

func (r *reader) Read(p []byte) (int, error) {
	if r.i < len(r.data) {
		p[0] = r.data[r.i]
		r.i++
		return 1, nil
	}
	return 0, r.eof
}
