package ws

import (
	"context"
	"github.com/graniticio/granitic/v3/ioc"
	"github.com/graniticio/granitic/v3/logging"
	"github.com/graniticio/granitic/v3/ws"
	"github.com/graniticio/granitic/v3/ws/handler"
	"net/http"
	"testing"
)

func TestWsHandlerDecorator_DecorateComponent(t *testing.T) {

	wd := new(wsHandlerDecorator)

	wd.FrameworkLogger = new(logging.ConsoleErrorLogger)

	wd.ResponseWriter = new(mrw)

	wd.FrameworkErrors = new(ws.FrameworkErrorGenerator)

	wd.Unmarshaller = new(mum)

	wd.QueryBinder = new(ws.ParamBinder)

	h := new(handler.WsHandler)

	c := ioc.NewComponent("", h)

	if !wd.OfInterest(c) {
		t.FailNow()
	}

	wd.DecorateComponent(c, nil)

	if h.Unmarshaller == nil {
		t.Fail()
	}

	if h.FrameworkErrors == nil {
		t.Fail()
	}

	if h.ResponseWriter == nil {
		t.Fail()
	}

	if h.ParamBinder == nil {
		t.Fail()
	}

}

type mrw struct{}

func (m *mrw) Write(ctx context.Context, state *ws.ProcessState, outcome ws.Outcome) error {
	return nil
}

type mum struct{}

func (m *mum) Unmarshall(ctx context.Context, req *http.Request, wsReq *ws.Request) error {
	return nil
}
