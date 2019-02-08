// Copyright 2016-2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package httpserver

import (
	"github.com/graniticio/granitic/v2/config"
	"github.com/graniticio/granitic/v2/instance"
	"github.com/graniticio/granitic/v2/ioc"
	"github.com/graniticio/granitic/v2/logging"
)

// The name of the HTTPServer component as stored in the IoC framework.
const HTTPServerComponentName = instance.FrameworkPrefix + "HTTPServer"
const contextIdDecoratorName = instance.FrameworkPrefix + "RequestIdContextDecorator"

// The field on the HTTPServer component into which a ws.AbnormalStatusWriter can be injected. Most applications will use either
// the JsonWs or XmlWs facility, in which case a AbnormalStatusWriter that will respond to requests with an abnormal result
// (404, 503 etc) by sending a JSON or XML response respectively.
//
// If this behaviour is undesirable, an alternative AbnormalStatusWriter can set by using the frameworkModifiers mechanism
// (see http://granitic.io/1.0/ref/components)
const HTTPServerAbnormalStatusFieldName = "AbnormalStatusWriter"
const accessLogWriterName = instance.FrameworkPrefix + "AccessLogWriter"

// Creates the components that make up the HTTPServer facility (the server and an access log writer).
type HTTPServerFacilityBuilder struct {
}

// See FacilityBuilder.BuildAndRegister
func (hsfb *HTTPServerFacilityBuilder) BuildAndRegister(lm *logging.ComponentLoggerManager, ca *config.Accessor, cn *ioc.ComponentContainer) error {

	httpServer := new(HTTPServer)
	ca.Populate("HTTPServer", httpServer)

	cn.WrapAndAddProto(HTTPServerComponentName, httpServer)

	if httpServer.AccessLogging {
		accessLogWriter := new(AccessLogWriter)
		ca.Populate("HTTPServer.AccessLog", accessLogWriter)

		httpServer.AccessLogWriter = accessLogWriter

		cn.WrapAndAddProto(accessLogWriterName, accessLogWriter)
	}

	idbd := new(contextBuilderDecorator)
	idbd.Server = httpServer
	cn.WrapAndAddProto(contextIdDecoratorName, idbd)

	return nil

}

// See FacilityBuilder.FacilityName
func (hsfb *HTTPServerFacilityBuilder) FacilityName() string {
	return "HTTPServer"
}

// See FacilityBuilder.DependsOnFacilities
func (hsfb *HTTPServerFacilityBuilder) DependsOnFacilities() []string {
	return []string{}
}
