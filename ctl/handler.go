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

func (cl *CommandLogic) UnmarshallTarget() interface{} {
	return new(CtlCommand)
}

type CtlCommand struct {
	Command    *types.NilableString
	Qualifiers []*types.NilableString
	Arguments  map[string]string
}
