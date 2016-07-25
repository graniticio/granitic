package ioc

const FrameworkPrefix = "gntc"

type ProtoComponent struct {
	Component      *Component
	Dependencies   map[string]string
	ConfigPromises map[string]string
}

func (pc *ProtoComponent) AddDependency(fieldName, componentName string) {

	if pc.Dependencies == nil {
		pc.Dependencies = make(map[string]string)
	}

	pc.Dependencies[fieldName] = componentName
}

func (pc *ProtoComponent) AddConfigPromise(fieldName, configPath string) {

	if pc.ConfigPromises == nil {
		pc.ConfigPromises = make(map[string]string)
	}

	pc.ConfigPromises[fieldName] = configPath
}

func CreateProtoComponent(componentInstance interface{}, componentName string) *ProtoComponent {

	proto := new(ProtoComponent)

	component := new(Component)
	component.Name = componentName
	component.Instance = componentInstance

	proto.Component = component

	return proto

}

type Component struct {
	Instance interface{}
	Name     string
}
