package runtimectl

import (
	"github.com/graniticio/granitic/ctl"
	"github.com/graniticio/granitic/instance"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/logging"
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

func (csd *ShutdownCommand) ExecuteCommand(qualifiers []string, args map[string]string) (*ctl.CommandOutcome, error) {

	go csd.startShutdown()

	return nil, nil
}

func (csd *ShutdownCommand) startShutdown() {
	csd.FrameworkLogger.LogInfof("Shutting down (runtime command)")

	csd.container.ShutdownComponents()
	instance.ExitNormal()
}

func (csd *ShutdownCommand) Name() string {
	return shutdownCommandName
}
