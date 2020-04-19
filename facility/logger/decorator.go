// Copyright 2016-2020 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package logger

import (
	"github.com/graniticio/granitic/v2/ioc"
	"github.com/graniticio/granitic/v2/logging"
	"github.com/graniticio/granitic/v2/reflecttools"
	"reflect"
)

const expectedApplicationLoggerFieldName string = "Log"
const expectedFrameworkLoggerFieldName string = "FrameworkLogger"

// Injects a Logger into any component with field of type logging.Logger and the name Log.
type applicationLogDecorator struct {
	// The application ComponentLoggerManager (as opposed to the framework ComponentLoggerManager)
	LoggerManager *logging.ComponentLoggerManager

	// Logger to allow this decorator to log messages.
	FrameworkLogger logging.Logger

	useNullLogger bool

	nullLogger logging.Logger
}

// OfInterest returns true if the subject component has a field of type logging.Logger and the name Log.
func (ald *applicationLogDecorator) OfInterest(subject *ioc.Component) bool {

	result := false
	fieldPresent := reflecttools.HasFieldOfName(subject.Instance, expectedApplicationLoggerFieldName)

	frameworkLog := ald.FrameworkLogger

	if fieldPresent {

		targetFieldType := reflecttools.TypeOfField(subject.Instance, expectedApplicationLoggerFieldName)
		typeOfLogger := reflect.TypeOf(ald.FrameworkLogger)

		v := reflect.ValueOf(subject.Instance).Elem().FieldByName(expectedApplicationLoggerFieldName)

		if typeOfLogger.AssignableTo(targetFieldType) && v.IsNil() {
			result = true
		}

	}

	if frameworkLog.IsLevelEnabled(logging.Trace) {
		if result {
			frameworkLog.LogTracef("%s NEEDS an ApplicationLogger", subject.Name)

		} else {
			frameworkLog.LogTracef("%s does not need an ApplicationLogger (either no field named %s or incompatible type)", subject.Name, expectedApplicationLoggerFieldName)
		}
	}

	return result
}

// DecorateComponent injects a newly created Logger into the Log field of the subject component.
func (ald *applicationLogDecorator) DecorateComponent(subject *ioc.Component, container *ioc.ComponentContainer) {

	var logger logging.Logger

	if ald.useNullLogger {
		logger = ald.nullLogger
	} else {
		logger = ald.LoggerManager.CreateLogger(subject.Name)
	}

	reflectComponent := reflect.ValueOf(subject.Instance).Elem()
	reflectComponent.FieldByName(expectedApplicationLoggerFieldName).Set(reflect.ValueOf(logger))

}

// FrameworkLogDecorator injects a framework logger into Granitic framework components.
type FrameworkLogDecorator struct {
	// The framework ComponentLoggerManager (as opposed to the application ComponentLoggerManager)
	LoggerManager *logging.ComponentLoggerManager

	// Logger to allow this decorator to log messages.
	FrameworkLogger logging.Logger
}

// OfInterest returns true if the subject component has a field of type logging.Logger and the name FrameworkLogger.
func (fld *FrameworkLogDecorator) OfInterest(component *ioc.Component) bool {

	hasField := reflecttools.HasFieldOfName(component.Instance, expectedFrameworkLoggerFieldName)

	v := reflect.ValueOf(component.Instance).Elem().FieldByName(expectedFrameworkLoggerFieldName)

	result := hasField && v.IsNil()

	frameworkLog := fld.FrameworkLogger

	if frameworkLog.IsLevelEnabled(logging.Trace) {
		if result {
			frameworkLog.LogTracef("%s NEEDS a %s", component.Name, expectedFrameworkLoggerFieldName)

		} else {
			frameworkLog.LogTracef("%s does not need a %s", component.Name, expectedFrameworkLoggerFieldName)
		}
	}

	return result
}

// DecorateComponent injects a newly created Logger into the FrameworkLogger field of the subject component.
func (fld *FrameworkLogDecorator) DecorateComponent(component *ioc.Component, container *ioc.ComponentContainer) {
	logger := fld.LoggerManager.CreateLogger(component.Name)

	targetFieldType := reflecttools.TypeOfField(component.Instance, expectedFrameworkLoggerFieldName)
	typeOfLogger := reflect.TypeOf(logger)

	if typeOfLogger.AssignableTo(targetFieldType) {
		reflectComponent := reflect.ValueOf(component.Instance).Elem()
		reflectComponent.FieldByName(expectedFrameworkLoggerFieldName).Set(reflect.ValueOf(logger))
	} else {
		fld.FrameworkLogger.LogErrorf("Unable to inject a FrameworkLogger into component %s because field %s is not of the expected type logger.Logger", component.Name, expectedFrameworkLoggerFieldName)
	}

}
