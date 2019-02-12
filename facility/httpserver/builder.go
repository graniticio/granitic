// Copyright 2016-2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package httpserver

import (
	"github.com/graniticio/granitic/v2/config"
	"github.com/graniticio/granitic/v2/instance"
	"github.com/graniticio/granitic/v2/ioc"
	"github.com/graniticio/granitic/v2/logging"
)

// HTTPServerComponentName is the name of the HTTPServer component as stored in the IoC framework.
const HTTPServerComponentName = instance.FrameworkPrefix + "HTTPServer"
const contextIDDecoratorName = instance.FrameworkPrefix + "RequestIDContextDecorator"

// HTTPServerAbnormalStatusFieldName is the field on the HTTPServer component into which a ws.AbnormalStatusWriter can be injected. Most applications will use either
// the JSONWs or XMLWs facility, in which case a AbnormalStatusWriter that will respond to requests with an abnormal result
// (404, 503 etc) by sending a JSON or XML response respectively.
//
// If this behaviour is undesirable, an alternative AbnormalStatusWriter can set by using the frameworkModifiers mechanism
// (see http://granitic.io/1.0/ref/components)
const HTTPServerAbnormalStatusFieldName = "AbnormalStatusWriter"
const accessLogWriterName = instance.FrameworkPrefix + "AccessLogWriter"

// FacilityBuilder creates the components that make up the HTTPServer facility (the server and an access log writer).
type FacilityBuilder struct {
}

// BuildAndRegister implements FacilityBuilder.BuildAndRegister
func (hsfb *FacilityBuilder) BuildAndRegister(lm *logging.ComponentLoggerManager, ca *config.Accessor, cn *ioc.ComponentContainer) error {

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
	cn.WrapAndAddProto(contextIDDecoratorName, idbd)

	return nil

}

// FacilityName implements FacilityBuilder.FacilityName
func (hsfb *FacilityBuilder) FacilityName() string {
	return "HTTPServer"
}

// DependsOnFacilities implements FacilityBuilder.DependsOnFacilities
func (hsfb *FacilityBuilder) DependsOnFacilities() []string {
	return []string{}
}
