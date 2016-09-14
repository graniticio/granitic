package runtimectl

import (
	"github.com/graniticio/granitic/config"
	"github.com/graniticio/granitic/ctl"
	"github.com/graniticio/granitic/facility/httpserver"
	"github.com/graniticio/granitic/facility/serviceerror"
	"github.com/graniticio/granitic/httpendpoint"
	"github.com/graniticio/granitic/instance"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/logging"
	"github.com/graniticio/granitic/validate"
	"github.com/graniticio/granitic/ws"
	"github.com/graniticio/granitic/ws/handler"
	"github.com/graniticio/granitic/ws/json"
)

const (
	runtimeCtlServer           = instance.FrameworkPrefix + "CtlServer"
	runtimeCtlResponseWriter   = instance.FrameworkPrefix + "CtlResponseWriter"
	runtimeCtlFrameworkErrors  = instance.FrameworkPrefix + "CtlFrameworkErrors"
	runtimeCtlCommandHandler   = instance.FrameworkPrefix + "CtlCommandHandler"
	runtimeCtlUnmarshaller     = instance.FrameworkPrefix + "CtlUnmarshaller"
	runtimeCtlValidator        = instance.FrameworkPrefix + "CtlValidator"
	runtimeCtlServiceErrors    = instance.FrameworkPrefix + "CtlServiceErrors"
	runtimeCtlCommandDecorator = instance.FrameworkPrefix + "CtlCommandDecorator"
	runtimeCtlCommandManager   = instance.FrameworkPrefix + "CtlCommandManager"
	shutdownCommand            = instance.FrameworkPrefix + "CommandShutdown"
	defaultValidationCode      = "INV_CTL_REQUEST"
)

type RuntimeCtlFacilityBuilder struct {
}

func (fb *RuntimeCtlFacilityBuilder) BuildAndRegister(lm *logging.ComponentLoggerManager, ca *config.ConfigAccessor, cc *ioc.ComponentContainer) error {

	sv := new(httpserver.HTTPServer)
	ca.Populate("RuntimeCtl.Server", sv)

	cc.WrapAndAddProto(runtimeCtlServer, sv)

	rw := new(ws.MarshallingResponseWriter)
	ca.Populate("RuntimeCtl.ResponseWriter", rw)
	rw.FrameworkLogger = lm.CreateLogger(runtimeCtlResponseWriter)
	sv.AbnormalStatusWriter = rw

	mw := new(json.JSONMarshalingWriter)
	ca.Populate("RuntimeCtl.Marshal", mw)
	rw.MarshalingWriter = mw

	wr := new(json.StandardJSONResponseWrapper)
	ca.Populate("RuntimeCtl.ResponseWrapper", wr)
	rw.ResponseWrapper = wr

	rw.ErrorFormatter = new(json.StandardJSONErrorFormatter)

	rw.StatusDeterminer = new(ws.DefaultHttpStatusCodeDeterminer)

	feg := new(ws.FrameworkErrorGenerator)
	feg.FrameworkLogger = lm.CreateLogger(runtimeCtlFrameworkErrors)

	ca.Populate("FrameworkServiceErrors", feg)

	rw.FrameworkErrors = feg

	//Handler
	h := new(handler.WsHandler)
	h.PreventAutoWiring = true
	ca.Populate("RuntimeCtl.CommandHandler", h)
	h.Log = lm.CreateLogger(runtimeCtlCommandHandler)
	h.Logic = new(ctl.CommandLogic)
	h.DisablePathParsing = true
	h.DisableQueryParsing = true
	h.ResponseWriter = rw
	h.FrameworkErrors = feg

	um := new(json.StandardJSONUnmarshaller)
	um.FrameworkLogger = lm.CreateLogger(runtimeCtlUnmarshaller)

	h.Unmarshaller = um

	handlers := make(map[string]httpendpoint.HttpEndpointProvider)
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
	sem := new(serviceerror.ServiceErrorManager)
	sem.PanicOnMissing = true

	e := new(errors)
	ca.SetField("Unparsed", "RuntimeCtl.Errors", e)

	sem.LoadErrors(e.Unparsed)

	cc.WrapAndAddProto(runtimeCtlServiceErrors, sem)

	h.ErrorFinder = sem

	//Command manager
	cm := new(ctl.CommandManager)
	cm.FrameworkLogger = lm.CreateLogger(runtimeCtlCommandManager)
	cc.WrapAndAddProto(runtimeCtlCommandManager, cm)

	//Command decorator
	cd := new(ctl.CommandDecorator)
	cd.CommandManager = cm
	cd.FrameworkLogger = lm.CreateLogger(runtimeCtlCommandDecorator)
	cc.WrapAndAddProto(runtimeCtlCommandDecorator, cd)

	fb.createBuiltinCommands(lm, cc)

	return nil
}

func (fb *RuntimeCtlFacilityBuilder) createBuiltinCommands(lm *logging.ComponentLoggerManager, cc *ioc.ComponentContainer) {

	sd := new(ShutdownCommand)
	cc.WrapAndAddProto(shutdownCommand, sd)

}

func (fb *RuntimeCtlFacilityBuilder) FacilityName() string {
	return "RuntimeCtl"
}

func (fb *RuntimeCtlFacilityBuilder) DependsOnFacilities() []string {
	return []string{}
}

type errors struct {
	Unparsed []interface{}
}
