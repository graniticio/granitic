package decorator

import (
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/logging"
	"github.com/graniticio/granitic/reflecttools"
	"reflect"
)

//TODO Rename application log var
const expectedApplicationLoggerFieldName string = "Log"
const expectedFrameworkLoggerFieldName string = "FrameworkLogger"

type ApplicationLogDecorator struct {
	LoggerManager   *logging.ComponentLoggerManager
	FrameworkLogger logging.Logger
}

func (ald *ApplicationLogDecorator) OfInterest(component *ioc.Component) bool {

	result := false
	fieldPresent := reflecttools.HasFieldOfName(component.Instance, expectedApplicationLoggerFieldName)

	frameworkLog := ald.FrameworkLogger

	if fieldPresent {

		targetFieldType := reflecttools.TypeOfField(component.Instance, expectedApplicationLoggerFieldName)
		typeOfLogger := reflect.TypeOf(ald.FrameworkLogger)

		if typeOfLogger.AssignableTo(targetFieldType) {
			result = true
		}

	}

	if frameworkLog.IsLevelEnabled(logging.Trace) {
		if result {
			frameworkLog.LogTracef("%s NEEDS an ApplicationLogger", component.Name)

		} else {
			frameworkLog.LogTracef("%s does not need an ApplicationLogger (either no field named %s or incompatible type)", component.Name, expectedApplicationLoggerFieldName)
		}
	}

	return result
}

func (ald *ApplicationLogDecorator) DecorateComponent(component *ioc.Component, container *ioc.ComponentContainer) {
	logger := ald.LoggerManager.CreateLogger(component.Name)

	reflectComponent := reflect.ValueOf(component.Instance).Elem()
	reflectComponent.FieldByName(expectedApplicationLoggerFieldName).Set(reflect.ValueOf(logger))

}

type FrameworkLogDecorator struct {
	LoggerManager   *logging.ComponentLoggerManager
	FrameworkLogger logging.Logger
}

func (fld *FrameworkLogDecorator) OfInterest(component *ioc.Component) bool {

	result := reflecttools.HasFieldOfName(component.Instance, expectedFrameworkLoggerFieldName)

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
