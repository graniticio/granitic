package runtimectl

import (
	"fmt"
	"github.com/graniticio/granitic/ctl"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/logging"
	"github.com/graniticio/granitic/ws"
)

const (
	stopCommandName = "stop"
	stopSummary     = "Stops one component or all components."
	stopUsage       = "stop [component] [-fw true] [-rc true]"
	stopHelp        = "Stops a component (or all components, if no component name is supplied) that implements the ioc.Stoppable interface."
	stopHelpTwo     = "If the '-fw true' argument is supplied when no component name is specified, built-in Granitic framework components will also be stopped (except for the runtime command control server)."
	stopHelpThree   = "If the '-rc true' AND '-fw true' arguments are supplied, the runtime command control server will also be stopped and no further runtime control of the application will be possible."
)

type StopCommand struct {
	FrameworkLogger logging.Logger
	container       *ioc.ComponentContainer
}

func (c *StopCommand) Container(container *ioc.ComponentContainer) {
	c.container = container
}

func (c *StopCommand) ExecuteCommand(qualifiers []string, args map[string]string) (*ctl.CommandOutcome, []*ws.CategorisedError) {

	if len(qualifiers) > 0 {
		return c.stopSingle(qualifiers[0])
	} else {
		return c.stopAll(args)
	}

}

func (c *StopCommand) stopAll(args map[string]string) (*ctl.CommandOutcome, []*ws.CategorisedError) {

	var includeFramework bool
	var allowStopCtlServer bool
	var err error

	if includeFramework, err = operateOnFramework(args); err != nil {
		return nil, []*ws.CategorisedError{ctl.NewCommandClientError(err.Error())}
	}

	sm := make(map[string]ioc.Stoppable)

	names := make([][]string, 0)

	if allowStopCtlServer, err = includeRuntime(args); err != nil {
		return nil, []*ws.CategorisedError{ctl.NewCommandClientError(err.Error())}
	}

	for _, c := range c.container.AllComponents() {

		fw := isFramework(c)

		if ((fw && includeFramework) || !fw) && matchesFilter(ioc.CanStop, c.Instance) {

			if c.Name == RuntimeCtlServer && !allowStopCtlServer {
				continue
			}

			sm[c.Name] = c.Instance.(ioc.Stoppable)
			names = append(names, []string{c.Name})
		}
	}

	if len(names) == 0 {
		return nil, []*ws.CategorisedError{ctl.NewCommandClientError("No stoppable components found.")}
	}

	co := new(ctl.CommandOutcome)
	co.OutputHeader = "Stopping:"
	co.OutputBody = names
	co.RenderHint = ctl.Columns

	go c.runStop(sm)

	return co, nil
}

func (c *StopCommand) stopSingle(name string) (*ctl.CommandOutcome, []*ws.CategorisedError) {

	comp := c.container.ComponentByName(name)

	if comp == nil {
		m := fmt.Sprintf("Unrecognised component %s", name)

		return nil, []*ws.CategorisedError{ctl.NewCommandClientError(m)}
	}

	if s, found := comp.Instance.(ioc.Stoppable); found {

		sm := make(map[string]ioc.Stoppable)
		sm[name] = s
		go c.runStop(sm)

		co := new(ctl.CommandOutcome)
		co.OutputHeader = "Stopping " + name

		return co, nil

	} else {
		m := fmt.Sprintf("Component %s is not stoppable (does not implement ioc.Stoppable)", name)

		return nil, []*ws.CategorisedError{ctl.NewCommandClientError(m)}
	}

}

func (c *StopCommand) runStop(comps map[string]ioc.Stoppable) {

	err := c.container.StopComponents(comps)

	if err != nil {
		c.FrameworkLogger.LogErrorf("Problem stopping components " + err.Error())
	}

}

func (c *StopCommand) Name() string {
	return stopCommandName
}

func (c *StopCommand) Summmary() string {
	return stopSummary
}

func (c *StopCommand) Usage() string {
	return stopUsage
}

func (c *StopCommand) Help() []string {
	return []string{stopHelp, stopHelpTwo, stopHelpThree}
}
