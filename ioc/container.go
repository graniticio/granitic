// Copyright 2016-2022 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package ioc

import (
	"errors"
	"fmt"
	config_access "github.com/graniticio/config-access"
	"github.com/graniticio/granitic/v3/config"
	"github.com/graniticio/granitic/v3/instance"
	"github.com/graniticio/granitic/v3/logging"
	"github.com/graniticio/granitic/v3/reflecttools"
	"github.com/graniticio/granitic/v3/types"
	"os"
	"sort"
)

const containerDecoratorComponentName = instance.FrameworkPrefix + "ContainerDecorator"
const containerComponentName = instance.FrameworkPrefix + "Container"
const lifecycleComponentName = instance.FrameworkPrefix + "LifecycleManager"

// ComponentLookup is used by components that have visibility of other components stored in the IOC container
type ComponentLookup interface {
	// ComponentByName returns the Component with the supplied name, or nil if it does not exist.
	ComponentByName(string) *Component
	// AllComponents returns all registered components
	AllComponents() []*Component
}

// NewComponentContainer creates a new instance of a Granitic IoC container.
func NewComponentContainer(logm *logging.ComponentLoggerManager, ca config_access.Selector, sys *instance.System) *ComponentContainer {

	cc := new(ComponentContainer)
	cc.protoComponents = make(map[string]*ProtoComponent)
	cc.FrameworkLogger = logm.CreateLogger(containerComponentName)
	cc.configAccessor = ca
	cc.modifiers = make(map[string]map[string]string)
	cc.byLifecycleSupport = make(map[LifecycleSupport][]*Component)
	cc.system = sys

	lcm := new(LifecycleManager)
	lcm.container = cc
	lcm.FrameworkLogger = logm.CreateLogger(lifecycleComponentName)
	lcm.system = sys
	cc.Lifecycle = lcm

	return cc

}

/*
ComponentContainer is the The Granitic IoC container. See the GoDoc for the ioc package for more information on how to interact with the container.

Most applications should never need to interact with the container programmatically.
*/
type ComponentContainer struct {
	allComponents      map[string]*Component
	protoComponents    map[string]*ProtoComponent
	FrameworkLogger    logging.Logger
	configAccessor     config_access.Selector
	byLifecycleSupport map[LifecycleSupport][]*Component
	modifiers          map[string]map[string]string
	Lifecycle          *LifecycleManager
	system             *instance.System
}

// ProtoComponentsByType returns any ProtoComponents whose Component.Instance field matches the against the supplied TypeMatcher function.
// If called after the container is 'Accessible' an empty slice will be returned.
func (cc *ComponentContainer) ProtoComponentsByType(tm TypeMatcher) []*ProtoComponent {

	results := make([]*ProtoComponent, 0)

	for _, pc := range cc.protoComponents {

		if tm(pc.Component.Instance) {
			results = append(results, pc)
		}

	}

	return results
}

// ProtoComponents returns all components that have been registered by the container but have not yet all their dependencies
// resolved. If called after the container is 'Accessible' an empty slice will be returned.
func (cc *ComponentContainer) ProtoComponents() map[string]*ProtoComponent {
	return cc.protoComponents
}

// ComponentByName implements ComponentLookup.ComponentByName
func (cc *ComponentContainer) ComponentByName(name string) *Component {
	return cc.allComponents[name]
}

// ByLifecycleSupport returns all components hosted by the container that have specific support for a lifecycle event
// (i.e. implement the associated lifecycle interface
func (cc *ComponentContainer) ByLifecycleSupport(ls LifecycleSupport) []*Component {
	return cc.byLifecycleSupport[ls]
}

// AllComponents returns all of the components hosted by the container.
func (cc *ComponentContainer) AllComponents() []*Component {

	ac := make([]*Component, len(cc.allComponents))

	i := 0

	for _, v := range cc.allComponents {

		ac[i] = v
		i++

	}

	sort.Sort(ByName{ac})

	return ac
}

// AddModifier is used to override a dependency on a component (normally a built-in Granitic component) during the
// configure phase of container startup.
func (cc *ComponentContainer) AddModifier(comp string, field string, dep string) {

	m := cc.modifiers
	cm := m[comp]

	if cm == nil {
		cm = make(map[string]string)
		m[comp] = cm
	}

	cm[field] = dep

}

// AddModifiers is used to override a dependency on set of components (normally built-in Granitic components) during the
// configure phase of container startup.
func (cc *ComponentContainer) AddModifiers(mods map[string]map[string]string) {

	for c, cm := range mods {

		for f, d := range cm {
			cc.AddModifier(c, f, d)
		}

	}

}

// ModifierExists checks to see if a modifier (see AddModifier) has previously been registered for a field on a component.
func (cc *ComponentContainer) ModifierExists(comp string, field string) bool {

	m := cc.modifiers[comp]

	return m != nil && m[field] != ""

}

// ModifiersExist returns true if any modifiers (see AddModifier) have been registered.
func (cc *ComponentContainer) ModifiersExist(comp string) bool {
	return cc.modifiers[comp] != nil
}

// Modifiers returns all registered modifiers (see AddModifier)
func (cc *ComponentContainer) Modifiers(comp string) map[string]string {
	return cc.modifiers[comp]
}

// AddProto registers an instantiated but un-configured proto-component.
func (cc *ComponentContainer) AddProto(proto *ProtoComponent) {

	cc.FrameworkLogger.LogTracef("Adding proto %s", proto.Component.Name)

	cc.protoComponents[proto.Component.Name] = proto
}

// WrapAndAddProto registers an instance and name as an un-configured proto-component.
func (cc *ComponentContainer) WrapAndAddProto(name string, instance interface{}) {
	p := CreateProtoComponent(instance, name)
	cc.AddProto(p)
}

// AddProtos registers a collection of proto-components (see AddProto)
func (cc *ComponentContainer) AddProtos(protos []*ProtoComponent) {
	for _, p := range protos {
		cc.AddProto(p)
	}
}

// Populate converts all registered proto-components into components and populates them with configuration and dependencies.
func (cc *ComponentContainer) Populate() error {

	defer func() {
		if r := recover(); r != nil {
			cc.FrameworkLogger.LogErrorfWithTrace("Panic recovered while configuring components %s", r)
			os.Exit(-1)
		}
	}()

	decorators := make(map[string]ComponentDecorator)

	containerDecorator := new(ContainerDecorator)
	containerDecorator.container = cc

	decorators[containerDecoratorComponentName] = containerDecorator

	cc.allComponents = make(map[string]*Component)

	for _, protoComponent := range cc.protoComponents {

		component := protoComponent.Component

		if !reflecttools.IsPointerToStruct(component.Instance) {
			m := fmt.Sprintf("Component %s is not a pointer to a struct.", component.Name)
			return errors.New(m)
		}

		cc.addComponent(component)
		cc.captureDecorator(component, decorators)
	}

	err := cc.resolveDependenciesAndConfig()

	if err != nil {
		cc.FrameworkLogger.LogFatalf(err.Error())
		cc.FrameworkLogger.LogInfof("Aborting startup")
		os.Exit(-1)
	}

	cc.runDecorators(decorators)

	cc.protoComponents = nil

	return nil
}

func (cc *ComponentContainer) resolveDependenciesAndConfig() error {

	fl := cc.FrameworkLogger
	pi := new(types.ParamValueInjector)

	for _, targetProto := range cc.protoComponents {

		compName := targetProto.Component.Name
		deps := cc.mergeDependencies(compName, targetProto.Dependencies)

		for fieldName, depName := range deps {

			fl.LogTracef("%s needs %s", compName, depName)

			requiredComponent := cc.allComponents[depName]

			if requiredComponent == nil {
				message := fmt.Sprintf("No component named %s available (required by %s.%s)", depName, compName, fieldName)
				return errors.New(message)
			}

			targetInstance := targetProto.Component.Instance
			requiredInstance := requiredComponent.Instance

			err := reflecttools.SetPtrToStruct(targetInstance, fieldName, requiredInstance)

			if err != nil {
				m := fmt.Sprintf("Problem injecting dependency '%s' into %s.%s: %s", depName, compName, fieldName, err.Error())
				return errors.New(m)
			}

		}

		for fieldName, configPath := range targetProto.ConfigPromises {
			fl.LogTracef("%s.%s needs %s", compName, fieldName, configPath)

			target := targetProto.Component.Instance

			conf := cc.configAccessor.Config()

			if err := config_access.SetField(fieldName, configPath, target, conf); err != nil {

				if _, found := err.(config.MissingPathError); found && targetProto.HasDefaultValue(fieldName) {

					fl.LogDebugf("Default value found for %s.%s - attempting to inject", compName, fieldName)

					df := targetProto.DefaultValue(fieldName)
					params := types.NewSingleValueParams(fieldName, df)

					if err = pi.BindValueToField(fieldName, fieldName, params, target, defaultValueInjectionError); err != nil {

						err = fmt.Errorf("problem using a default value to populate component %s.%s. "+
							"Check your component definition files and rebuild or set a valid value in configuration at %s: %s", compName, fieldName, configPath, err.Error())

					}

				}

				if err != nil {
					return err
				}
			}

		}

	}

	return nil
}

// Combines dependencies attached to the proto components with any available framework modifiers
func (cc *ComponentContainer) mergeDependencies(comp string, cd map[string]string) map[string]string {

	merged := make(map[string]string)

	for k, v := range cd {
		merged[k] = v
	}

	if cc.ModifiersExist(comp) {
		for k, v := range cc.Modifiers(comp) {
			merged[k] = v
		}
	}

	return merged
}

func (cc *ComponentContainer) runDecorators(decorators map[string]ComponentDecorator) {

	decs := len(decorators)
	done := make(chan string, decs)

	for n, d := range decorators {

		go cc.runDecorator(n, d, done)
	}

	doneCount := 0

	for {
		<-done
		doneCount++

		if doneCount >= decs {
			break
		}

	}

	for n := range decorators {
		delete(cc.allComponents, n)
	}
}

func (cc *ComponentContainer) runDecorator(name string, cd ComponentDecorator, ch chan<- string) {

	for _, component := range cc.allComponents {
		if cd.OfInterest(component) {
			cd.DecorateComponent(component, cc)
		}
	}

	ch <- name
}

func (cc *ComponentContainer) captureDecorator(component *Component, decorators map[string]ComponentDecorator) {

	decorator, isDecorator := component.Instance.(ComponentDecorator)

	if isDecorator {
		cc.FrameworkLogger.LogTracef("Found decorator %s", component.Name)
		decorators[component.Name] = decorator
	}
}

func (cc *ComponentContainer) addComponent(component *Component) {
	cc.allComponents[component.Name] = component

	l := cc.FrameworkLogger

	n, nameable := component.Instance.(ComponentNamer)

	if nameable {
		n.SetComponentName(component.Name)
	}

	if _, startable := component.Instance.(Startable); startable {
		l.LogTracef("%s is Startable", component.Name)
		cc.addBySupport(component, CanStart)
	}

	if _, stoppable := component.Instance.(Stoppable); stoppable {
		l.LogTracef("%s is Stoppable", component.Name)
		cc.addBySupport(component, CanStop)
	}

	if _, blocker := component.Instance.(AccessibilityBlocker); blocker {
		l.LogTracef("%s is an AvailabilityBlocker", component.Name)
		cc.addBySupport(component, CanBlockStart)
	}

	if _, accessible := component.Instance.(Accessible); accessible {
		l.LogTracef("%s is a Accesible", component.Name)
		cc.addBySupport(component, CanBeAccessed)
	}

	if _, suspendable := component.Instance.(Suspendable); suspendable {
		l.LogTracef("%s is a Suspendable", component.Name)
		cc.addBySupport(component, CanSuspend)
	}

}

func (cc *ComponentContainer) addBySupport(c *Component, ls LifecycleSupport) {

	a := cc.byLifecycleSupport[ls]

	if a == nil {
		a = make([]*Component, 1)
		a[0] = c
	} else {
		a = append(a, c)
	}

	cc.byLifecycleSupport[ls] = a

}

// TypeMatcher implementations return true if the supplied interface is (or implements) an expected type
type TypeMatcher func(i interface{}) bool

func defaultValueInjectionError(paramName string, fieldName string, typeName string, params *types.Params) error {

	return fmt.Errorf("unable to inject default value of field %s which is of type %s", fieldName, typeName)

}
