package ioc

type ComponentDecorator interface {
	OfInterest(component *Component) bool
	DecorateComponent(component *Component, container *ComponentContainer)
}

type ContainerAccessor interface {
	Container(container *ComponentContainer)
}

type ContainerDecorator struct {
	container *ComponentContainer
}

func (cd *ContainerDecorator) OfInterest(component *Component) bool {
	result := false

	switch component.Instance.(type) {
	case ContainerAccessor:
		result = true
	}

	return result
}

func (cd *ContainerDecorator) DecorateComponent(component *Component, container *ComponentContainer) {

	accessor := component.Instance.(ContainerAccessor)
	accessor.Container(container)

}
