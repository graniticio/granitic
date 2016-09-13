package runtimectl

import (
	"github.com/graniticio/granitic/config"
	"github.com/graniticio/granitic/ctl"
	"github.com/graniticio/granitic/facility/httpserver"
	"github.com/graniticio/granitic/httpendpoint"
	"github.com/graniticio/granitic/instance"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/logging"
	"github.com/graniticio/granitic/ws"
	"github.com/graniticio/granitic/ws/handler"
	"github.com/graniticio/granitic/ws/json"
)

const (
	runtimeCtlServerName          = instance.FrameworkPrefix + "CtlServer"
	runtimeCtlResponseWriterName  = instance.FrameworkPrefix + "CtlResponseWriter"
	runtimeCtlFrameworkErrorsName = instance.FrameworkPrefix + "CtlFrameworkErrors"
	runtimeCtlCommandHandlerName  = instance.FrameworkPrefix + "CtlCommandHandler"
	runtimeCtlUnmarshallerName    = instance.FrameworkPrefix + "CtlUnmarshaller"
)

type RuntimeCtlFacilityBuilder struct {
}

func (fb *RuntimeCtlFacilityBuilder) BuildAndRegister(lm *logging.ComponentLoggerManager, ca *config.ConfigAccessor, cn *ioc.ComponentContainer) error {

	sv := new(httpserver.HTTPServer)
	ca.Populate("RuntimeCtl.Server", sv)

	cn.WrapAndAddProto(runtimeCtlServerName, sv)

	rw := new(ws.MarshallingResponseWriter)
	ca.Populate("RuntimeCtl.ResponseWriter", rw)
	rw.FrameworkLogger = lm.CreateLogger(runtimeCtlResponseWriterName)
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
	feg.FrameworkLogger = lm.CreateLogger(runtimeCtlFrameworkErrorsName)

	ca.Populate("FrameworkServiceErrors", feg)

	rw.FrameworkErrors = feg

	h := new(handler.WsHandler)
	h.PreventAutoWiring = true
	ca.Populate("RuntimeCtl.CommandHandler", h)
	h.Log = lm.CreateLogger(runtimeCtlCommandHandlerName)
	h.Logic = new(ctl.CommandLogic)
	h.DisablePathParsing = true
	h.DisableQueryParsing = true
	h.ResponseWriter = rw
	h.FrameworkErrors = feg

	um := new(json.StandardJSONUnmarshaller)
	um.FrameworkLogger = lm.CreateLogger(runtimeCtlUnmarshallerName)

	h.Unmarshaller = um

	handlers := make(map[string]httpendpoint.HttpEndpointProvider)
	handlers[runtimeCtlCommandHandlerName] = h
	sv.SetProvidersManually(handlers)

	cn.WrapAndAddProto(runtimeCtlCommandHandlerName, h)

	return nil
}

func (fb *RuntimeCtlFacilityBuilder) FacilityName() string {
	return "RuntimeCtl"
}

func (fb *RuntimeCtlFacilityBuilder) DependsOnFacilities() []string {
	return []string{}
}
