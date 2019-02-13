package ctl

import (
	"context"
	"github.com/graniticio/granitic/v2/types"
	"github.com/graniticio/granitic/v2/ws"
	"testing"
)

func TestLogicExecution(t *testing.T) {

	m := createManager()

	err := m.Register(new(mockCommand))

	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}

	cl := new(CommandLogic)
	cl.FrameworkLogger = m.FrameworkLogger

	cl.CommandManager = m

	cr := cl.UnmarshallTarget().(*ctlCommandRequest)

	cr.Command = types.NewNilableString("mock")
	cr.Qualifiers = []string{}
	cr.Arguments = make(map[string]string)

	req := new(ws.Request)

	req.RequestBody = cr

	res := new(ws.Response)
	res.Errors = new(ws.ServiceErrors)

	ctx := context.Background()

	cl.Process(ctx, req, res)

	if len(res.Errors.Errors) > 0 {
		t.Fatalf("Unexpected errors %v", res.Errors.Errors)
	}

}
