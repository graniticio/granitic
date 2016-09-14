package ctl

type Command interface {
	ExecuteCommand(qualifiers []string, args map[string]string) (*CommandOutcome, error)
	Name() string
}

type CommandOutcome struct {
}
