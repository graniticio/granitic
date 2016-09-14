package runtimectl

import (
	"github.com/graniticio/granitic/ctl"
	"github.com/graniticio/granitic/ioc"
)

const (
	shutdownCommandName = "shutdown"
)

type ShutdownCommand struct {
	container *ioc.ComponentContainer
}

func (csd *ShutdownCommand) Container(container *ioc.ComponentContainer) {
	csd.container = container
}

func (csd *ShutdownCommand) ExecuteCommand(qualifiers []string, args map[string]string) (*ctl.CommandOutcome, error) {
	return nil, nil
}

func (csd *ShutdownCommand) Name() string {
	return shutdownCommandName
}
