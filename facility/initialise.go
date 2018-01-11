// Copyright 2016-2018 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package facility

import (
	"errors"
	"fmt"
	"github.com/graniticio/granitic/config"
	"github.com/graniticio/granitic/facility/httpserver"
	"github.com/graniticio/granitic/facility/logger"
	"github.com/graniticio/granitic/facility/querymanager"
	"github.com/graniticio/granitic/facility/rdbms"
	"github.com/graniticio/granitic/facility/runtimectl"
	"github.com/graniticio/granitic/facility/serviceerror"
	"github.com/graniticio/granitic/facility/taskscheduler"
	"github.com/graniticio/granitic/facility/ws"
	"github.com/graniticio/granitic/instance"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/logging"
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
// corresponding FacilityBuilder to initialise and configure them.
type FacilitiesInitialisor struct {
	// Access to the merged view of application configuration.
	ConfigAccessor *config.ConfigAccessor

	//A ComponentLoggerManager able to create Loggers for built-in Granitic components.
	FrameworkLoggingManager *logging.ComponentLoggerManager

	//Allows the FacilitiesInitialisor to log problems during facility initialisation.
	Logger         logging.Logger
	container      *ioc.ComponentContainer
	facilities     []FacilityBuilder
	facilityStatus map[string]interface{}
}

func (fi *FacilitiesInitialisor) addFacility(f FacilityBuilder) {
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

// Initialise creates a FacilityBuilder for each of the built-in Granitic facilities and then
// builds those facilities that have been enabled by the user.
func (fi *FacilitiesInitialisor) Initialise(ca *config.ConfigAccessor) error {
	fi.ConfigAccessor = ca

	fc, err := ca.ObjectVal("Facilities")

	if err != nil {
		return err
	}

	fi.facilityStatus = fc
	fi.updateFrameworkLogLevel()

	if fc["ApplicationLogging"].(bool) {
		fi.addFacility(new(logger.ApplicationLoggingFacilityBuilder))
	}

	fi.addFacility(new(querymanager.QueryManagerFacilityBuilder))
	fi.addFacility(new(httpserver.HttpServerFacilityBuilder))
	fi.addFacility(new(ws.JsonWsFacilityBuilder))
	fi.addFacility(new(ws.XmlWsFacilityBuilder))
	fi.addFacility(new(serviceerror.ServiceErrorManagerFacilityBuilder))
	fi.addFacility(new(rdbms.RdbmsAccessFacilityBuilder))
	fi.addFacility(new(runtimectl.RuntimeCtlFacilityBuilder))
	fi.addFacility(new(taskscheduler.TaskSchedulerFacilityBuilder))

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
