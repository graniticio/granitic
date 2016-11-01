// Copyright 2016 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package rdbms

import (
	"errors"
	"fmt"
	"github.com/graniticio/granitic/reflecttools"
	"reflect"
)

const (
	// The name of a Go tag on struct fields that can be used to map that field to a parameter name
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
