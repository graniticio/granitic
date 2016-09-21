package rdbms

import (
	"errors"
	"github.com/graniticio/granitic/reflecttools"
	"reflect"
)

const (
	DBParamTag = "dbparam"
)

func ParamsFromTags(i interface{}) (map[string]interface{}, error) {

	to := reflect.TypeOf(i)
	vo := reflect.ValueOf(i)

	k := to.Kind()

	if k != reflect.Struct && !reflecttools.IsPointerToStruct(i) {
		return nil, errors.New("Argument to ParamsFromTags must be a struct or pointer to a struct")
	}

	if to.Kind() == reflect.Ptr {
		to = to.Elem()
		vo = vo.Elem()
	}

	p := make(map[string]interface{})
	for i := 0; i < to.NumField(); i++ {

		f := to.Field(i)
		v := f.Tag.Get(DBParamTag)

		if v != "" {

			p[v] = vo.FieldByName(f.Name).Interface()

		}

	}

	return p, nil

}
