// Copyright 2016-2022 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

/*
Package logger provides the FrameworkLogging and ApplicationLogging facilities which control logging from framework and application components.

Full documentation for this facility can be found at https://granitic.io/ref/logging and GoDoc for the logging types that your application
will interact with are detailed in the logging package.
*/
package logger

import (
	"errors"
	"fmt"
	config_access "github.com/graniticio/config-access"
	"github.com/graniticio/granitic/v3/facility/runtimectl"
	"github.com/graniticio/granitic/v3/instance"
	"github.com/graniticio/granitic/v3/ioc"
	"github.com/graniticio/granitic/v3/logging"
)

const applicationLoggingDecoratorName = instance.FrameworkPrefix + "ApplicationLoggingDecorator"
const applicationLoggingManagerName = instance.FrameworkPrefix + "ApplicationLoggingManager"
const applicationLoggingFormatterName = instance.FrameworkPrefix + "ApplicationLoggingEntryFormatter"

const textEntryMode = "TEXT"
const jsonEntryMode = "JSON"

// FacilityBuilder creates a new logging.ComponentLoggerManager for application components and updates the framework's ComponentLoggerManager
// (which was bootstraped with a command-line supplied global log level) with the application's logging configuration.
type FacilityBuilder struct {
}

// BuildAndRegister implements FacilityBuilder.BuildAndRegister
func (alfb *FacilityBuilder) BuildAndRegister(lm *logging.ComponentLoggerManager, ca config_access.Selector, cn *ioc.ComponentContainer) error {
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

	writers, err := BuildWritersFromConfig(ca)

	if err != nil {
		return alfb.error(err.Error())
	}

	formatter, err := BuildFormatterFromConfig(ca)
	cn.WrapAndAddProto(applicationLoggingFormatterName, formatter)

	if err != nil {
		return alfb.error(err.Error())
	}

	//Update the bootstrapped framework logger with the newly configured writers and formatter
	lm.UpdateWritersAndFormatter(writers, formatter)

	alm := logging.CreateComponentLoggerManager(defaultLogLevel, initialLogLevelsByComponent, writers, formatter, false)
	cn.WrapAndAddProto(applicationLoggingManagerName, alm)

	ald := new(applicationLogDecorator)
	ald.LoggerManager = alm
	ald.FrameworkLogger = lm.CreateLogger(applicationLoggingDecoratorName)

	cn.WrapAndAddProto(applicationLoggingDecoratorName, ald)

	AddRuntimeCommandsForLogging(ca, alm, lm, cn)

	return nil
}

// AddRuntimeCommandsForFrameworkLogging registers the runtime control commands related to logging with stubbed
// out application logging
func AddRuntimeCommandsForFrameworkLogging(ca config_access.Selector, flm *logging.ComponentLoggerManager, cn *ioc.ComponentContainer) {

	alm := logging.CreateComponentLoggerManager(logging.Fatal, map[string]interface{}{}, []logging.LogWriter{}, nil, false)

	AddRuntimeCommandsForLogging(ca, alm, flm, cn)

}

// AddRuntimeCommandsForLogging registers the runtime control commands related to logging
func AddRuntimeCommandsForLogging(ca config_access.Selector, alm *logging.ComponentLoggerManager, flm *logging.ComponentLoggerManager, cn *ioc.ComponentContainer) {

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

// BuildFormatterFromConfig uses configuration to determine the format for application logs
func BuildFormatterFromConfig(ca config_access.Selector) (logging.StringFormatter, error) {

	var mode string
	var err error

	entryPath := "LogWriting.Format.Entry"

	if mode, err = ca.StringVal(entryPath); err != nil {
		return nil, err
	}

	if mode == textEntryMode {

		lmf := new(logging.LogMessageFormatter)

		if err := config_access.Populate("LogWriting.Format", lmf, ca.Config()); err != nil {
			return nil, err
		}

		if lmf.PrefixFormat == "" && lmf.PrefixPreset == "" {
			lmf.PrefixPreset = logging.FrameworkPresetPrefix
		}

		return lmf, lmf.Init()
	} else if mode == jsonEntryMode {

		jmf := new(logging.JSONLogFormatter)

		cfg := new(logging.JSONConfig)

		config_access.Populate("LogWriting.Format.JSON", cfg, ca.Config())
		jmf.Config = cfg

		cfg.UTC, _ = ca.BoolVal("LogWriting.Format.UtcTimes")

		cfg.ParsedFields = logging.ConvertFields(cfg.Fields)

		if err := logging.ValidateJSONFields(cfg.ParsedFields); err != nil {
			return nil, err
		}

		if mb, err := logging.CreateMapBuilder(cfg); err == nil {
			jmf.MapBuilder = mb
		} else {
			return nil, err
		}

		return jmf, nil
	}

	return nil, fmt.Errorf("%s is a not a supported value for %s. Should be %s or %s", mode, entryPath, textEntryMode, jsonEntryMode)

}

// BuildWritersFromConfig uses configuration to determine the writers for logging
func BuildWritersFromConfig(ca config_access.Selector) ([]logging.LogWriter, error) {
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

		if err = config_access.Populate("LogWriting.File", fileWriter, ca.Config()); err != nil {
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
