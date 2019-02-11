// Copyright 2016-2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package logger

import (
	"fmt"
	"github.com/graniticio/granitic/v2/ctl"
	"github.com/graniticio/granitic/v2/facility/runtimectl"
	"github.com/graniticio/granitic/v2/instance"
	"github.com/graniticio/granitic/v2/logging"
	"github.com/graniticio/granitic/v2/ws"
)

const (
	GLLComponentName = instance.FrameworkPrefix + "CommandGlobalLevel"
	gllCommandName   = "global-level"
	gllSummary       = "Views or sets the global logging threshold for application or framework components."
	gllUsage         = "global-level [level] [-fw true]"
	gllHelp          = "With no qualifier, this command shows the current global log threshold for application components. When a level is specified, the global log threshold will be set to the supplied level."
	gllHelpTwo       = "If the '-fw true' argument is supplied, the command will display or set the built-in Granitic framework's global log threshold."
)

type globalLogLevelCommand struct {
	FrameworkLogger    logging.Logger
	FrameworkManager   *logging.ComponentLoggerManager
	ApplicationManager *logging.ComponentLoggerManager
}

func (c *globalLogLevelCommand) ExecuteCommand(qualifiers []string, args map[string]string) (*ctl.CommandOutput, []*ws.CategorisedError) {

	if len(qualifiers) == 0 {
		return c.showCurrentLevel(args)
	}

	return c.setLevel(qualifiers[0], args)
}

func (c *globalLogLevelCommand) setLevel(label string, args map[string]string) (*ctl.CommandOutput, []*ws.CategorisedError) {
	var err error
	var setFramework bool
	var ll logging.LogLevel

	if setFramework, err = runtimectl.OperateOnFramework(args); err != nil {
		return nil, []*ws.CategorisedError{ctl.NewCommandClientError(err.Error())}
	}

	if ll, err = logging.LogLevelFromLabel(label); err != nil {
		return nil, []*ws.CategorisedError{ctl.NewCommandClientError(err.Error())}
	}

	if setFramework {
		c.FrameworkManager.SetGlobalThreshold(ll)
	} else {
		c.ApplicationManager.SetGlobalThreshold(ll)
	}

	return new(ctl.CommandOutput), nil
}

func (c *globalLogLevelCommand) showCurrentLevel(args map[string]string) (*ctl.CommandOutput, []*ws.CategorisedError) {

	var m string
	var err error
	var showFrameworkThreshold bool

	if showFrameworkThreshold, err = runtimectl.OperateOnFramework(args); err != nil {
		return nil, []*ws.CategorisedError{ctl.NewCommandClientError(err.Error())}
	}

	if showFrameworkThreshold {

		cl := c.FrameworkManager.GlobalLevel()
		label := logging.LabelFromLevel(cl)

		m = fmt.Sprintf("Global logging threshold for Granitic framwork components is set to %s", label)

	} else {

		cl := c.ApplicationManager.GlobalLevel()
		label := logging.LabelFromLevel(cl)

		m = fmt.Sprintf("Global logging threshold for application components is set to %s", label)

	}

	co := new(ctl.CommandOutput)
	co.OutputHeader = m

	return co, nil

}

func (c *globalLogLevelCommand) Name() string {
	return gllCommandName
}

func (c *globalLogLevelCommand) Summmary() string {
	return gllSummary
}

func (c *globalLogLevelCommand) Usage() string {
	return gllUsage
}

func (c *globalLogLevelCommand) Help() []string {
	return []string{gllHelp, gllHelpTwo}
}
