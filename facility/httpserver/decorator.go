package httpserver

import (
	"github.com/graniticio/granitic/logging"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/ws"
	"github.com/graniticio/granitic/httpendpoint"
)

type AbnormalStatusWriterDecorator struct {
	FrameworkLogger logging.Logger
	HttpServer      *HTTPServer
}

func (d *AbnormalStatusWriterDecorator) OfInterest(component *ioc.Component) bool {

	i := component.Instance

	_, found := i.(ws.AbnormalStatusWriter)

	return found

}

func (d *AbnormalStatusWriterDecorator) DecorateComponent(component *ioc.Component, container *ioc.ComponentContainer) {

	i := component.Instance.(ws.AbnormalStatusWriter)

	d.HttpServer.RegisterAbnormalStatusWriter(component.Name, i)
}


type VersionExtractorDecorator struct {
	FrameworkLogger logging.Logger
	HttpServer      *HTTPServer
	ServerName string
}

func (v *VersionExtractorDecorator) OfInterest(component *ioc.Component) bool {

	i := component.Instance

	_, found := i.(httpendpoint.RequestedVersionExtractor)

	return found

}

func (v *VersionExtractorDecorator) DecorateComponent(component *ioc.Component, container *ioc.ComponentContainer) {

	v.FrameworkLogger.LogInfof("%s will use %s as a RequestedVersionExtractor", v.ServerName, component.Name)


	i := component.Instance.(httpendpoint.RequestedVersionExtractor)

	v.HttpServer.VersionExtractor = i
}
