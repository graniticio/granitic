// Copyright 2016-2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package ioc

// ComponentDecorator is implemented by special temporary components that only exist while the IoC container is being populated.
//
// A ComponentDecorator's job is to examine another component (the subject) to see if it is suitable for modification by the ComponentDecorator.
// Typically this will involve the ComponentDecorator injecting another object into the subject if the component implements
// a particular interface or has a writable field of a particular name or type.
// A number of built-in decorators exist to accomplish tasks like automatically adding Loggers to components with a particular field.
type ComponentDecorator interface {
	// OfInterest should return true if the ComponentDecorator decides that the supplied component needs to be decorated.
	OfInterest(subject *Component) bool

	// DecorateComponent modifies the subject component.
	DecorateComponent(subject *Component, container *ComponentContainer)
}

// ContainerAccessor is implemented by any component that wants direct access to the IoC container.
type ContainerAccessor interface {
	// Container accepts a reference to the Granitic IoC container.
	Container(container *ComponentContainer)
}

// ContainerDecorator injects a reference to the IoC container into any component implementing the ContainerAccessor interface.
type ContainerDecorator struct {
	container *ComponentContainer
}

// OfInterest returns true if the subject component implements ContainerAccessor
func (cd *ContainerDecorator) OfInterest(subject *Component) bool {
	result := false

	switch subject.Instance.(type) {
	case ContainerAccessor:
		result = true
	}

	return result
}

// DecorateComponent injects a reference to the IoC container to a component that has alredy been determined to implement ContainerAccessor.
func (cd *ContainerDecorator) DecorateComponent(subject *Component, cc *ComponentContainer) {

	accessor := subject.Instance.(ContainerAccessor)
	accessor.Container(cc)

}
