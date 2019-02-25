// Copyright 2016-2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package main

import (
	"github.com/graniticio/granitic/v2/cmd/grnc-bind/binder"
	"github.com/graniticio/granitic/v2/logging"
	"github.com/graniticio/granitic/v2/test"
	"os"
	"path/filepath"
	"testing"
)

func TestBindProcess(t *testing.T) {

	tmp := os.TempDir()

	bindOut := filepath.Join(tmp, "bindings.go")

	compDir := test.FilePath("complete")
	merged := ""

	b := new(binder.Binder)
	b.ToolName = "bind-test"
	b.Loader = new(jsonDefinitionLoader)

	s := binder.Settings{
		CompDefLocation: &compDir,
		BindingsFile:    &bindOut,
		MergedDebugFile: &merged,
	}

	b.Log = new(logging.ConsoleErrorLogger)

	b.Bind(s)

	if _, err := os.Stat(bindOut); os.IsNotExist(err) {
		t.Fatalf("Expected bindings file %s does not exist: %v", bindOut, err)
	}

}

func TestOutputMerged(t *testing.T) {

	tmp := os.TempDir()

	bindOut := filepath.Join(tmp, "bindings.go")
	mergeOut := filepath.Join(tmp, "merged.json")

	compDir := test.FilePath("complete")

	b := new(binder.Binder)
	b.ToolName = "bind-test"
	b.Loader = new(jsonDefinitionLoader)

	b.Log = new(logging.ConsoleErrorLogger)

	s := binder.Settings{
		CompDefLocation: &compDir,
		BindingsFile:    &bindOut,
		MergedDebugFile: &mergeOut,
	}

	b.Bind(s)

	if _, err := os.Stat(mergeOut); os.IsNotExist(err) {
		t.Fatalf("Expected bindings file %s does not exist: %v", mergeOut, err)
	}

}
