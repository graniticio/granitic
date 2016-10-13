// Copyright 2016 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

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

type lifecycleCommand struct {
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

func (c *lifecycleCommand) Container(container *ioc.ComponentContainer) {
	c.container = container
}

func (c *lifecycleCommand) ExecuteCommand(qualifiers []string, args map[string]string) (*ctl.CommandOutput, []*ws.CategorisedError) {

	if len(qualifiers) > 0 {
		return c.invokeSingle(qualifiers[0])
	} else {
		return c.invokeAll(args)
	}

}

func (c *lifecycleCommand) invokeAll(args map[string]string) (*ctl.CommandOutput, []*ws.CategorisedError) {

	var includeFramework bool
	var allowStopCtlServer bool
	var err error
	var sm []*ioc.Component

	if includeFramework, err = OperateOnFramework(args); err != nil {
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

	co := new(ctl.CommandOutput)
	co.OutputHeader = c.outputPrefix + ":"
	co.OutputBody = c.names(sm)
	co.RenderHint = ctl.Columns

	go c.invokeFunc(sm, c.FrameworkLogger, c.container)

	return co, nil
}

func (c *lifecycleCommand) names(cs []*ioc.Component) [][]string {
	n := make([][]string, len(cs))

	for i, c := range cs {
		n[i] = []string{c.Name}
	}

	return n
}

func (c *lifecycleCommand) invokeSingle(name string) (*ctl.CommandOutput, []*ws.CategorisedError) {

	comp := c.container.ComponentByName(name)

	if comp == nil {
		m := fmt.Sprintf("Unrecognised component %s", name)

		return nil, []*ws.CategorisedError{ctl.NewCommandClientError(m)}
	}

	if _, err := c.checkFunc(comp.Instance); err == nil {

		go c.invokeFunc([]*ioc.Component{comp}, c.FrameworkLogger, c.container)

		co := new(ctl.CommandOutput)
		co.OutputHeader = c.outputPrefix + " " + name

		return co, nil

	} else {
		return nil, []*ws.CategorisedError{ctl.NewCommandClientError(err.Error())}
	}

}

func (c *lifecycleCommand) Name() string {
	return c.commandName
}

func (c *lifecycleCommand) Summmary() string {
	return c.commandSummary
}

func (c *lifecycleCommand) Usage() string {
	return c.commandUsage
}

func (c *lifecycleCommand) Help() []string {
	return c.commandHelp
}
