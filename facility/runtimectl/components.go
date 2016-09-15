package runtimectl

import (
	"errors"
	"github.com/graniticio/granitic/ctl"
	"github.com/graniticio/granitic/instance"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/logging"
	"github.com/graniticio/granitic/ws"
	"strconv"
	"strings"
)

const (
	compCommandName = "components"
	compSummary     = "Show a list of the names of components managed by the IoC container."
	compUsage       = "components [-fw true]"
	compHelp        = "Lists the name of all of the user-defined components currently present in the IoC Container."
	compHelpTwo     = "If the '-fw true' argument is supplied, the list will show built-in Granitic framework components instead of user-defined components."
	fwArg           = "fw"
)

type ComponentsCommand struct {
	FrameworkLogger logging.Logger
	container       *ioc.ComponentContainer
}

func (c *ComponentsCommand) Container(container *ioc.ComponentContainer) {
	c.container = container
}

func (c *ComponentsCommand) ExecuteCommand(qualifiers []string, args map[string]string) (*ctl.CommandOutcome, []*ws.CategorisedError) {

	var frameworkOnly bool
	var err error

	if frameworkOnly, err = showBuiltin(args); err != nil {
		return nil, []*ws.CategorisedError{ctl.NewCommandClientError(err.Error())}
	}

	names := make([][]string, 0)

	for _, c := range c.container.AllComponents() {

		fw := strings.HasPrefix(c.Name, instance.FrameworkPrefix)

		if (fw && frameworkOnly) || (!fw && !frameworkOnly) {
			names = append(names, []string{c.Name})
		}
	}

	co := new(ctl.CommandOutcome)
	co.OutputBody = names
	co.RenderHint = ctl.Columns

	return co, nil
}

func showBuiltin(args map[string]string) (bool, error) {

	if args == nil || len(args) == 0 {
		return false, nil
	}

	for k, v := range args {

		if k != fwArg {
			return false, errors.New("fw is the only argument supported by the components command.")
		}

		if choice, err := strconv.ParseBool(v); err == nil {
			return choice, nil
		} else {
			return false, errors.New("Value of fw argument cannot be interpreted as a bool")
		}

	}

	return false, nil

}

func (c *ComponentsCommand) Name() string {
	return compCommandName
}

func (c *ComponentsCommand) Summmary() string {
	return compSummary
}

func (c *ComponentsCommand) Usage() string {
	return compUsage
}

func (c *ComponentsCommand) Help() []string {
	return []string{compHelp, compHelpTwo}
}
