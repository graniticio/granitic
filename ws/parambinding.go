package ws

import (
	"github.com/graniticio/granitic/logging"
	rt "github.com/graniticio/granitic/reflecttools"
	"reflect"
	"strconv"
)

type bindError func(string, string, string) *WsFrameworkError

type ParamBinder struct {
	FrameworkLogger logging.Logger
	FrameworkErrors *FrameworkErrorGenerator
}

func (pb *ParamBinder) AutoBindPathParameters(wsReq *WsRequest, p *WsParams) {

	t := wsReq.RequestBody

	for _, fieldName := range p.ParamNames() {

		if rt.HasFieldOfName(t, fieldName) {
			fErr := pb.bindValueToField(fieldName, fieldName, p, t, pb.pathParamError)

			if fErr != nil {
				wsReq.AddFrameworkError(fErr)
			} else {
				wsReq.RecordFieldAsPopulated(fieldName)
			}

		} else {
			pb.FrameworkLogger.LogWarnf("No field %s exists on a target object to bind a path parameter into.", fieldName)
		}

	}

}

func (pb *ParamBinder) AutoBindQueryParameters(wsReq *WsRequest) {

	t := wsReq.RequestBody
	p := wsReq.QueryParams

	for _, paramName := range p.ParamNames() {

		if rt.HasFieldOfName(t, paramName) {

			if !rt.TargetFieldIsArray(t, paramName) && p.MultipleValues(paramName) {
				m, c := pb.FrameworkErrors.MessageCode(QueryTargetNotArray, paramName)


				wsReq.AddFrameworkError(NewQueryBindFrameworkError(m, c, paramName, paramName))
				continue
			}

			fErr := pb.bindValueToField(paramName, paramName, p, t, pb.queryParamError)

			if fErr != nil {
				wsReq.AddFrameworkError(fErr)
			} else {
				wsReq.RecordFieldAsPopulated(paramName)
			}

		}

	}
}

func (pb *ParamBinder) queryParamError(paramName string, fieldName string, typeName string) *WsFrameworkError {

	m, c := pb.FrameworkErrors.MessageCode(QueryWrongType, paramName, typeName)
	return NewQueryBindFrameworkError(m, c, paramName, fieldName)

}

func (pb *ParamBinder) pathParamError(paramName string, fieldName string, typeName string) *WsFrameworkError {

	m, c  := pb.FrameworkErrors.MessageCode(PathWrongType, typeName)
	return NewPathBindFrameworkError(m, c, fieldName)

}

func (pb *ParamBinder) bindValueToField(paramName string, fieldName string, p *WsParams, t interface{}, errorFn bindError) *WsFrameworkError {

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

	}

	return nil

}

func (pb *ParamBinder) setStringField(paramName string, fieldName string, qp *WsParams, t interface{}, errorFn bindError) *WsFrameworkError {
	s, err := qp.StringValue(paramName)

	if err != nil {
		return errorFn(paramName, fieldName, "string")
	}

	rt.SetString(t, fieldName, s)

	return nil
}

func (pb *ParamBinder) setBoolField(paramName string, fieldName string, qp *WsParams, t interface{}, errorFn bindError) *WsFrameworkError {
	b, err := qp.BoolValue(paramName)

	if err != nil {
		return errorFn(paramName, fieldName, "boolean")
	}

	rt.SetBool(t, fieldName, b)
	return nil
}

func (pb *ParamBinder) setIntNField(paramName string, fieldName string, qp *WsParams, t interface{}, bits int, errorFn bindError) *WsFrameworkError {
	i, err := qp.IntNValue(paramName, bits)

	if err != nil {
		return errorFn(paramName, fieldName, pb.intTypeName("integer", bits))
	}

	rt.SetInt64(t, fieldName, i)
	return nil
}

func (pb *ParamBinder) setFloatNField(paramName string, fieldName string, qp *WsParams, t interface{}, bits int, errorFn bindError) *WsFrameworkError {
	i, err := qp.FloatNValue(paramName, bits)

	if err != nil {
		return errorFn(paramName, fieldName, pb.intTypeName("floating-point", bits))
	}

	rt.SetFloat64(t, fieldName, i)
	return nil
}

func (pb *ParamBinder) setUintNField(paramName string, fieldName string, qp *WsParams, t interface{}, bits int, errorFn bindError) *WsFrameworkError {
	i, err := qp.UIntNValue(paramName, bits)

	if err != nil {
		return errorFn(paramName, fieldName, pb.intTypeName("unsigned-integer", bits))
	}

	rt.SetUint64(t, fieldName, i)
	return nil
}

func (pb *ParamBinder) intTypeName(prefix string, bits int) string {
	if bits == 0 {
		return prefix
	} else {
		return prefix + " (" + strconv.Itoa(bits) + "-bit)"
	}
}
