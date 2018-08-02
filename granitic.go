// Copyright 2016-2018 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

/*
Package granitic provides methods for configuring and starting a Granitic application.

Granitic is a framework for building micro-services in Go

To get started with Granitic, visit http://www.granitic.io/getting-started-installing-granitic
All of the documentation you will need is included in the GoDoc, you can find an index and recommended reading order here: http://www.granitic.io/documentation

Entry points

This package provides entry point functions for your application to hand control over to Granitic. Typically your application
will have a single, minimal file in its main package similar to:

	package main

	import "github.com/graniticio/granitic"
	import "github.com/yourUser/yourPackage/bindings"

	func main() {
		granitic.StartGranitic(bindings.Components())
	}

You can build a skeleton Granitic application by running the grnc-project tool, which will generate a main file, empty
configuration file and empty component definition file. The uses and syntax of these files are described in the config and ioc packages respectively.

Components and configuration

A Granitic application needs two things to start:

1. A list of components to host in its IoC container.

2. One or more JSON configuration files containing environment-specific settings for your application (passwords, hostnames etc.)

Configuration files

Folders and files containing configuration are by default expected to be stored in

	resource/config

This folder can contain any number of files or sub-directories. This location can be overridden by using the -c argument
when starting your application from the command line. This argument is expected to be a comma separated list of file paths,
directories or HTTP URLs to JSON files or any mixture of the above.

Command line arguments

When starting your application from the command, Granitic takes control of processing command line arguments. By
default your application will support the following arguments.

	-c A comma separated list of files, directories or HTTP URIs in any combination (default resource/config)
	-l The level of messages that will be logged by the framework while bootstrapping (before logging configuration is loaded; default INFO)
	-i An optional string that can be used to uniquely identify this instance of your application

If your application needs to perform command line processing and you want to prevent Granitic from attempting to parse command line arguments,
you should start Granitic using the alternative:

	StartGraniticWithSettings(cs *ioc.ProtoComponents, is *config.InitialSettings)

where you are expected to programmatically define the initial settings.

*/
package granitic

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"github.com/graniticio/granitic/config"
	"github.com/graniticio/granitic/facility"
	"github.com/graniticio/granitic/instance"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/logging"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

const (
	initiatorComponentName      string = instance.FrameworkPrefix + "Init"
	systemPath                         = "System"
	configAccessorComponentName string = instance.FrameworkPrefix + "ConfigAccessor"
	instanceIdDecoratorName            = instance.FrameworkPrefix + "InstanceIdDecorator"
)

// StartGranitic starts the IoC container and populates it with the supplied list of prototype components. Any settings
// required during the initial startup of the container are expected to be provided via command line arguments (see
// this page's header for more details). This function will run until the application is halted by an interrupt (ctrl+c) or
// a runtime control shutdown command.
func StartGranitic(cs *ioc.ProtoComponents) {

	is := config.InitialSettingsFromEnvironment()
	is.BuiltInConfig = cs.FrameworkConfig

	StartGraniticWithSettings(cs, is)
}

// StartGraniticWithSettings starts the IoC container and populates it with the supplied list of prototype components and using the
// provided intial settings. This function will run until the application is halted by an interrupt (ctrl+c) or
// a runtime control shutdown command.
func StartGraniticWithSettings(cs *ioc.ProtoComponents, is *config.InitialSettings) {
	i := new(initiator)
	i.Start(cs, is)
}

type initiator struct {
	logger logging.Logger
}

func (i *initiator) Start(customComponents *ioc.ProtoComponents, is *config.InitialSettings) {

	container := i.buildContainer(customComponents, is)
	customComponents.Clear()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)

	go func() {
		<-c
		i.shutdown(container)
		instance.ExitNormal()
	}()

	for {
		time.Sleep(10 * time.Second)
	}
}

// Creates and populate a Granitic IoC container using the user components and configuration files provided
func (i *initiator) buildContainer(ac *ioc.ProtoComponents, is *config.InitialSettings) *ioc.ComponentContainer {

	//Bootstrap the logging framework
	frameworkLoggingManager, logManageProto := facility.BootstrapFrameworkLogging(is.FrameworkLogLevel)
	i.logger = frameworkLoggingManager.CreateLogger(initiatorComponentName)

	l := i.logger
	l.LogInfof("Starting components")

	//Merge all configuration files and create a container
	ca := i.createConfigAccessor(is.BuiltInConfig, is.Configuration, frameworkLoggingManager)

	//Load system settings from config
	ss := i.loadSystemsSettings(ca)

	//Create the IoC container
	cc := ioc.NewComponentContainer(frameworkLoggingManager, ca, ss)
	cc.AddProto(logManageProto)

	//Assign an identity to this instance of the application
	i.createInstanceIdentifier(is, cc)

	//Register user components with container
	cc.AddProtos(ac.Components)
	cc.AddModifiers(ac.FrameworkDependencies)

	//Instantiate those facilities required by user and register as components in container
	fi := facility.NewFacilitiesInitialisor(cc, frameworkLoggingManager)

	err := fi.Initialise(ca)
	i.shutdownIfError(err, cc)

	//Inject configuration and dependencies into all components
	err = cc.Populate()
	i.shutdownIfError(err, cc)

	//Proto components no longer needed
	if ss.FlushMergedConfig {
		ca.Flush()
	}

	if ss.GCAfterConfigure {
		runtime.GC()
	}

	//Start all startable components
	err = cc.Lifecycle.StartAll()
	i.shutdownIfError(err, cc)

	elapsed := time.Since(is.StartTime)
	l.LogInfof("Ready (startup time %s)", elapsed)

	return cc
}

func (i *initiator) createInstanceIdentifier(is *config.InitialSettings, cc *ioc.ComponentContainer) {
	id := is.InstanceId

	if id != "" {
		ii := new(instance.InstanceIdentifier)
		ii.Id = id
		cc.WrapAndAddProto(instance.InstanceIdComponent, ii)

		iidd := new(facility.InstanceIdDecorator)
		iidd.InstanceId = ii
		cc.WrapAndAddProto(instanceIdDecoratorName, iidd)

		i.logger.LogInfof("Instance Id: %s", id)
	}

}

// Cleanly stop the container and any running components in the event of an error
// during startup.
func (i *initiator) shutdownIfError(err error, cc *ioc.ComponentContainer) {

	if err != nil {
		i.logger.LogFatalf(err.Error())
		i.shutdown(cc)
		instance.ExitError()
	}

}

// Log that the container is stopping and let the container stop its
// components gracefully
func (i *initiator) shutdown(cc *ioc.ComponentContainer) {
	i.logger.LogInfof("Shutting down (system signal)")

	cc.Lifecycle.StopAll()
}

// Merge together all of the local and remote JSON configuration files and wrap them in a *config.ConfigAccessor
// which allows programmatic access to the merged config.
func (i *initiator) createConfigAccessor(builtIn64 *string, configPaths []string, flm *logging.ComponentLoggerManager) *config.ConfigAccessor {

	builtIn := map[string]interface{}{}

	bz, err := base64.StdEncoding.DecodeString(*builtIn64)

	if err != nil {
		i.logger.LogFatalf("Unable to deserialize the copy of Grantic's configuration created by grnc-bind. Re-run grnc-bind and re-build: %s", err.Error())
		instance.ExitError()
	}

	b := bytes.Buffer{}
	b.Write(bz)

	gob.Register(map[string]interface{}{})
	gob.Register([]interface{}{})

	dc := gob.NewDecoder(&b)

	err = dc.Decode(&builtIn)

	if err != nil {
		i.logger.LogFatalf("Unable to deserialize the copy of Grantic's configuration created by grnc-bind. Re-run grnc-bind and re-build: %s", err.Error())
		instance.ExitError()
	}

	i.logConfigLocations(configPaths)

	fl := flm.CreateLogger(configAccessorComponentName)

	jm := config.NewJsonMerger(flm)

	mergedJson, err := jm.LoadAndMergeConfigWithBase(builtIn, configPaths)

	if err != nil {
		i.logger.LogFatalf(err.Error())
		instance.ExitError()
	}

	return &config.ConfigAccessor{mergedJson, fl}
}

// Record the files and URLs used to create a merged configuration (in the order in which they will be merged)
func (i *initiator) logConfigLocations(configPaths []string) {
	if i.logger.IsLevelEnabled(logging.Debug) {

		i.logger.LogDebugf("Loading configuration from: ")

		for _, fileName := range configPaths {
			i.logger.LogDebugf(fileName)
		}
	}
}

// Load system settings covering memory management and start/stop behaviour from configuration
func (i *initiator) loadSystemsSettings(ca *config.ConfigAccessor) *instance.System {

	s := new(instance.System)
	l := i.logger

	if ca.PathExists(systemPath) {

		if err := ca.Populate(systemPath, s); err != nil {
			l.LogFatalf("Problem loading system settings from config: " + err.Error())
			instance.ExitError()
		}

	} else {
		l.LogFatalf("Cannot find path %s in configuration.", systemPath)
		instance.ExitError()
	}

	return s
}
