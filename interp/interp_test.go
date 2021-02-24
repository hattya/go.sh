//
// go.sh/interp :: interp_test.go
//
//   Copyright (c) 2021 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
//

package interp_test

import (
	"reflect"
	"testing"

	"github.com/hattya/go.sh/interp"
)

func TestVar(t *testing.T) {
	env := interp.NewExecEnv()
	if _, set := env.Get("FOO"); set {
		t.Fatal("expected unset")
	}
	// set
	e := interp.Var{
		Name:  "FOO",
		Value: "foo",
	}
	env.Set("FOO", "foo")
	switch g, set := env.Get("FOO"); {
	case !set:
		t.Fatal("expected set")
	case !reflect.DeepEqual(g, e):
		t.Fatalf("expected %#v, got %#v", e, g)
	}
	// walk
	n := 0
	export := 0
	env.Walk(func(v interp.Var) {
		if v.Export {
			export++
		}
		n++
	})
	if export == n {
		t.Errorf("expected export < n; got export = %d, n = %d", export, n)
	}
	// unset
	env.Unset("FOO")
	if _, set := env.Get("FOO"); set {
		t.Fatalf("expected unset")
	}
	// walk
	n = 0
	export = 0
	env.Walk(func(v interp.Var) {
		if v.Export {
			export++
		}
		n++
	})
	if export != n {
		t.Errorf("expected export == n; got export = %d, n = %d", export, n)
	}
}
