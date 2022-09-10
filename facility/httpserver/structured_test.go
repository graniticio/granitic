// Copyright 2016-2022 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package httpserver

import "testing"

func TestInterfaceImplementation(t *testing.T) {

	var lb LineBuilder

	lb = new(JSONLineBuilder)

	if lb == nil {
		t.FailNow()
	}

	if _, ok := lb.(*JSONLineBuilder); !ok {
		t.FailNow()
	}
}
