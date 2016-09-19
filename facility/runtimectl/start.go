package runtimectl

import (
	"errors"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/logging"
)

const (
	startCommandName = "start"
	startSummary     = "Starts one component or all components."
	startUsage       = "start [component] [-fw true] [-rc true]"
	startHelp        = "Starts a component (or all components, if no component name is supplied) that implements the ioc.Startable interface."
	startHelpTwo     = "If the '-fw true' argument is supplied when no component name is specified, built-in Granitic framework components will also be started."
)

func NewStartCommand() *LifecycleCommand {

	sc := new(LifecycleCommand)

	sc.checkFunc = isStartable
	sc.filterFunc = findStartable
	sc.invokeFunc = invokeStart
	sc.commandHelp = []string{startHelp, startHelpTwo}
	sc.commandName = startCommandName
	sc.commandSummary = startSummary
	sc.commandUsage = startUsage

	sc.outputPrefix = "Starting"
	sc.noneFoundMessage = "No startable components found."

	return sc
}

func invokeStart(comps []*ioc.Component, l logging.Logger, cc *ioc.ComponentContainer) {

	defer func() {
		if r := recover(); r != nil {
			l.LogErrorfWithTrace("Panic recovered while starting components components %s", r)
		}
	}()

	if err := cc.Lifecycle.Start(comps); err != nil {
		l.LogErrorf("Problem starting components from remote command", err.Error())
	}

}

func isStartable(i interface{}) (bool, error) {

	if _, found := i.(ioc.Startable); found {
		return found, nil
	} else {

		return false, errors.New("Component does not implement ioc.Startable")

	}

}

func findStartable(cc *ioc.ComponentContainer, frameworkMode bool, exclude ...string) []*ioc.Component {

	var of ownershipFilter

	if frameworkMode {
		of = FrameworkOwned
	} else {
		of = ApplicationOwned
	}

	return filteredComponents(cc, ioc.CanStart, of, true)

}
