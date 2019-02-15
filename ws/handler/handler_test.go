// Copyright 2016-2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package handler

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/graniticio/granitic/v2/httpendpoint"
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

	gf := filepath.Join("ws", "get")

	getFilePath := test.FilePath(gf)

	fmt.Println(getFilePath)

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
