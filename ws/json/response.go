package json

import (
	"encoding/json"
	"github.com/graniticio/granitic/logging"
	"github.com/graniticio/granitic/ws"
	"github.com/graniticio/granitic/httpendpoint"
	"errors"
)

type DefaultJsonResponseWriter struct {
	FrameworkLogger  logging.Logger
	StatusDeterminer ws.HttpStatusCodeDeterminer
	FrameworkErrors  *ws.FrameworkErrorGenerator
	DefaultHeaders   map[string]string
	WrapResponse     bool
}

func (djrw *DefaultJsonResponseWriter)  Write(state *ws.WsProcessState, outcome ws.WsOutcome) error{

	switch outcome {
	case ws.Normal:
		return djrw.write(state.WsResponse, state.HTTPResponseWriter)
	case ws.Error:
		return djrw.writeErrors(state.ServiceErrors, state.HTTPResponseWriter)
	case ws.Abnormal:
		return djrw.writeAbnormalStatus(state.Status, state.HTTPResponseWriter)
	}


	return errors.New("Unsuported ws.WsOutcome value")
}


func (djrw *DefaultJsonResponseWriter) write(res *ws.WsResponse, w *httpendpoint.HTTPResponseWriter) error {

	if w.DataSent {
		//This HTTP response has already been written to by another component - not safe to continue
		if djrw.FrameworkLogger.IsLevelEnabled(logging.Debug) {
			djrw.FrameworkLogger.LogDebugf("Response already written to.")
		}

		return nil
	}


	ws.WriteMetaData(w, res, djrw.DefaultHeaders)

	s := djrw.StatusDeterminer.DetermineCode(res)
	w.WriteHeader(s)

	e := res.Errors

	if res.Body == nil && !e.HasErrors() {
		return nil
	}

	var wrapper interface{}

	if e.HasErrors() && res.Body != nil {
		wrapper = wrapJsonResponse(djrw.formatErrors(e), res.Body)
	} else if e.HasErrors() {
		wrapper = wrapJsonResponse(djrw.formatErrors(e), nil)
	} else {
		wrapper = wrapJsonResponse(nil, res.Body)
	}

	data, err := json.Marshal(wrapper)

	if err != nil {
		return err
	}

	_, err = w.Write(data)

	return err
}

func (djrw *DefaultJsonResponseWriter) WriteAbnormalStatus(state *ws.WsProcessState) error {
	return djrw.Write(state, ws.Abnormal)
}


func (djrw *DefaultJsonResponseWriter) writeAbnormalStatus(status int, w *httpendpoint.HTTPResponseWriter) error {

	res := new(ws.WsResponse)
	res.HttpStatus = status
	var errors ws.ServiceErrors

	e := djrw.FrameworkErrors.HttpError(status)
	errors.AddError(e)

	res.Errors = &errors

	return djrw.write(res, w)

}

func (djrw *DefaultJsonResponseWriter) writeErrors(errors *ws.ServiceErrors, w *httpendpoint.HTTPResponseWriter) error {

	res := new(ws.WsResponse)
	res.Errors = errors

	return djrw.write(res, w)
}

func (djrw *DefaultJsonResponseWriter) formatErrors(errors *ws.ServiceErrors) interface{} {

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

func wrapJsonResponse(errors interface{}, body interface{}) interface{} {
	f := make(map[string]interface{})

	if errors != nil {
		f["Errors"] = errors
	}

	if body != nil {
		f["Response"] = body
	}

	return f
}
