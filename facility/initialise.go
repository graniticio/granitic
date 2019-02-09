// Copyright 2016-2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package facility

import (
	"errors"
	"fmt"
	"github.com/graniticio/granitic/v2/config"
	"github.com/graniticio/granitic/v2/facility/httpserver"
	"github.com/graniticio/granitic/v2/facility/logger"
	"github.com/graniticio/granitic/v2/facility/querymanager"
	"github.com/graniticio/granitic/v2/facility/rdbms"
	"github.com/graniticio/granitic/v2/facility/runtimectl"
	"github.com/graniticio/granitic/v2/facility/serviceerror"
	"github.com/graniticio/granitic/v2/facility/taskscheduler"
	"github.com/graniticio/granitic/v2/facility/ws"
	"github.com/graniticio/granitic/v2/instance"
	"github.com/graniticio/granitic/v2/ioc"
	"github.com/graniticio/granitic/v2/logging"
)

const frameworkLoggingManagerName = instance.FrameworkPrefix + "FrameworkLoggingManager"
const frameworkLoggerDecoratorName = instance.FrameworkPrefix + "FrameworkLoggingDecorator"
const facilityInitialisorComponentName string = instance.FrameworkPrefix + "FacilityInitialisor"
const configErrorPrefix = "Unable to configure framework logging: "

// BootstrapFrameworkLogging creates a ComponentLoggerManager that can be used to create Loggers used by internal
// Granitic components during the bootstrap (pre-configuration load) phase of application startup.
func BootstrapFrameworkLogging(bootStrapLogLevel logging.LogLevel) (*logging.ComponentLoggerManager, *ioc.ProtoComponent) {

	flm := logging.CreateComponentLoggerManager(bootStrapLogLevel, nil,
		[]logging.LogWriter{new(logging.ConsoleWriter)}, logging.NewFrameworkLogMessageFormatter())
	proto := ioc.CreateProtoComponent(flm, frameworkLoggingManagerName)

	return flm, proto

}

// NewFacilitiesInitialisor creates a new FacilitiesInitialisor with access to the IoC container and a ComponentLoggerManager
func NewFacilitiesInitialisor(cc *ioc.ComponentContainer, flm *logging.ComponentLoggerManager) *FacilitiesInitialisor {
	fi := new(FacilitiesInitialisor)
	fi.container = cc
	fi.FrameworkLoggingManager = flm

	fi.Logger = flm.CreateLogger(facilityInitialisorComponentName)

	return fi
}

// The FacilitiesInitialisor finds all the facilities that have been enabled by the user application and invokes their
// corresponding Builder to initialise and configure them.
type FacilitiesInitialisor struct {
	// Access to the merged view of application configuration.
	ConfigAccessor *config.Accessor

	//A ComponentLoggerManager able to create Loggers for built-in Granitic components.
	FrameworkLoggingManager *logging.ComponentLoggerManager

	//Allows the FacilitiesInitialisor to log problems during facility initialisation.
	Logger         logging.Logger
	container      *ioc.ComponentContainer
	facilities     []Builder
	facilityStatus map[string]interface{}
}

func (fi *FacilitiesInitialisor) addFacility(f Builder) {
	fi.facilities = append(fi.facilities, f)
}

func (fi *FacilitiesInitialisor) buildEnabledFacilities() error {

	for _, fb := range fi.facilities {

		name := fb.FacilityName()

		if fi.facilityStatus[name] == nil {

			fi.Logger.LogWarnf("No setting for facility %s in the Facilities configuration object - will not enable this facility", name)
			continue

		}

		if fi.facilityStatus[name].(bool) {

			for _, dep := range fb.DependsOnFacilities() {

				if fi.facilityStatus[dep] == nil || fi.facilityStatus[dep].(bool) == false {
					message := fmt.Sprintf("Facility %s depends on facility %s, but %s is not enabled in configuration.", name, dep, dep)
					return errors.New(message)
				}

			}

			err := fb.BuildAndRegister(fi.FrameworkLoggingManager, fi.ConfigAccessor, fi.container)

			if err != nil {
				return err
			}
		}
	}

	return nil

}

// Initialise creates a Builder for each of the built-in Granitic facilities and then
// builds those facilities that have been enabled by the user.
func (fi *FacilitiesInitialisor) Initialise(ca *config.Accessor) error {
	fi.ConfigAccessor = ca

	fc, err := ca.ObjectVal("Facilities")

	if err != nil {
		return err
	}

	fi.facilityStatus = fc
	fi.updateFrameworkLogLevel()

	if fc["ApplicationLogging"].(bool) {
		fi.addFacility(new(logger.FacilityBuilder))
	}

	fi.addFacility(new(querymanager.FacilityBuilder))
	fi.addFacility(new(httpserver.FacilityBuilder))
	fi.addFacility(new(ws.JSONFacilityBuilder))
	fi.addFacility(new(ws.XMLFacilityBuilder))
	fi.addFacility(new(serviceerror.FacilityBuilder))
	fi.addFacility(new(rdbms.FacilityBuilder))
	fi.addFacility(new(runtimectl.FacilityBuilder))
	fi.addFacility(new(taskscheduler.FacilityBuilder))

	err = fi.buildEnabledFacilities()

	return err
}

func (fi *FacilitiesInitialisor) updateFrameworkLogLevel() error {

	flm := fi.FrameworkLoggingManager

	defaultLogLevelLabel, err := fi.ConfigAccessor.StringVal("FrameworkLogger.GlobalLogLevel")

	if err != nil {
		return errors.New(configErrorPrefix + err.Error())
	}

	defaultLogLevel, err := logging.LogLevelFromLabel(defaultLogLevelLabel)

	if err != nil {
		return errors.New(configErrorPrefix + err.Error())
	}

	il, err := fi.ConfigAccessor.ObjectVal("FrameworkLogger.ComponentLogLevels")

	if err != nil {
		return err
	}

	flm.SetInitialLogLevels(il)
	flm.SetGlobalThreshold(defaultLogLevel)

	fld := new(logger.FrameworkLogDecorator)
	fld.FrameworkLogger = flm.CreateLogger(frameworkLoggerDecoratorName)
	fld.LoggerManager = flm

	fi.container.WrapAndAddProto(frameworkLoggerDecoratorName, fld)

	return nil

}
