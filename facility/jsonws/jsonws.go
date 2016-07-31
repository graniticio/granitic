package jsonws

import (
	"github.com/graniticio/granitic/config"
	"github.com/graniticio/granitic/facility/serviceerror"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/logging"
	"github.com/graniticio/granitic/ws"
	"github.com/graniticio/granitic/ws/json"
)

const jsonResponseWriterComponentName = ioc.FrameworkPrefix + "JsonResponseWriter"
const jsonUnmarshallerComponentName = ioc.FrameworkPrefix + "JsonUnmarshaller"
const jsonAbnormalResponseWriterComponentName = ioc.FrameworkPrefix + "JsonAbnormalResponseWriter"
const jsonHandlerDecoratorComponentName = ioc.FrameworkPrefix + "JsonHandlerDecorator"
const wsHttpStatusDeterminerComponentName = ioc.FrameworkPrefix + "HttpStatusDeterminer"
const wsQueryBinderComponentName = ioc.FrameworkPrefix + "QueryBinder"
const wsFrameworkErrorGenerator = ioc.FrameworkPrefix + "FrameworkErrorGenerator"

type JsonWsFacilityBuilder struct {
}

func (fb *JsonWsFacilityBuilder) BuildAndRegister(lm *logging.ComponentLoggerManager, ca *config.ConfigAccessor, cn *ioc.ComponentContainer) error {

	responseWriter := new(json.DefaultJsonResponseWriter)
	cn.WrapAndAddProto(jsonResponseWriterComponentName, responseWriter)

	abnormalResponseWriter := new(json.DefaultAbnormalResponseWriter)
	cn.WrapAndAddProto(jsonAbnormalResponseWriterComponentName, abnormalResponseWriter)

	queryBinder := new(ws.ParamBinder)
	cn.WrapAndAddProto(wsQueryBinderComponentName, queryBinder)

	statusDeterminer := new(ws.DefaultHttpStatusCodeDeterminer)
	cn.WrapAndAddProto(wsHttpStatusDeterminerComponentName, statusDeterminer)

	jsonUnmarshaller := new(json.DefaultJsonUnmarshaller)
	cn.WrapAndAddProto(jsonUnmarshallerComponentName, jsonUnmarshaller)

	frameworkErrors := new(serviceerror.FrameworkErrorGenerator)
	ca.Populate("FrameworkServiceErrors", frameworkErrors)
	cn.WrapAndAddProto(wsFrameworkErrorGenerator, frameworkErrors)

	decoratorLogger := lm.CreateLogger(jsonHandlerDecoratorComponentName)
	decorator := JsonWsHandlerDecorator{decoratorLogger, responseWriter, abnormalResponseWriter, statusDeterminer, jsonUnmarshaller, queryBinder, frameworkErrors}
	cn.WrapAndAddProto(jsonHandlerDecoratorComponentName, &decorator)

	return nil
}

func (fb *JsonWsFacilityBuilder) FacilityName() string {
	return "JsonWs"
}

func (fb *JsonWsFacilityBuilder) DependsOnFacilities() []string {
	return []string{}
}

type JsonWsHandlerDecorator struct {
	FrameworkLogger      logging.Logger
	ResponseWriter       ws.WsResponseWriter
	ErrorResponseWriter  ws.WsAbnormalResponseWriter
	StatusCodeDeterminer ws.HttpStatusCodeDeterminer
	Unmarshaller         ws.WsUnmarshaller
	QueryBinder          *ws.ParamBinder
	FrameworkErrors      *serviceerror.FrameworkErrorGenerator
}

func (jwhd *JsonWsHandlerDecorator) OfInterest(component *ioc.Component) bool {
	switch component.Instance.(type) {
	default:
		jwhd.FrameworkLogger.LogTracef("No interest %s", component.Name)
		return false
	case *ws.WsHandler:
		return true
	}
}

func (jwhd *JsonWsHandlerDecorator) DecorateComponent(component *ioc.Component, container *ioc.ComponentContainer) {
	h := component.Instance.(*ws.WsHandler)
	l := jwhd.FrameworkLogger
	l.LogTracef("Decorating component %s", component.Name)

	if h.StatusDeterminer == nil {
		h.StatusDeterminer = jwhd.StatusCodeDeterminer
	}

	if h.ResponseWriter == nil {
		h.ResponseWriter = jwhd.ResponseWriter
	}

	if h.ErrorResponseWriter == nil {
		h.ErrorResponseWriter = jwhd.ErrorResponseWriter
	}

	if h.Unmarshaller == nil {
		l.LogTracef("%s needs Unmarshaller", component.Name)
		h.Unmarshaller = jwhd.Unmarshaller
	}

	if h.ParamBinder == nil {
		h.ParamBinder = jwhd.QueryBinder
	}

}
