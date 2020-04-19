// Copyright 2016-2020 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package logger

import (
	"github.com/graniticio/granitic/v2/config"
	"github.com/graniticio/granitic/v2/ioc"
	"github.com/graniticio/granitic/v2/logging"
)

//NullLoggingFacilityBuilder creates a minimal set of components to allow applications to run even if the application logging facility has been disabled
type NullLoggingFacilityBuilder struct {
}

// BuildAndRegister creates a decorator that will inject a 'null' logger into any component that asks for a an application logger
func (nlfb *NullLoggingFacilityBuilder) BuildAndRegister(lm *logging.ComponentLoggerManager, ca *config.Accessor, cn *ioc.ComponentContainer) error {

	ald := new(applicationLogDecorator)
	ald.FrameworkLogger = lm.CreateLogger(applicationLoggingDecoratorName)

	ald.useNullLogger = true
	ald.nullLogger = new(logging.NullLogger)

	cn.WrapAndAddProto(applicationLoggingDecoratorName, ald)

	return nil

}

// FacilityName implements FacilityBuilder.FacilityName
func (nlfb *NullLoggingFacilityBuilder) FacilityName() string {
	return "ApplicationLogging"
}

// DependsOnFacilities implements FacilityBuilder.DependsOnFacilities
func (nlfb *NullLoggingFacilityBuilder) DependsOnFacilities() []string {
	return []string{}
}
