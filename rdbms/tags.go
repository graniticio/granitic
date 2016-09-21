package rdbms

import (
	"errors"
	"fmt"
	"github.com/graniticio/granitic/reflecttools"
	"reflect"
)

const (
	DBParamTag = "dbparam"
)

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
