// Copyright 2016-2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

/*
	Package ws provides the JsonWs and XmlWs facilities which support JSON and XML web services.

	This facility is documented in detail at http://granitic.io/1.0/ref/web-services

	Web-services

	Enabling the JsonWs or XmlWs facility allows the creation of web service endpoints where inbound and outbound data is automatically converted from and to JSON/XML.

	An endpoint is created by adding an instance of handler.WsHandler with a corresponding implementation of handler.WsPostProcessor
	(generally referred to as handler logic) to your component definition file. For example:

		"createRecordLogic": {
		  "type": "inventory.CreateRecordLogic"
		},

		"createRecordHandler": {
		  "type": "handler.WsHandler",
		  "HTTPMethod": "POST",
		  "Logic": "ref:createRecordLogic",
		  "PathPattern": "^/record$",
		}

	See the ws and ws/handler package documentation for more information.


	JSON

	If JsonWs is enabled, any requests to a registered endpoint will have their request body parsed as JSON and any response rendered as JSON.
	This is handled with Go's built-in json package.

	Many aspects of the parsing and rendering process (including content types, formatting of errors, pretty-printing and
	camel-case mapping) is configurable. Refer to http://granitic.io/1.0/ref/json for more details.

	XML

	Once the XmlWs facility is enabled, requests to an endpoint will, by default, be parsed as XML and rendered using
	user defined templates. Refer to http://granitic.io/1.0/ref/xml#templates for more details.

	Alternatively, the XmlWs facility can be configured to automatically render responses as XML using Go's built-in
	XML marshalling components by setting the following configuration in your application's configuration files:

		{
		  "XmlWs": {
			"ResponseMode": "MARSHAL"
		  }
		}


	Refer to http://granitic.io/1.0/ref/xml#marshalled for more information.

	Many aspects of the parsing and rendering process (including content types and formatting of errors) is configurable.
	Refer to http://granitic.io/1.0/ref/xml for more details.
*/
package ws

import (
	"github.com/graniticio/granitic/v2/config"
	"github.com/graniticio/granitic/v2/facility/httpserver"
	"github.com/graniticio/granitic/v2/instance"
	"github.com/graniticio/granitic/v2/ioc"
	"github.com/graniticio/granitic/v2/logging"
	"github.com/graniticio/granitic/v2/ws"
	"github.com/graniticio/granitic/v2/ws/handler"
)

const wsHTTPStatusDeterminerComponentName = instance.FrameworkPrefix + "HTTPStatusDeterminer"
const wsParamBinderComponentName = instance.FrameworkPrefix + "ParamBinder"
const wsFrameworkErrorGenerator = instance.FrameworkPrefix + "FrameworkErrorGenerator"
const wsHandlerDecoratorName = instance.FrameworkPrefix + "WsHandlerDecorator"

func offerAbnormalStatusWriter(arw ws.AbnormalStatusWriter, cc *ioc.ComponentContainer, name string) {

	if !cc.ModifierExists(httpserver.HTTPServerComponentName, httpserver.HTTPServerAbnormalStatusFieldName) {
		//The HTTP server does not have an AbnormalStatusWriter defined
		cc.AddModifier(httpserver.HTTPServerComponentName, httpserver.HTTPServerAbnormalStatusFieldName, name)
	}
}

func buildAndRegisterWsCommon(lm *logging.ComponentLoggerManager, ca *config.Accessor, cn *ioc.ComponentContainer) *wsCommon {

	scd := new(ws.GraniticHTTPStatusCodeDeterminer)
	cn.WrapAndAddProto(wsHTTPStatusDeterminerComponentName, scd)

	pb := new(ws.ParamBinder)
	cn.WrapAndAddProto(wsParamBinderComponentName, pb)

	feg := new(ws.FrameworkErrorGenerator)
	ca.Populate("FrameworkServiceErrors", feg)
	cn.WrapAndAddProto(wsFrameworkErrorGenerator, feg)

	pb.FrameworkErrors = feg

	return newWsCommon(pb, feg, scd)

}

func newWsCommon(pb *ws.ParamBinder, feg *ws.FrameworkErrorGenerator, sd *ws.GraniticHTTPStatusCodeDeterminer) *wsCommon {

	wc := new(wsCommon)
	wc.ParamBinder = pb
	wc.FrameworkErrors = feg
	wc.StatusDeterminer = sd

	return wc

}

type wsCommon struct {
	ParamBinder      *ws.ParamBinder
	FrameworkErrors  *ws.FrameworkErrorGenerator
	StatusDeterminer *ws.GraniticHTTPStatusCodeDeterminer
}

func buildRegisterWsDecorator(cc *ioc.ComponentContainer, rw ws.WsResponseWriter, um ws.WsUnmarshaller, wc *wsCommon, lm *logging.ComponentLoggerManager) {

	decoratorLogger := lm.CreateLogger(wsHandlerDecoratorName)
	decorator := wsHandlerDecorator{decoratorLogger, rw, um, wc.ParamBinder, wc.FrameworkErrors}
	cc.WrapAndAddProto(wsHandlerDecoratorName, &decorator)
}

type wsHandlerDecorator struct {
	FrameworkLogger logging.Logger
	ResponseWriter  ws.WsResponseWriter
	Unmarshaller    ws.WsUnmarshaller
	QueryBinder     *ws.ParamBinder
	FrameworkErrors *ws.FrameworkErrorGenerator
}

func (jwhd *wsHandlerDecorator) OfInterest(component *ioc.Component) bool {
	switch h := component.Instance.(type) {
	default:
		jwhd.FrameworkLogger.LogTracef("No interest %s", component.Name)
		return false
	case *handler.WsHandler:
		return h.AutoWireable()
	}
}

func (jwhd *wsHandlerDecorator) DecorateComponent(component *ioc.Component, container *ioc.ComponentContainer) {
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
