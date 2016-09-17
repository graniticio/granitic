package runtimectl

import (
	"github.com/graniticio/granitic/ctl"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/logging"
	"github.com/graniticio/granitic/ws"
)

const (
	compCommandName = "components"
	compSummary     = "Show a list of the names of components managed by the IoC container."
	compUsage       = "components [-fw true] [-lc start|stop|suspend]"
	compHelp        = "Lists the name of all of the user-defined components currently present in the IoC Container."
	compHelpTwo     = "If the '-fw true' argument is supplied, the list will show built-in Granitic framework components instead of user-defined components."
	compHelpThree   = "If the '-lc true' argument is supplied with one of the values start/stop/suspend then only those components that implement the corresponding " +
		"lifecycle interface (ioc.Startable, ioc.Stoppable, ioc.Suspendable) will be displayed"
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
	var lcFilter ioc.LifecycleSupport
	var err error

	if frameworkOnly, err = operateOnFramework(args); err != nil {
		return nil, []*ws.CategorisedError{ctl.NewCommandClientError(err.Error())}
	}

	if lcFilter, err = findLifecycleFilter(args); err != nil {
		return nil, []*ws.CategorisedError{ctl.NewCommandClientError(err.Error())}
	}

	names := make([][]string, 0)

	for _, c := range c.container.AllComponents() {

		fw := isFramework(c)

		if ((fw && frameworkOnly) || (!fw && !frameworkOnly)) && matchesFilter(lcFilter, c.Instance) {
			names = append(names, []string{c.Name})
		}
	}

	co := new(ctl.CommandOutcome)
	co.OutputBody = names
	co.RenderHint = ctl.Columns

	return co, nil
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
