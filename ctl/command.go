// Copyright 2016-2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

/*
Package ctl provides functionality for the runtime control of Granitic applications.

If the RuntimeControl facility is enabled for a Granitic application (see https://granitic.io/ref/runtime-control ),
the grnc-ctl command line utility can be used to issue commands to that application via HTTP. Note that the HTTP server
and handlers instantiated to facilitate runtime control are completely separate from the HTTP server that handles user-defined
web services.

Each command is associated with a component hosted in the IoC container that implements the ctl.Command interface described
below. Granitic includes a number of built-in commands for common administration tasks, but users can create their own
commands by implementing the Command interface on any component.

*/
package ctl

import "github.com/graniticio/granitic/v2/ws"

// A Command represents an instruction that can be sent to Granitic to operate on a running instance of an application.
type Command interface {
	// ExecuteCommand is called when grnc-ctl is used to invoke a command that matches this Command's Name() method.
	// It is expected to execute whatever functionality is represented by the Command and describe the outcome in a
	// *CommandOutput. If errors are encountered, they should be formatted and returned as a slice of *ws.CategorisedError.
	ExecuteCommand(qualifiers []string, args map[string]string) (*CommandOutput, []*ws.CategorisedError)

	// A unique name for this command. Names should be short and only contain letters, numbers, underscores and dashes.
	Name() string

	// A short description (recommended less than 80 characters) briefly explaining the purpose of the Command. This text
	// will be shown when a user runs grnc-ctl help
	Summmary() string

	// A formal description of how this Command should be called e.g. start [component] [-fw true] [-rc true].
	Usage() string

	// A detailed, free-text description of the purpose of this Command. This text will be shown when a user runs
	// grnc-ctl help command-name and each element of the slice will be rendered as a new paragraph.
	Help() []string
}

// Commands is a convenience type to support sorting of slices of Commands.
type Commands []Command

// Len returns the number of Commands in the slice.
func (s Commands) Len() int { return len(s) }

// Swap swaps the position of two Commands in a slice.
func (s Commands) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// ByName is a wrapping type to allow sorting of Commands by name
type ByName struct{ Commands }

// Less returns true if the Command at index i has a name lexicographically earlier than the Command at index j.
func (s ByName) Less(i, j int) bool { return s.Commands[i].Name() < s.Commands[j].Name() }

type renderMode string

// A hint to the grnc-ctl command on how to render the output of a Command - either as paragraphs of free text or as two columns.
const (
	Columns   = "COLUMNS"
	Paragraph = "PARAGRAPH"
)

const commandError = "COMMAND_ERROR"

// A CommandOutput contains any output from a Command that should be displayed to the user of grnc-ctl.
type CommandOutput struct {
	// An optional text header that will be rendered as paragraph of text before the text of the OutputBody is displayed.
	OutputHeader string

	// Optional output text that will be rendered by grnc-ctl according to the hint in RenderHint. If column output is used,
	// each element in the outer slice is a row and each inner slice will be rendered as an indented column (up to a maxiumu of two
	// columns.
	OutputBody [][]string

	// Whether grnc-ctl should render the OutputBody as Columns or Paragraph
	RenderHint renderMode
}

// NewCommandClientError creates a new *ws.CategorisedError of type ws.Client with the supplied message.
func NewCommandClientError(message string) *ws.CategorisedError {
	return ws.NewCategorisedError(ws.Client, commandError, message)
}

// NewCommandLogicError creates a new *ws.CategorisedError of type ws.Logic with the supplied message.
func NewCommandLogicError(message string) *ws.CategorisedError {
	return ws.NewCategorisedError(ws.Logic, commandError, message)
}

// NewCommandUnexpectedError creates a new *ws.CategorisedError of type ws.Unexpected with the supplied message.
func NewCommandUnexpectedError(message string) *ws.CategorisedError {
	return ws.NewCategorisedError(ws.Unexpected, commandError, message)
}
