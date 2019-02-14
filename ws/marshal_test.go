package ws

import (
	"bytes"
	"context"
	"github.com/graniticio/granitic/v2/httpendpoint"
	"github.com/graniticio/granitic/v2/logging"
	"net/http"
	"testing"
)

func TestMarshalAbnormal(t *testing.T) {

	mrw := new(MarshallingResponseWriter)

	ps := new(ProcessState)
	ps.Status = 400
	ps.HTTPResponseWriter = httpendpoint.NewHTTPResponseWriter(new(resWriter))

	feg := new(FrameworkErrorGenerator)
	feg.HTTPMessages = map[string]string{"400": "Bad request"}
	feg.FrameworkLogger = new(logging.ConsoleErrorLogger)

	mrw.FrameworkErrors = feg
	mrw.FrameworkLogger = new(logging.ConsoleErrorLogger)

	scd := new(GraniticHTTPStatusCodeDeterminer)

	mrw.StatusDeterminer = scd

	mrw.ErrorFormatter = new(mockErrorFormatter)
	mrw.ResponseWrapper = new(mockResponseWrapper)
	mrw.MarshalingWriter = new(mockWriter)

	mrw.WriteAbnormalStatus(context.Background(), ps)

}

type resWriter struct {
	sw bytes.Buffer
}

func (rw *resWriter) Header() http.Header {
	return nil
}

func (rw *resWriter) Write(b []byte) (int, error) {
	return rw.sw.Write(b)
}

func (rw *resWriter) WriteHeader(statusCode int) {

}

type mockErrorFormatter struct{}

func (mef *mockErrorFormatter) FormatErrors(errors *ServiceErrors) interface{} {
	return "ERROR"
}

type mockResponseWrapper struct{}

func (mr mockResponseWrapper) WrapResponse(body interface{}, errors interface{}) interface{} {
	return "WRAPPED"
}

type mockWriter struct{}

func (mw *mockWriter) MarshalAndWrite(data interface{}, w http.ResponseWriter) error {
	return nil
}
