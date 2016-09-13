package ctl

import (
	"github.com/graniticio/granitic/ws"
	"golang.org/x/net/context"
)

type CommandLogic struct {
}

func (cl *CommandLogic) Process(ctx context.Context, req *ws.WsRequest, res *ws.WsResponse) {

	res.Errors.AddNewError(ws.Client, "LABEL", "MESSAGE")

}

func (cl *CommandLogic) UnmarshallTarget() interface{} {
	return new(CtlCommand)
}

type CtlCommand struct {
	Command    string
	Qualifiers []string
	Arguments  map[string]string
}
