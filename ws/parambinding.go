package ws

import (
	"github.com/graniticio/granitic/logging"
	rt "github.com/graniticio/granitic/reflecttools"
	"github.com/graniticio/granitic/types"
	"reflect"
	"strconv"
)

type bindError func(string, string, string, *WsParams) *WsFrameworkError

type ParamBinder struct {
	FrameworkLogger logging.Logger
	FrameworkErrors *FrameworkErrorGenerator
}

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
}

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

	switch tv.(type) {
	case *types.NillableString, *types.NillableBool, *types.NillableFloat64, *types.NillableInt64:
		return pb.setNillableField(paramName, fieldName, qp, tf, tv, errorFn, t)
	}

	return nil
}

func (pb *ParamBinder) setNillableField(paramName string, fieldName string, p *WsParams, tf reflect.Value, tv interface{}, errorFn bindError, parent interface{}) *WsFrameworkError {
	np := new(nillableProxy)

	var e *WsFrameworkError
	var nv interface{}

	switch tv.(type) {

	case *types.NillableString:
		e = pb.setStringField(paramName, "S", p, np, errorFn)
		nv = types.NewNillableString(np.S)

	case *types.NillableBool:
		if p.NotEmpty(paramName) {
			e = pb.setBoolField(paramName, "B", p, np, errorFn)
			nv = types.NewNillableBool(np.B)
		}

	case *types.NillableInt64:
		if p.NotEmpty(paramName) {
			e = pb.setIntNField(paramName, "I", p, np, 64, errorFn)
			nv = types.NewNillableInt64(np.I)
		}

	case *types.NillableFloat64:
		if p.NotEmpty(paramName) {
			e = pb.setFloatNField(paramName, "F", p, np, 64, errorFn)
			nv = types.NewNillableFloat64(np.F)
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
