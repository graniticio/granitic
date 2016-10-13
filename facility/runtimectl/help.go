// Copyright 2016 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package runtimectl

import (
	"fmt"
	"github.com/graniticio/granitic/ctl"
	"github.com/graniticio/granitic/logging"
	"github.com/graniticio/granitic/ws"
)

const (
	helpCommandName = "help"
	helpSummary     = "Show a list of all available commands or show help on a specific command."
	helpUsage       = "help [command]"
	helpHelp        = "Without qualifiers, help will display a list of available runtime commands. " +
		"With a command name as a qualifier, it will display detailed help and usage for that command."
	helpListHeader = "Available commands:"
)

type helpCommand struct {
	FrameworkLogger logging.Logger
	commandManager  *ctl.CommandManager
}

func (c *helpCommand) ExecuteCommand(qualifiers []string, args map[string]string) (*ctl.CommandOutput, []*ws.CategorisedError) {

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

func (c *helpCommand) detail(command ctl.Command) *ctl.CommandOutput {
	co := new(ctl.CommandOutput)
	co.RenderHint = ctl.Paragraph
	co.OutputHeader = "Command usage: " + command.Usage()
	co.OutputBody = [][]string{command.Help()}

	return co
}

func (c *helpCommand) listing() *ctl.CommandOutput {
	co := new(ctl.CommandOutput)
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

func (c *helpCommand) Name() string {
	return helpCommandName
}

func (c *helpCommand) Summmary() string {
	return helpSummary
}

func (c *helpCommand) Usage() string {
	return helpUsage
}

func (c *helpCommand) Help() []string {
	return []string{helpHelp}
}
