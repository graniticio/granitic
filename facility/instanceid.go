// Copyright 2016 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package facility

import (
	"github.com/graniticio/granitic/instance"
	"github.com/graniticio/granitic/ioc"
)

// Decorator to inject an InstanceIdentifier into components that need to be aware of the current application instance's
// ID.
type InstanceIDDecorator struct {
	// The instance identity that will be injected into components.
	InstanceID *instance.InstanceIdentifier
}

// OfInterest returns true if the supplied component implements instance.InstanceIdentifierReceiver
func (id *InstanceIDDecorator) OfInterest(component *ioc.Component) bool {

	i := component.Instance

	_, found := i.(instance.InstanceIdentifierReceiver)

	return found

}

// DecorateComponent injects the InstanceIdentifier
func (id *InstanceIDDecorator) DecorateComponent(component *ioc.Component, container *ioc.ComponentContainer) {

	r := component.Instance.(instance.InstanceIdentifierReceiver)

	r.RegisterInstanceID(id.InstanceID)
}
