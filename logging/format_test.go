// Copyright 2016 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package logging

import (
	"fmt"
	"github.com/graniticio/granitic/test"
	"golang.org/x/net/context"
	"testing"
)

func TestNoPlaceholdersFormat(t *testing.T) {

	lf := new(LogMessageFormatter)
	lf.PrefixFormat = "PLAINTEXT"

	err := lf.Init()
	test.ExpectNil(t, err)

	m := lf.Format(context.Background(), "DEBUG", "NAME", "MESSAGE")

	fmt.Println(m)

}

func TestPlaceHolders(t *testing.T) {

	lf := new(LogMessageFormatter)
	lf.PrefixFormat = "%P %L %l %c %% "

	err := lf.Init()
	test.ExpectNil(t, err)

	m := lf.Format(context.Background(), "INFO", "NAME", "MESSAGE")

	test.ExpectString(t, m, "INFO  INFO I NAME % MESSAGE\n")

	fmt.Println(m)

}
