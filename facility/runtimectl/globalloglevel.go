package runtimectl

import (
	"fmt"
	"github.com/graniticio/granitic/ctl"
	"github.com/graniticio/granitic/instance"
	"github.com/graniticio/granitic/logging"
	"github.com/graniticio/granitic/ws"
)

const (
	GLLComponentName = instance.FrameworkPrefix + "CommandGlobalLevel"
	gllCommandName   = "global-level"
	gllSummary       = "Views or sets the global logging threshold for application or framework components."
	gllUsage         = "global-level [level] [-fw true]"
	gllHelp          = "With no qualifier, this command shows the current global log threshold for application components. When a level is specified, the global log threshold will be set to the supplied level."
	gllHelpTwo       = "If the '-fw true' argument is supplied, the command will display or set the built-in Granitic framework's global log threshold."
)

type GlobalLogLevelCommand struct {
	FrameworkLogger    logging.Logger
	FrameworkManager   *logging.ComponentLoggerManager
	ApplicationManager *logging.ComponentLoggerManager
}

func (c *GlobalLogLevelCommand) ExecuteCommand(qualifiers []string, args map[string]string) (*ctl.CommandOutcome, []*ws.CategorisedError) {

	if len(qualifiers) == 0 {
		return c.showCurrentLevel(args)
	} else {
		return c.setLevel(qualifiers[0], args)
	}

	return nil, nil
}

func (c *GlobalLogLevelCommand) setLevel(label string, args map[string]string) (*ctl.CommandOutcome, []*ws.CategorisedError) {
	var err error
	var setFramework bool
	var ll logging.LogLevel

	if setFramework, err = operateOnFramework(args); err != nil {
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

	return new(ctl.CommandOutcome), nil
}

func (c *GlobalLogLevelCommand) showCurrentLevel(args map[string]string) (*ctl.CommandOutcome, []*ws.CategorisedError) {

	var m string
	var err error
	var showFrameworkThreshold bool

	if showFrameworkThreshold, err = operateOnFramework(args); err != nil {
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

	co := new(ctl.CommandOutcome)
	co.OutputHeader = m

	return co, nil

}

func (c *GlobalLogLevelCommand) Name() string {
	return gllCommandName
}

func (c *GlobalLogLevelCommand) Summmary() string {
	return gllSummary
}

func (c *GlobalLogLevelCommand) Usage() string {
	return gllUsage
}

func (c *GlobalLogLevelCommand) Help() []string {
	return []string{gllHelp, gllHelpTwo}
}
