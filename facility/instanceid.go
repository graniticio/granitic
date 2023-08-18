// Copyright 2016-2023 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package facility

import (
	"github.com/graniticio/granitic/v2/instance"
	"github.com/graniticio/granitic/v2/ioc"
)

// InstanceIDDecorator injects an Identifier into components that need to be aware of the current application instance's
// ID.
type InstanceIDDecorator struct {
	// The instance identity that will be injected into components.
	InstanceID *instance.Identifier
}

// OfInterest returns true if the supplied component implements instance.Receiver
func (id *InstanceIDDecorator) OfInterest(subject *ioc.Component) bool {

	i := subject.Instance

	_, found := i.(instance.Receiver)

	return found

}

// DecorateComponent injects the Identifier in to the subject component.
func (id *InstanceIDDecorator) DecorateComponent(subject *ioc.Component, container *ioc.ComponentContainer) {

	r := subject.Instance.(instance.Receiver)

	r.RegisterInstanceID(id.InstanceID)
}
