package json

import (
	"encoding/json"
	"errors"
	"github.com/graniticio/granitic/httpendpoint"
	"github.com/graniticio/granitic/logging"
	"github.com/graniticio/granitic/ws"
	"golang.org/x/net/context"
)

type StandardJSONResponseWriter struct {
	FrameworkLogger  logging.Logger
	StatusDeterminer ws.HttpStatusCodeDeterminer
	FrameworkErrors  *ws.FrameworkErrorGenerator
	DefaultHeaders   map[string]string
	ResponseWrapper  ws.ResponseWrapper
	HeaderBuilder    ws.WsCommonResponseHeaderBuilder
	ErrorFormatter   ws.ErrorFormatter
	PrettyPrint      bool
	IndentString     string
	PrefixString     string
}

func (rw *StandardJSONResponseWriter) Write(ctx context.Context, state *ws.WsProcessState, outcome ws.WsOutcome) error {

	var ch map[string]string

	if rw.HeaderBuilder != nil {
		ch = rw.HeaderBuilder.BuildHeaders(ctx, state)
	}

	switch outcome {
	case ws.Normal:
		return rw.write(ctx, state.WsResponse, state.HTTPResponseWriter, ch)
	case ws.Error:
		return rw.writeErrors(ctx, state.ServiceErrors, state.HTTPResponseWriter, ch)
	case ws.Abnormal:
		return rw.writeAbnormalStatus(ctx, state.Status, state.HTTPResponseWriter, ch)
	}

	return errors.New("Unsuported ws.WsOutcome value")
}

func (rw *StandardJSONResponseWriter) write(ctx context.Context, res *ws.WsResponse, w *httpendpoint.HTTPResponseWriter, ch map[string]string) error {

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

	ef := rw.ErrorFormatter
	wrap := rw.ResponseWrapper

	fe := ef.FormatErrors(e)
	wrapper := wrap.WrapResponse(res.Body, fe)

	var data []byte
	var err error

	if rw.PrettyPrint {
		data, err = json.MarshalIndent(wrapper, rw.PrefixString, rw.IndentString)
	} else {
		data, err = json.Marshal(wrapper)
	}

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

func (rw *StandardJSONResponseWriter) WriteAbnormalStatus(ctx context.Context, state *ws.WsProcessState) error {
	return rw.Write(ctx, state, ws.Abnormal)
}

func (rw *StandardJSONResponseWriter) writeAbnormalStatus(ctx context.Context, status int, w *httpendpoint.HTTPResponseWriter, ch map[string]string) error {

	res := new(ws.WsResponse)
	res.HttpStatus = status
	var errors ws.ServiceErrors

	e := rw.FrameworkErrors.HttpError(status)
	errors.AddError(e)

	res.Errors = &errors

	return rw.write(ctx, res, w, ch)

}

func (rw *StandardJSONResponseWriter) writeErrors(ctx context.Context, errors *ws.ServiceErrors, w *httpendpoint.HTTPResponseWriter, ch map[string]string) error {

	res := new(ws.WsResponse)
	res.Errors = errors

	return rw.write(ctx, res, w, ch)
}

type StandardJSONErrorFormatter struct{}

func (ef *StandardJSONErrorFormatter) FormatErrors(errors *ws.ServiceErrors) interface{} {

	if errors == nil || !errors.HasErrors() {
		return nil
	}

	f := make(map[string]interface{})

	generalErrors := make([]errorWrapper, 0)
	fieldErrors := make(map[string][]errorWrapper, 0)

	for _, error := range errors.Errors {

		c := ws.CategoryToCode(error.Category)
		displayCode := c + "-" + error.Label

		field := error.Field

		if field == "" {
			generalErrors = append(generalErrors, errorWrapper{displayCode, error.Message})
		} else {

			fe := fieldErrors[field]

			if fe == nil {
				fe = make([]errorWrapper, 0)

			}

			fe = append(fe, errorWrapper{displayCode, error.Message})
			fieldErrors[field] = fe

		}
	}

	if len(generalErrors) > 0 {
		f["General"] = generalErrors
	}

	if len(fieldErrors) > 0 {
		f["ByField"] = fieldErrors
	}

	return f
}

type errorWrapper struct {
	Code    string
	Message string
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
