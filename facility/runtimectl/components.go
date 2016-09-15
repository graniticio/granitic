package runtimectl

import (
	"errors"
	"fmt"
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
	compUsage       = "components [-fw true] [-lc start|stop|suspend]"
	compHelp        = "Lists the name of all of the user-defined components currently present in the IoC Container."
	compHelpTwo     = "If the '-fw true' argument is supplied, the list will show built-in Granitic framework components instead of user-defined components."
	compHelpThree   = "If the '-lc true' argument is supplied with one of the values start/stop/suspend then only those components that implement the corresponding " +
		"lifecycle interface (ioc.Startable, ioc.Stoppable, ioc.Suspendable) will be displayed"
	fwArg = "fw"
	lcArg = "lc"
)

type lifecycleFilter int

const (
	all = iota
	stop
	start
	suspend
)

func fromFilterArg(arg string) (lifecycleFilter, error) {

	s := strings.ToLower(arg)

	switch s {
	case "", "all":
		return all, nil
	case "stop":
		return stop, nil
	case "start":
		return start, nil
	case "suspend":
		return suspend, nil
	}

	m := fmt.Sprintf("%s is not a recognised lifecycle filter (all, stop, start, suspend)", arg)

	return all, errors.New(m)

}

type ComponentsCommand struct {
	FrameworkLogger logging.Logger
	container       *ioc.ComponentContainer
}

func (c *ComponentsCommand) Container(container *ioc.ComponentContainer) {
	c.container = container
}

func (c *ComponentsCommand) ExecuteCommand(qualifiers []string, args map[string]string) (*ctl.CommandOutcome, []*ws.CategorisedError) {

	var frameworkOnly bool
	var lcFilter lifecycleFilter
	var err error

	if frameworkOnly, err = showBuiltin(args); err != nil {
		return nil, []*ws.CategorisedError{ctl.NewCommandClientError(err.Error())}
	}

	if lcFilter, err = findLifecycleFilter(args); err != nil {
		return nil, []*ws.CategorisedError{ctl.NewCommandClientError(err.Error())}
	}

	names := make([][]string, 0)

	for _, c := range c.container.AllComponents() {

		fw := strings.HasPrefix(c.Name, instance.FrameworkPrefix)

		if ((fw && frameworkOnly) || (!fw && !frameworkOnly)) && matchesFilter(lcFilter, c.Instance) {
			names = append(names, []string{c.Name})
		}
	}

	co := new(ctl.CommandOutcome)
	co.OutputBody = names
	co.RenderHint = ctl.Columns

	return co, nil
}

func matchesFilter(f lifecycleFilter, i interface{}) bool {

	switch f {
	case all:
		return true

	case start:
		_, found := i.(ioc.Startable)
		return found

	case stop:
		_, found := i.(ioc.Stoppable)
		return found

	case suspend:
		_, found := i.(ioc.Suspendable)
		return found

	}

	return true
}

func findLifecycleFilter(args map[string]string) (lifecycleFilter, error) {

	if args == nil || len(args) == 0 || args[lcArg] == "" {
		return all, nil
	}

	v := args[lcArg]

	return fromFilterArg(v)

}

func showBuiltin(args map[string]string) (bool, error) {

	if args == nil || len(args) == 0 || args[fwArg] == "" {
		return false, nil
	}

	v := args[fwArg]

	if choice, err := strconv.ParseBool(v); err == nil {
		return choice, nil
	} else {
		return false, errors.New("Value of fw argument cannot be interpreted as a bool")
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
	return []string{compHelp, compHelpTwo, compHelpThree}
}
