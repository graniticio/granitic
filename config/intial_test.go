// Copyright 2016-2023 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package config

import (
	"github.com/graniticio/granitic/v2/test"
	"testing"
)

func TestExpandToFilesAndURLs(t *testing.T) {

	p := test.FilePath("folders")
	u := "http://www.example.com/json"

	r, err := ExpandToFilesAndURLs([]string{u, p})

	test.ExpectNil(t, err)

	test.ExpectInt(t, len(r), 6)

}
