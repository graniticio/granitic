package ioc

type ComponentDecorator interface {
	OfInterest(component *Component) bool
	DecorateComponent(component *Component, container *ComponentContainer)
}
