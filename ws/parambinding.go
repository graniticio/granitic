// Copyright 2016 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package ws

import (
	"github.com/graniticio/granitic/logging"
	rt "github.com/graniticio/granitic/reflecttools"
	"github.com/graniticio/granitic/types"
	"reflect"
	"strconv"
)

type bindError func(string, string, string, *WsParams) *WsFrameworkError

// Takes string parameters extracted from an HTTP request, converts them to Go native or Granitic nilable types and
// injects them into the RequestBody on a WsRequest.
type ParamBinder struct {
	// Injected by Granitic
	FrameworkLogger logging.Logger

	// Source of service errors for errors encountered while binding.
	FrameworkErrors *FrameworkErrorGenerator
}

// BindPathParameters takes strings extracted from an HTTP's request path (using regular expression groups) and
// injects them into fields on the WsRequest.RequestBody. Any errors encountered are recorded as framework errors in
// the WsRequest.
func (pb *ParamBinder) BindPathParameters(wsReq *WsRequest, p *WsParams) {

	t := wsReq.RequestBody

	for i, fieldName := range p.ParamNames() {

		if rt.HasFieldOfName(t, fieldName) {
			fErr := pb.bindValueToField(strconv.Itoa(i), fieldName, p, t, pb.pathParamError)

			if fErr != nil {
				fErr.Position = i
				wsReq.AddFrameworkError(fErr)
			} else {
				wsReq.RecordFieldAsBound(fieldName)
			}

		} else {
			pb.FrameworkLogger.LogWarnf("No field %s exists on a target object to bind a path parameter into.", fieldName)
		}

	}

}

// BindPathParameters takes the query parameters from an HTTP request and
// injects them into fields on the WsRequest.RequestBody using the keys of the supplied map as the name of the target fields.
// Any errors encountered are recorded as framework errors in the WsRequest.
func (pb *ParamBinder) BindQueryParameters(wsReq *WsRequest, targets map[string]string) {

	t := wsReq.RequestBody
	p := wsReq.QueryParams
	l := pb.FrameworkLogger

	for field, param := range targets {

		if rt.HasFieldOfName(t, field) {

			if p.Exists(param) {
				l.LogTracef("Binding parameter %s to field %s", param, field)

				fErr := pb.bindValueToField(param, field, p, t, pb.queryParamError)

				if fErr != nil {
					wsReq.AddFrameworkError(fErr)
				} else {
					wsReq.RecordFieldAsBound(field)
				}
			}

		} else {
			l.LogErrorf("No field named %s exists to bind a query parameter into", field)
			m, c := pb.FrameworkErrors.MessageCode(QueryNoTargetField, field, param)
			wsReq.AddFrameworkError(NewQueryBindFrameworkError(m, c, param, field))
		}
	}

	pb.initialiseUnsetNilables(t)
}

// BindPathParameters takes the query parameters from an HTTP request and
// injects them into fields on the WsRequest.RequestBody assuming the parameters have exactly the same name as the target
// fields. Any errors encountered are recorded as framework errors in the WsRequest.
func (pb *ParamBinder) AutoBindQueryParameters(wsReq *WsRequest) {

	t := wsReq.RequestBody
	p := wsReq.QueryParams

	for _, paramName := range p.ParamNames() {

		if rt.HasFieldOfName(t, paramName) {

			fErr := pb.bindValueToField(paramName, paramName, p, t, pb.queryParamError)

			if fErr != nil {
				wsReq.AddFrameworkError(fErr)
			} else {
				wsReq.RecordFieldAsBound(paramName)
			}

		}

	}

	pb.initialiseUnsetNilables(t)
}

func (pb *ParamBinder) initialiseUnsetNilables(t interface{}) {

	vt := reflect.ValueOf(t).Elem()

FieldLoop:
	for i := 0; i < vt.NumField(); i++ {

		f := vt.Field(i)

		if !rt.NilPointer(f) {
			//Not nil
			continue FieldLoop
		}

		var nv interface{}

		switch f.Interface().(type) {

		case *types.NilableString:
			nv = new(types.NilableString)
		case *types.NilableBool:
			nv = new(types.NilableBool)
		case *types.NilableInt64:
			nv = new(types.NilableInt64)
		case *types.NilableFloat64:
			nv = new(types.NilableFloat64)
		default:
			continue FieldLoop
		}

		if err := rt.SetFieldPtrToStruct(f, nv); err != nil {
			pb.FrameworkLogger.LogErrorf("Problem initialising a nilable field %s", err.Error())
		}

	}

}

func (pb *ParamBinder) queryParamError(paramName string, fieldName string, typeName string, p *WsParams) *WsFrameworkError {

	var v = ""

	if p.Exists(paramName) {
		v, _ = p.StringValue(paramName)
	}

	m, c := pb.FrameworkErrors.MessageCode(QueryWrongType, paramName, typeName, v)
	return NewQueryBindFrameworkError(m, c, paramName, fieldName)

}

func (pb *ParamBinder) pathParamError(paramName string, fieldName string, typeName string, p *WsParams) *WsFrameworkError {

	var v = ""

	if p.Exists(paramName) {
		v, _ = p.StringValue(paramName)
	}

	m, c := pb.FrameworkErrors.MessageCode(PathWrongType, paramName, typeName, v)
	return NewPathBindFrameworkError(m, c, fieldName)

}

func (pb *ParamBinder) bindValueToField(paramName string, fieldName string, p *WsParams, t interface{}, errorFn bindError) *WsFrameworkError {

	if !rt.TargetFieldIsArray(t, fieldName) && p.MultipleValues(paramName) {
		m, c := pb.FrameworkErrors.MessageCode(QueryTargetNotArray, fieldName)
		return NewQueryBindFrameworkError(m, c, paramName, fieldName)
	}

	switch rt.TypeOfField(t, fieldName).Kind() {
	case reflect.Int:
		return pb.setIntNField(paramName, fieldName, p, t, 0, errorFn)
	case reflect.Int8:
		return pb.setIntNField(paramName, fieldName, p, t, 8, errorFn)
	case reflect.Int16:
		return pb.setIntNField(paramName, fieldName, p, t, 16, errorFn)
	case reflect.Int32:
		return pb.setIntNField(paramName, fieldName, p, t, 32, errorFn)
	case reflect.Int64:
		return pb.setIntNField(paramName, fieldName, p, t, 64, errorFn)
	case reflect.Bool:
		return pb.setBoolField(paramName, fieldName, p, t, errorFn)
	case reflect.String:
		return pb.setStringField(paramName, fieldName, p, t, errorFn)
	case reflect.Uint8:
		return pb.setUintNField(paramName, fieldName, p, t, 8, errorFn)
	case reflect.Uint16:
		return pb.setUintNField(paramName, fieldName, p, t, 16, errorFn)
	case reflect.Uint32:
		return pb.setUintNField(paramName, fieldName, p, t, 32, errorFn)
	case reflect.Uint64:
		return pb.setUintNField(paramName, fieldName, p, t, 64, errorFn)
	case reflect.Float32:
		return pb.setFloatNField(paramName, fieldName, p, t, 32, errorFn)
	case reflect.Float64:
		return pb.setFloatNField(paramName, fieldName, p, t, 64, errorFn)
	case reflect.Ptr:
		return pb.considerStructField(paramName, fieldName, p, t, errorFn)

	}

	return nil

}

func (pb *ParamBinder) considerStructField(paramName string, fieldName string, qp *WsParams, t interface{}, errorFn bindError) *WsFrameworkError {

	tf := reflect.ValueOf(t).Elem().FieldByName(fieldName)
	tv := tf.Interface()

	_, found := tv.(types.Nilable)

	if found {
		return pb.setNilableField(paramName, fieldName, qp, tf, tv, errorFn, t)
	}

	return nil
}

func (pb *ParamBinder) setNilableField(paramName string, fieldName string, p *WsParams, tf reflect.Value, tv interface{}, errorFn bindError, parent interface{}) *WsFrameworkError {
	np := new(nillableProxy)

	var e *WsFrameworkError
	var nv interface{}

	switch tv.(type) {

	case *types.NilableString:
		e = pb.setStringField(paramName, "S", p, np, errorFn)
		nv = types.NewNilableString(np.S)

	case *types.NilableBool:
		if p.NotEmpty(paramName) {
			e = pb.setBoolField(paramName, "B", p, np, errorFn)
			nv = types.NewNilableBool(np.B)
		}

	case *types.NilableInt64:
		if p.NotEmpty(paramName) {
			e = pb.setIntNField(paramName, "I", p, np, 64, errorFn)
			nv = types.NewNilableInt64(np.I)
		}

	case *types.NilableFloat64:
		if p.NotEmpty(paramName) {
			e = pb.setFloatNField(paramName, "F", p, np, 64, errorFn)
			nv = types.NewNilableFloat64(np.F)
		}
	}

	if e == nil {
		rt.SetPtrToStruct(parent, fieldName, nv)
	} else {
		e.TargetField = fieldName
	}

	return e
}

func (pb *ParamBinder) setStringField(paramName string, fieldName string, qp *WsParams, t interface{}, errorFn bindError) *WsFrameworkError {
	s, err := qp.StringValue(paramName)

	if err != nil {
		return errorFn(paramName, fieldName, "string", qp)
	}

	rt.SetString(t, fieldName, s)

	return nil
}

func (pb *ParamBinder) setBoolField(paramName string, fieldName string, qp *WsParams, t interface{}, errorFn bindError) *WsFrameworkError {
	b, err := qp.BoolValue(paramName)

	if err != nil {
		return errorFn(paramName, fieldName, "bool", qp)
	}

	rt.SetBool(t, fieldName, b)
	return nil
}

func (pb *ParamBinder) setIntNField(paramName string, fieldName string, qp *WsParams, t interface{}, bits int, errorFn bindError) *WsFrameworkError {
	i, err := qp.IntNValue(paramName, bits)

	if err != nil {
		return errorFn(paramName, fieldName, pb.intTypeName("int", bits), qp)
	}

	rt.SetInt64(t, fieldName, i)
	return nil
}

func (pb *ParamBinder) setFloatNField(paramName string, fieldName string, qp *WsParams, t interface{}, bits int, errorFn bindError) *WsFrameworkError {
	i, err := qp.FloatNValue(paramName, bits)

	if err != nil {
		return errorFn(paramName, fieldName, pb.intTypeName("float", bits), qp)
	}

	rt.SetFloat64(t, fieldName, i)
	return nil
}

func (pb *ParamBinder) setUintNField(paramName string, fieldName string, qp *WsParams, t interface{}, bits int, errorFn bindError) *WsFrameworkError {
	i, err := qp.UIntNValue(paramName, bits)

	if err != nil {
		return errorFn(paramName, fieldName, pb.intTypeName("uint", bits), qp)
	}

	rt.SetUint64(t, fieldName, i)
	return nil
}

func (pb *ParamBinder) intTypeName(prefix string, bits int) string {
	if bits == 0 {
		return prefix
	} else {
		return prefix + strconv.Itoa(bits)
	}
}

type nillableProxy struct {
	S string
	B bool
	I int64
	F float64
}
