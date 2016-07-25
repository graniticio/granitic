package ioc

import (
	"errors"
	"fmt"
	"github.com/graniticio/granitic/config"
	"github.com/graniticio/granitic/logging"
	"os"
	"reflect"
	"time"
)

const containerDecoratorComponentName = FrameworkPrefix + "ContainerDecorator"
const containerComponentName = FrameworkPrefix + "Container"

type ComponentContainer struct {
	allComponents    map[string]*Component
	protoComponents  map[string]*ProtoComponent
	componentsByType map[string][]interface{}
	FrameworkLogger  logging.Logger
	configAccessor   *config.ConfigAccessor
	startable        []*Component
	stoppable        []*Component
	blocker          []*Component
	accessible       []*Component
}

func (cc *ComponentContainer) AllComponents() map[string]*Component {
	return cc.allComponents
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

func (cc *ComponentContainer) FindByType(typeName string) []interface{} {
	return cc.componentsByType[typeName]
}

func (cc *ComponentContainer) StartComponents() error {

	defer func() {
		if r := recover(); r != nil {
			cc.FrameworkLogger.LogErrorfWithTrace("Panic recovered while starting components components %s", r)
			os.Exit(-1)
		}
	}()

	for _, component := range cc.startable {

		startable := component.Instance.(Startable)

		err := startable.StartComponent()

		if err != nil {
			message := fmt.Sprintf("Unable to start %s: %s", component.Name, err)
			return errors.New(message)
		}

	}

	if len(cc.blocker) != 0 {
		err := cc.waitForBlockers(5*time.Second, 12, 0)

		if err != nil {
			return err
		}

	}

	for _, component := range cc.accessible {

		accessible := component.Instance.(Accessible)
		err := accessible.AllowAccess()

		if err != nil {
			return err
		}

	}

	return nil
}

func (cc *ComponentContainer) waitForBlockers(retestInterval time.Duration, maxTries int, warnAfterTries int) error {

	var names []string

	for i := 0; i < maxTries; i++ {

		notReady, cNames := cc.countBlocking(i > warnAfterTries)
		names = cNames

		if notReady == 0 {
			return nil
		} else {
			time.Sleep(retestInterval)
		}
	}

	message := fmt.Sprintf("Startup blocked by %v", names)

	return errors.New(message)

}

func (cc *ComponentContainer) ShutdownComponents() error {

	for _, c := range cc.stoppable {

		s := c.Instance.(Stoppable)
		s.PrepareToStop()
	}

	cc.waitForReadyToStop(5*time.Second, 10, 3)

	for _, c := range cc.stoppable {

		s := c.Instance.(Stoppable)
		err := s.Stop()

		if err != nil {
			cc.FrameworkLogger.LogErrorf("%s did not stop cleanly %s", c.Name, err)
		}

	}

	return nil

}

func (cc *ComponentContainer) waitForReadyToStop(retestInterval time.Duration, maxTries int, warnAfterTries int) {

	for i := 0; i < maxTries; i++ {

		notReady := cc.countNotReady(i > warnAfterTries)

		if notReady == 0 {
			return
		} else {
			time.Sleep(retestInterval)
		}
	}

	cc.FrameworkLogger.LogFatalf("Some components not ready to stop, stopping anyway")

}

func (cc *ComponentContainer) countBlocking(warn bool) (int, []string) {

	notReady := 0
	names := []string{}

	for _, c := range cc.blocker {
		ab := c.Instance.(AccessibilityBlocker)

		block, err := ab.BlockAccess()

		if block {
			notReady += 1
			names = append(names, c.Name)
			if warn {
				if err != nil {
					cc.FrameworkLogger.LogErrorf("%s blocking startup: %s", c.Name, err)
				} else {
					cc.FrameworkLogger.LogErrorf("%s blocking startup (no reason given)", c.Name)
				}

			}
		}

	}

	return notReady, names
}

func (cc *ComponentContainer) countNotReady(warn bool) int {

	notReady := 0

	for _, c := range cc.stoppable {
		s := c.Instance.(Stoppable)

		ready, err := s.ReadyToStop()

		if !ready {
			notReady += 1

			if warn {
				if err != nil {
					cc.FrameworkLogger.LogWarnf("%s is not ready to stop: %s", c.Name, err)
				} else {
					cc.FrameworkLogger.LogWarnf("%s is not ready to stop (no reason given)", c.Name)
				}

			}
		}

	}

	return notReady
}

func (cc *ComponentContainer) Populate() error {

	defer func() {
		if r := recover(); r != nil {
			cc.FrameworkLogger.LogErrorfWithTrace("Panic recovered while configuring components %s", r)
			os.Exit(-1)
		}
	}()

	decorators := make([]ComponentDecorator, 1)

	containerDecorator := new(ContainerDecorator)
	containerDecorator.container = cc

	decorators[0] = containerDecorator

	cc.allComponents = make(map[string]*Component)
	cc.componentsByType = make(map[string][]interface{})

	for _, protoComponent := range cc.protoComponents {

		component := protoComponent.Component

		cc.addComponent(component)
		decorators = cc.captureDecorator(component, decorators)

	}

	err := cc.resolveDependenciesAndConfig()

	if err != nil {
		cc.FrameworkLogger.LogFatalf(err.Error())
		cc.FrameworkLogger.LogInfof("Aborting startup")
		os.Exit(-1)
	}

	cc.decorateComponents(decorators)

	return nil
}

func (cc *ComponentContainer) resolveDependenciesAndConfig() error {

	fl := cc.FrameworkLogger

	for _, proto := range cc.protoComponents {

		for fieldName, depName := range proto.Dependencies {

			fl.LogTracef("%s needs %s", proto.Component.Name, depName)

			requiredComponent := cc.allComponents[depName]

			if requiredComponent == nil {
				message := fmt.Sprintf("No component named %s available (required by %s.%s)", depName, proto.Component.Name, fieldName)
				return errors.New(message)
			}

			requiredInstance := requiredComponent.Instance

			targetReflect := reflect.ValueOf(proto.Component.Instance).Elem()

			defer func() {
				if r := recover(); r != nil {
					fl.LogFatalf("Problem setting %s.%s: %s ", proto.Component.Name, fieldName, r)
				}
			}()

			targetReflect.FieldByName(fieldName).Set(reflect.ValueOf(requiredInstance))
		}

		for fieldName, configPath := range proto.ConfigPromises {
			fl.LogTracef("%s needs %s", proto.Component.Name, fieldName, configPath)

			cc.configAccessor.SetField(fieldName, configPath, proto.Component.Instance)

		}

	}

	return nil
}

func (cc *ComponentContainer) decorateComponents(decorators []ComponentDecorator) {

	for _, component := range cc.allComponents {
		for _, decorator := range decorators {

			if decorator.OfInterest(component) {
				decorator.DecorateComponent(component, cc)
			}
		}
	}

}

func (cc *ComponentContainer) captureDecorator(component *Component, decorators []ComponentDecorator) []ComponentDecorator {

	decorator, isDecorator := component.Instance.(ComponentDecorator)

	if isDecorator {
		cc.FrameworkLogger.LogTracef("Found decorator %s", component.Name)
		return append(decorators, decorator)
	} else {
		return decorators
	}
}

func (cc *ComponentContainer) addComponent(component *Component) {
	cc.allComponents[component.Name] = component
	cc.mapComponentToType(component)

	l := cc.FrameworkLogger

	_, startable := component.Instance.(Startable)

	if startable {
		l.LogTracef("%s is Startable", component.Name)
		cc.startable = append(cc.startable, component)
	}

	_, stoppable := component.Instance.(Stoppable)

	if stoppable {
		l.LogTracef("%s is Stoppable", component.Name)
		cc.stoppable = append(cc.stoppable, component)
	}

	_, blocker := component.Instance.(AccessibilityBlocker)

	if blocker {
		l.LogTracef("%s is an AvailabilityBlocker", component.Name)
		cc.blocker = append(cc.blocker, component)
	}

	_, accessible := component.Instance.(Accessible)

	if accessible {
		l.LogTracef("%s is a Accesible", component.Name)
		cc.accessible = append(cc.accessible, component)
	}

}

func (cc *ComponentContainer) mapComponentToType(component *Component) {
	componentType := reflect.TypeOf(component.Instance)
	typeName := componentType.String()

	cc.FrameworkLogger.LogTracef("Storing component %s of type %s", component.Name, componentType.String())

	componentsOfSameType := cc.componentsByType[typeName]

	if componentsOfSameType == nil {
		componentsOfSameType = make([]interface{}, 1)
		componentsOfSameType[0] = component.Instance
		cc.componentsByType[typeName] = componentsOfSameType
	} else {
		cc.componentsByType[typeName] = append(componentsOfSameType, component.Instance)
	}

}

func NewContainer(loggingManager *logging.ComponentLoggerManager, configAccessor *config.ConfigAccessor) *ComponentContainer {

	container := new(ComponentContainer)
	container.protoComponents = make(map[string]*ProtoComponent)
	container.FrameworkLogger = loggingManager.CreateLogger(containerComponentName)
	container.configAccessor = configAccessor

	return container

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
