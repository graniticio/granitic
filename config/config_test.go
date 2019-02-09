// Copyright 2016-2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package config

import (
	"encoding/json"
	"github.com/graniticio/granitic/v2/logging"
	"github.com/graniticio/granitic/v2/test"
	"io/ioutil"
	"path/filepath"
	"testing"
)

type SimpleConfig struct {
	String         string
	Bool           bool
	Int            int
	Float          float64
	StringArray    []string
	FloatArray     []float64
	IntArray       []int
	StringMap      map[string]string
	Unsupported    *SimpleConfig
	StringArrayMap map[string][]string
}

func TestTypeDetection(t *testing.T) {

	if JSONType("") != JSONString {
		t.FailNow()
	}

	if JSONType(true) != JSONBool {
		t.FailNow()
	}

	if JSONType(make(map[string]interface{})) != JSONMap {
		t.FailNow()
	}

	if JSONType([]interface{}{}) != JSONArray {
		t.FailNow()
	}

	if JSONType(1) != JSONUnknown {
		t.FailNow()
	}
}

func LoadConfigFromFile(f string) *Accessor {

	osp := filepath.Join("config", f)

	p := test.FilePath(osp)
	l := logging.CreateAnonymousLogger("config_test", 0)

	var d interface{}
	b, _ := ioutil.ReadFile(p)
	json.Unmarshal(b, &d)

	ca := new(Accessor)
	ca.FrameworkLogger = l
	ca.JSONData = d.(map[string]interface{})

	return ca

}

func TestFindConfigFiles(t *testing.T) {

	p := test.FilePath("folders")

	r, err := FindConfigFilesInDir(p)

	test.ExpectNil(t, err)

	test.ExpectInt(t, len(r), 4)

}

func TestSimpleConfig(t *testing.T) {

	ca := LoadConfigFromFile("simple.json")

	s, err := ca.StringVal("simpleOne.String")
	test.ExpectString(t, "abc", s)
	test.ExpectNil(t, err)

	b, err := ca.BoolVal("simpleOne.Bool")
	test.ExpectBool(t, true, b)
	test.ExpectNil(t, err)

	i, err := ca.IntVal("simpleOne.Int")
	test.ExpectNil(t, err)
	test.ExpectInt(t, 32, i)

	f, err := ca.Float64Val("simpleOne.Float")
	test.ExpectNil(t, err)
	test.ExpectFloat(t, 32.22, f)

	sa, err := ca.Array("simpleOne.StringArray")
	test.ExpectNil(t, err)
	test.ExpectString(t, sa[1].(string), "b")

	sa, err = ca.Array("simpleOne.StringArrayX")
	test.ExpectNil(t, err)

	sa, err = ca.Array("simpleOne.Bool")
	test.ExpectNotNil(t, err)

}

func TestUnset(t *testing.T) {

	ca := LoadConfigFromFile("simple.json")

	ca.StringVal("unset.String")

	ca.BoolVal("unset.Bool")

	ca.IntVal("unset.Int")

	ca.Float64Val("unset.Float")

}

func TestPathExistence(t *testing.T) {

	ca := LoadConfigFromFile("simple.json")

	test.ExpectBool(t, true, ca.PathExists("simpleOne.Bool"))

	test.ExpectBool(t, false, ca.PathExists("simpleX.Bool"))
	test.ExpectBool(t, false, ca.PathExists(""))
	test.ExpectBool(t, false, ca.PathExists("....."))

}

func TestWrongType(t *testing.T) {
	ca := LoadConfigFromFile("simple.json")

	i, err := ca.IntVal("simpleOne.String")
	test.ExpectInt(t, 0, i)
	test.ExpectNotNil(t, err)

	b, err := ca.BoolVal("simpleOne.String")
	test.ExpectBool(t, false, b)
	test.ExpectNotNil(t, err)

	f, err := ca.Float64Val("simpleOne.String")
	test.ExpectFloat(t, 0, f)
	test.ExpectNotNil(t, err)

	s, err := ca.StringVal("simpleOne.Bool")
	test.ExpectString(t, "", s)
	test.ExpectNotNil(t, err)
}

func TestPopulateObject(t *testing.T) {

	ca := LoadConfigFromFile("simple.json")

	var sc SimpleConfig

	err := ca.Populate("simpleOne", &sc)

	test.ExpectNil(t, err)

	test.ExpectString(t, "abc", sc.String)

	test.ExpectBool(t, true, sc.Bool)

	test.ExpectInt(t, 32, sc.Int)

	test.ExpectFloat(t, 32.22, sc.Float)

	m := sc.StringMap

	test.ExpectNotNil(t, m)

	test.ExpectInt(t, 3, len(sc.FloatArray))

}

func TestSetField(t *testing.T) {

	ca := LoadConfigFromFile("simple.json")
	ca.FrameworkLogger = new(logging.ConsoleErrorLogger)

	var sc SimpleConfig

	if err := ca.SetField("String", "simpleOne.String", &sc); err != nil {
		t.FailNow()
	}

	if err := ca.SetField("Bool", "simpleOne.Bool", &sc); err != nil {
		t.FailNow()
	}

	if err := ca.SetField("Int", "simpleOne.Int", &sc); err != nil {
		t.FailNow()
	}

	if err := ca.SetField("Float", "simpleOne.Float", &sc); err != nil {
		t.FailNow()
	}

	if err := ca.SetField("IntArray", "simpleOne.IntArray", &sc); err != nil {
		t.FailNow()
	}

	if err := ca.SetField("StringMap", "simpleOne.StringMap", &sc); err != nil {
		t.FailNow()
	}

	if err := ca.SetField("Unsupported", "simpleOne.IntArray", &sc); err == nil {
		t.FailNow()
	}

	if err := ca.SetField("StringMap", "missing.path", &sc); err == nil {
		t.FailNow()
	}

	if err := ca.SetField("StringMap", "simpleOne.Bool", &sc); err == nil {
		t.FailNow()
	}

	if err := ca.SetField("StringMap", "simpleOne.BoolA", &sc); err == nil {
		t.FailNow()
	}

	if _, err := ca.ObjectVal("simpleOne.Bool"); err == nil {
		t.FailNow()
	}

	if err := ca.SetField("StringArrayMap", "simpleOne.StringArrayMap", &sc); err != nil {
		t.FailNow()
	}

	if err := ca.SetField("StringArrayMap", "simpleOne.EmptyStringArrayMap", &sc); err == nil {
		t.FailNow()
	}

	if err := ca.SetField("StringArrayMap", "simpleOne.BoolArrayMap", &sc); err == nil {
		t.FailNow()
	}
}

func TestPopulateObjectMissingPath(t *testing.T) {
	ca := LoadConfigFromFile("simple.json")

	var sc SimpleConfig

	err := ca.Populate("undefined", &sc)

	test.ExpectNotNil(t, err)

}

func TestPopulateInvalid(t *testing.T) {

	ca := LoadConfigFromFile("simple.json")

	var sc SimpleConfig

	ca.Populate("invalidConfig", &sc)

}
