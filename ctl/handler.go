package ctl

import (
	"github.com/graniticio/granitic/types"
	"github.com/graniticio/granitic/ws"
	"golang.org/x/net/context"
)

type CommandLogic struct {
}

func (cl *CommandLogic) Process(ctx context.Context, req *ws.WsRequest, res *ws.WsResponse) {

}

const (
	maxArgs     = 32
	tooManyArgs = "TOO_MANY_ARGS"
)

func (cl *CommandLogic) Validate(ctx context.Context, se *ws.ServiceErrors, request *ws.WsRequest) {

	sub := request.RequestBody.(*CtlCommand)

	cl.validateArgs(se, sub)
	cl.validateArgs(se, sub)

}

func (cl *CommandLogic) validateArgs(se *ws.ServiceErrors, sub *CtlCommand) {
	if sub.Arguments != nil {
		ac := len(sub.Arguments)

		if ac > maxArgs {
			se.AddPredefinedError(tooManyArgs)
		}
	}
}

func (cl *CommandLogic) UnmarshallTarget() interface{} {
	return new(CtlCommand)
}

type CtlCommand struct {
	Command    *types.NilableString
	Qualifiers []*types.NilableString
	Arguments  map[string]string
}
