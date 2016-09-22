/*
Package ioc provides an Inversion of Control component container and lifecycle hooks.
*/
package ioc

import (
	"errors"
	"fmt"
	"github.com/graniticio/granitic/config"
	"github.com/graniticio/granitic/instance"
	"github.com/graniticio/granitic/logging"
	"github.com/graniticio/granitic/reflecttools"
	"os"
	"sort"
)

const containerDecoratorComponentName = instance.FrameworkPrefix + "ContainerDecorator"
const containerComponentName = instance.FrameworkPrefix + "Container"
const lifecycleComponentName = instance.FrameworkPrefix + "LifecycleManager"

type ComponentByNameFinder interface {
	ComponentByName(string) *Component
}

func NewComponentContainer(logm *logging.ComponentLoggerManager, ca *config.ConfigAccessor, sys *instance.System) *ComponentContainer {

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

type ComponentContainer struct {
	allComponents      map[string]*Component
	protoComponents    map[string]*ProtoComponent
	FrameworkLogger    logging.Logger
	configAccessor     *config.ConfigAccessor
	byLifecycleSupport map[LifecycleSupport][]*Component
	modifiers          map[string]map[string]string
	Lifecycle          *LifecycleManager
	system             *instance.System
}

func (cc *ComponentContainer) ComponentByName(name string) *Component {
	return cc.allComponents[name]
}

func (cc *ComponentContainer) ByLifecycleSupport(ls LifecycleSupport) []*Component {
	return cc.byLifecycleSupport[ls]
}

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

func (cc *ComponentContainer) AddModifier(comp string, field string, dep string) {

	m := cc.modifiers
	cm := m[comp]

	if cm == nil {
		cm = make(map[string]string)
		m[comp] = cm
	}

	cm[field] = dep

}

func (cc *ComponentContainer) AddModifiers(mods map[string]map[string]string) {

	for c, cm := range mods {

		for f, d := range cm {
			cc.AddModifier(c, f, d)
		}

	}

}

func (cc *ComponentContainer) ModifierExists(comp string, field string) bool {

	m := cc.modifiers[comp]

	return m != nil && m[field] != ""

}

func (cc *ComponentContainer) ModifiersExist(comp string) bool {
	return cc.modifiers[comp] != nil
}

func (cc *ComponentContainer) Modifiers(comp string) map[string]string {
	return cc.modifiers[comp]
}

func (cc *ComponentContainer) AddProto(proto *ProtoComponent) {

	cc.FrameworkLogger.LogTracef("Adding proto %s", proto.Component.Name)

	cc.protoComponents[proto.Component.Name] = proto
}

func (cc *ComponentContainer) WrapAndAddProto(name string, instance interface{}) {
	p := CreateProtoComponent(instance, name)
	cc.AddProto(p)
}

func (cc *ComponentContainer) AddProtos(protos []*ProtoComponent) {
	for _, p := range protos {
		cc.AddProto(p)
	}
}

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
			fl.LogTracef("%s needs %s", targetProto.Component.Name, fieldName, configPath)

			cc.configAccessor.SetField(fieldName, configPath, targetProto.Component.Instance)

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
		doneCount += 1

		if doneCount >= decs {
			break
		}

	}

	for n, _ := range decorators {
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
