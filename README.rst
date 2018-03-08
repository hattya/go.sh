go.sh
=====

A parser of the `Shell Command Language`_.

.. image:: https://godoc.org/github.com/hattya/go.sh?status.svg
   :target: https://godoc.org/github.com/hattya/go.sh

.. image:: https://semaphoreci.com/api/v1/hattya/go-sh/branches/master/badge.svg
   :target: https://semaphoreci.com/hattya/go-sh

.. image:: https://ci.appveyor.com/api/projects/status/ptsv6es9dq1nt3k9?svg=true
   :target: https://ci.appveyor.com/project/hattya/go-sh/branch/master

.. image:: https://codecov.io/gh/hattya/go.sh/branch/master/graph/badge.svg
   :target: https://codecov.io/gh/hattya/go.sh

.. _Shell Command Language: http://pubs.opengroup.org/onlinepubs/9699919799/utilities/V3_chap02.html


Installation
------------

.. code:: console

   $ go get -u github.com/hattya/go.sh


Usage
-----

.. code:: go

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


License
-------

go.sh is distributed under the terms of the MIT License.
