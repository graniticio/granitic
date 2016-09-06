/*
Package jsonws builds the components required to support JSON-based web-services.
*/
package jsonws

import (
	"github.com/graniticio/granitic/config"
	"github.com/graniticio/granitic/facility/httpserver"
	"github.com/graniticio/granitic/instance"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/logging"
	"github.com/graniticio/granitic/ws"
	"github.com/graniticio/granitic/ws/handler"
	"github.com/graniticio/granitic/ws/json"
)

const jsonResponseWriterComponentName = instance.FrameworkPrefix + "JsonResponseWriter"
const jsonUnmarshallerComponentName = instance.FrameworkPrefix + "JsonUnmarshaller"
const jsonHandlerDecoratorComponentName = instance.FrameworkPrefix + "JsonHandlerDecorator"
const wsHttpStatusDeterminerComponentName = instance.FrameworkPrefix + "HttpStatusDeterminer"
const wsQueryBinderComponentName = instance.FrameworkPrefix + "QueryBinder"
const wsFrameworkErrorGenerator = instance.FrameworkPrefix + "FrameworkErrorGenerator"

type JsonWsFacilityBuilder struct {
}

func (fb *JsonWsFacilityBuilder) BuildAndRegister(lm *logging.ComponentLoggerManager, ca *config.ConfigAccessor, cn *ioc.ComponentContainer) error {

	responseWriter := new(json.StandardJSONResponseWriter)
	ca.Populate("JsonWs.ResponseWriter", responseWriter)
	cn.WrapAndAddProto(jsonResponseWriterComponentName, responseWriter)

	queryBinder := new(ws.ParamBinder)
	cn.WrapAndAddProto(wsQueryBinderComponentName, queryBinder)

	statusDeterminer := new(ws.DefaultHttpStatusCodeDeterminer)
	cn.WrapAndAddProto(wsHttpStatusDeterminerComponentName, statusDeterminer)
	responseWriter.StatusDeterminer = statusDeterminer

	jsonUnmarshaller := new(json.StandardJSONUnmarshaller)
	cn.WrapAndAddProto(jsonUnmarshallerComponentName, jsonUnmarshaller)

	frameworkErrors := new(ws.FrameworkErrorGenerator)
	ca.Populate("FrameworkServiceErrors", frameworkErrors)
	cn.WrapAndAddProto(wsFrameworkErrorGenerator, frameworkErrors)

	queryBinder.FrameworkErrors = frameworkErrors
	responseWriter.FrameworkErrors = frameworkErrors

	decoratorLogger := lm.CreateLogger(jsonHandlerDecoratorComponentName)
	decorator := JsonWsHandlerDecorator{decoratorLogger, responseWriter, jsonUnmarshaller, queryBinder, frameworkErrors}
	cn.WrapAndAddProto(jsonHandlerDecoratorComponentName, &decorator)

	if !cn.ModifierExists(jsonResponseWriterComponentName, "ErrorFormatter") {
		responseWriter.ErrorFormatter = new(json.StandardJSONErrorFormatter)
	}

	if !cn.ModifierExists(jsonResponseWriterComponentName, "ResponseWrapper") {
		wrap := new(json.StandardJSONResponseWrapper)
		ca.Populate("JsonWs.ResponseWrapper", wrap)
		responseWriter.ResponseWrapper = wrap
	}

	if !cn.ModifierExists(httpserver.HttpServerComponentName, httpserver.HttpServerAbnormalStatusFieldName) {
		//The HTTP server does not have an AbnormalStatusWriter defined
		cn.AddModifier(httpserver.HttpServerComponentName, httpserver.HttpServerAbnormalStatusFieldName, jsonResponseWriterComponentName)
	}

	return nil
}

func (fb *JsonWsFacilityBuilder) FacilityName() string {
	return "JsonWs"
}

func (fb *JsonWsFacilityBuilder) DependsOnFacilities() []string {
	return []string{}
}

type JsonWsHandlerDecorator struct {
	FrameworkLogger logging.Logger
	ResponseWriter  ws.WsResponseWriter
	Unmarshaller    ws.WsUnmarshaller
	QueryBinder     *ws.ParamBinder
	FrameworkErrors *ws.FrameworkErrorGenerator
}

func (jwhd *JsonWsHandlerDecorator) OfInterest(component *ioc.Component) bool {
	switch component.Instance.(type) {
	default:
		jwhd.FrameworkLogger.LogTracef("No interest %s", component.Name)
		return false
	case *handler.WsHandler:
		return true
	}
}

func (jwhd *JsonWsHandlerDecorator) DecorateComponent(component *ioc.Component, container *ioc.ComponentContainer) {
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
