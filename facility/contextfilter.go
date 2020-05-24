// Copyright 2016-2020 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package facility

import (
	"github.com/graniticio/granitic/v2/config"
	"github.com/graniticio/granitic/v2/instance"
	"github.com/graniticio/granitic/v2/ioc"
	"github.com/graniticio/granitic/v2/logging"
	"github.com/graniticio/granitic/v2/reflecttools"
	"reflect"
)

// ContextFilterBuilder adds decorators required to inject an instance of logging.ContextFilter into those framework
// components that might use them
type ContextFilterBuilder struct{}

//BuildAndRegister constructs the components that together constitute the facility and stores them in the IoC container.
func (cf *ContextFilterBuilder) BuildAndRegister(lm *logging.ComponentLoggerManager, ca *config.Accessor, cn *ioc.ComponentContainer) error {

	log := lm.CreateLogger(instance.FrameworkPrefix + "ContextFilterBuilder")

	//Try and find an instance of a component that is a ContextFilter
	matches := cn.ProtoComponentsByType(cfTypeMatcher)

	var filter logging.ContextFilter

	matchCount := len(matches)

	if matchCount == 0 {
		// No matching components

		log.LogDebugf("No components found that match logging.ContextFilter")

		return nil
	}

	if matchCount == 1 {
		//Use the filter without wrapping it

		fc := matches[0].Component

		log.LogDebugf("Found one component %s which implements logging.ContextFilter ", fc.Name)

		filter = fc.Instance.(logging.ContextFilter)
	}

	if matchCount > 1 {

		log.LogDebugf("Found %d components that implements logging.ContextFilter. Will wrap in a PrioritisedContextFilter ", matchCount)

		pcf := new(logging.PrioritisedContextFilter)

		for _, v := range matches {

			log.LogDebugf("Adding %s", v.Component.Name)

			pcf.Add(v.Component.Instance.(logging.ContextFilter))
		}

		filter = pcf
	}

	dn := instance.FrameworkPrefix + "ContextFilterDecorator"

	fl := lm.CreateLogger(dn)

	d := new(filterDecorator)
	d.FrameworkLogger = fl
	d.Filter = filter

	cn.WrapAndAddProto(dn, d)

	return nil

}

//FacilityName returns the facility's unique name. Used to check whether the facility is enabled in configuration.
func (cf *ContextFilterBuilder) FacilityName() string {
	return "ContextFilterDecorator"
}

//DependsOnFacilities returns the names of other facilities that must be enabled in order for this facility to run correctly.
func (cf *ContextFilterBuilder) DependsOnFacilities() []string {
	return []string{}
}

func cfTypeMatcher(i interface{}) bool {

	_, okay := i.(logging.ContextFilter)

	return okay

}

const expectedFilterFieldName string = "ContextFilter"

type filterDecorator struct {
	Filter          logging.ContextFilter
	FrameworkLogger logging.Logger
}

// OfInterest returns true if the supplied component is an instance of instrument.RequestInstrumentationManager
func (id *filterDecorator) OfInterest(subject *ioc.Component) bool {
	result := false
	fieldPresent := reflecttools.HasFieldOfName(subject.Instance, expectedFilterFieldName)

	if fieldPresent {

		targetFieldType := reflecttools.TypeOfField(subject.Instance, expectedFilterFieldName)
		typeOfFilter := reflect.TypeOf(id.Filter)

		v := reflect.ValueOf(subject.Instance).Elem().FieldByName(expectedFilterFieldName)

		if typeOfFilter.AssignableTo(targetFieldType) && v.IsNil() {
			result = true
		}

	}

	if result {
		id.FrameworkLogger.LogDebugf("%s needs a logging.ContextFilter", subject.Name)
	}

	return result

}

// DecorateComponent injects the instrument.RequestInstrumentationManager into the HTTP server
func (id *filterDecorator) DecorateComponent(subject *ioc.Component, cc *ioc.ComponentContainer) {

	reflectComponent := reflect.ValueOf(subject.Instance).Elem()
	reflectComponent.FieldByName(expectedFilterFieldName).Set(reflect.ValueOf(id.Filter))
}
