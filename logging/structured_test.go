// Copyright 2016-2020 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package logging

import (
	"context"
	"testing"
)

func TestUnsupportedContent(t *testing.T) {

	f := JSONField{
		Content: "XXXX",
		Name:    "Unsupported",
	}

	err := ValidateJSONFields([]*JSONField{&f})

	if err == nil {
		t.Fatalf("Failed to detect invalid content type")
	}
}

func TestMissingName(t *testing.T) {

	f := JSONField{
		Content: "MESSAGE",
		Name:    "",
	}

	err := ValidateJSONFields([]*JSONField{&f})

	if err == nil {
		t.Fatalf("Failed to detect invalid content type")
	}
}

func TestMissingContextValueKey(t *testing.T) {

	f := JSONField{
		Content: "CONTEXT_VALUE",
		Name:    "MissingArg",
	}

	err := ValidateJSONFields([]*JSONField{&f})

	if err == nil {
		t.Fatalf("Failed to detect missing context value key")
	}
}

func TestMissingTimestampLayout(t *testing.T) {

	f := JSONField{
		Content: "TIMESTAMP",
		Name:    "MissingArg",
	}

	err := ValidateJSONFields([]*JSONField{&f})

	if err == nil {
		t.Fatalf("Failed to detect missing timestamp layout")
	}
}

func TestTimestampLayout(t *testing.T) {

	f := JSONField{
		Content: "TIMESTAMP",
		Name:    "Stamp",
		Arg:     "Mon Jan 2 15:04:05 MST 2006",
	}

	err := ValidateJSONFields([]*JSONField{&f})

	if err != nil {
		t.Fatalf("Did not accept valid layout")
	}
}

/*
func TestInvalidTimestampLayout(t *testing.T) {

	f := JSONField{
		Content: "TIMESTAMP",
		Name:    "Stamp",
		Arg:     "Mon Jan 32 15:04:05 MST 2006",
	}

	err := ValidateJSONFields([]*JSONField{&f})

	if err == nil {
		t.Fatalf("Did not reject invalid layout")
	}
}*/

func TestMapBuilder(t *testing.T) {

	cfg := new(JSONConfig)

	f := new(JSONField)

	f.Content = "MESSAGE"
	f.Name = "message"

	cfg.Fields = []*JSONField{f}

	mb, err := CreateMapBuilder(cfg)

	if err != nil {
		t.Fatalf(err.Error())
	}

	m := mb.Build(context.Background(), "TRACE", "MyComp", "some message")

	if m == nil {
		t.FailNow()
	}

}

func TestMissingContextFilter(t *testing.T) {

	fields := []*JSONField{
		{Name: "CtxVal", Content: "CONTEXT_VALUE", Arg: "someKey"}}

	cfg := JSONConfig{
		Prefix: "",
		Fields: fields,
		Suffix: "\n",
		UTC:    true,
	}

	mb, _ := CreateMapBuilder(&cfg)

	jf := new(JSONLogFormatter)

	jf.Config = &cfg
	jf.MapBuilder = mb

	if jf.StartComponent() == nil {
		t.Fatalf("Failed to detect missing context filter")
	}

}

func TestContextVal(t *testing.T) {

	fields := []*JSONField{
		{Name: "CtxVal", Content: "CONTEXT_VALUE", Arg: "someKey"}}

	cfg := JSONConfig{
		Prefix: "",
		Fields: fields,
		Suffix: "\n",
		UTC:    true,
	}

	mb, _ := CreateMapBuilder(&cfg)

	cf := testFilter{m: make(FilteredContextData)}

	cf.m["someKey"] = "someVal"

	jf := new(JSONLogFormatter)
	jf.Config = &cfg
	jf.MapBuilder = mb

	jf.SetContextFilter(cf)

	if jf.StartComponent() != nil {
		t.Fatalf("Failed to detect supplied context filter")
	}

	s := jf.Format(context.Background(), "", "", "")

	if s != "{\"CtxVal\":\"someVal\"}\n" {
		t.Fatalf("Unexpected message")
	}
}

func BenchmarkDefaultJSONFormatter(b *testing.B) {

	fields := []*JSONField{
		{Name: "Timestamp", Content: "TIMESTAMP", Arg: "02/Jan/2006:15:04:05 Z0700"},
		{Name: "Level", Content: "LEVEL"},
		{Name: "Source", Content: "COMPONENT_NAME"},
		{Name: "Message", Content: "MESSAGE"},
	}

	cfg := JSONConfig{
		Prefix: "",
		Fields: fields,
		Suffix: "\n",
		UTC:    true,
	}

	mb, _ := CreateMapBuilder(&cfg)

	jf := new(JSONLogFormatter)

	jf.Config = &cfg
	jf.MapBuilder = mb

	for i := 0; i < b.N; i++ {
		jf.Format(nil, "INFO", "someComp", "A benchmark test message of fixed length")
	}
}
