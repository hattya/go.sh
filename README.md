# go.sh

A parser of the [Shell Command Language](https://pubs.opengroup.org/onlinepubs/9699919799/utilities/V3_chap02.html).

[![pkg.go.dev](https://pkg.go.dev/badge/github.com/hattya/go.sh)](https://pkg.go.dev/github.com/hattya/go.sh)
[![GitHub Actions](https://github.com/hattya/go.sh/workflows/CI/badge.svg)](https://github.com/hattya/go.sh/actions?query=workflow:CI)
[![Semaphore](https://semaphoreci.com/api/v1/hattya/go-sh/branches/master/badge.svg)](https://semaphoreci.com/hattya/go-sh)
[![Appveyor](https://ci.appveyor.com/api/projects/status/ptsv6es9dq1nt3k9/branch/master?svg=true)](https://ci.appveyor.com/project/hattya/go-sh)
[![Codecov](https://codecov.io/gh/hattya/go.sh/branch/master/graph/badge.svg)](https://codecov.io/gh/hattya/go.sh)


## Installation

```console
$ go get -u github.com/hattya/go.sh
```


## Usage

```go
package main

import (
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/hattya/go.sh/parser"
)

func main() {
	cmd, comments, err := parser.ParseCommand("<stdin>", "echo Hello, World!")
	if err != nil {
		fmt.Println(err)
		return
	}
	spew.Dump(cmd)
	spew.Dump(comments)
}
```


## License

go.sh is distributed under the terms of the MIT License.
