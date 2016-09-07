package handler

import (
	"bufio"
	"bytes"
	"github.com/graniticio/granitic/httpendpoint"
	"github.com/graniticio/granitic/test"
	"github.com/graniticio/granitic/ws"
	"golang.org/x/net/context"
	"net/http"
	"os"
	"testing"
)

func TestMinimal(t *testing.T) {

	l := new(ProcessOnlyLogic)

	h, req := GetHandler(t)

	h.Logic = l
	err := h.StartComponent()

	test.ExpectNil(t, err)

	uw := NewStringBufferResponseWriter()
	w := httpendpoint.NewHTTPResponseWriter(uw)

	h.ServeHTTP(context.Background(), w, req)

	test.ExpectBool(t, l.Called, true)

}

func TestAllOptionalPhases(t *testing.T) {

	l := new(AllPhasesLogic)

	h, req := GetHandler(t)

	h.Logic = l
	err := h.StartComponent()

	test.ExpectNil(t, err)

	uw := NewStringBufferResponseWriter()
	w := httpendpoint.NewHTTPResponseWriter(uw)
	h.PreValidateManipulator = l
	h.PostProcessor = l

	h.ServeHTTP(context.Background(), w, req)

	test.ExpectBool(t, l.ProcessCalled, true)
	test.ExpectBool(t, l.UnmarshallTargetCalled, true)
	test.ExpectBool(t, l.ValidateCalled, true)
	test.ExpectBool(t, l.PostProcessCalled, true)
	test.ExpectBool(t, l.PreValidateCalled, true)

}

func GetHandler(t *testing.T) (*WsHandler, *http.Request) {

	getFilePath := test.TestFilePath("ws/get")
	fr, err := os.Open(getFilePath)
	test.ExpectNil(t, err)

	req, err := http.ReadRequest(bufio.NewReader(fr))

	test.ExpectNil(t, err)
	h := new(WsHandler)
	h.PathMatchPattern = "/test$"
	h.HttpMethod = "GET"
	h.ResponseWriter = new(NilResponseWriter)
	h.componentName = "testHandler"

	return h, req
}

type ProcessOnlyLogic struct {
	Called bool
}

func (l *ProcessOnlyLogic) Process(ctx context.Context, request *ws.WsRequest, response *ws.WsResponse) {
	l.Called = true
}

type StringBufferResponseWriter struct {
	h      http.Header
	buffer bytes.Buffer
}

func (w *StringBufferResponseWriter) Header() http.Header {
	return w.h
}

func (w *StringBufferResponseWriter) Write(b []byte) (int, error) {
	return w.buffer.Write(b)
}

func (w *StringBufferResponseWriter) WriteHeader(i int) {

}

func NewStringBufferResponseWriter() *StringBufferResponseWriter {
	w := new(StringBufferResponseWriter)
	w.h = make(http.Header)

	return w
}

type NilResponseWriter struct{}

func (rw *NilResponseWriter) Write(state *ws.WsProcessState, outcome ws.WsOutcome) error {
	return nil
}

type AllPhasesLogic struct {
	ProcessCalled          bool
	UnmarshallTargetCalled bool
	ValidateCalled         bool
	PostProcessCalled      bool
	PreValidateCalled      bool
}

func (l *AllPhasesLogic) Process(ctx context.Context, request *ws.WsRequest, response *ws.WsResponse) {
	l.ProcessCalled = true
}

func (l *AllPhasesLogic) UnmarshallTarget() interface{} {
	l.UnmarshallTargetCalled = true

	return new(Body)
}

func (l *AllPhasesLogic) Validate(ctx context.Context, errors *ws.ServiceErrors, request *ws.WsRequest) {
	l.ValidateCalled = true
}

func (l *AllPhasesLogic) PostProcess(ctx context.Context, handlerName string, request *ws.WsRequest, response *ws.WsResponse) {
	l.PostProcessCalled = true
}

func (l *AllPhasesLogic) PreValidate(ctx context.Context, request *ws.WsRequest, errors *ws.ServiceErrors) (proceed bool) {
	l.PreValidateCalled = true

	return true
}

type Body struct{}
