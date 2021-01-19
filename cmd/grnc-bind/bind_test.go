// Copyright 2016-2020 Granitic. All rights reserved.
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
	ignore := ""
	extFac := false

	b := new(binder.Binder)
	b.ToolName = "bind-test"
	b.Loader = binder.NewJsonDefinitionLoader()

	s := binder.Settings{
		CompDefLocation:    &compDir,
		BindingsFile:       &bindOut,
		MergedDebugFile:    &merged,
		IgnoreModules:      &ignore,
		ExternalFacilities: &extFac,
	}

	b.Log = new(logging.ConsoleErrorLogger)

	b.Bind(s)

	if _, err := os.Stat(bindOut); os.IsNotExist(err) {
		t.Fatalf("Expected bindings file %s does not exist: %v", bindOut, err)
	}

	if b.Failed() {
		t.Fail()
	}
}

func TestManifestParse(t *testing.T) {

	manifestDir := test.FilePath("manifest")

	mfp := filepath.Join(manifestDir, "valid.json")

	b := binder.NewJsonDefinitionLoader()
	m, err := b.FacilityManifest(mfp)

	if m == nil {
		t.Errorf("Unexpected nil")
	}

	if err != nil {
		t.Errorf("Unexpected error %s", err.Error())
	}

	if m.ExternalFacilities == nil || len(m.ExternalFacilities) == 0 {
		t.Errorf("Expected definitions")
	}

	pm := m.ExternalFacilities["FacilityA"]

	if pm == nil {
		t.Errorf("Expected a definition")
	}

}

func TestManifestBadPath(t *testing.T) {

	manifestDir := test.FilePath("manifest")

	mfp := filepath.Join(manifestDir, "missing.json")

	b := binder.NewJsonDefinitionLoader()

	m, err := b.FacilityManifest(mfp)

	if m != nil {
		t.Errorf("Unexpected not nil")
	}

	if err == nil {
		t.Errorf("Expected error")
	}
}

func TestManifestMalformed(t *testing.T) {

	manifestDir := test.FilePath("manifest")

	mfp := filepath.Join(manifestDir, "malformed.json")

	b := binder.NewJsonDefinitionLoader()

	m, err := b.FacilityManifest(mfp)

	if m != nil {
		t.Errorf("Unexpected not nil")
	}

	if err == nil {
		t.Errorf("Expected error")
	}
}

func TestOutputMerged(t *testing.T) {

	tmp := os.TempDir()

	bindOut := filepath.Join(tmp, "bindings.go")
	mergeOut := filepath.Join(tmp, "merged.json")

	compDir := test.FilePath("complete")
	findExternal := false
	b := new(binder.Binder)
	b.ToolName = "bind-test"
	b.Loader = binder.NewJsonDefinitionLoader()
	b.Log = new(logging.ConsoleErrorLogger)

	s := binder.Settings{
		CompDefLocation:    &compDir,
		BindingsFile:       &bindOut,
		MergedDebugFile:    &mergeOut,
		ExternalFacilities: &findExternal,
	}

	b.Bind(s)

	if b.Failed() {
		t.Fail()
	}

	if _, err := os.Stat(mergeOut); os.IsNotExist(err) {
		t.Fatalf("Expected bindings file %s does not exist: %v", mergeOut, err)
	}

}
