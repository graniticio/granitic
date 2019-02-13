// Copyright 2016-2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package config

import (
	"github.com/graniticio/granitic/v2/test"
	"testing"
)

func TestFindJSONFilesInDir(t *testing.T) {

	p := test.FilePath("../")

	j, err := FindJSONFilesInDir(p)

	if err != nil {
		t.Fatalf("Problem finding files: %v", err)
	}

	if len(j) == 0 {
		t.Errorf("Expected to find files under resource folder, found none")
	}

}

func TestFileListFromPath(t *testing.T) {

	p := test.FilePath("../")

	j, err := FileListFromPath(p)

	if err != nil {
		t.Fatalf("Problem finding files: %v", err)
	}

	if len(j) == 0 {
		t.Errorf("Expected to find files under resource folder, found none")
	}

}
