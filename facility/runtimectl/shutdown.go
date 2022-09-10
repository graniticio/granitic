// Copyright 2016-2022 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package runtimectl

import (
	"github.com/graniticio/granitic/v3/ctl"
	"github.com/graniticio/granitic/v3/instance"
	"github.com/graniticio/granitic/v3/ioc"
	"github.com/graniticio/granitic/v3/logging"
	"github.com/graniticio/granitic/v3/ws"
)

const (
	shutdownCommandName = "shutdown"
	shutdownSummary     = "Stops all components then exits the application."
	shutdownUsage       = "shutdown"
	shutdownHelp        = "Causes the IoC container to stop each component according to the lifecyle interfaces they implement. " +
		"The Granitic application will exit once all components have stopped."
)

type shutdownCommand struct {
	FrameworkLogger logging.Logger
	container       *ioc.ComponentContainer
	disableExit     bool
}

func (csd *shutdownCommand) Container(container *ioc.ComponentContainer) {
	csd.container = container
}

func (csd *shutdownCommand) ExecuteCommand(qualifiers []string, args map[string]string) (*ctl.CommandOutput, []*ws.CategorisedError) {

	go csd.startShutdown()

	co := new(ctl.CommandOutput)
	co.OutputHeader = "Shutdown initiated"

	return co, nil
}

func (csd *shutdownCommand) startShutdown() {
	csd.FrameworkLogger.LogInfof("Shutting down (runtime command)")

	csd.container.Lifecycle.StopAll()

	if !csd.disableExit {
		instance.ExitNormal()
	} else {
		csd.FrameworkLogger.LogInfof("Application exit disabled")
	}
}

func (csd *shutdownCommand) Name() string {
	return shutdownCommandName
}

func (csd *shutdownCommand) Summmary() string {
	return shutdownSummary
}

func (csd *shutdownCommand) Usage() string {
	return shutdownUsage
}

func (csd *shutdownCommand) Help() []string {
	return []string{shutdownHelp}
}
