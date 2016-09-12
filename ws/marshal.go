package ws

import (
	"errors"
	"github.com/graniticio/granitic/httpendpoint"
	"github.com/graniticio/granitic/logging"
	"golang.org/x/net/context"
	"net/http"
)

type MarshalingWriter interface {
	MarshalAndWrite(data interface{}, w http.ResponseWriter) error
}

type MarshallingResponseWriter struct {
	FrameworkLogger  logging.Logger
	StatusDeterminer HttpStatusCodeDeterminer
	FrameworkErrors  *FrameworkErrorGenerator
	DefaultHeaders   map[string]string
	ResponseWrapper  ResponseWrapper
	HeaderBuilder    WsCommonResponseHeaderBuilder
	ErrorFormatter   ErrorFormatter
	MarshalingWriter MarshalingWriter
}

func (rw *MarshallingResponseWriter) Write(ctx context.Context, state *WsProcessState, outcome WsOutcome) error {

	var ch map[string]string

	if rw.HeaderBuilder != nil {
		ch = rw.HeaderBuilder.BuildHeaders(ctx, state)
	}

	switch outcome {
	case Normal:
		return rw.write(ctx, state.WsResponse, state.HTTPResponseWriter, ch)
	case Error:
		return rw.writeErrors(ctx, state.ServiceErrors, state.HTTPResponseWriter, ch)
	case Abnormal:
		return rw.writeAbnormalStatus(ctx, state.Status, state.HTTPResponseWriter, ch)
	}

	return errors.New("Unsuported WsOutcome value")
}

func (rw *MarshallingResponseWriter) write(ctx context.Context, res *WsResponse, w *httpendpoint.HTTPResponseWriter, ch map[string]string) error {

	if w.DataSent {
		//This HTTP response has already been written to by another component - not safe to continue
		if rw.FrameworkLogger.IsLevelEnabled(logging.Debug) {
			rw.FrameworkLogger.LogDebugfCtx(ctx, "Response already written to.")
		}

		return nil
	}

	headers := MergeHeaders(res, ch, rw.DefaultHeaders)
	WriteHeaders(w, headers)

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

	return rw.MarshalingWriter.MarshalAndWrite(wrapper, w)
}

func (rw *MarshallingResponseWriter) WriteAbnormalStatus(ctx context.Context, state *WsProcessState) error {
	return rw.Write(ctx, state, Abnormal)
}

func (rw *MarshallingResponseWriter) writeAbnormalStatus(ctx context.Context, status int, w *httpendpoint.HTTPResponseWriter, ch map[string]string) error {

	res := new(WsResponse)
	res.HttpStatus = status
	var errors ServiceErrors

	e := rw.FrameworkErrors.HttpError(status)
	errors.AddError(e)

	res.Errors = &errors

	return rw.write(ctx, res, w, ch)

}

func (rw *MarshallingResponseWriter) writeErrors(ctx context.Context, errors *ServiceErrors, w *httpendpoint.HTTPResponseWriter, ch map[string]string) error {

	res := new(WsResponse)
	res.Errors = errors

	return rw.write(ctx, res, w, ch)
}
