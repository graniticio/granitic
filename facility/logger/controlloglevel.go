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
	"sort"
)

const (
	// LogLevelComponentName is the name of the component able to alter log levels at runtime
	LogLevelComponentName = instance.FrameworkPrefix + "CommandLogLevel"
	llCommandName         = "log-level"
	llSummary             = "Views or sets a specific logging threshold for application or framework components."
	llUsage               = "log-level [component level] [-fw true]"
	llHelp                = "With no qualifier, this command shows a list of application components that have a specific logging threshold set. When a " +
		"component and a level are specified as qualifiers, the component's logging threshold is set at the specified level."
	llHelpTwo   = "Valid values for level are ALL, TRACE, DEBUG, INFO, WARN, ERROR, FATAL (case insensitive)."
	llHelpThree = "Setting the level to ALL has special behaviour - it efffectively removes the specific logging threshold for the component. The global log threshold will then apply to that component."
	llHelpFour  = "If the '-fw true' argument is supplied without qualifiers, a list of built-in framework components and their associated log levels will be shown."
)

type logLevelCommand struct {
	FrameworkLogger    logging.Logger
	FrameworkManager   *logging.ComponentLoggerManager
	ApplicationManager *logging.ComponentLoggerManager
}

func (c *logLevelCommand) ExecuteCommand(qualifiers []string, args map[string]string) (*ctl.CommandOutput, []*ws.CategorisedError) {

	if len(qualifiers) == 0 {
		return c.showCurrentLevel(args)
	}

	return c.setLevel(qualifiers)
}

func (c *logLevelCommand) setLevel(qualifiers []string) (*ctl.CommandOutput, []*ws.CategorisedError) {

	var err error
	var ll logging.LogLevel

	if len(qualifiers) < 2 {
		return nil, []*ws.CategorisedError{ctl.NewCommandClientError("You must provide a component name and a loging level.")}
	}

	name := qualifiers[0]
	label := qualifiers[1]

	if ll, err = logging.LogLevelFromLabel(label); err != nil {
		return nil, []*ws.CategorisedError{ctl.NewCommandClientError(err.Error())}
	}

	logger := c.ApplicationManager.LoggerByName(name)

	if logger == nil {
		logger = c.FrameworkManager.LoggerByName(name)
	}

	if logger == nil {
		m := fmt.Sprintf("Component %s does not exist or does not have a logger attached.", name)
		return nil, []*ws.CategorisedError{ctl.NewCommandClientError(m)}
	}

	logger.SetLocalThreshold(ll)

	return new(ctl.CommandOutput), nil

}

func (c *logLevelCommand) showCurrentLevel(args map[string]string) (*ctl.CommandOutput, []*ws.CategorisedError) {

	var comps []*logging.ComponentLevel
	var err error
	var frameworkLevels bool

	if frameworkLevels, err = runtimectl.OperateOnFramework(args); err != nil {
		return nil, []*ws.CategorisedError{ctl.NewCommandClientError(err.Error())}
	}

	if frameworkLevels {
		comps = c.FrameworkManager.CurrentLevels()
	} else {
		comps = c.ApplicationManager.CurrentLevels()
	}

	sort.Sort(logging.ByName{ComponentLevels: comps})

	filtered := make([][]string, 0)

	for _, cl := range comps {

		if cl.Level != logging.All {
			filtered = append(filtered, []string{cl.Name, logging.LabelFromLevel(cl.Level)})
		}

	}

	co := new(ctl.CommandOutput)
	co.OutputBody = filtered
	co.RenderHint = ctl.Columns

	return co, nil
}

func (c *logLevelCommand) Name() string {
	return llCommandName
}

func (c *logLevelCommand) Summmary() string {
	return llSummary
}

func (c *logLevelCommand) Usage() string {
	return llUsage
}

func (c *logLevelCommand) Help() []string {
	return []string{llHelp, llHelpTwo, llHelpThree, llHelpFour}
}
