// Copyright 2016-2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

/*
Package config provides functionality for working with configuration files and command line arguments to a Granitic application.

Grantic uses JSON files to store component definitions (declarations of, and relationships between, components to
run in the IoC container) and configuration (variables used by IoC components that may vary between environments and settings
for Grantic's built-in facilities). A defintion of the use and syntax of these files are outside of the scope of a GoDoc page,
but are described in detail at http://granitic.io/1.0/ref/components and http://granitic.io/1.0/ref/config

This package defines functionality for loading a JSON file (from a filesystem or via HTTP) and merging multiple files into
a single view. This is a key concept in Granitic.

Given a folder of configuration files called conf:
	conf/x.json
	conf/sub/a.json
	conf/sub/b.json

starting a Grantic application with:

	-c http://example.com/base.json,conf,http://example.com/myinstance.json

The following will take place. Firstly the files would be expanded into a flat list of paths/URIs

	http://example.com/base.json
	conf/sub/a.json
	conf/sub/b.json
	conf/x.json
	http://example.com/myinstance.json

The the files will be merged together from left, using the the first file as a base. In this example,  http://example.com/base.json
and conf/sub/a.json will be merged together, then result of that merge will be merged with conf/sub/b.json and so on.

For named fields (in a JSON object/map), the process of merging is fairly obvious. When merging files A and B, a field that
is defined in both files will have the value of the field used in file B in the merged output. For example,

	a.json

	{
		"database": {
			"host": "localhost",
			"port": 3306,
			"flags": ["a", "b", "c"]
		}
	}

and

	b.json

	{
		"database": {
			"host": "remotehost",
			"flags": ["d"]
		}
	}

woud merge to:

	{
		"database": {
			"host": "remotehost",
			"port": 3306,
			"flags": ["d"]
		}
	}

The merging of configuration files occurs exactly above, but when component definition files are merged, arrays are joined, not overwritten.
For example:

	{ "methods": ["GET"] }

merged with;

	{ "methods": ["POST"] }

would result in:

	{ "methods": ["GET", "POST"] }

Another core concept used by the types in this package is a config path. This is the absolute path to field in the
eventual merged configuration file with a dot-delimited notation. E.g "database.host".
*/
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/graniticio/granitic/v2/logging"
	"reflect"
	"strings"
)

// The character used to delimit paths to config values.
const JSONPathSeparator string = "."

// Used by code that needs to know what type of JSON data structure resides at a particular path
// before operating on it.
const (
	Unset       = -2
	JSONUnknown = -1
	JSONString  = 1
	JSONArray   = 2
	JSONMap     = 3
	JSONBool    = 4
)

// A ConfigAccessor provides access to a merged view of configuration files during the initialisation and
// configuration of the Granitic IoC container.
type ConfigAccessor struct {
	// The merged JSON configuration in object form.
	JSONData map[string]interface{}

	// Logger used by Granitic framework components. Automatically injected.
	FrameworkLogger logging.Logger
}

// Flush removes internal references to the (potentially very large) merged JSON data so the associated
// memory can be recovered during the next garbage collection.
func (c *ConfigAccessor) Flush() {
	c.JSONData = nil
}

// PathExists check to see whether the supplied dot-delimited path exists in the configuration and points to a non-null JSON value.
func (c *ConfigAccessor) PathExists(path string) bool {
	value := c.Value(path)

	return value != nil
}

// Value returns the JSON value at the supplied path or nil if the path does not exist of points to a null JSON value.
func (c *ConfigAccessor) Value(path string) interface{} {

	splitPath := strings.Split(path, JSONPathSeparator)

	return c.configVal(splitPath, c.JSONData)

}

// ObjectVal returns a map representing a JSON object or nil if the path does not exist of points to a null JSON value. An error
// is returned if the value cannot be interpreted as a JSON object.
func (c *ConfigAccessor) ObjectVal(path string) (map[string]interface{}, error) {

	value := c.Value(path)

	if value == nil {
		return nil, nil
	} else {
		if v, found := value.(map[string]interface{}); found {
			return v, nil
		} else {
			m := fmt.Sprintf("Unable to convert the value at %s to a JSON map/object.", path)
			return nil, errors.New(m)
		}
	}
}

// ObjectVal returns the string value of the JSON string at the supplied path. Does not convert other types to
// a string, so will return an error if the value is not a JSON string.
func (c *ConfigAccessor) StringVal(path string) (string, error) {

	v := c.Value(path)

	if v == nil {
		return "", errors.New("No string value found at " + path)
	}

	s, found := v.(string)

	if found {
		return s, nil
	} else {
		message := fmt.Sprintf("Value at %s is %q and cannot be converted to a string", path, v)
		return "", errors.New(message)
	}

}

// IntVal returns the int value of the JSON number at the supplied path. JSON numbers
// are internally represented by Go as a float64, so no error will be returned, but data might be lost
// if the JSON number does not actually represent an int. An error will be returned if the value is not a JSON number
// or cannot be converted to an int.
func (c *ConfigAccessor) IntVal(path string) (int, error) {

	v := c.Value(path)

	if v == nil {
		return 0, errors.New("No such path " + path)
	}

	if f, found := v.(float64); found {
		return int(f), nil
	} else {
		message := fmt.Sprintf("Value at %s is %q and cannot be converted to an int", path, v)
		return 0, errors.New(message)

	}
}

// IntVal returns the float64 value of the JSON number at the supplied path. An error will be returned if the value is not a JSON number.
func (c *ConfigAccessor) Float64Val(path string) (float64, error) {

	v := c.Value(path)

	if v == nil {
		return 0, errors.New("No such path " + path)
	}

	if f, found := v.(float64); found {
		return f, nil
	} else {
		message := fmt.Sprintf("Value at %s is %q and cannot be converted to a float64", path, v)
		return 0, errors.New(message)

	}
}

// Array returns the value of an array of JSON obects at the supplied path. Caution should be used when calling this method
// as behaviour is undefined for JSON arrays of JSON types other than object.
func (c *ConfigAccessor) Array(path string) ([]interface{}, error) {

	value := c.Value(path)

	if value == nil {
		return nil, nil
	} else {

		if v, found := value.([]interface{}); found {
			return v, nil
		} else {
			m := fmt.Sprintf("Unable to convert the value at %s to a JSON array.", path)
			return nil, errors.New(m)
		}
	}
}

// BoolVal returns the bool value of the JSON bool at the supplied path. An error will be returned if the value is not a JSON bool.
// Note this method only suports the JSON definition of bools (true, false) not the Go definition (true, false, 1, 0 etc).
func (c *ConfigAccessor) BoolVal(path string) (bool, error) {

	v := c.Value(path)

	if v == nil {
		return false, errors.New("No such path " + path)
	}

	if b, found := v.(bool); found {
		return b, nil
	} else {
		message := fmt.Sprintf("Value at %s is %q and cannot be converted to a bool", path, v)
		return false, errors.New(message)

	}

}

// Determines the apparent JSON type of the supplied Go interface.
func JSONType(value interface{}) int {

	switch value.(type) {
	case string:
		return JSONString
	case map[string]interface{}:
		return JSONMap
	case bool:
		return JSONBool
	case []interface{}:
		return JSONArray
	default:
		return JSONUnknown
	}
}

func (c *ConfigAccessor) configVal(path []string, jsonMap map[string]interface{}) interface{} {

	var result interface{}
	result = jsonMap[path[0]]

	if result == nil {
		return nil
	}

	if len(path) == 1 {
		return result
	} else {
		remainPath := path[1:len(path)]
		return c.configVal(remainPath, result.(map[string]interface{}))
	}
}

// SetField takes a target Go interface and uses the data a the supplied path to populated the named field on the
// target. The target must be a pointer to a struct. The field must be a string, bool, int, float63, string[interface{}] map
// or a slice of one of those types. An eror will be returned if the target field, is missing, not settable or incompatiable
// with the JSON value at the supplied path.
func (ca *ConfigAccessor) SetField(fieldName string, path string, target interface{}) error {

	if !ca.PathExists(path) {
		return errors.New("No value found at " + path)
	}

	targetReflect := reflect.ValueOf(target).Elem()
	targetField := targetReflect.FieldByName(fieldName)

	k := targetField.Type().Kind()

	switch k {
	case reflect.String:
		s, _ := ca.StringVal(path)
		targetField.SetString(s)
	case reflect.Bool:
		b, _ := ca.BoolVal(path)
		targetField.SetBool(b)
	case reflect.Int:
		i, _ := ca.IntVal(path)
		targetField.SetInt(int64(i))
	case reflect.Float64:
		f, _ := ca.Float64Val(path)
		targetField.SetFloat(f)
	case reflect.Map:

		if v, err := ca.ObjectVal(path); err == nil {
			if err = ca.populateMapField(targetField, v); err != nil {
				return err
			}
		} else {
			return err
		}
	case reflect.Slice:
		ca.populateSlice(targetField, path, target)

	default:
		m := fmt.Sprintf("Unable to use value at path %s as target field %s is not a suppported type (%s)", path, fieldName, k)
		ca.FrameworkLogger.LogErrorf(m)

		return errors.New(m)
	}

	return nil
}

func (ca *ConfigAccessor) populateSlice(targetField reflect.Value, path string, target interface{}) {

	v := ca.Value(path)

	data, _ := json.Marshal(v)

	vt := targetField.Type()
	nt := reflect.New(vt)

	jTarget := nt.Interface()
	json.Unmarshal(data, &jTarget)

	vr := reflect.ValueOf(jTarget)
	targetField.Set(vr.Elem())

}

func (ca *ConfigAccessor) populateMapField(targetField reflect.Value, contents map[string]interface{}) error {
	var err error

	m := reflect.MakeMap(targetField.Type())
	targetField.Set(m)

	for k, v := range contents {

		kVal := reflect.ValueOf(k)
		vVal := reflect.ValueOf(v)

		if vVal.Kind() == reflect.Slice {
			vVal, err = ca.arrayVal(vVal)

			if err != nil {
				return err
			}
		}

		m.SetMapIndex(kVal, vVal)

	}

	return nil
}

func (ca *ConfigAccessor) arrayVal(a reflect.Value) (reflect.Value, error) {

	v := a.Interface().([]interface{})
	l := len(v)

	if l == 0 {

		return reflect.Zero(reflect.TypeOf(ca)), errors.New("Cannot use an empty array as a value in a Map.")

	}

	var s reflect.Value

	switch t := v[0].(type) {
	case string:
		s = reflect.MakeSlice(reflect.TypeOf([]string{}), 0, 0)
	default:
		m := fmt.Sprintf("Cannot use an array of %T as a value in a Map.", t)
		return reflect.Zero(reflect.TypeOf(ca)), errors.New(m)
	}

	for _, elem := range v {

		s = reflect.Append(s, reflect.ValueOf(elem))

	}

	return s, nil
}

// Populate sets the fields on the supplied target object using the JSON data
// at the supplied path. This is acheived using Go's json.Marshal to convert the data
// back into text JSON and then json.Unmarshal to unmarshal back into the target.
func (ca *ConfigAccessor) Populate(path string, target interface{}) error {
	exists := ca.PathExists(path)

	if !exists {
		return errors.New("No such path: " + path)
	}

	//Already check if path exists
	object, _ := ca.ObjectVal(path)

	if data, err := json.Marshal(object); err != nil {
		m := fmt.Sprintf("%T cannot be marshalled to JSON", object)
		return errors.New(m)
	} else {

		if json.Unmarshal(data, target); err != nil {
			m := fmt.Sprintf("%T cannot be populated with %v to JSON", object, data)
			return errors.New(m)
		}
	}

	return nil

}
