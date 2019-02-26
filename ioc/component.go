// Copyright 2016-2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

/*
Package ioc provides an Inversion of Control component container and lifecycle hooks.

This package provides the types that define and support Granitic component container, which allows your application's
objects and Granitic's framework facilities to follow the inversion of control (IoC) pattern by having their lifecycle
and dependencies managed for them. An in-depth discussion of Granitic IoC can be found at http://granitic.io/1.0/ref/ioc
but a description of the core concepts follows.

Components

A component is defined by Granitic as an instance of a Go struct with a name that is unique within an application.
Each component in your application requires a entry in the components section of your application's component definition file like:

	{
	  "components": {
		"componentName": {
		  "type": "package.structType"
		}
	  }
	}

e.g:

	{
	  "components": {
		"createRecordLogic": {
		  "type": "inventory.CreateRecordLogic"
		}
	  }
	}

For complete information on defining components, refer to http://granitic.io/1.0/ref/components

Granitic's documentation will use the term component and component instance interchangeably. For example, 'component field'
means 'a field on the instance of the Go struct associated with that component'.


The container

When Granitic starts, it will create an instance of ioc.ComponentContainer - a structure that holds references to
each of the components in your application. It is responsible for injecting dependencies and configuration into your
components (see below) and managing lifecycle events (also see below). The container is also referred to as the component
container or IoC container.

Framework components

For each Granitic facility that you enable in your application, one or more framework components will be created and added to the container.
A framework component is exactly the same as any other component - an instance of a struct with a name. Depending on the
complexity of the facility, multiple components may be created.

A very simple name-spacing is used to separate the names of framework components from application components -
framework components' names all start with the string grnc

You are strongly encouraged to make sure your application components' names do not start with this string.


Dependencies and configuration

As part of its definition, your components can request that other components are injected into its fields. Your definition
can also include configuration (actual values to be set when the component is instantiated) or configuration promises
(values to be injected once all sources of configuration have been merged together).


	{
	  "components": {
		"createRecordLogic": {
		  "type": "inventory.CreateRecordLogic",
		  "MaxTracks": 20,
		  "ArtistMustExist": "conf:record.disableAutoArtist",
		  "DAO": "ref:inventoryDAO"
		}
	  }
	}

In the above example, the field CreateRecordLogic.MaxTracks is set to 20 when the struct is instantiated, ArtistMustExist
is set to the config element 'record.disableAutoArtist' and DAO is set to a reference to another component's instance. Note that c: and r:
can be used as shorthand for config: and ref: See http://granitic.io/1.0/ref/components for more information

Any error such as type mismatches or missing configuration will cause an error that will halt application startup.

Component templates

A template mechanism exists to allow multiple components that share a type, dependencies or configuration items to
only have those elements defined once. This is especially useful for web service handlers. See
http://granitic.io/1.0/ref/components#templates for more details.

Binding

Unlike JVM/CLR languages, Go has no runtime 'instance-for-type-name' mechanism for creating instances of a struct. As
a result, unlike JVM/CLR IoC containers you may have used, the container does not instantiate the actual instances
of the Go structs behind application components. Instead a 'binding' proces is used - refer to the pacakage documentation
for the grnc-bind tool for more information.

Container lifecycle

The process of starting the container transitions through a number of distinct phase. It is possible for your code
to be explicitly invoked during some of these phases by implementing one more lifecycle interfaces.

		Populate    Application and framework components are stored in the container.

		Configure   Configuration and dependencies are resolved and injected into components.

		Decorate    Components implementing the ioc.ComponentDecorator are given access to all other components
					to potentially modify.

		Start       Components implementing the ioc.Startable interface are invoked.

		Access	    Components implementing ioc.AccessibilityBlocker and ioc.Accessible are interacted with.

		Ready       Granitic application is running.

When the container is ready, a log message similar to

		grncInit Ready (startup time 6.28ms)

will be logged.

There are several other possible lifecycle phases after the container is ready:

		Suspend     Components implementing ioc.Suspendable  have their Suspend method invoked.

		Resume      Components implementing ioc.Suspendable have their Resume method invoked.

		Stop        Components implementing ioc.Stoppable are allowed to stop gracefully before the application exits.


Decorators

Decorators are special components implementing ioc.ComponentDecorator. Their main purpose is to inject dynamically
created objects into other components (such as Loggers). Decorators are destroyed after the Decorate phase of container
startup.

Stopping

Components that need to perform some shutdown process before an application exits should implement the Stoppable
interface. See the GoDoc for ioc.Stoppable below for more detail.

Container settings

The file $GRANITIC_HOME/resource/facility-config/system.json contains configuration, that can be overridden in your
application's configuration file, affecting startup, garbage collection and shutdown behaviour of the container.

More information can be found at http://granitic.io/1.0/ref/system-settings

Gaining access to the container

If your application component needs direct access to the container it should implement the ioc.ContainerAccessor. A
reference to the container will be injected into your component during the decorate phase.

External interaction with the container

If your application enables the RuntimeCtl facility, you can interact with the container and its components by using
the grnc-ctl command line utility. See the package documentation for grnc-ctl for more information.

*/
package ioc

// ComponentState represents what state (stopped, running) or transition between states (stopping, starting) a component is currently in.
type ComponentState int

const (
	//StoppedState indicates that a component has stopped
	StoppedState = iota
	//StoppingState indicates that a component is in the process of stopping
	StoppingState
	//StartingState indicates that a component is in the process of starting
	StartingState
	//AwaitingAccessState indicates that a component is available for connections from external sources
	AwaitingAccessState
	//RunningState indicates that a component is running normally
	RunningState
	//SuspendingState indicates that a component is in the process of being suspended
	SuspendingState
	//SuspendedState indicates that a component has been suspended and is effectively paused
	SuspendedState
	//ResumingState indicates that a component in the process of being resumed from a suspended state
	ResumingState
)

// ProtoComponents is a wrapping structure for a list of ProtoComponents and FrameworkDependencies that is required when starting Granitic.
// A ProtoComponents structure is built by the grnc-bind tool.
type ProtoComponents struct {
	// ProtoComponents to be finalised and stored in the IoC container.
	Components []*ProtoComponent

	// FrameworkDependencies are instructions to inject components into built-in Granitic components to alter their behaviour.
	// The structure is map[
	FrameworkDependencies map[string]map[string]string

	//A Base64 encoded version of the JSON files found in resource/facility-confg
	FrameworkConfig *string
}

// Clear removes the reference to the ProtoComponent objects held in this object, encouraging garbage collection.
func (pc *ProtoComponents) Clear() {
	pc.Components = nil
}

// NewProtoComponents creates a wrapping structure for a list of ProtoComponents
func NewProtoComponents(pc []*ProtoComponent, fd map[string]map[string]string, ser *string) *ProtoComponents {
	p := new(ProtoComponents)
	p.Components = pc
	p.FrameworkDependencies = fd
	p.FrameworkConfig = ser
	return p
}

// CreateProtoComponent creates a new ProtoComponent.
func CreateProtoComponent(componentInstance interface{}, componentName string) *ProtoComponent {

	proto := new(ProtoComponent)

	component := new(Component)
	component.Name = componentName
	component.Instance = componentInstance

	proto.Component = component

	return proto

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

	// A map of default values for fields if a config promise is not fulfiled
	DefaultValues map[string]string
}

// AddDependency requests that the container injects another component into the specified field during the configure phase of
// container startup
func (pc *ProtoComponent) AddDependency(fieldName, componentName string) {

	if pc.Dependencies == nil {
		pc.Dependencies = make(map[string]string)
	}

	pc.Dependencies[fieldName] = componentName
}

// AddConfigPromise requests that the container injects the config value at the specified path into the specified field during the configure phase of
// container startup.
func (pc *ProtoComponent) AddConfigPromise(fieldName, configPath string) {

	if pc.ConfigPromises == nil {
		pc.ConfigPromises = make(map[string]string)
	}

	pc.ConfigPromises[fieldName] = configPath
}

// AddDefaultValue records an untyped default value to use if a config promise is not fulfiled
func (pc *ProtoComponent) AddDefaultValue(fieldName, value string) {

	if pc.DefaultValues == nil {
		pc.DefaultValues = make(map[string]string)
	}

	pc.DefaultValues[fieldName] = value
}

// A Component is an instance of a struct with a name that is unique within your application.
type Component struct {
	// A pointer to a struct
	Instance interface{}

	// A name for this component that is unique within your application
	Name string
}

// Components is a type definition for a slice of components to allow sorting.
type Components []*Component

// Len returns the number of components in the slice
func (s Components) Len() int { return len(s) }

// Swap exchanges the position of the components at the specified indexes
func (s Components) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// ByName allows a slice of components to be sorted by name
type ByName struct{ Components }

// Less returns true if the the component at index i has a name the lexicographically proceeds that at j
func (s ByName) Less(i, j int) bool { return s.Components[i].Name < s.Components[j].Name }

// ComponentNamer is implemented by components where the component's instance needs to be aware of its own component name.
type ComponentNamer interface {
	// ComponentName returns the name of the component
	ComponentName() string

	// SetComponentName injects the component's name
	SetComponentName(name string)
}

// NewComponent creates a new Component with the supplied name and instance
func NewComponent(name string, instance interface{}) *Component {
	c := new(Component)
	c.Instance = instance
	c.Name = name

	return c
}
