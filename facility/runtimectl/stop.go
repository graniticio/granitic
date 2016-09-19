package runtimectl

import (
	"errors"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/logging"
)

const (
	stopCommandName = "stop"
	stopSummary     = "Stops one component or all components."
	stopUsage       = "stop [component] [-fw true] [-rc true]"
	stopHelp        = "Stops a component (or all components, if no component name is supplied) that implements the ioc.Stoppable interface."
	stopHelpTwo     = "If the '-fw true' argument is supplied when no component name is specified, built-in Granitic framework components will also be stopped (except for the runtime command control server)."
	stopHelpThree   = "If the '-rc true' AND '-fw true' arguments are supplied, the runtime command control server will also be stopped and no further runtime control of the application will be possible."
)

func NewStopCommand() *LifecycleCommand {

	sc := new(LifecycleCommand)

	sc.checkFunc = isStoppable
	sc.filterFunc = findStoppable
	sc.invokeFunc = invokeStop
	sc.commandHelp = []string{stopHelp, stopHelpTwo, stopHelpThree}
	sc.commandName = stopCommandName
	sc.commandSummary = stopSummary
	sc.commandUsage = stopUsage

	sc.outputPrefix = "Stopping"
	sc.noneFoundMessage = "No stoppable components found."

	return sc
}

func invokeStop(comps []*ioc.Component, l logging.Logger, cc *ioc.ComponentContainer) {
	cc.Lifecycle.StopComponents(comps)
}

func isStoppable(i interface{}) (bool, error) {

	if _, found := i.(ioc.Stoppable); found {
		return found, nil
	} else {

		return false, errors.New("Component does not implement ioc.Startable")

	}

}

func findStoppable(cc *ioc.ComponentContainer, frameworkMode bool, exclude ...string) []*ioc.Component {

	var of ownershipFilter

	if frameworkMode {
		of = FrameworkOwned
	} else {
		of = ApplicationOwned
	}

	return filteredComponents(cc, ioc.CanStop, of, true)

}
