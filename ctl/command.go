package ctl

import "github.com/graniticio/granitic/ws"

type Command interface {
	ExecuteCommand(qualifiers []string, args map[string]string) (*CommandOutcome, []*ws.CategorisedError)
	Name() string
}

type renderMode string

const (
	columns   = "COLUMNS"
	paragraph = "PARAGRAPH"
)

type CommandOutcome struct {
	OutputHeader string
	OutputBody   [][]string
	RenderHint   renderMode
}
