package ioc

// What state (stopped, running) or state-tranistion a component is currently in.
type ComponentState int

const (
	StoppedState = iota
	StoppingState
	StartingState
	AwaitingAccessState
	RunningState
	SuspendingState
	SuspendedState
	ResumingState
)

// A wrapping structure for a list of ProtoComponents and FrameworkDependencies that is required when starting Granitic.
// A ProtoComponents structure is built by the grnc-bind tool.
type ProtoComponents struct {
	// ProtoComponents to be finalised and stored in the IoC container.
	Components []*ProtoComponent

	// FrameworkDependencies are instructions to inject components into built-in Granitic components to alter their behaviour.
	// The structure is map[
	FrameworkDependencies map[string]map[string]string
}

func (pc *ProtoComponents) Clear() {
	pc.Components = nil
}

// NewProtoComponents creates a wrapping structure for a list of ProtoComponents
func NewProtoComponents(pc []*ProtoComponent, fd map[string]map[string]string) *ProtoComponents {
	p := new(ProtoComponents)
	p.Components = pc
	p.FrameworkDependencies = fd
	return p
}

// A ProtoComponent is a partially configured component that will be hosted in the Granitic IoC container once
// it is fully configured. Typically ProtoComponents are created using the grnc-bind tool.
type ProtoComponent struct {
	// The name of a component and the component instance (a pointer to an instantiated struct).
	Component *Component

	// A map of fields on the component instance and the names of other components that should be injected into those fields.
	Dependencies map[string]string

	// A map of fields on the component instance and the config-path that will contain the configuration that shoud be inject into the field.
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

type Components []*Component

func (s Components) Len() int      { return len(s) }
func (s Components) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type ByName struct{ Components }

func (s ByName) Less(i, j int) bool { return s.Components[i].Name < s.Components[j].Name }

type ComponentNamer interface {
	ComponentName() string
	SetComponentName(name string)
}

func NewComponent(name string, instance interface{}) *Component {
	c := new(Component)
	c.Instance = instance
	c.Name = name

	return c
}
