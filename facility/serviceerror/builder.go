// Copyright 2016-2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package serviceerror

import (
	"errors"
	"fmt"
	"github.com/graniticio/granitic/v2/config"
	ge "github.com/graniticio/granitic/v2/grncerror"
	"github.com/graniticio/granitic/v2/instance"
	"github.com/graniticio/granitic/v2/ioc"
	"github.com/graniticio/granitic/v2/logging"
)

const (
	serviceErrorManagerComponentName      = instance.FrameworkPrefix + "ServiceErrorManager"
	serviceErrorDecoratorComponentName    = instance.FrameworkPrefix + "ServiceErrorSourceDecorator"
	errorCodeSourceDecoratorComponentName = instance.FrameworkPrefix + "errorCodeSourceDecorator"
)

// FacilityBuilder constructs an instance of ServiceErrorManager and registers it as a component.
type FacilityBuilder struct {
}

// BuildAndRegister implements FacilityBuilder.BuildAndRegister
func (fb *FacilityBuilder) BuildAndRegister(lm *logging.ComponentLoggerManager, ca *config.Accessor, cn *ioc.ComponentContainer) error {

	manager := new(ge.ServiceErrorManager)
	manager.FrameworkLogger = lm.CreateLogger(serviceErrorManagerComponentName)

	panicOnMissing, err := ca.BoolVal("ServiceErrorManager.PanicOnMissing")

	if err != nil {
		return errors.New("Unable to build service error manager " + err.Error())
	}

	manager.PanicOnMissing = panicOnMissing

	cn.WrapAndAddProto(serviceErrorManagerComponentName, manager)

	errorDecorator := new(consumerDecorator)
	errorDecorator.ErrorSource = manager
	cn.WrapAndAddProto(serviceErrorDecoratorComponentName, errorDecorator)

	codeDecorator := new(errorCodeSourceDecorator)
	codeDecorator.ErrorSource = manager
	cn.WrapAndAddProto(errorCodeSourceDecoratorComponentName, codeDecorator)

	definitionsPath, err := ca.StringVal("ServiceErrorManager.ErrorDefinitions")

	if err != nil {
		return errors.New("Unable to load service error messages from configuration: " + err.Error())
	}

	if messages, err := fb.loadMessagesFromConfig(definitionsPath, ca); err == nil {
		manager.LoadErrors(messages)

	} else {
		return err
	}

	return nil
}

// FacilityName implements FacilityBuilder.FacilityName
func (fb *FacilityBuilder) FacilityName() string {
	return "ServiceErrorManager"
}

// DependsOnFacilities implements FacilityBuilder.DependsOnFacilities
func (fb *FacilityBuilder) DependsOnFacilities() []string {
	return []string{}
}

func (fb *FacilityBuilder) loadMessagesFromConfig(dPath string, ca *config.Accessor) ([]interface{}, error) {

	if !ca.PathExists(dPath) {
		return nil, fmt.Errorf("no error definitions found at config path %s", dPath)
	}

	i := ca.Value(dPath)

	if config.JSONType(i) != config.JSONArray {
		e := fmt.Errorf("couldn't load error messages from config path %s. Make sure the path exists in your configuration and that %s is an array of string arrays ([][]string)", dPath, dPath)
		return nil, e
	}

	return ca.Array(dPath)

}
