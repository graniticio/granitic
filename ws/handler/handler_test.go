// Copyright 2016-2020 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package handler

import (
	"bufio"
	"bytes"
	"context"
	"github.com/graniticio/granitic/v2/httpendpoint"
	"github.com/graniticio/granitic/v2/iam"
	"github.com/graniticio/granitic/v2/instrument"
	"github.com/graniticio/granitic/v2/test"
	"github.com/graniticio/granitic/v2/ws"
	"net/http"
	"os"
	"path/filepath"
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

func TestInstrumentation(t *testing.T) {

	l := new(ProcessOnlyLogic)

	h, req := GetHandler(t)

	h.Logic = l
	err := h.StartComponent()

	test.ExpectNil(t, err)

	uw := NewStringBufferResponseWriter()
	w := httpendpoint.NewHTTPResponseWriter(uw)

	ti := new(testInstrumentor)

	ctx := instrument.AddInstrumentorToContext(context.Background(), ti)

	h.ServeHTTP(ctx, w, req)

	test.ExpectBool(t, l.Called, true)

	test.ExpectBool(t, ti.identitySeen, true)
	test.ExpectBool(t, ti.requestSeen, true)
	test.ExpectBool(t, ti.handlerSeen, true)

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

func TestHandlerWithProcessPayload(t *testing.T) {

	l := new(mockLogic)

	h, req := GetHandler(t)

	h.Logic = l
	err := h.StartComponent()

	test.ExpectNil(t, err)

	uw := NewStringBufferResponseWriter()
	w := httpendpoint.NewHTTPResponseWriter(uw)

	h.ServeHTTP(context.Background(), w, req)
	h.ServeHTTP(context.Background(), w, req)

}

func TestProcessPayloadDeclarationValidation(t *testing.T) {

	ml := new(mockLogic)

	h := new(WsHandler)
	h.Logic = ml

	err := h.validateProcessPayload()

	if err != nil {
		t.Fatalf(err.Error())
	}

	h.Logic = new(mockLogicInvalid)

	err = h.validateProcessPayload()

	if err == nil {
		t.Fatalf("Expected validation to fail")
	}

}

func TestFactoryFunctionFromLogic(t *testing.T) {

	h := new(WsHandler)

	h.Logic = new(mockLogic)

	f := h.extractFactoryFromLogic()

	if f == nil {
		t.Fatalf("Expected function to be extracted")
	}

	x := f()

	if _, found := x.(*mockTarget); !found {
		t.Fatalf("Expected a *mockTarget, was %T", x)
	}
}

func TestHandlerStart(t *testing.T) {

	wh, _ := GetHandler(t)

	wh.PathPattern = ""
	wh.HTTPMethod = ""
	wh.Logic = nil

	err := wh.StartComponent()

	if err == nil {
		t.Fatalf("Expected error")
	}

	wh, _ = GetHandler(t)

	wh.HTTPMethod = ""
	wh.Logic = nil

	err = wh.StartComponent()

	if err == nil {
		t.Fatalf("Expected error")
	}

	wh, _ = GetHandler(t)

	wh.Logic = nil

	err = wh.StartComponent()

	if err == nil {
		t.Fatalf("Expected error")
	}

	wh, _ = GetHandler(t)

	wh.Logic = new(mockLogic)

	err = wh.StartComponent()

	if err != nil {
		t.Fatalf(err.Error())
	}

	if wh.createTarget == nil {
		t.Fatalf("Expected create function to have been extracted")
	}

}

func GetHandler(t *testing.T) (*WsHandler, *http.Request) {

	gf := filepath.Join("ws", "get")

	getFilePath := test.FilePath(gf)

	fr, err := os.Open(getFilePath)
	test.ExpectNil(t, err)

	req, err := http.ReadRequest(bufio.NewReader(fr))

	test.ExpectNil(t, err)
	h := new(WsHandler)
	h.PathPattern = "/test$"
	h.HTTPMethod = "GET"
	h.ResponseWriter = new(NilResponseWriter)
	h.componentName = "testHandler"

	return h, req
}

type ProcessOnlyLogic struct {
	Called bool
}

func (l *ProcessOnlyLogic) Process(ctx context.Context, request *ws.Request, response *ws.Response) {
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

func (rw *NilResponseWriter) Write(ctx context.Context, state *ws.ProcessState, outcome ws.Outcome) error {
	return nil
}

type AllPhasesLogic struct {
	ProcessCalled          bool
	UnmarshallTargetCalled bool
	ValidateCalled         bool
	PostProcessCalled      bool
	PreValidateCalled      bool
}

func (l *AllPhasesLogic) Process(ctx context.Context, request *ws.Request, response *ws.Response) {
	l.ProcessCalled = true
}

func (l *AllPhasesLogic) UnmarshallTarget() interface{} {
	l.UnmarshallTargetCalled = true

	return new(Body)
}

func (l *AllPhasesLogic) Validate(ctx context.Context, errors *ws.ServiceErrors, request *ws.Request) {
	l.ValidateCalled = true
}

func (l *AllPhasesLogic) PostProcess(ctx context.Context, handlerName string, request *ws.Request, response *ws.Response) {
	l.PostProcessCalled = true
}

func (l *AllPhasesLogic) PreValidate(ctx context.Context, request *ws.Request, errors *ws.ServiceErrors) (proceed bool) {
	l.PreValidateCalled = true

	return true
}

type Body struct{}

type mockTarget struct {
	Outcome string
}

type mockLogic struct {
}

func (ml *mockLogic) ProcessPayload(ctx context.Context, request *ws.Request, response *ws.Response, target *mockTarget) {
	target.Outcome = "PROCESSED"

}

type mockLogicInvalid struct {
}

func (ml *mockLogicInvalid) ProcessPayload(ctx context.Context, request *ws.Request, response *ws.Response, target mockTarget) {

}

type testInstrumentor struct {
	requestSeen  bool
	identitySeen bool
	handlerSeen  bool
}

func (ni *testInstrumentor) StartEvent(id string, metadata ...interface{}) instrument.EndEvent {
	return ni.endEvent
}

func (ni *testInstrumentor) endEvent() {

	return

}

func (ni *testInstrumentor) Fork(ctx context.Context) (context.Context, instrument.Instrumentor) {
	return ctx, ni
}

func (ni *testInstrumentor) Integrate(instrumentor instrument.Instrumentor) {
	return
}

func (ni *testInstrumentor) Amend(additional instrument.Additional, value interface{}) {

	if additional == instrument.Handler {
		if _, okay := value.(*WsHandler); okay {
			ni.handlerSeen = true
		}
	} else if additional == instrument.Request {
		if _, okay := value.(*ws.Request); okay {
			ni.requestSeen = true
		}
	} else if additional == instrument.UserIdentity {
		if _, okay := value.(iam.ClientIdentity); okay {
			ni.identitySeen = true
		}
	}

	return
}
