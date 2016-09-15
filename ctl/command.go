package ctl

import "github.com/graniticio/granitic/ws"

type Command interface {
	ExecuteCommand(qualifiers []string, args map[string]string) (*CommandOutcome, []*ws.CategorisedError)
	Name() string
	Summmary() string
	Usage() string
	Help() []string
}

type Commands []Command

func (s Commands) Len() int      { return len(s) }
func (s Commands) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type ByName struct{ Commands }

func (s ByName) Less(i, j int) bool { return s.Commands[i].Name() < s.Commands[j].Name() }

type renderMode string

const (
	Columns      = "COLUMNS"
	Paragraph    = "PARAGRAPH"
	commandError = "COMMAND_ERROR"
)

type CommandOutcome struct {
	OutputHeader string
	OutputBody   [][]string
	RenderHint   renderMode
}

func NewCommandClientError(message string) *ws.CategorisedError {
	return ws.NewCategorisedError(ws.Client, commandError, message)
}

func NewCommandLogicError(message string) *ws.CategorisedError {
	return ws.NewCategorisedError(ws.Logic, commandError, message)
}

func NewCommandUnexpectedError(message string) *ws.CategorisedError {
	return ws.NewCategorisedError(ws.Unexpected, commandError, message)
}
