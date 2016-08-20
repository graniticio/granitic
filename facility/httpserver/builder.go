package httpserver

import (
	"github.com/graniticio/granitic/config"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/logging"
)

const HttpServerComponentName = ioc.FrameworkPrefix + "HttpServer"
const HttpServerAbnormalStatusFieldName = "AbnormalStatusWriter"
const accessLogWriterName = ioc.FrameworkPrefix + "AccessLogWriter"

type HttpServerFacilityBuilder struct {
}

func (hsfb *HttpServerFacilityBuilder) BuildAndRegister(lm *logging.ComponentLoggerManager, ca *config.ConfigAccessor, cn *ioc.ComponentContainer) error {

	httpServer := new(HTTPServer)
	ca.Populate("HttpServer", httpServer)

	cn.WrapAndAddProto(HttpServerComponentName, httpServer)


	if !httpServer.AccessLogging {
		return nil
	}

	accessLogWriter := new(AccessLogWriter)
	ca.Populate("HttpServer.AccessLog", accessLogWriter)

	httpServer.AccessLogWriter = accessLogWriter

	cn.WrapAndAddProto(accessLogWriterName, accessLogWriter)

	return nil

}

func (hsfb *HttpServerFacilityBuilder) FacilityName() string {
	return "HttpServer"
}

func (hsfb *HttpServerFacilityBuilder) DependsOnFacilities() []string {
	return []string{}
}
