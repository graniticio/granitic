// Copyright 2016 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

/*
Package reflecttools provides utility functions for working with Go's reflect package.

These functions are highly specific to Granitic's internal use of reflection and are not recommended for use in user
applications.
*/
package reflecttools

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

const dotPathSep = "."

// SetPtrToStruct is used to inject an object into the specified field on a another object. The target object, supplied value and the type
// of the named target field must all be a pointer to a struct.
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


func SetFieldPtrToStruct(tfv reflect.Value , valuePointer interface{}) error {

	vp := reflect.ValueOf(valuePointer)

	if !tfv.CanSet() {
		m := fmt.Sprintf("Field cannot be set")
		return errors.New(m)
	}

	if tfv.Kind() == reflect.Interface {
		if vp.Type().Implements(tfv.Type()) {
			tfv.Set(vp)
		} else {
			m := fmt.Sprintf("Supplied value (type %s) does not implement the interface (%s) required by the target field", vp.Elem().Type().Name(), tfv.Type().Name())
			return errors.New(m)
		}

	}

	if vp.Type().AssignableTo(tfv.Type()) {
		tfv.Set(vp)
	}

	return nil
}

// NilPointer returns true if the supplied reflect value is a pointer that does not point a valid value.
func NilPointer(v reflect.Value) bool {

	return v.Kind() == reflect.Ptr && !v.Elem().IsValid()

}

// NilPointer returns true is the supplied reflect value is a Map and is nil.
func NilMap(v reflect.Value) bool {

	return v.Kind() == reflect.Map && v.IsNil()

}

// IsPointerToStruct returns true if the supplied interfaces is a pointer to a struct.
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

// HasFieldOfName assumes the supplied interface is a pointer to a struct and checks to see if the underlying struct
// has a field of the supplied name. It does not check to see if the field is writable.
func HasFieldOfName(i interface{}, fieldName string) bool {
	r := reflect.ValueOf(i).Elem()
	f := r.FieldByName(fieldName)

	return f.IsValid()
}

// HasFieldOfName assumes the supplied interface is a pointer to a struct and checks to see if the underlying struct
// has a writable field of the supplied name.
func HasWritableFieldOfName(i interface{}, fieldName string) bool {
	r := reflect.ValueOf(i).Elem()
	f := r.FieldByName(fieldName)

	return f.IsValid() && f.CanSet()
}

// TypeOfField assumes the supplied interface is a pointer to a struct and that the supplied field name exists on that struct, then
// finds the reflect type of that field.
func TypeOfField(i interface{}, name string) reflect.Type {
	r := reflect.ValueOf(i).Elem()
	return r.FieldByName(name).Type()
}

// SetInt64 assumes that the supplied interface is a pointer to a struct and has a writable int64 field of the supplied name, then
// sets the field of the supplied value.
func SetInt64(i interface{}, name string, v int64) {
	t := FieldValue(i, name)
	t.SetInt(v)
}

// SetFloat64 assumes that the supplied interface is a pointer to a struct and has a writable float64 field of the supplied name, then
// sets the field of the supplied value.
func SetFloat64(i interface{}, name string, v float64) {
	t := FieldValue(i, name)
	t.SetFloat(v)
}

// SetUint64 assumes that the supplied interface is a pointer to a struct and has a writable uint64 field of the supplied name, then
// sets the field of the supplied value.
func SetUint64(i interface{}, name string, v uint64) {
	t := FieldValue(i, name)
	t.SetUint(v)
}

// SetBool assumes that the supplied interface is a pointer to a struct and has a writable bool field of the supplied name, then
// sets the field of the supplied value.
func SetBool(i interface{}, name string, b bool) {
	t := FieldValue(i, name)
	t.SetBool(b)
}

// SetString assumes that the supplied interface is a pointer to a struct and has a writable string field of the supplied name, then
// sets the field of the supplied value.
func SetString(i interface{}, name string, s string) {
	t := FieldValue(i, name)
	t.SetString(s)
}

// FieldValue assumes the supplied interface is a pointer to a struct, an interface or a struct and has a valid field of the supplied
// name, then returns the reflect value of that field.
func FieldValue(i interface{}, name string) reflect.Value {

	var r reflect.Value

	r = reflect.ValueOf(i)

	k := r.Kind()

	if k == reflect.Interface || k == reflect.Ptr {
		r = r.Elem()
	}

	return r.FieldByName(name)
}

// TargetFieldIsArray assumes the supplied interface is a pointer to a struct, an interface or a struct
// and has a valid field of the supplied name, then returns true if the reflect type of that field is Array. Note that
// this method will return false for Slice fields.
func TargetFieldIsArray(i interface{}, name string) bool {
	return TypeOfField(i, name).Kind() == reflect.Array
}

// ExtractDotPath converts a dot-delimited path into a string array of its constiuent parts. E.g. "a.b.c" becomes
// ["a","b","c"]
func ExtractDotPath(path string) []string {
	return strings.SplitN(path, dotPathSep, -1)
}

// FindNestedField take the output of ExtractDotPath and uses it to traverse an object graph to find a value. Apart from
// final value, each intermediate step in the graph must be a struct or pointer to a struct.
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

// IsZero returns true if i is set to the zero value of i's type
func IsZero(i interface{}) bool {
	return i == reflect.Zero(reflect.TypeOf(i)).Interface()
}
