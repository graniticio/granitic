// Copyright 2016-2020 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package rdbms

import (
	"fmt"
	"github.com/graniticio/granitic/v2/test"
	"github.com/graniticio/granitic/v2/types"
	"testing"
)

func TestTagReading(t *testing.T) {

	tt := new(TagTest)

	tt.NoTag = "none"
	tt.ExplicitTag = "exp"

	p, err := ParamsFromTags(tt)

	fmt.Printf("%v\n", p)

	test.ExpectNil(t, err)
	test.ExpectInt(t, len(p), 1)
	test.ExpectString(t, p["explicit"].(string), "exp")

}

func TestNonStructTags(t *testing.T) {
	_, err := ParamsFromTags(1)
	test.ExpectNotNil(t, err)
}

type TagTest struct {
	NoTag       string
	ExplicitTag string `dbparam:"explicit"`
}

func TestSingleMapToParams(t *testing.T) {

	m := make(map[string]interface{})

	m["a"] = 1
	m["b"] = new(types.NilableString)

	p, err := ParamsFromFieldsOrTags(m)

	test.ExpectNil(t, err)
	test.ExpectInt(t, len(p), 2)

}

func TestIllegalArgsToParams(t *testing.T) {

	m := make(map[string]interface{})

	m["a"] = 1
	m["b"] = new(types.NilableString)

	_, err := ParamsFromFieldsOrTags(m, 1)
	test.ExpectNotNil(t, err)

	_, err = ParamsFromFieldsOrTags(1, m)
	test.ExpectNotNil(t, err)

	_, err = ParamsFromFieldsOrTags(1)
	test.ExpectNotNil(t, err)

}

type mixArgTypesTest struct {
	A *types.NilableString
	B *types.NilableString
	C *types.NilableString
	D *types.NilableString `dbparam:"C"`
	E *types.NilableString
}

func TestMixTypes(t *testing.T) {

	m := make(map[string]interface{})

	m["A"] = types.NewNilableString("A1")
	m["B"] = types.NewNilableString("B1")

	m2 := new(mixArgTypesTest)

	m2.A = types.NewNilableString("A2")
	m2.C = types.NewNilableString("C2")

	m3 := make(map[string]interface{})

	m3["B"] = types.NewNilableString("B3")

	m4 := new(mixArgTypesTest)
	m4.D = types.NewNilableString("C4")

	p, err := ParamsFromFieldsOrTags(m, m2, m3, m4)
	test.ExpectNil(t, err)

	test.ExpectInt(t, len(p), 3)

	test.ExpectString(t, p["A"].(*types.NilableString).String(), "A2")
	test.ExpectString(t, p["B"].(*types.NilableString).String(), "B3")
	test.ExpectString(t, p["C"].(*types.NilableString).String(), "C4")
	test.ExpectNil(t, p["D"])
	test.ExpectNil(t, p["E"])

}
