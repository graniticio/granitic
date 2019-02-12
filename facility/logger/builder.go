// Copyright 2016-2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

/*
Package logger provides the FrameworkLogging and ApplicationLogging facilities which control logging from framework and application components.

Full documentation for this facility can be found at http://granitic.io/1.0/ref/logging and GoDoc for the logging types that your application
will interact with are detailled in the logging package.
*/
package logger

import (
	"errors"
	"github.com/graniticio/granitic/v2/config"
	"github.com/graniticio/granitic/v2/facility/runtimectl"
	"github.com/graniticio/granitic/v2/instance"
	"github.com/graniticio/granitic/v2/ioc"
	"github.com/graniticio/granitic/v2/logging"
)

const applicationLoggingDecoratorName = instance.FrameworkPrefix + "ApplicationLoggingDecorator"
const applicationLoggingManagerName = instance.FrameworkPrefix + "ApplicationLoggingManager"

// FacilityBuilder creates a new logging.ComponentLoggerManager for application components and updates the framework's ComponentLoggerManager
// (which was bootstraped with a command-line supplied global log level) with the application's logging configuration.
type FacilityBuilder struct {
}

// BuildAndRegister implements FacilityBuilder.BuildAndRegister
func (alfb *FacilityBuilder) BuildAndRegister(lm *logging.ComponentLoggerManager, ca *config.Accessor, cn *ioc.ComponentContainer) error {
	globalLogLevelLabel, err := ca.StringVal("ApplicationLogger.GlobalLogLevel")

	if err != nil {
		return alfb.error(err.Error())
	}

	defaultLogLevel, err := logging.LogLevelFromLabel(globalLogLevelLabel)

	if err != nil {
		return alfb.error(err.Error())
	}

	initialLogLevelsByComponent, err := ca.ObjectVal("ApplicationLogger.ComponentLogLevels")

	if err != nil {
		return err
	}

	writers, err := alfb.buildWriters(ca)
	formatter, err := alfb.buildFormatter(ca)

	if err != nil {
		return alfb.error(err.Error())
	}

	//Update the bootstrapped framework logger with the newly configured writers and formatter
	lm.UpdateWritersAndFormatter(writers, formatter)

	alm := logging.CreateComponentLoggerManager(defaultLogLevel, initialLogLevelsByComponent, writers, formatter)
	cn.WrapAndAddProto(applicationLoggingManagerName, alm)

	ald := new(applicationLogDecorator)
	ald.LoggerManager = alm
	ald.FrameworkLogger = lm.CreateLogger(applicationLoggingDecoratorName)

	cn.WrapAndAddProto(applicationLoggingDecoratorName, ald)

	alfb.addRuntimeCommands(ca, alm, lm, cn)

	return nil
}

func (alfb *FacilityBuilder) addRuntimeCommands(ca *config.Accessor, alm *logging.ComponentLoggerManager, flm *logging.ComponentLoggerManager, cn *ioc.ComponentContainer) {

	if !runtimectl.Enabled(ca) {
		return
	}

	gll := new(globalLogLevelCommand)
	gll.ApplicationManager = alm
	gll.FrameworkManager = flm

	cn.WrapAndAddProto(GlobalLogCommand, gll)

	llc := new(logLevelCommand)
	llc.ApplicationManager = alm
	llc.FrameworkManager = flm

	cn.WrapAndAddProto(LogLevelComponentName, llc)

}

func (alfb *FacilityBuilder) buildFormatter(ca *config.Accessor) (*logging.LogMessageFormatter, error) {

	lmf := new(logging.LogMessageFormatter)

	if err := ca.Populate("LogWriting.Format", lmf); err != nil {
		return nil, err
	}

	if lmf.PrefixFormat == "" && lmf.PrefixPreset == "" {
		lmf.PrefixPreset = logging.FrameworkPresetPrefix
	}

	return lmf, lmf.Init()

}

func (alfb *FacilityBuilder) buildWriters(ca *config.Accessor) ([]logging.LogWriter, error) {
	writers := make([]logging.LogWriter, 0)

	if console, err := ca.BoolVal("LogWriting.EnableConsoleLogging"); err != nil {
		return nil, err
	} else if console {
		writers = append(writers, new(logging.ConsoleWriter))
	}

	if file, err := ca.BoolVal("LogWriting.EnableFileLogging"); err != nil {
		return nil, err
	} else if file {
		fileWriter := new(logging.AsynchFileWriter)

		if err = ca.Populate("LogWriting.File", fileWriter); err != nil {
			return nil, err
		}

		if err = fileWriter.Init(); err != nil {
			return nil, err
		}

		writers = append(writers, fileWriter)
	}

	return writers, nil
}

func (alfb *FacilityBuilder) error(suffix string) error {

	return errors.New("Unable to initialise application logging: " + suffix)

}

// FacilityName implements FacilityBuilder.FacilityName
func (alfb *FacilityBuilder) FacilityName() string {
	return "ApplicationLogging"
}

// DependsOnFacilities implements FacilityBuilder.DependsOnFacilities
func (alfb *FacilityBuilder) DependsOnFacilities() []string {
	return []string{}
}
