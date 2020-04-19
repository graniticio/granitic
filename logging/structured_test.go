// Copyright 2016-2020 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package logging

import (
	"testing"
)

func TestUnsupportedContent(t *testing.T) {

	f := JSONField{
		Content: "XXXX",
		Name:    "Unsupported",
	}

	err := ValidateJSONFields([]JSONField{f})

	if err == nil {
		t.Fatalf("Failed to detect invalid content type")
	}
}

func TestMissingName(t *testing.T) {

	f := JSONField{
		Content: "MESSAGE",
		Name:    "",
	}

	err := ValidateJSONFields([]JSONField{f})

	if err == nil {
		t.Fatalf("Failed to detect invalid content type")
	}
}

func TestMissingContextValueKey(t *testing.T) {

	f := JSONField{
		Content: "CONTEXT_VALUE",
		Name:    "MissingArg",
	}

	err := ValidateJSONFields([]JSONField{f})

	if err == nil {
		t.Fatalf("Failed to detect missing context value key")
	}
}
