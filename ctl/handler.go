package ctl

import (
	"fmt"
	"github.com/graniticio/granitic/types"
	"github.com/graniticio/granitic/ws"
	"golang.org/x/net/context"
	"regexp"
)

type CommandLogic struct {
}

func (cl *CommandLogic) Process(ctx context.Context, req *ws.WsRequest, res *ws.WsResponse) {

}

const (
	maxArgs      = 32
	tooManyArgs  = "TOO_MANY_ARGS"
	argKeyFormat = "ARG_PATTERN"
	argMaxLength = 256
	tooLongArg   = "ARG_TOO_LONG"
)

func (cl *CommandLogic) Validate(ctx context.Context, se *ws.ServiceErrors, request *ws.WsRequest) {

	sub := request.RequestBody.(*CtlCommand)

	cl.validateArgs(se, sub)

}

func (cl *CommandLogic) validateArgs(se *ws.ServiceErrors, sub *CtlCommand) {
	if sub.Arguments != nil {
		ac := len(sub.Arguments)

		keyRegex := regexp.MustCompile("^[a-zA-Z]{1}[\\w-]{0,19}$")

		field := "Arguments"

		if ac > maxArgs {
			se.AddPredefinedError(tooManyArgs, field)
		}

		for k, v := range sub.Arguments {

			fn := fmt.Sprintf("%s[%q]", field, k)

			if !keyRegex.MatchString(k) {
				se.AddPredefinedError(argKeyFormat, fn)
			}

			al := len(v)

			if al == 0 || al > argMaxLength {
				se.AddPredefinedError(tooLongArg, fn)
			}

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
