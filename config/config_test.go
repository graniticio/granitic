// Copyright 2016 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package config

import (
	"encoding/json"
	"github.com/graniticio/granitic/logging"
	"github.com/graniticio/granitic/test"
	"io/ioutil"
	"testing"
	"path/filepath"
)

type SimpleConfig struct {
	String      string
	Bool        bool
	Int         int
	Float       float64
	StringArray []string
	FloatArray  []float64
	IntArray    []int
	StringMap   map[string]string
}

func LoadConfigFromFile(f string) *ConfigAccessor {

	osp := filepath.Join("config", f)

	p := test.TestFilePath(osp)
	l := logging.CreateAnonymousLogger("config_test", 0)

	var d interface{}
	b, _ := ioutil.ReadFile(p)
	json.Unmarshal(b, &d)

	ca := new(ConfigAccessor)
	ca.FrameworkLogger = l
	ca.JsonData = d.(map[string]interface{})

	return ca

}

func TestFindConfigFiles(t *testing.T) {

	p := test.TestFilePath("folders")

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
