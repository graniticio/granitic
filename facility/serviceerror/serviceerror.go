package serviceerror

import (
	"github.com/graniticio/granitic/error"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/ws"
)

type ServiceErrorConsumerDecorator struct {
	ErrorSource *error.ServiceErrorManager
}

func (secd *ServiceErrorConsumerDecorator) OfInterest(component *ioc.Component) bool {
	_, found := component.Instance.(ws.ServiceErrorConsumer)

	return found
}

func (secd *ServiceErrorConsumerDecorator) DecorateComponent(component *ioc.Component, container *ioc.ComponentContainer) {
	c := component.Instance.(ws.ServiceErrorConsumer)
	c.ProvideErrorFinder(secd.ErrorSource)
}

type ErrorCodeSourceDecorator struct {
	ErrorSource *error.ServiceErrorManager
}

func (ecs *ErrorCodeSourceDecorator) OfInterest(component *ioc.Component) bool {
	s, found := component.Instance.(error.ErrorCodeSource)

	if found {
		return s.ValidateMissing()
	}

	return found
}

func (ecs *ErrorCodeSourceDecorator) DecorateComponent(component *ioc.Component, container *ioc.ComponentContainer) {
	c := component.Instance.(error.ErrorCodeSource)

	ecs.ErrorSource.RegisterCodeSource(c)
}
