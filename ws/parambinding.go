// Copyright 2016-2023 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package ws

import (
	"github.com/graniticio/granitic/v2/logging"
	rt "github.com/graniticio/granitic/v2/reflecttools"
	"github.com/graniticio/granitic/v2/types"
	"reflect"
	"strconv"
)

// ParamBinder takes string parameters extracted from an HTTP request, converts them to Go native or Granitic nilable types and
// injects them into the RequestBody on a Request.
type ParamBinder struct {
	// Injected by Granitic
	FrameworkLogger logging.Logger

	// Source of service errors for errors encountered while binding.
	FrameworkErrors *FrameworkErrorGenerator
}

// BindPathParameters takes strings extracted from an HTTP's request path (using regular expression groups) and
// injects them into fields on the Request.RequestBody. Any errors encountered are recorded as framework errors in
// the Request.
func (pb *ParamBinder) BindPathParameters(wsReq *Request, p *types.Params) {

	t := wsReq.RequestBody

	for i, fieldName := range p.ParamNames() {

		if rt.HasFieldOfName(t, fieldName) {
			err := pb.bindValueToField(strconv.Itoa(i), fieldName, p, t, pb.pathParamError)

			if err != nil {

				if fe, okay := err.(*FrameworkError); okay {
					fe.Position = i
					wsReq.AddFrameworkError(fe)
				} else {
					pb.FrameworkLogger.LogErrorf("Unexpected error of type %t (was expecting *FrameworkError). Message was: %s", err, err.Error())
				}
			} else {
				wsReq.RecordFieldAsBound(fieldName)
			}

		} else {
			pb.FrameworkLogger.LogWarnf("No field %s exists on a target object to bind a path parameter into.", fieldName)
		}

	}

}

// BindQueryParameters takes the query parameters from an HTTP request and
// injects them into fields on the Request.RequestBody using the keys of the supplied map as the name of the target fields.
// Any errors encountered are recorded as framework errors in the Request.
func (pb *ParamBinder) BindQueryParameters(wsReq *Request, targets map[string]string) {

	t := wsReq.RequestBody
	p := wsReq.QueryParams
	l := pb.FrameworkLogger

	for field, param := range targets {

		if rt.HasFieldOfName(t, field) {

			if p.Exists(param) {
				l.LogTracef("Binding parameter %s to field %s", param, field)

				err := pb.bindValueToField(param, field, p, t, pb.queryParamError)

				if err != nil {
					if fe, okay := err.(*FrameworkError); okay {
						wsReq.AddFrameworkError(fe)
					} else {
						pb.FrameworkLogger.LogErrorf("Unexpected error of type %t (was expecting *FrameworkError). Message was: %s", err, err.Error())
					}
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

// AutoBindQueryParameters takes the query parameters from an HTTP request and
// injects them into fields on the Request.RequestBody assuming the parameters have exactly the same name as the target
// fields. Any errors encountered are recorded as framework errors in the Request.
func (pb *ParamBinder) AutoBindQueryParameters(wsReq *Request) {

	t := wsReq.RequestBody
	p := wsReq.QueryParams

	for _, paramName := range p.ParamNames() {

		if rt.HasFieldOfName(t, paramName) {

			err := pb.bindValueToField(paramName, paramName, p, t, pb.queryParamError)

			if err != nil {

				if fe, okay := err.(*FrameworkError); okay {
					wsReq.AddFrameworkError(fe)
				} else {
					pb.FrameworkLogger.LogErrorf("Unexpected error of type %t (was expecting *FrameworkError). Message was: %s", err, err.Error())
				}

			} else {
				wsReq.RecordFieldAsBound(paramName)
			}

		}

	}

	pb.initialiseUnsetNilables(t)
}

func (pb *ParamBinder) bindValueToField(paramName string, fieldName string, p *types.Params, t interface{}, errorFn types.GenerateMappingError) error {

	if !rt.TargetFieldIsArray(t, fieldName) && p.MultipleValues(paramName) {
		m, c := pb.FrameworkErrors.MessageCode(QueryTargetNotArray, fieldName)
		return NewQueryBindFrameworkError(m, c, paramName, fieldName)
	}

	pi := new(types.ParamValueInjector)

	return pi.BindValueToField(paramName, fieldName, p, t, errorFn)

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

func (pb *ParamBinder) queryParamError(paramName string, fieldName string, typeName string, p *types.Params) error {

	var v = ""

	if p.Exists(paramName) {
		v, _ = p.StringValue(paramName)
	}

	m, c := pb.FrameworkErrors.MessageCode(QueryWrongType, paramName, typeName, v)
	return NewQueryBindFrameworkError(m, c, paramName, fieldName)

}

func (pb *ParamBinder) pathParamError(paramName string, fieldName string, typeName string, p *types.Params) error {

	var v = ""

	if p.Exists(paramName) {
		v, _ = p.StringValue(paramName)
	}

	m, c := pb.FrameworkErrors.MessageCode(PathWrongType, paramName, typeName, v)
	return NewPathBindFrameworkError(m, c, fieldName)

}
