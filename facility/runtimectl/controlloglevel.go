package runtimectl

import (
	"fmt"
	"github.com/graniticio/granitic/ctl"
	"github.com/graniticio/granitic/instance"
	"github.com/graniticio/granitic/logging"
	"github.com/graniticio/granitic/ws"
	"sort"
)

const (
	LLComponentName = instance.FrameworkPrefix + "CommandLogLevel"
	llCommandName   = "log-level"
	llSummary       = "Views or sets a specific logging threshold for application or framework components."
	llUsage         = "log-level [component level] [-fw true]"
	llHelp          = "With no qualifier, this command shows a list of application components that have a specific logging threshold set. When a " +
		"component and a level are specified as qualifiers, the component's logging threshold is set at the specified level."
	llHelpTwo   = "Valid values for level are ALL, TRACE, DEBUG, INFO, WARN, ERROR, FATAL (case insensitive)."
	llHelpThree = "Setting the level to ALL has special behaviour - it efffectively removes the specific logging threshold for the component. The global log threshold will then apply to that component."
	llHelpFour  = "If the '-fw true' argument is supplied without qualifiers, a list of built-in framework components and their associated log levels will be shown."
)

type LogLevelCommand struct {
	FrameworkLogger    logging.Logger
	FrameworkManager   *logging.ComponentLoggerManager
	ApplicationManager *logging.ComponentLoggerManager
}

func (c *LogLevelCommand) ExecuteCommand(qualifiers []string, args map[string]string) (*ctl.CommandOutcome, []*ws.CategorisedError) {

	if len(qualifiers) == 0 {
		return c.showCurrentLevel(args)
	} else {
		return c.setLevel(qualifiers)
	}
}

func (c *LogLevelCommand) setLevel(qualifiers []string) (*ctl.CommandOutcome, []*ws.CategorisedError) {

	var err error
	var ll logging.LogLevel

	if len(qualifiers) < 2 {
		return nil, []*ws.CategorisedError{ctl.NewCommandClientError("You must provide a component name and a loging level.")}
	}

	name := qualifiers[0]
	label := qualifiers[1]

	if ll, err = logging.LogLevelFromLabel(label); err != nil {
		return nil, []*ws.CategorisedError{ctl.NewCommandClientError(err.Error())}
	}

	logger := c.ApplicationManager.LoggerByName(name)

	if logger == nil {
		logger = c.FrameworkManager.LoggerByName(name)
	}

	if logger == nil {
		m := fmt.Sprintf("Component %s does not exist or does not have a logger attached.", name)
		return nil, []*ws.CategorisedError{ctl.NewCommandClientError(m)}
	}

	logger.SetLocalThreshold(ll)

	return new(ctl.CommandOutcome), nil

}

func (c *LogLevelCommand) showCurrentLevel(args map[string]string) (*ctl.CommandOutcome, []*ws.CategorisedError) {

	var comps []*logging.ComponentLevel
	var err error
	var frameworkLevels bool

	if frameworkLevels, err = operateOnFramework(args); err != nil {
		return nil, []*ws.CategorisedError{ctl.NewCommandClientError(err.Error())}
	}

	if frameworkLevels {
		comps = c.FrameworkManager.CurrentLevels()
	} else {
		comps = c.ApplicationManager.CurrentLevels()
	}

	sort.Sort(logging.ByName{comps})

	filtered := make([][]string, 0)

	for _, cl := range comps {

		if cl.Level != logging.All {
			filtered = append(filtered, []string{cl.Name, logging.LabelFromLevel(cl.Level)})
		}

	}

	co := new(ctl.CommandOutcome)
	co.OutputBody = filtered
	co.RenderHint = ctl.Columns

	return co, nil
}

func (c *LogLevelCommand) Name() string {
	return llCommandName
}

func (c *LogLevelCommand) Summmary() string {
	return llSummary
}

func (c *LogLevelCommand) Usage() string {
	return llUsage
}

func (c *LogLevelCommand) Help() []string {
	return []string{llHelp, llHelpTwo, llHelpThree, llHelpFour}
}
