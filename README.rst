go.sh
=====

A parser of the `Shell Command Language`_.

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
   	cmd, err := parser.ParseCommand("<stdin>", "echo Hello, World!")
   	if err != nil {
   		fmt.Println(err)
   		return
   	}
   	spew.Dump(cmd)
   }


License
-------

go.sh is distributed under the terms of the MIT License.
