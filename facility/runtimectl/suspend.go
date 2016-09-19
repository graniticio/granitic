package runtimectl

import (
	"errors"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/logging"
)

const (
	suspendCommandName = "suspend"
	suspendSummary     = "Suspends one component or all components."
	suspendUsage       = "suspend [component] [-fw true] [-rc true]"
	suspendHelp        = "Suspends a component (or all components, if no component name is supplied) that implements the ioc.Suspendable interface."
	suspendHelpTwo     = "If the '-fw true' argument is supplied when no component name is specified, built-in Granitic framework components will also be suspended (except for the runtime command control server)."
	suspendHelpThree   = "If the '-rc true' AND '-fw true' arguments are supplied, the runtime command control server will also be suspended and no further runtime control of the application will be possible."

	resumeCommandName = "resume"
	resumeSummary     = "Resumes one component or all components that have previously been suspended."
	resumeUsage       = "resume [component] [-fw true] [-rc true]"
	resumeHelp        = "Resumes a component (or all components, if no component name is supplied) that implements the ioc.Suspendable interface."
	resumeHelpTwo     = "If the '-fw true' argument is supplied when no component name is specified, built-in Granitic framework components will also be resumed."
)

func NewSuspendCommand() *LifecycleCommand {

	sc := new(LifecycleCommand)

	sc.checkFunc = isSuspendable
	sc.filterFunc = findSuspendable
	sc.invokeFunc = invokeSuspendable
	sc.commandHelp = []string{suspendHelp, suspendHelpTwo, suspendHelpThree}
	sc.commandName = suspendCommandName
	sc.commandSummary = suspendSummary
	sc.commandUsage = suspendUsage

	sc.outputPrefix = "Suspending"
	sc.noneFoundMessage = "No suspendable components found."

	return sc
}

func NewResumeCommand() *LifecycleCommand {

	sc := new(LifecycleCommand)

	sc.checkFunc = isSuspendable
	sc.filterFunc = findSuspendable
	sc.invokeFunc = invokeResumable
	sc.commandHelp = []string{resumeHelp, resumeHelpTwo}
	sc.commandName = resumeCommandName
	sc.commandSummary = resumeSummary
	sc.commandUsage = resumeUsage

	sc.outputPrefix = "Resuming"
	sc.noneFoundMessage = "No resumable components found."

	return sc
}

func invokeResumable(comps []*ioc.Component, l logging.Logger, cc *ioc.ComponentContainer) {

	defer func() {
		if r := recover(); r != nil {
			l.LogErrorfWithTrace("Panic recovered while resuming components %s", r)
		}
	}()

	if err := cc.Lifecycle.ResumeComponents(comps); err != nil {
		l.LogErrorf("Problem resuming components from remote command", err.Error())
	}
}

func invokeSuspendable(comps []*ioc.Component, l logging.Logger, cc *ioc.ComponentContainer) {

	defer func() {
		if r := recover(); r != nil {
			l.LogErrorfWithTrace("Panic recovered while suspending components %s", r)
		}
	}()

	if err := cc.Lifecycle.SuspendComponents(comps); err != nil {
		l.LogErrorf("Problem suspending components from remote command", err.Error())
	}
}

func isSuspendable(i interface{}) (bool, error) {

	if _, found := i.(ioc.Suspendable); found {
		return found, nil
	} else {

		return false, errors.New("Component does not implement ioc.Suspendable")

	}

}

func findSuspendable(cc *ioc.ComponentContainer, frameworkMode bool, exclude ...string) []*ioc.Component {

	var of ownershipFilter

	if frameworkMode {
		of = FrameworkOwned
	} else {
		of = ApplicationOwned
	}

	return filteredComponents(cc, ioc.CanSuspend, of, true)

}
