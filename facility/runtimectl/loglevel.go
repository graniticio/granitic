package runtimectl

import (
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

	return nil, nil
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
