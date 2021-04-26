//
// go.sh/interp :: interp_test.go
//
//   Copyright (c) 2021 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
//

package interp_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/hattya/go.sh/interp"
)

const name = "go.sh"

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

func tempDir() (string, error) {
	return ioutil.TempDir("", "go.sh")
}

func touch(s ...string) error {
	return ioutil.WriteFile(filepath.Join(s...), []byte{}, 0o666)
}
