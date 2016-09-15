package runtimectl

import (
	"github.com/graniticio/granitic/ctl"
	"github.com/graniticio/granitic/logging"
	"github.com/graniticio/granitic/ws"
)

const (
	helpCommandName = "help"
	helpSummary     = "List of all available commands or help on a specific command"
	helpUsage       = "help [command]"
	helpHelp        = "Without qualifiers, help will display a list of available runtime commands. " +
		"With a command name as a qualifier, it will display detailled help and usage for that command"
)

type HelpCommand struct {
	FrameworkLogger logging.Logger
	commandManager  *ctl.CommandManager
}

func (csd *HelpCommand) ExecuteCommand(qualifiers []string, args map[string]string) (*ctl.CommandOutcome, []*ws.CategorisedError) {

	co := new(ctl.CommandOutcome)

	return co, nil
}

func (csd *HelpCommand) Name() string {
	return helpCommandName
}

func (csd *HelpCommand) Summmary() string {
	return helpSummary
}

func (csd *HelpCommand) Usage() string {
	return helpUsage
}

func (csd *HelpCommand) Help() []string {
	return []string{helpHelp}
}
