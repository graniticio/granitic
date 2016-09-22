package facility

import (
	"github.com/graniticio/granitic/instance"
	"github.com/graniticio/granitic/ioc"
)

const InstanceIDDecoratorName = instance.FrameworkPrefix + "InstanceIDDecorator"

type InstanceIDDecorator struct {
	InstanceID *instance.InstanceIdentifier
}

func (id *InstanceIDDecorator) OfInterest(component *ioc.Component) bool {

	i := component.Instance

	_, found := i.(instance.InstanceIdentifierReceiver)

	return found

}

func (id *InstanceIDDecorator) DecorateComponent(component *ioc.Component, container *ioc.ComponentContainer) {

	r := component.Instance.(instance.InstanceIdentifierReceiver)

	r.RegisterInstanceID(id.InstanceID)
}
