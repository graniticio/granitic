package reflecttools

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

const dotPathSep = "."

func SetPtrToStruct(target interface{}, field string, valuePointer interface{}) error {

	if !IsPointerToStruct(target) {
		return errors.New("Target is not a pointer to a struct.")
	}

	if !IsPointerToStruct(valuePointer) {
		return errors.New("Value supplied to set on the target is not a pointer to a struct.")
	}

	tv := reflect.ValueOf(target).Elem()
	vp := reflect.ValueOf(valuePointer)

	if !HasFieldOfName(target, field) {
		m := fmt.Sprintf("Target does not have a field called %s", field)
		return errors.New(m)
	}

	tfv := tv.FieldByName(field)

	if !tfv.CanSet() {
		m := fmt.Sprintf("Field %s on target cannot be set. Has the field been exported?", field)
		return errors.New(m)
	}

	if tfv.Kind() == reflect.Interface {
		if vp.Type().Implements(tfv.Type()) {
			tfv.Set(vp)
		} else {
			m := fmt.Sprintf("Supplied value (type %s) does not implement the interface (%s) required by the target field %s", vp.Elem().Type().Name(), tfv.Type().Name(), field)
			return errors.New(m)
		}

	}

	if vp.Type().AssignableTo(tfv.Type()) {
		tfv.Set(vp)
	}

	return nil
}

func NilPointer(v reflect.Value) bool {

	return v.Kind() == reflect.Ptr && !v.Elem().IsValid()

}

func NilMap(v reflect.Value) bool {

	return v.Kind() == reflect.Map && v.IsNil()

}

func IsPointerToStruct(p interface{}) bool {

	pv := reflect.ValueOf(p)
	pvk := pv.Kind()

	if pvk != reflect.Ptr {
		return false
	}

	vv := pv.Elem()
	vvk := vv.Kind()

	if vvk != reflect.Struct {
		return false
	}

	return true
}

func HasFieldOfName(i interface{}, fieldName string) bool {
	r := reflect.ValueOf(i).Elem()
	f := r.FieldByName(fieldName)

	return f.IsValid()
}

func HasWritableFieldOfName(i interface{}, fieldName string) bool {
	r := reflect.ValueOf(i).Elem()
	f := r.FieldByName(fieldName)

	return f.IsValid() && f.CanSet()
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

	var r reflect.Value

	r = reflect.ValueOf(i)

	k := r.Kind()

	if k == reflect.Interface || k == reflect.Ptr {
		r = r.Elem()
	}

	return r.FieldByName(name)
}

func TargetFieldIsArray(i interface{}, name string) bool {
	return TypeOfField(i, name).Kind() == reflect.Array
}

func ExtractDotPath(path string) []string {
	return strings.SplitN(path, dotPathSep, -1)
}

func FindNestedField(path []string, v interface{}) (reflect.Value, error) {

	pl := len(path)
	head := path[0]

	if pl == 1 {
		return FieldValue(v, head), nil
	} else {
		fv := FieldValue(v, head)
		next := fv.Interface()

		if !IsPointerToStruct(next) && fv.Kind() != reflect.Struct {
			m := fmt.Sprintf("%s is not a struct or a pointer to a struct", head)
			var zero reflect.Value

			return zero, errors.New(m)
		}

		return FindNestedField(path[1:], next)

	}

}
