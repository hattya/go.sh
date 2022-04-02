//
// go.sh/interp :: interp_test.go
//
//   Copyright (c) 2021-2022 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
//

package interp_test

import (
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"

	"github.com/hattya/go.sh/interp"
)

const name = "go.sh"

var optionTests = []struct {
	o interp.Option
	s string
}{
	{interp.AllExport, "a"},
	{interp.ErrExit, "e"},
	{interp.IgnoreEOF, ""},
	{interp.Monitor, "m"},
	{interp.NoClobber, "C"},
	{interp.NoGlob, "f"},
	{interp.NoExec, "n"},
	{interp.NoLog, ""},
	{interp.Notify, "b"},
	{interp.NoUnset, "u"},
	{interp.Verbose, "v"},
	{interp.Vi, ""},
	{interp.XTrace, "x"},
	{interp.AllExport | interp.IgnoreEOF | interp.NoClobber | interp.NoExec | interp.Notify | interp.Verbose | interp.XTrace, "aCnbvx"},
	{interp.ErrExit | interp.Monitor | interp.NoGlob | interp.NoLog | interp.NoUnset | interp.Vi, "emfu"},
}

func TestOption(t *testing.T) {
	for _, tt := range optionTests {
		if g, e := tt.o.String(), tt.s; g != e {
			t.Errorf("expected %q, got %q", e, g)
		}
	}
}

func TestVar(t *testing.T) {
	env := interp.NewExecEnv(name)
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

func TestSpParam(t *testing.T) {
	env := interp.NewExecEnv(name, "1")
	for _, tt := range []struct {
		name, value string
	}{
		{"#", "1"},
		{"?", "0"},
		{"-", ""},
		{"$", strconv.Itoa(os.Getpid())},
		{"!", ""},
		{"0", name},
	} {
		// get
		if v, _ := env.Get(tt.name); v.Value != tt.value {
			t.Errorf("expected $%v = %q, got %q", tt.name, tt.value, v.Value)
		}
		// set
		env.Set(tt.name, "_")
		if v, _ := env.Get(tt.name); v.Value != tt.value {
			t.Errorf("expected $%v = %q, got %q", tt.name, tt.value, v.Value)
		}
	}
}

func TestPosParam(t *testing.T) {
	env := interp.NewExecEnv(name, "1")
	// get
	e := interp.Var{
		Name:  "1",
		Value: "1",
	}
	switch g, set := env.Get("1"); {
	case !set:
		t.Errorf("expected set")
	case !reflect.DeepEqual(g, e):
		t.Errorf("expected %#v, got %#v", e, g)
	}
	// set
	env.Set("1", "01")
	switch g, set := env.Get("1"); {
	case !set:
		t.Errorf("expected set")
	case !reflect.DeepEqual(g, e):
		t.Errorf("expected %#v, got %#v", e, g)
	}
	env.Set("2", "2")
	if _, set := env.Get("2"); set {
		t.Errorf("expected unset")
	}
}

func pushd(path string) (func() error, error) {
	wd, err := os.Getwd()
	popd := func() error {
		if err != nil {
			return err
		}
		return os.Chdir(wd)
	}
	return popd, os.Chdir(path)
}

func touch(s ...string) error {
	return os.WriteFile(filepath.Join(s...), []byte{}, 0o666)
}
