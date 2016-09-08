/*
Package jsonws builds the components required to support JSON-based web-services.
*/
package ws

import (
	"github.com/graniticio/granitic/config"
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

type JSONWsFacilityBuilder struct {
}

func (fb *JSONWsFacilityBuilder) BuildAndRegister(lm *logging.ComponentLoggerManager, ca *config.ConfigAccessor, cn *ioc.ComponentContainer) error {

	wc := BuildAndRegisterWsCommon(lm, ca, cn)

	responseWriter := new(json.StandardJSONResponseWriter)
	ca.Populate("JsonWs.ResponseWriter", responseWriter)
	cn.WrapAndAddProto(jsonResponseWriterComponentName, responseWriter)

	responseWriter.StatusDeterminer = wc.StatusDeterminer

	jsonUnmarshaller := new(json.StandardJSONUnmarshaller)
	cn.WrapAndAddProto(jsonUnmarshallerComponentName, jsonUnmarshaller)

	responseWriter.FrameworkErrors = wc.FrameworkErrors

	decoratorLogger := lm.CreateLogger(jsonHandlerDecoratorComponentName)
	decorator := JSONWsHandlerDecorator{decoratorLogger, responseWriter, jsonUnmarshaller, wc.ParamBinder, wc.FrameworkErrors}
	cn.WrapAndAddProto(jsonHandlerDecoratorComponentName, &decorator)

	if !cn.ModifierExists(jsonResponseWriterComponentName, "ErrorFormatter") {
		responseWriter.ErrorFormatter = new(json.StandardJSONErrorFormatter)
	}

	if !cn.ModifierExists(jsonResponseWriterComponentName, "ResponseWrapper") {
		wrap := new(json.StandardJSONResponseWrapper)
		ca.Populate("JsonWs.ResponseWrapper", wrap)
		responseWriter.ResponseWrapper = wrap
	}

	OfferAbnormalStatusWriter(responseWriter, cn)

	return nil
}

func (fb *JSONWsFacilityBuilder) FacilityName() string {
	return "JsonWs"
}

func (fb *JSONWsFacilityBuilder) DependsOnFacilities() []string {
	return []string{}
}

type JSONWsHandlerDecorator struct {
	FrameworkLogger logging.Logger
	ResponseWriter  ws.WsResponseWriter
	Unmarshaller    ws.WsUnmarshaller
	QueryBinder     *ws.ParamBinder
	FrameworkErrors *ws.FrameworkErrorGenerator
}

func (jwhd *JSONWsHandlerDecorator) OfInterest(component *ioc.Component) bool {
	switch component.Instance.(type) {
	default:
		jwhd.FrameworkLogger.LogTracef("No interest %s", component.Name)
		return false
	case *handler.WsHandler:
		return true
	}
}

func (jwhd *JSONWsHandlerDecorator) DecorateComponent(component *ioc.Component, container *ioc.ComponentContainer) {
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
