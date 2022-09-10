// Copyright 2016-2022 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package ctl

import (
	"context"
	"fmt"
	"github.com/graniticio/granitic/v3/logging"
	"github.com/graniticio/granitic/v3/types"
	"github.com/graniticio/granitic/v3/ws"
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

// CommandLogic handles the validation of an HTTP call from grnc-ctl, matching that call to a ctl.Command and invoking that command.
// An instance of CommandLogic is automatically created as part of the RuntimeCtl facility and it is not required (or recommended) that
// user applications create components of this type.
type CommandLogic struct {
	// Logger used by Granitic framework components. Automatically injected.
	FrameworkLogger logging.Logger

	// An instance of CommandManager that contains all known commands.
	CommandManager *CommandManager
}

// Process implements handler.WsRequestProcessor.Process Whenever a valid request is handled, the name of the invoked command is logged at INFO level
// to assist with auditing.
func (cl *CommandLogic) Process(ctx context.Context, req *ws.Request, res *ws.Response) {

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

// Validate implements handler.WsRequestValidator.Validate
func (cl *CommandLogic) Validate(ctx context.Context, se *ws.ServiceErrors, request *ws.Request) {

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

// UnmarshallTarget Implements handler.WsUnmarshallTarget.UnmarshallTarget
func (cl *CommandLogic) UnmarshallTarget() interface{} {
	return new(ctlCommandRequest)
}

type ctlCommandRequest struct {
	Command    *types.NilableString
	Qualifiers []string
	Arguments  map[string]string
}
