package runtimectl

import (
	"fmt"
	"github.com/graniticio/granitic/ctl"
	"github.com/graniticio/granitic/logging"
	"github.com/graniticio/granitic/ws"
)

const (
	helpCommandName = "help"
	helpSummary     = "Show all available commands or show help on a specific command"
	helpUsage       = "help [command]"
	helpHelp        = "Without qualifiers, help will display a list of available runtime commands. " +
		"With a command name as a qualifier, it will display detailed help and usage for that command."
	helpListHeader = "Available commands:"
)

type HelpCommand struct {
	FrameworkLogger logging.Logger
	commandManager  *ctl.CommandManager
}

func (c *HelpCommand) ExecuteCommand(qualifiers []string, args map[string]string) (*ctl.CommandOutcome, []*ws.CategorisedError) {

	if len(qualifiers) > 0 {
		cn := qualifiers[0]

		command := c.commandManager.Find(cn)

		if command == nil {
			m := fmt.Sprintf("%s is not a recognised command. Run help without qualifiers to see a list of all available commands.", cn)
			e := ctl.NewCommandClientError(m)

			return nil, []*ws.CategorisedError{e}

		}

		return c.detail(command), nil

	}

	return c.listing(), nil
}

func (c *HelpCommand) detail(command ctl.Command) *ctl.CommandOutcome {
	co := new(ctl.CommandOutcome)
	co.RenderHint = ctl.Paragraph
	co.OutputHeader = "Command usage: " + command.Usage()
	co.OutputBody = [][]string{command.Help()}

	return co
}

func (c *HelpCommand) listing() *ctl.CommandOutcome {
	co := new(ctl.CommandOutcome)
	co.RenderHint = ctl.Columns
	co.OutputHeader = helpListHeader

	all := c.commandManager.All()
	la := len(all)

	o := make([][]string, la)

	for i := 0; i < la; i++ {

		o[i] = []string{all[i].Name(), all[i].Summmary()}

	}

	co.OutputBody = o

	return co

}

func (c *HelpCommand) Name() string {
	return helpCommandName
}

func (c *HelpCommand) Summmary() string {
	return helpSummary
}

func (c *HelpCommand) Usage() string {
	return helpUsage
}

func (c *HelpCommand) Help() []string {
	return []string{helpHelp}
}
