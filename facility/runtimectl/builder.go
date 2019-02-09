// Copyright 2016-2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

/*
	Package runtimectl provides the RuntimeCtl facility which allows external runtime control of Granitic applications.

	This facility is described in detail at http://granitic.io/1.0/ref/runtime-ctl. Refer to the ctl package documentation
	for information on how to implement your own commands.

	Enabling runtime control

	Enabling the RuntimeCtl facility creates an HTTP server that allows instructions to be issued to
	any component in the IoC container which implements the ctl.Command interface from the grnc-ctl command line tool.
	See http://granitic.io/1.0/ref/runtime-ctl#builtin for documentation on Granitic's built-in commands.

	The HTTP server that listens for commands is separate to the HTTP server created by the XMLWs and JSONWs facilities and runs on a
	different port. The listen port defaults to 9099 but can be changed with the following configuration:

		{
		  "RuntimeCtl": {
			"Server":{
			  "Port": 9099,
			  "Address": "127.0.0.1"
			}
		  }
		}

	Note that by default the server only listens on the IPV4 localhost. To listen on all interfaces, change address to ""

	Disabling individual commands

	You can disable individual commands (either builtin commands or your own application commands) with configuration. For
	example:

	{
	  "RuntimeCtl": {
	    "Manager":{
	      "Disabled": ["shutdown"]
	    }
	  }
	}

	Disables the shutdown command, preventing your application being stopped remotely.

*/
package runtimectl

import (
	"github.com/graniticio/granitic/v2/config"
	"github.com/graniticio/granitic/v2/ctl"
	"github.com/graniticio/granitic/v2/facility/httpserver"
	ge "github.com/graniticio/granitic/v2/grncerror"
	"github.com/graniticio/granitic/v2/httpendpoint"
	"github.com/graniticio/granitic/v2/instance"
	"github.com/graniticio/granitic/v2/ioc"
	"github.com/graniticio/granitic/v2/logging"
	"github.com/graniticio/granitic/v2/types"
	"github.com/graniticio/granitic/v2/validate"
	"github.com/graniticio/granitic/v2/ws"
	"github.com/graniticio/granitic/v2/ws/handler"
	"github.com/graniticio/granitic/v2/ws/json"
)

const (
	RuntimeCtlServer           = instance.FrameworkPrefix + "CtlServer"
	runtimeCtlLogic            = instance.FrameworkPrefix + "CtlLogic"
	runtimeCtlResponseWriter   = instance.FrameworkPrefix + "CtlResponseWriter"
	runtimeCtlFrameworkErrors  = instance.FrameworkPrefix + "CtlFrameworkErrors"
	runtimeCtlCommandHandler   = instance.FrameworkPrefix + "CtlCommandHandler"
	runtimeCtlUnmarshaller     = instance.FrameworkPrefix + "CtlUnmarshaller"
	runtimeCtlValidator        = instance.FrameworkPrefix + "CtlValidator"
	runtimeCtlServiceErrors    = instance.FrameworkPrefix + "CtlServiceErrors"
	runtimeCtlCommandDecorator = instance.FrameworkPrefix + "CtlCommandDecorator"
	runtimeCtlCommandManager   = instance.FrameworkPrefix + "CtlCommandManager"
	shutdownCommandComp        = instance.FrameworkPrefix + "CommandShutdown"
	helpCommandComp            = instance.FrameworkPrefix + "CommandHelp"
	componentsCommandComp      = instance.FrameworkPrefix + "CommandComponents"
	stopCommandComp            = instance.FrameworkPrefix + "CommandStop"
	suspendCommandComp         = instance.FrameworkPrefix + "CommandSuspend"
	resumeCommandComp          = instance.FrameworkPrefix + "CommandResume"
	defaultValidationCode      = "INV_CTL_REQUEST"
)

type FacilityBuilder struct {
}

func (fb *FacilityBuilder) BuildAndRegister(lm *logging.ComponentLoggerManager, ca *config.Accessor, cc *ioc.ComponentContainer) error {

	sv := new(httpserver.HTTPServer)
	ca.Populate("RuntimeCtl.Server", sv)

	cc.WrapAndAddProto(RuntimeCtlServer, sv)

	rw := new(ws.MarshallingResponseWriter)
	ca.Populate("RuntimeCtl.ResponseWriter", rw)
	rw.FrameworkLogger = lm.CreateLogger(runtimeCtlResponseWriter)
	sv.AbnormalStatusWriter = rw

	mw := new(json.MarshalingWriter)
	ca.Populate("RuntimeCtl.Marshal", mw)
	rw.MarshalingWriter = mw

	wr := new(json.GraniticJSONResponseWrapper)
	ca.Populate("RuntimeCtl.ResponseWrapper", wr)
	rw.ResponseWrapper = wr

	rw.ErrorFormatter = new(json.GraniticJSONErrorFormatter)

	rw.StatusDeterminer = new(ws.GraniticHTTPStatusCodeDeterminer)

	feg := new(ws.FrameworkErrorGenerator)
	feg.FrameworkLogger = lm.CreateLogger(runtimeCtlFrameworkErrors)

	ca.Populate("FrameworkServiceErrors", feg)

	rw.FrameworkErrors = feg

	//Handler
	h := new(handler.WsHandler)
	h.PreventAutoWiring = true
	ca.Populate("RuntimeCtl.CommandHandler", h)
	h.Log = lm.CreateLogger(runtimeCtlCommandHandler)
	h.DisablePathParsing = true
	h.DisableQueryParsing = true
	h.ResponseWriter = rw
	h.FrameworkErrors = feg

	um := new(json.Unmarshaller)
	um.FrameworkLogger = lm.CreateLogger(runtimeCtlUnmarshaller)

	h.Unmarshaller = um

	handlers := make(map[string]httpendpoint.Provider)
	handlers[runtimeCtlCommandHandler] = h
	sv.SetProvidersManually(handlers)

	cc.WrapAndAddProto(runtimeCtlCommandHandler, h)

	//Validator
	v := new(validate.RuleValidator)
	v.ComponentFinder = cc
	v.DefaultErrorCode = defaultValidationCode
	v.Log = lm.CreateLogger(runtimeCtlValidator)
	v.DisableCodeValidation = true

	ca.SetField("Rules", "RuntimeCtl.CommandValidation", v)

	cc.WrapAndAddProto(runtimeCtlValidator, v)

	h.AutoValidator = v

	rm := new(validate.UnparsedRuleManager)
	ca.SetField("Rules", "RuntimeCtl.SharedRules", rm)

	v.RuleManager = rm

	//Error finder
	sem := new(ge.ServiceErrorManager)
	sem.PanicOnMissing = true

	e := new(errorsWrapper)
	ca.SetField("Unparsed", "RuntimeCtl.Errors", e)

	sem.LoadErrors(e.Unparsed)

	cc.WrapAndAddProto(runtimeCtlServiceErrors, sem)

	h.ErrorFinder = sem

	//Command manager
	cm := new(ctl.CommandManager)

	ca.Populate("RuntimeCtl.Manager", cm)

	if cm.Disabled == nil {
		cm.DisabledLookup = types.NewEmptyOrderedStringSet()
	} else {
		cm.DisabledLookup = types.NewOrderedStringSet(cm.Disabled)
	}

	cm.FrameworkLogger = lm.CreateLogger(runtimeCtlCommandManager)
	cc.WrapAndAddProto(runtimeCtlCommandManager, cm)

	//Command decorator
	cd := new(ctl.CommandDecorator)
	cd.CommandManager = cm
	cd.FrameworkLogger = lm.CreateLogger(runtimeCtlCommandDecorator)
	cc.WrapAndAddProto(runtimeCtlCommandDecorator, cd)

	fb.createBuiltinCommands(lm, cc, cm)

	//Command logic
	cl := new(ctl.CommandLogic)
	cl.FrameworkLogger = lm.CreateLogger(runtimeCtlLogic)
	cl.CommandManager = cm
	h.Logic = cl

	return nil
}

func (fb *FacilityBuilder) createBuiltinCommands(lm *logging.ComponentLoggerManager, cc *ioc.ComponentContainer, cm *ctl.CommandManager) {

	sd := new(shutdownCommand)
	fb.addCommand(cc, shutdownCommandComp, sd)

	hc := new(helpCommand)
	hc.commandManager = cm
	fb.addCommand(cc, helpCommandComp, hc)

	cs := new(componentsCommand)
	fb.addCommand(cc, componentsCommandComp, cs)

	stopc := NewStopCommand()
	fb.addCommand(cc, stopCommandComp, stopc)

	startc := NewStartCommand()
	fb.addCommand(cc, startCommandName, startc)

	suspendc := NewSuspendCommand()
	fb.addCommand(cc, suspendCommandName, suspendc)

	resumec := NewResumeCommand()
	fb.addCommand(cc, resumeCommandName, resumec)

}

func (fb *FacilityBuilder) addCommand(cc *ioc.ComponentContainer, name string, c ctl.Command) {
	cc.WrapAndAddProto(name, c)
}

func (fb *FacilityBuilder) FacilityName() string {
	return "RuntimeCtl"
}

func (fb *FacilityBuilder) DependsOnFacilities() []string {
	return []string{}
}

type errorsWrapper struct {
	Unparsed []interface{}
}
