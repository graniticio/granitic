package reflecttools

import (
	"reflect"
)

func HasFieldOfName(i interface{}, fieldName string) bool {
	r := reflect.ValueOf(i).Elem()
	f := r.FieldByName(fieldName)

	return f.IsValid()
}

func TypeOfField(i interface{}, name string) reflect.Type {
	r := reflect.ValueOf(i).Elem()
	return r.FieldByName(name).Type()
}

func SetInt64(i interface{}, name string, v int64) {
	t := FieldValue(i, name)
	t.SetInt(v)
}

func SetFloat64(i interface{}, name string, v float64) {
	t := FieldValue(i, name)
	t.SetFloat(v)
}

func SetUint64(i interface{}, name string, v uint64) {
	t := FieldValue(i, name)
	t.SetUint(v)
}

func SetBool(i interface{}, name string, b bool) {
	t := FieldValue(i, name)
	t.SetBool(b)
}

func SetString(i interface{}, name string, s string) {
	t := FieldValue(i, name)
	t.SetString(s)
}

func FieldValue(i interface{}, name string) reflect.Value {
	r := reflect.ValueOf(i).Elem()
	return r.FieldByName(name)
}

func TargetFieldIsArray(i interface{}, name string) bool {
	return TypeOfField(i, name).Kind() == reflect.Array
}
