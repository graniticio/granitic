// Copyright 2016-2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package rdbms

import (
	"errors"
	"fmt"
	"github.com/graniticio/granitic/v2/reflecttools"
	"reflect"
)

const (
	// DBParamTag is the name of a Go tag on struct fields that can be used to map that field to a parameter name
	DBParamTag = "dbparam"
)

/*
ParamsFromTags takes one or more structs whose fields might have the dbparam tag set. Those fields that do have the
tag set are added to the the returned map, where the tag value is used as the map key and the field value is used as the map value.
*/
func ParamsFromTags(sources ...interface{}) (map[string]interface{}, error) {

	p := make(map[string]interface{})

	for _, i := range sources {

		to := reflect.TypeOf(i)
		vo := reflect.ValueOf(i)

		k := to.Kind()

		if k != reflect.Struct && !reflecttools.IsPointerToStruct(i) {

			m := fmt.Sprintf("Argument to ParamsFromTags must be a struct or pointer to a struct is %v", k)

			return nil, errors.New(m)
		}

		if to.Kind() == reflect.Ptr {
			to = to.Elem()
			vo = vo.Elem()
		}

		for i := 0; i < to.NumField(); i++ {

			f := to.Field(i)
			v := f.Tag.Get(DBParamTag)

			if v != "" {

				p[v] = vo.FieldByName(f.Name).Interface()

			}

		}
	}

	return p, nil

}

/*
ParamsFromFieldsOrTags takes one or more objects (that must be a map[string]interface{} or a pointer to a struct) and
returns a single map[string]interface{}. Keys and values are copied from supplied map[string]interface{}s as-is. For
pointers to structs, the object will have its fields added to the map  using field names as keys (unless the dbparam tag is set)
and the field value as the map value.

An error is returned if one of the arguments is not a map[string]interface{} pointer to a struct.
*/
func ParamsFromFieldsOrTags(sources ...interface{}) (map[string]interface{}, error) {

	p := make(map[string]interface{})

ArgLoop:
	for _, arg := range sources {

		if asMap, found := arg.(map[string]interface{}); found {
			//Argument is a map
			mergeMapInto(asMap, p)
			continue ArgLoop

		}

		argType := reflect.TypeOf(arg)
		argVal := reflect.ValueOf(arg)

		argKind := argType.Kind()

		if argKind != reflect.Struct && !reflecttools.IsPointerToStruct(arg) {

			m := fmt.Sprintf("Argument argType ParamsFromFieldsOrTags must be a struct or pointer argType a struct is %v", argKind)

			return nil, errors.New(m)
		}

		if argType.Kind() == reflect.Ptr {
			//Arg is a pointer to struct - deference it before proceeding
			argType = argType.Elem()
			argVal = argVal.Elem()
		}

	FieldLoop:
		for fieldIndex := 0; fieldIndex < argType.NumField(); fieldIndex++ {
			//Loop over the fields on the struct and store any fields with a non-zero value in the param map
			field := argType.Field(fieldIndex)
			tagVal := field.Tag.Get(DBParamTag)
			fieldValInterface := argVal.FieldByName(field.Name).Interface()

			if reflecttools.IsSliceOrArray(fieldValInterface) || reflecttools.IsZero(fieldValInterface) {
				continue FieldLoop
			}

			if tagVal != "" {
				//Use tag value as key
				p[tagVal] = fieldValInterface

			} else {
				//Use field name as key
				fn := argVal.Type().Field(fieldIndex).Name
				p[fn] = fieldValInterface
			}

		}

	}

	return p, nil
}

func mergeMapInto(source, target map[string]interface{}) {

	for k, v := range source {
		target[k] = v
	}

}
