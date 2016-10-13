package serviceerror

import (
	"github.com/graniticio/granitic/grncerror"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/ws"
)

type ServiceErrorConsumerDecorator struct {
	ErrorSource *grncerror.ServiceErrorManager
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
	ErrorSource *grncerror.ServiceErrorManager
}

func (ecs *ErrorCodeSourceDecorator) OfInterest(component *ioc.Component) bool {
	s, found := component.Instance.(grncerror.ErrorCodeUser)

	if found {
		return s.ValidateMissing()
	}

	return found
}

func (ecs *ErrorCodeSourceDecorator) DecorateComponent(component *ioc.Component, container *ioc.ComponentContainer) {
	c := component.Instance.(grncerror.ErrorCodeUser)

	ecs.ErrorSource.RegisterCodeUser(c)
}
