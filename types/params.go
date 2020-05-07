package types

import (
	"errors"
	"fmt"
	rt "github.com/graniticio/granitic/v2/reflecttools"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

// NewSingleValueParams creates a Params with only one value in it
func NewSingleValueParams(name string, value string) *Params {

	contents := make(url.Values)
	contents[name] = []string{value}
	names := []string{name}

	return &Params{paramNames: names, values: contents}

}

// NewParams creates a Params with the supplied contents
func NewParams(v url.Values, pn []string) *Params {
	p := new(Params)
	p.values = v
	p.paramNames = pn

	return p
}

// Params is an abstraction of the HTTP query parameters or path parameters with type-safe accessors.
type Params struct {
	values     url.Values
	paramNames []string
}

// ParamNames returns the names of all of the parameters stored
func (wp *Params) ParamNames() []string {
	return wp.paramNames
}

// NotEmpty returns true if a parameter with the supplied name exists and has a non-empty string representation.
func (wp *Params) NotEmpty(key string) bool {

	if !wp.Exists(key) {
		return false
	}

	s, _ := wp.StringValue(key)

	return s != ""

}

// Exists returns true if a parameter with the supplied name exists, even if that parameter's value is an empty string.
func (wp *Params) Exists(key string) bool {
	return wp.values[key] != nil
}

// MultipleValues returns true if the parameter with the supplied name was set more than once (allowed for HTTP query parameters).
func (wp *Params) MultipleValues(key string) bool {

	value := wp.values[key]

	return value != nil && len(value) > 1

}

// StringValue returns the string representation of the specified parameter or an error if no value exists for that parameter.
func (wp *Params) StringValue(key string) (string, error) {

	s := wp.values[key]

	if s == nil {
		return "", wp.noVal(key)
	}

	return s[len(s)-1], nil

}

// BoolValue returns the bool representation of the specified parameter (using Go's bool conversion rules) or an error if no value exists for that parameter.
func (wp *Params) BoolValue(key string) (bool, error) {

	v := wp.values[key]

	if v == nil {
		return false, wp.noVal(key)
	}

	b, err := strconv.ParseBool(v[len(v)-1])

	return b, err

}

// FloatNValue returns a float representation of the specified parameter with the specified bit size, or an error if no value exists for that parameter or
// if the value could not be converted to a float.
func (wp *Params) FloatNValue(key string, bits int) (float64, error) {

	v := wp.values[key]

	if v == nil {
		return 0.0, wp.noVal(key)
	}

	i, err := strconv.ParseFloat(v[len(v)-1], bits)

	return i, err

}

// IntNValue returns a signed int representation of the specified parameter with the specified bit size, or an error if no value exists for that parameter or
// if the value could not be converted to an int.
func (wp *Params) IntNValue(key string, bits int) (int64, error) {

	v := wp.values[key]

	if v == nil {
		return 0, wp.noVal(key)
	}

	i, err := strconv.ParseInt(v[len(v)-1], 10, bits)

	return i, err

}

// IntNInterfaceValue returns a signed int representation of the specified parameter as the correct basic type (int8, int32 etc), or an error if no value exists for that parameter or
// if the value could not be converted to an int.
func (wp *Params) IntNInterfaceValue(i64 int64, bits int) interface{} {

	switch bits {
	case 8:
		return int8(i64)
	case 16:
		return int16(i64)
	case 32:
		return int32(i64)
	}

	return i64

}

// UIntNInterfaceValue returns an unsigned int representation of the specified parameter as the correct basic type (uint8, uint32 etc), or an error if no value exists for that parameter or
// if the value could not be converted to an int.
func (wp *Params) UIntNInterfaceValue(u64 uint64, bits int) interface{} {

	switch bits {
	case 8:
		return uint8(u64)
	case 16:
		return uint16(u64)
	case 32:
		return uint32(u64)
	}

	return u64

}

// UIntNValue returns an unsigned int representation of the specified parameter with the specified bit size, or an error if no value exists for that parameter or
// if the value could not be converted to an unsigned int.
func (wp *Params) UIntNValue(key string, bits int) (uint64, error) {

	v := wp.values[key]

	if v == nil {
		return 0, wp.noVal(key)
	}

	i, err := strconv.ParseUint(v[len(v)-1], 10, bits)

	return i, err

}

func (wp *Params) noVal(key string) error {
	message := fmt.Sprintf("No value available for key %s", key)
	return errors.New(message)
}

// GenerateMappingError is used to create an error in the context of mapping a parameter into a struct field
type GenerateMappingError func(paramName string, fieldName string, typeName string, params *Params) error

// FieldAssociatedError is implemented by types that can record which field on a struct caused a problem
type FieldAssociatedError interface {
	// RecordField captures the field that was involved in the error
	RecordField(string)
}

// ParamValueInjector takes a series of key/value (string/string) parameters and tries to inject them into the fields
// on a target struct
type ParamValueInjector struct {
}

// BindValueToField attempts to take a named parameter from the supplied set of parameters and inject it into a field on the supplied target,
// converting to the correct data type as it goes.
func (pb *ParamValueInjector) BindValueToField(paramName string, fieldName string, p *Params, t interface{}, errorFn GenerateMappingError, index ...int) error {

	ft := rt.TypeOfField(t, fieldName)
	k := ft.Kind()

	if k == reflect.Slice && index != nil {
		//We're setting a value in a slice
		k = ft.Elem().Kind()
	}

	switch k {
	case reflect.Int:
		return pb.setIntNField(paramName, fieldName, p, t, 0, errorFn, index...)
	case reflect.Int8:
		return pb.setIntNField(paramName, fieldName, p, t, 8, errorFn, index...)
	case reflect.Int16:
		return pb.setIntNField(paramName, fieldName, p, t, 16, errorFn, index...)
	case reflect.Int32:
		return pb.setIntNField(paramName, fieldName, p, t, 32, errorFn, index...)
	case reflect.Int64:
		return pb.setIntNField(paramName, fieldName, p, t, 64, errorFn, index...)
	case reflect.Bool:
		return pb.setBoolField(paramName, fieldName, p, t, errorFn, index...)
	case reflect.String:
		return pb.setStringField(paramName, fieldName, p, t, errorFn, index...)
	case reflect.Uint8:
		return pb.setUintNField(paramName, fieldName, p, t, 8, errorFn, index...)
	case reflect.Uint16:
		return pb.setUintNField(paramName, fieldName, p, t, 16, errorFn, index...)
	case reflect.Uint32:
		return pb.setUintNField(paramName, fieldName, p, t, 32, errorFn, index...)
	case reflect.Uint64:
		return pb.setUintNField(paramName, fieldName, p, t, 64, errorFn, index...)
	case reflect.Float32:
		return pb.setFloatNField(paramName, fieldName, p, t, 32, errorFn, index...)
	case reflect.Float64:
		return pb.setFloatNField(paramName, fieldName, p, t, 64, errorFn, index...)
	case reflect.Ptr:
		return pb.considerStructField(paramName, fieldName, p, t, errorFn, index...)
	case reflect.Slice:
		return pb.populateSlice(paramName, fieldName, p, t, errorFn)

	}

	return nil

}

func (pb *ParamValueInjector) populateSlice(paramName string, fieldName string, qp *Params, t interface{}, errorFn GenerateMappingError) error {

	tf := reflect.ValueOf(t).Elem().FieldByName(fieldName)

	paramValue, err := qp.StringValue(paramName)

	paramValue = strings.TrimSpace(paramValue)

	if err != nil {
		return err
	}

	values := strings.Split(paramValue, ",")
	vlen := len(values)

	if vlen == 1 && len(paramValue) == 0 {
		vlen = 0
	}

	refSlice := reflect.MakeSlice(tf.Type(), vlen, vlen)
	tf.Set(refSlice)

	if vlen == 0 {
		return nil
	}

	for i, v := range values {
		p := NewSingleValueParams(paramName, strings.TrimSpace(v))

		if err := pb.BindValueToField(paramName, fieldName, p, t, errorFn, i); err != nil {
			return err
		}

	}

	return nil

}

func (pb *ParamValueInjector) considerStructField(paramName string, fieldName string, qp *Params, t interface{}, errorFn GenerateMappingError, index ...int) error {

	tf := reflect.ValueOf(t).Elem().FieldByName(fieldName)
	tv := tf.Interface()

	_, found := tv.(Nilable)

	if found {
		return pb.setNilableField(paramName, fieldName, qp, tf, tv, errorFn, t)
	}

	return nil
}

func (pb *ParamValueInjector) setNilableField(paramName string, fieldName string, p *Params, tf reflect.Value, tv interface{}, errorFn GenerateMappingError, parent interface{}) error {
	np := new(nillableProxy)

	var e error
	var nv interface{}

	switch tv.(type) {

	case *NilableString:
		e = pb.setStringField(paramName, "S", p, np, errorFn)
		nv = NewNilableString(np.S)

	case *NilableBool:
		if p.NotEmpty(paramName) {
			e = pb.setBoolField(paramName, "B", p, np, errorFn)
			nv = NewNilableBool(np.B)
		}

	case *NilableInt64:
		if p.NotEmpty(paramName) {
			e = pb.setIntNField(paramName, "I", p, np, 64, errorFn)
			nv = NewNilableInt64(np.I)
		}

	case *NilableFloat64:
		if p.NotEmpty(paramName) {
			e = pb.setFloatNField(paramName, "F", p, np, 64, errorFn)
			nv = NewNilableFloat64(np.F)
		}
	}

	if e == nil {
		rt.SetPtrToStruct(parent, fieldName, nv)
	} else {

		if fe, okay := e.(FieldAssociatedError); okay {
			fe.RecordField(fieldName)
		}

	}

	return e
}

func (pb *ParamValueInjector) setStringField(paramName string, fieldName string, qp *Params, t interface{}, errorFn GenerateMappingError, index ...int) error {
	s, err := qp.StringValue(paramName)

	if err != nil {
		return errorFn(paramName, fieldName, "string", qp)
	}

	if index != nil {
		rt.SetSliceElem(t, fieldName, s, index[0])
	} else {
		rt.SetString(t, fieldName, s)
	}

	return nil
}

func (pb *ParamValueInjector) setBoolField(paramName string, fieldName string, qp *Params, t interface{}, errorFn GenerateMappingError, index ...int) error {
	b, err := qp.BoolValue(paramName)

	if err != nil {
		return errorFn(paramName, fieldName, "bool", qp)
	}

	if index != nil {
		rt.SetSliceElem(t, fieldName, b, index[0])
	} else {
		rt.SetBool(t, fieldName, b)
	}

	return nil
}

func (pb *ParamValueInjector) setIntNField(paramName string, fieldName string, qp *Params, t interface{}, bits int, errorFn GenerateMappingError, index ...int) error {
	i, err := qp.IntNValue(paramName, bits)

	if err != nil {
		return errorFn(paramName, fieldName, pb.intTypeName("int", bits), qp)
	}

	if index != nil {

		rt.SetSliceElem(t, fieldName, qp.IntNInterfaceValue(i, bits), index[0])
	} else {
		rt.SetInt64(t, fieldName, i)
	}

	return nil
}

func (pb *ParamValueInjector) setFloatNField(paramName string, fieldName string, qp *Params, t interface{}, bits int, errorFn GenerateMappingError, index ...int) error {
	f, err := qp.FloatNValue(paramName, bits)

	if err != nil {
		return errorFn(paramName, fieldName, pb.intTypeName("float", bits), qp)
	}

	if index != nil {
		rt.SetSliceElem(t, fieldName, f, index[0])
	} else {
		rt.SetFloat64(t, fieldName, f)
	}

	return nil
}

func (pb *ParamValueInjector) setUintNField(paramName string, fieldName string, qp *Params, t interface{}, bits int, errorFn GenerateMappingError, index ...int) error {
	i, err := qp.UIntNValue(paramName, bits)

	if err != nil {
		return errorFn(paramName, fieldName, pb.intTypeName("uint", bits), qp)
	}

	if index != nil {
		rt.SetSliceElem(t, fieldName, qp.UIntNInterfaceValue(i, bits), index[0])
	} else {
		rt.SetUint64(t, fieldName, i)
	}

	return nil
}

func (pb *ParamValueInjector) intTypeName(prefix string, bits int) string {
	if bits == 0 {
		return prefix
	}

	return prefix + strconv.Itoa(bits)
}

type nillableProxy struct {
	S string
	B bool
	I int64
	F float64
}
