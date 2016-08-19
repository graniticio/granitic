package httpserver

import (
	"github.com/graniticio/granitic/config"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/logging"
)

const httpServerName = ioc.FrameworkPrefix + "HttpServer"
const accessLogWriterName = ioc.FrameworkPrefix + "AccessLogWriter"
const abnormalStatusWriterDecoratorName = ioc.FrameworkPrefix + "AbnormalStatusWriterDecorator"
const versionAssessorDecoratorName = ioc.FrameworkPrefix + "VersionAssessorDecorator"

type HttpServerFacilityBuilder struct {
}

func (hsfb *HttpServerFacilityBuilder) BuildAndRegister(lm *logging.ComponentLoggerManager, ca *config.ConfigAccessor, cn *ioc.ComponentContainer) error {

	httpServer := new(HTTPServer)
	ca.Populate("HttpServer", httpServer)

	cn.WrapAndAddProto(httpServerName, httpServer)

	writerDecorator := new(AbnormalStatusWriterDecorator)
	writerDecorator.HttpServer = httpServer
	cn.WrapAndAddProto(abnormalStatusWriterDecoratorName, writerDecorator)

	versionDecorator := new(VersionExtractorDecorator)
	versionDecorator.HttpServer = httpServer
	versionDecorator.ServerName = httpServerName
	versionDecorator.FrameworkLogger = lm.CreateLogger(versionAssessorDecoratorName)
	cn.WrapAndAddProto(versionAssessorDecoratorName, versionDecorator)


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
