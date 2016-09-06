package granitic

import (
	"github.com/graniticio/granitic/config"
	"github.com/graniticio/granitic/facility"
	"github.com/graniticio/granitic/facility/jsonmerger"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/logging"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

const (
	initiatorComponentName           string = ioc.FrameworkPrefix + "FrameworkInitiator"
	jsonMergerComponentName          string = ioc.FrameworkPrefix + "JsonMerger"
	configAccessorComponentName      string = ioc.FrameworkPrefix + "ConfigAccessor"
	facilityInitialisorComponentName string = ioc.FrameworkPrefix + "FacilityInitialisor"
)

func StartGranitic(customComponents *ioc.ProtoComponents) {
	i := new(Initiator)

	is := config.InitialSettingsFromEnvironment()

	i.Start(customComponents, is)
}

type Initiator struct {
	logger logging.Logger
}

func (i *Initiator) Start(customComponents *ioc.ProtoComponents, is *config.InitialSettings) {

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

func (i *Initiator) buildContainer(ac *ioc.ProtoComponents, is *config.InitialSettings) *ioc.ComponentContainer {

	start := time.Now()

	frameworkLoggingManager, logManageProto := facility.BootstrapFrameworkLogging(is.FrameworkLogLevel)
	i.logger = frameworkLoggingManager.CreateLogger(initiatorComponentName)

	i.logger.LogInfof("Starting components")

	configAccessor := i.loadConfigIntoAccessor(is.Configuration, frameworkLoggingManager)
	c := ioc.NewContainer(frameworkLoggingManager, configAccessor)

	c.AddProto(logManageProto)
	c.AddProtos(ac.Components)
	c.AddModifiers(ac.FrameworkDependencies)

	fi := facility.NewFacilitiesInitialisor(c, frameworkLoggingManager)
	fi.Logger = frameworkLoggingManager.CreateLogger(facilityInitialisorComponentName)

	err := fi.Initialise(configAccessor)
	i.shutdownIfError(err, c)

	err = c.Populate()
	i.shutdownIfError(err, c)

	runtime.GC()

	err = c.StartComponents()
	i.shutdownIfError(err, c)

	elapsed := time.Since(start)
	i.logger.LogInfof("Ready (startup time %s)", elapsed)

	return c
}

func (i *Initiator) shutdownIfError(err error, c *ioc.ComponentContainer) {

	if err != nil {
		i.logger.LogFatalf(err.Error())
		i.shutdown(c)
		os.Exit(-1)
	}

}

func (i *Initiator) shutdown(container *ioc.ComponentContainer) {
	i.logger.LogInfof("Shutting down")

	container.ShutdownComponents()

}

func (i *Initiator) loadConfigIntoAccessor(configPaths []string, lm *logging.ComponentLoggerManager) *config.ConfigAccessor {

	fl := lm.CreateLogger(configAccessorComponentName)

	if i.logger.IsLevelEnabled(logging.Debug) {

		i.logger.LogDebugf("Loading configuration from: ")

		for _, fileName := range configPaths {
			i.logger.LogDebugf(fileName)
		}
	}

	jm := new(jsonmerger.JsonMerger)
	jm.Logger = lm.CreateLogger(jsonMergerComponentName)

	mergedJson := jm.LoadAndMergeConfig(configPaths)

	return &config.ConfigAccessor{mergedJson, fl}
}
