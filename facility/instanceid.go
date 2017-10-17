// Copyright 2016 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package facility

import (
	"github.com/graniticio/granitic/instance"
	"github.com/graniticio/granitic/ioc"
)

// Decorator to inject an InstanceIdentifier into components that need to be aware of the current application instance's
// ID.
type InstanceIdDecorator struct {
	// The instance identity that will be injected into components.
	InstanceId *instance.InstanceIdentifier
}

// OfInterest returns true if the supplied component implements instance.InstanceIdentifierReceiver
func (id *InstanceIdDecorator) OfInterest(subject *ioc.Component) bool {

	i := subject.Instance

	_, found := i.(instance.InstanceIdentifierReceiver)

	return found

}

// DecorateComponent injects the InstanceIdentifier in to the subject component.
func (id *InstanceIdDecorator) DecorateComponent(subject *ioc.Component, container *ioc.ComponentContainer) {

	r := subject.Instance.(instance.InstanceIdentifierReceiver)

	r.RegisterInstanceId(id.InstanceId)
}
