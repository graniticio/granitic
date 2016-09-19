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
	shutdownSummary     = "Stops all components then exits the application."
	shutdownUsage       = "shutdown"
	shutdownHelp        = "Causes the IoC container to stop each component according to the lifecyle interfaces they implement. " +
		"The Granitic application will exit once all components have stopped."
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

	csd.container.Lifecycle.StopAll()
	instance.ExitNormal()
}

func (csd *ShutdownCommand) Name() string {
	return shutdownCommandName
}

func (csd *ShutdownCommand) Summmary() string {
	return shutdownSummary
}

func (csd *ShutdownCommand) Usage() string {
	return shutdownUsage
}

func (csd *ShutdownCommand) Help() []string {
	return []string{shutdownHelp}
}
