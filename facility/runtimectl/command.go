package runtimectl

import (
	"github.com/graniticio/granitic/ctl"
	"github.com/graniticio/granitic/instance"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/logging"
	"github.com/graniticio/granitic/ws"
)

const (
	shutdownCommandName = "shutdown"
)

type ShutdownCommand struct {
	FrameworkLogger logging.Logger
	container       *ioc.ComponentContainer
}

func (csd *ShutdownCommand) Container(container *ioc.ComponentContainer) {
	csd.container = container
}

func (csd *ShutdownCommand) ExecuteCommand(qualifiers []string, args map[string]string) (*ctl.CommandOutcome, []*ws.CategorisedError) {

	go csd.startShutdown()

	co := new(ctl.CommandOutcome)
	co.OutputHeader = "Shutdown initiated"

	return co, nil
}

func (csd *ShutdownCommand) startShutdown() {
	csd.FrameworkLogger.LogInfof("Shutting down (runtime command)")

	csd.container.ShutdownComponents()
	instance.ExitNormal()
}

func (csd *ShutdownCommand) Name() string {
	return shutdownCommandName
}
