// Copyright 2016-2018 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package httpserver

import (
	"github.com/graniticio/granitic/config"
	"github.com/graniticio/granitic/instance"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/logging"
)

// The name of the HttpServer component as stored in the IoC framework.
const HttpServerComponentName = instance.FrameworkPrefix + "HttpServer"
const contextIdDecoratorName = instance.FrameworkPrefix + "RequestIdContextDecorator"

// The field on the HttpServer component into which a ws.AbnormalStatusWriter can be injected. Most applications will use either
// the JsonWs or XmlWs facility, in which case a AbnormalStatusWriter that will respond to requests with an abnormal result
// (404, 503 etc) by sending a JSON or XML response respectively.
//
// If this behaviour is undesirable, an alternative AbnormalStatusWriter can set by using the frameworkModifiers mechanism
// (see http://granitic.io/1.0/ref/components)
const HttpServerAbnormalStatusFieldName = "AbnormalStatusWriter"
const accessLogWriterName = instance.FrameworkPrefix + "AccessLogWriter"

// Creates the components that make up the HttpServer facility (the server and an access log writer).
type HttpServerFacilityBuilder struct {
}

// See FacilityBuilder.BuildAndRegister
func (hsfb *HttpServerFacilityBuilder) BuildAndRegister(lm *logging.ComponentLoggerManager, ca *config.ConfigAccessor, cn *ioc.ComponentContainer) error {

	httpServer := new(HttpServer)
	ca.Populate("HttpServer", httpServer)

	cn.WrapAndAddProto(HttpServerComponentName, httpServer)

	if httpServer.AccessLogging {
		accessLogWriter := new(AccessLogWriter)
		ca.Populate("HttpServer.AccessLog", accessLogWriter)

		httpServer.AccessLogWriter = accessLogWriter

		cn.WrapAndAddProto(accessLogWriterName, accessLogWriter)
	}

	idbd := new(contextBuilderDecorator)
	cn.WrapAndAddProto(contextIdDecoratorName, idbd)

	return nil

}

// See FacilityBuilder.FacilityName
func (hsfb *HttpServerFacilityBuilder) FacilityName() string {
	return "HttpServer"
}

// See FacilityBuilder.DependsOnFacilities
func (hsfb *HttpServerFacilityBuilder) DependsOnFacilities() []string {
	return []string{}
}
