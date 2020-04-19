// Copyright 2016-2020 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package logging

import (
	"context"
	"fmt"
	"github.com/graniticio/granitic/v2/test"
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
	lf.Unset = "-"
	lf.PrefixFormat = "%P %L %l %c %% %{CTX}X "

	err := lf.Init()
	test.ExpectNil(t, err)

	d := make(FilteredContextData)
	d["CTX"] = "someValue"

	lf.SetContextFilter(testFilter{d})

	m := lf.Format(context.Background(), "INFO", "NAME", "MESSAGE")

	test.ExpectString(t, m, "INFO  INFO I NAME % someValue MESSAGE\n")

}

func TestComponentNameTrunctation(t *testing.T) {

	lf := new(LogMessageFormatter)
	lf.Unset = "-"
	lf.PrefixFormat = "%{3}C "

	err := lf.Init()
	test.ExpectNil(t, err)

	m := lf.Format(context.Background(), "INFO", "COMP", "MESSAGE")

	test.ExpectString(t, m, "COM MESSAGE\n")

	lf = new(LogMessageFormatter)
	lf.Unset = "-"
	lf.PrefixFormat = "[%{5}C] "

	err = lf.Init()
	test.ExpectNil(t, err)

	m = lf.Format(context.Background(), "INFO", "COMP", "MESSAGE")

	test.ExpectString(t, m, "[COMP ] MESSAGE\n")

}

func TestUnsupportedPlaceholder(t *testing.T) {

	lf := new(LogMessageFormatter)
	lf.Unset = "-"
	lf.PrefixFormat = "%{3}K "

	err := lf.Init()
	test.ExpectNotNil(t, err)

}

// Clashes with golint
/*func TestDirectCtxValue(t *testing.T) {

	lf := new(LogMessageFormatter)
	lf.Unset = "-"
	lf.PrefixFormat = "%{INCTX}X "

	err := lf.Init()
	test.ExpectNil(t, err)

	ctx := context.WithValue(context.Background(), "INCTX", "FROMCTX")

	m := lf.Format(ctx, "INFO", "COMP", "MESSAGE")

	test.ExpectString(t, m, "FROMCTX MESSAGE\n")
}*/

type testFilter struct {
	m FilteredContextData
}

func (tf testFilter) Extract(ctx context.Context) FilteredContextData {
	return tf.m
}

func TestInvalidInit(t *testing.T) {

	lf := new(LogMessageFormatter)

	if err := lf.Init(); err == nil {
		t.Fatalf("Failed to detect missing prefix AND pattern")

	}

	lf = new(LogMessageFormatter)

	lf.PrefixPreset = "!!!!!!!"
	lf.PrefixFormat = "AAAA"

	if err := lf.Init(); err == nil {
		t.Fatalf("Failed to detect supplied prefix AND pattern")

	}

	lf = new(LogMessageFormatter)

	lf.PrefixPreset = "!!!!!!!"
	lf.PrefixFormat = ""

	if err := lf.Init(); err == nil {
		t.Fatalf("Failed to detect invalid prefix")

	}
}

func TestDefaultInitialised(t *testing.T) {
	lf := NewFrameworkLogMessageFormatter()

	m := lf.Format(context.Background(), "DEBUG", "NAME", "MESSAGE")

	fmt.Println(m)

}
