// Copyright 2016-2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package runtimectl

import (
	"errors"
	"github.com/graniticio/granitic/v2/ioc"
	"github.com/graniticio/granitic/v2/logging"
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

// Create a variant of the lifecycleCommand configured as the suspend command.
func newSuspendCommand() *lifecycleCommand {

	sc := new(lifecycleCommand)

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

// Create a variant of the lifecycleCommand configured as the resume command.
func newResumeCommand() *lifecycleCommand {

	sc := new(lifecycleCommand)

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
	}

	return false, errors.New("component does not implement ioc.Suspendable")
}

func findSuspendable(cc *ioc.ComponentContainer, frameworkMode bool, exclude ...string) []*ioc.Component {

	var of ownershipFilter

	if frameworkMode {
		of = All
	} else {
		of = ApplicationOwned
	}

	return filteredComponents(cc, ioc.CanSuspend, of, true, exclude...)

}
