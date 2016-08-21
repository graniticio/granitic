package json

import (
	"encoding/json"
	"errors"
	"github.com/graniticio/granitic/httpendpoint"
	"github.com/graniticio/granitic/logging"
	"github.com/graniticio/granitic/ws"
)

type StandardJSONResponseWriter struct {
	FrameworkLogger  logging.Logger
	StatusDeterminer ws.HttpStatusCodeDeterminer
	FrameworkErrors  *ws.FrameworkErrorGenerator
	DefaultHeaders   map[string]string
	ResponseWrapper  ws.ResponseWrapper
	HeaderBuilder    ws.WsCommonResponseHeaderBuilder
	ErrorFormatter   ws.ErrorFormatter
}

func (rw *StandardJSONResponseWriter) Write(state *ws.WsProcessState, outcome ws.WsOutcome) error {

	var ch map[string]string

	if rw.HeaderBuilder != nil {
		ch = rw.HeaderBuilder.BuildHeaders(state)
	}

	switch outcome {
	case ws.Normal:
		return rw.write(state.WsResponse, state.HTTPResponseWriter, ch)
	case ws.Error:
		return rw.writeErrors(state.ServiceErrors, state.HTTPResponseWriter, ch)
	case ws.Abnormal:
		return rw.writeAbnormalStatus(state.Status, state.HTTPResponseWriter, ch)
	}

	return errors.New("Unsuported ws.WsOutcome value")
}

func (rw *StandardJSONResponseWriter) write(res *ws.WsResponse, w *httpendpoint.HTTPResponseWriter, ch map[string]string) error {

	if w.DataSent {
		//This HTTP response has already been written to by another component - not safe to continue
		if rw.FrameworkLogger.IsLevelEnabled(logging.Debug) {
			rw.FrameworkLogger.LogDebugf("Response already written to.")
		}

		return nil
	}

	headers := rw.mergeHeaders(res, ch)
	ws.WriteHeaders(w, headers)

	s := rw.StatusDeterminer.DetermineCode(res)
	w.WriteHeader(s)

	e := res.Errors

	if res.Body == nil && !e.HasErrors() {
		return nil
	}

	var wrapper interface{}

	ef := rw.ErrorFormatter
	wrap := rw.ResponseWrapper

	if e.HasErrors() && res.Body != nil {
		wrapper = wrap.WrapResponse(ef.FormatErrors(e), res.Body)
	} else if e.HasErrors() {
		wrapper = wrap.WrapResponse(ef.FormatErrors(e), nil)
	} else {
		wrapper = wrap.WrapResponse(nil, res.Body)
	}

	data, err := json.Marshal(wrapper)

	if err != nil {
		return err
	}

	_, err = w.Write(data)

	return err
}

// Merges together the headers that have been defined on the WsResponse, the static default headers attache to this writer
// and (optionally) those constructed by the  ws.WsCommonResponseHeaderBuilder attached to this writer. The order of precedence,
// from lowest to highest, is static headers, constructed headers, headers in the WsResponse.
func (rw *StandardJSONResponseWriter) mergeHeaders(res *ws.WsResponse, ch map[string]string) map[string]string {

	merged := make(map[string]string)

	if rw.DefaultHeaders != nil {
		for k, v := range rw.DefaultHeaders {
			merged[k] = v
		}
	}

	if ch != nil {
		for k, v := range ch {
			merged[k] = v
		}
	}

	if res.Headers != nil {
		for k, v := range res.Headers {
			merged[k] = v
		}
	}

	return merged
}

func (rw *StandardJSONResponseWriter) WriteAbnormalStatus(state *ws.WsProcessState) error {
	return rw.Write(state, ws.Abnormal)
}

func (rw *StandardJSONResponseWriter) writeAbnormalStatus(status int, w *httpendpoint.HTTPResponseWriter, ch map[string]string) error {

	res := new(ws.WsResponse)
	res.HttpStatus = status
	var errors ws.ServiceErrors

	e := rw.FrameworkErrors.HttpError(status)
	errors.AddError(e)

	res.Errors = &errors

	return rw.write(res, w, ch)

}

func (rw *StandardJSONResponseWriter) writeErrors(errors *ws.ServiceErrors, w *httpendpoint.HTTPResponseWriter, ch map[string]string) error {

	res := new(ws.WsResponse)
	res.Errors = errors

	return rw.write(res, w, ch)
}

type StandardJSONErrorFormatter struct{}

func (ef *StandardJSONErrorFormatter) FormatErrors(errors *ws.ServiceErrors) interface{} {

	f := make(map[string][]string)

	for _, error := range errors.Errors {

		var c string

		switch error.Category {
		default:
			c = "?"
		case ws.Unexpected:
			c = "U"
		case ws.Security:
			c = "S"
		case ws.Logic:
			c = "L"
		case ws.Client:
			c = "C"
		case ws.HTTP:
			c = "H"
		}

		k := c + "-" + error.Label
		a := f[k]

		f[k] = append(a, error.Message)
	}

	return f
}

type StandardJSONResponseWrapper struct {
	ErrorsFieldName string
	BodyFieldName   string
}

func (rw *StandardJSONResponseWrapper) WrapResponse(body interface{}, errors interface{}) interface{} {
	f := make(map[string]interface{})

	if errors != nil {
		f[rw.ErrorsFieldName] = errors
	}

	if body != nil {
		f[rw.BodyFieldName] = body
	}

	return f
}
