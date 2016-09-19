package runtimectl

import (
	"fmt"
	"github.com/graniticio/granitic/ctl"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/logging"
	"github.com/graniticio/granitic/ws"
)

type ownershipFilter int

const (
	All = iota
	FrameworkOwned
	ApplicationOwned
)

type invokeOperation func([]*ioc.Component, logging.Logger, *ioc.ComponentContainer)
type filterComponents func(*ioc.ComponentContainer, bool, ...string) []*ioc.Component
type supportsOperation func(interface{}) (bool, error)

type LifecycleCommand struct {
	FrameworkLogger  logging.Logger
	container        *ioc.ComponentContainer
	invokeFunc       invokeOperation
	checkFunc        supportsOperation
	filterFunc       filterComponents
	commandName      string
	commandSummary   string
	commandUsage     string
	commandHelp      []string
	outputPrefix     string
	noneFoundMessage string
}

func (c *LifecycleCommand) Container(container *ioc.ComponentContainer) {
	c.container = container
}

func (c *LifecycleCommand) ExecuteCommand(qualifiers []string, args map[string]string) (*ctl.CommandOutcome, []*ws.CategorisedError) {

	if len(qualifiers) > 0 {
		return c.invokeSingle(qualifiers[0])
	} else {
		return c.invokeAll(args)
	}

}

func (c *LifecycleCommand) invokeAll(args map[string]string) (*ctl.CommandOutcome, []*ws.CategorisedError) {

	var includeFramework bool
	var allowStopCtlServer bool
	var err error
	var sm []*ioc.Component

	if includeFramework, err = operateOnFramework(args); err != nil {
		return nil, []*ws.CategorisedError{ctl.NewCommandClientError(err.Error())}
	}

	if allowStopCtlServer, err = includeRuntime(args); err != nil {
		return nil, []*ws.CategorisedError{ctl.NewCommandClientError(err.Error())}
	} else if allowStopCtlServer {
		sm = c.filterFunc(c.container, includeFramework)
	} else {
		sm = c.filterFunc(c.container, includeFramework, RuntimeCtlServer)
	}

	if len(sm) == 0 {
		return nil, []*ws.CategorisedError{ctl.NewCommandClientError(c.noneFoundMessage)}
	}

	co := new(ctl.CommandOutcome)
	co.OutputHeader = c.outputPrefix + ":"
	co.OutputBody = c.names(sm)
	co.RenderHint = ctl.Columns

	go c.invokeFunc(sm, c.FrameworkLogger, c.container)

	return co, nil
}

func (c *LifecycleCommand) names(cs []*ioc.Component) [][]string {
	n := make([][]string, len(cs))

	for i, c := range cs {
		n[i] = []string{c.Name}
	}

	return n
}

func (c *LifecycleCommand) invokeSingle(name string) (*ctl.CommandOutcome, []*ws.CategorisedError) {

	comp := c.container.ComponentByName(name)

	if comp == nil {
		m := fmt.Sprintf("Unrecognised component %s", name)

		return nil, []*ws.CategorisedError{ctl.NewCommandClientError(m)}
	}

	if _, err := c.checkFunc(comp.Instance); err == nil {

		go c.invokeFunc([]*ioc.Component{comp}, c.FrameworkLogger, c.container)

		co := new(ctl.CommandOutcome)
		co.OutputHeader = c.outputPrefix + " " + name

		return co, nil

	} else {
		return nil, []*ws.CategorisedError{ctl.NewCommandClientError(err.Error())}
	}

}

func (c *LifecycleCommand) Name() string {
	return c.commandName
}

func (c *LifecycleCommand) Summmary() string {
	return c.commandSummary
}

func (c *LifecycleCommand) Usage() string {
	return c.commandUsage
}

func (c *LifecycleCommand) Help() []string {
	return c.commandHelp
}
