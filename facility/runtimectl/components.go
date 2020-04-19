// Copyright 2016-2020 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package runtimectl

import (
	"github.com/graniticio/granitic/v2/ctl"
	"github.com/graniticio/granitic/v2/ioc"
	"github.com/graniticio/granitic/v2/logging"
	"github.com/graniticio/granitic/v2/ws"
)

const (
	compCommandName = "components"
	compSummary     = "Show a list of the names of components managed by the IoC container."
	compUsage       = "components [-fw true] [-lc start|stop|suspend]"
	compHelp        = "Lists the name of all of the user-defined components currently present in the IoC Container."
	compHelpTwo     = "If the '-fw true' argument is supplied, the list will show built-in Granitic framework components instead of user-defined components."
	compHelpThree   = "If the '-lc true' argument is supplied with one of the values start/stop/suspend then only those components that implement the corresponding " +
		"lifecycle interface (ioc.Startable, ioc.Stoppable, ioc.Suspendable) will be displayed"
)

type componentsCommand struct {
	FrameworkLogger logging.Logger
	container       *ioc.ComponentContainer
}

func (c *componentsCommand) Container(container *ioc.ComponentContainer) {
	c.container = container
}

func (c *componentsCommand) ExecuteCommand(qualifiers []string, args map[string]string) (*ctl.CommandOutput, []*ws.CategorisedError) {

	var frameworkOnly bool
	var lcFilter ioc.LifecycleSupport
	var err error

	if frameworkOnly, err = OperateOnFramework(args); err != nil {
		return nil, []*ws.CategorisedError{ctl.NewCommandClientError(err.Error())}
	}

	if lcFilter, err = findLifecycleFilter(args); err != nil {
		return nil, []*ws.CategorisedError{ctl.NewCommandClientError(err.Error())}
	}

	names := make([][]string, 0)

	for _, c := range c.container.AllComponents() {

		fw := isFramework(c)

		if ((fw && frameworkOnly) || (!fw && !frameworkOnly)) && matchesFilter(lcFilter, c.Instance) {
			names = append(names, []string{c.Name})
		}
	}

	co := new(ctl.CommandOutput)
	co.OutputBody = names
	co.RenderHint = ctl.Columns

	return co, nil
}

func (c *componentsCommand) Name() string {
	return compCommandName
}

func (c *componentsCommand) Summmary() string {
	return compSummary
}

func (c *componentsCommand) Usage() string {
	return compUsage
}

func (c *componentsCommand) Help() []string {
	return []string{compHelp, compHelpTwo, compHelpThree}
}
