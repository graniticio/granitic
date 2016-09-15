package ctl

import (
	"fmt"
	"github.com/graniticio/granitic/logging"
	"github.com/graniticio/granitic/types"
	"github.com/graniticio/granitic/ws"
	"golang.org/x/net/context"
	"regexp"
	"strings"
)

const (
	maxArgs        = 32
	tooManyArgs    = "TOO_MANY_ARGS"
	unknownCommand = "UNKNOWN_COMMAND"
	argKeyFormat   = "ARG_PATTERN"
	argMaxLength   = 256
	tooLongArg     = "ARG_TOO_LONG"
)

type CommandLogic struct {
	FrameworkLogger logging.Logger
	CommandManager  *CommandManager
}

func (cl *CommandLogic) Process(ctx context.Context, req *ws.WsRequest, res *ws.WsResponse) {

	cr := req.RequestBody.(*ctlCommandRequest)

	name := cl.normaliseCommandName(cr)
	comm := cl.CommandManager.Find(name)

	cl.FrameworkLogger.LogInfof("Executing runtime command '%s'", name)

	co, errs := comm.ExecuteCommand(cr.Qualifiers, cr.Arguments)

	if errs != nil {
		for _, e := range errs {
			res.Errors.AddError(e)
		}
	}

	res.Body = co

}

func (cl *CommandLogic) Validate(ctx context.Context, se *ws.ServiceErrors, request *ws.WsRequest) {

	sub := request.RequestBody.(*ctlCommandRequest)

	cl.validateArgs(se, sub)

	if se.HasErrors() {
		return
	}

	name := cl.normaliseCommandName(sub)
	comm := cl.CommandManager.Find(name)

	if comm == nil {
		se.AddPredefinedError(unknownCommand, "Command")
	}

}

func (cl *CommandLogic) validateArgs(se *ws.ServiceErrors, sub *ctlCommandRequest) {
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

func (cl *CommandLogic) normaliseCommandName(cr *ctlCommandRequest) string {
	return strings.ToLower(cr.Command.String())
}

func (cl *CommandLogic) UnmarshallTarget() interface{} {
	return new(ctlCommandRequest)
}

type ctlCommandRequest struct {
	Command    *types.NilableString
	Qualifiers []string
	Arguments  map[string]string
}
