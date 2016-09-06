/*
Package granitic provides methods for starting an instance of Granitic.
*/
package granitic

import (
	"github.com/graniticio/granitic/config"
	"github.com/graniticio/granitic/facility"
	"github.com/graniticio/granitic/facility/jsonmerger"
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
	initiatorComponentName  string = instance.FrameworkPrefix + "FrameworkInitiator"
	jsonMergerComponentName string = instance.FrameworkPrefix + "JsonMerger"
)

func StartGranitic(customComponents *ioc.ProtoComponents) {
	i := new(initiator)

	is := config.InitialSettingsFromEnvironment()

	i.Start(customComponents, is)
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
		config.ExitNormal()
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
	ca := i.createConfigAccessor(is.Configuration, frameworkLoggingManager)
	cc := ioc.NewComponentContainer(frameworkLoggingManager, ca)
	cc.AddProto(logManageProto)

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

	//Proto components and config no longer needed
	runtime.GC()

	//Start all startable components
	err = cc.StartComponents()
	i.shutdownIfError(err, cc)

	elapsed := time.Since(is.StartTime)
	l.LogInfof("Ready (startup time %s)", elapsed)

	return cc
}

// Cleanly stop the container and any running components in the event of an error
// during startup.
func (i *initiator) shutdownIfError(err error, cc *ioc.ComponentContainer) {

	if err != nil {
		i.logger.LogFatalf(err.Error())
		i.shutdown(cc)
		config.ExitError()
	}

}

// Log that the container is stopping and let the container stop its
// components gracefully
func (i *initiator) shutdown(cc *ioc.ComponentContainer) {
	i.logger.LogInfof("Shutting down")

	cc.ShutdownComponents()
}

// Merge together all of the local and remote JSON configuration files and wrap them in a *config.ConfigAccessor
// which allows programmatic access to the merged config.
func (i *initiator) createConfigAccessor(configPaths []string, lm *logging.ComponentLoggerManager) *config.ConfigAccessor {

	i.logConfigLocations(configPaths)

	fl := lm.CreateLogger(config.ConfigAccessorComponentName)

	jm := new(jsonmerger.JsonMerger)
	jm.Logger = lm.CreateLogger(jsonMergerComponentName)

	mergedJson := jm.LoadAndMergeConfig(configPaths)

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
