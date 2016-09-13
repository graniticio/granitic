package ws

import (
	"github.com/graniticio/granitic/config"
	"github.com/graniticio/granitic/facility/httpserver"
	"github.com/graniticio/granitic/instance"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/logging"
	"github.com/graniticio/granitic/ws"
	"github.com/graniticio/granitic/ws/handler"
)

const wsHttpStatusDeterminerComponentName = instance.FrameworkPrefix + "HttpStatusDeterminer"
const wsParamBinderComponentName = instance.FrameworkPrefix + "ParamBinder"
const wsFrameworkErrorGenerator = instance.FrameworkPrefix + "FrameworkErrorGenerator"
const wsHandlerDecoratorName = instance.FrameworkPrefix + "WsHandlerDecorator"

func OfferAbnormalStatusWriter(arw ws.AbnormalStatusWriter, cc *ioc.ComponentContainer, name string) {

	if !cc.ModifierExists(httpserver.HttpServerComponentName, httpserver.HttpServerAbnormalStatusFieldName) {
		//The HTTP server does not have an AbnormalStatusWriter defined
		cc.AddModifier(httpserver.HttpServerComponentName, httpserver.HttpServerAbnormalStatusFieldName, name)
	}
}

func BuildAndRegisterWsCommon(lm *logging.ComponentLoggerManager, ca *config.ConfigAccessor, cn *ioc.ComponentContainer) *WsCommon {

	scd := new(ws.DefaultHttpStatusCodeDeterminer)
	cn.WrapAndAddProto(wsHttpStatusDeterminerComponentName, scd)

	pb := new(ws.ParamBinder)
	cn.WrapAndAddProto(wsParamBinderComponentName, pb)

	feg := new(ws.FrameworkErrorGenerator)
	ca.Populate("FrameworkServiceErrors", feg)
	cn.WrapAndAddProto(wsFrameworkErrorGenerator, feg)

	pb.FrameworkErrors = feg

	return NewWsCommon(pb, feg, scd)

}

func NewWsCommon(pb *ws.ParamBinder, feg *ws.FrameworkErrorGenerator, sd *ws.DefaultHttpStatusCodeDeterminer) *WsCommon {

	wc := new(WsCommon)
	wc.ParamBinder = pb
	wc.FrameworkErrors = feg
	wc.StatusDeterminer = sd

	return wc

}

type WsCommon struct {
	ParamBinder      *ws.ParamBinder
	FrameworkErrors  *ws.FrameworkErrorGenerator
	StatusDeterminer *ws.DefaultHttpStatusCodeDeterminer
}

func BuildRegisterWsDecorator(cc *ioc.ComponentContainer, rw ws.WsResponseWriter, um ws.WsUnmarshaller, wc *WsCommon, lm *logging.ComponentLoggerManager) {

	decoratorLogger := lm.CreateLogger(wsHandlerDecoratorName)
	decorator := WsHandlerDecorator{decoratorLogger, rw, um, wc.ParamBinder, wc.FrameworkErrors}
	cc.WrapAndAddProto(wsHandlerDecoratorName, &decorator)
}

type WsHandlerDecorator struct {
	FrameworkLogger logging.Logger
	ResponseWriter  ws.WsResponseWriter
	Unmarshaller    ws.WsUnmarshaller
	QueryBinder     *ws.ParamBinder
	FrameworkErrors *ws.FrameworkErrorGenerator
}

func (jwhd *WsHandlerDecorator) OfInterest(component *ioc.Component) bool {
	switch h := component.Instance.(type) {
	default:
		jwhd.FrameworkLogger.LogTracef("No interest %s", component.Name)
		return false
	case *handler.WsHandler:
		return h.AutoWireable()
	}
}

func (jwhd *WsHandlerDecorator) DecorateComponent(component *ioc.Component, container *ioc.ComponentContainer) {
	h := component.Instance.(*handler.WsHandler)
	l := jwhd.FrameworkLogger
	l.LogTracef("Decorating component %s", component.Name)

	if h.ResponseWriter == nil {
		h.ResponseWriter = jwhd.ResponseWriter
	}

	if h.Unmarshaller == nil {
		l.LogTracef("%s needs Unmarshaller", component.Name)
		h.Unmarshaller = jwhd.Unmarshaller
	}

	if h.ParamBinder == nil {
		h.ParamBinder = jwhd.QueryBinder
	}

	if h.FrameworkErrors == nil {
		h.FrameworkErrors = jwhd.FrameworkErrors
	}

}
