package schedule

import (
	"github.com/graniticio/granitic/v2/instance"
	"github.com/graniticio/granitic/v2/logging"
)

const (
	//LLComponentName is the name of the component that handles the management of scheduled tasks
	LLComponentName = instance.FrameworkPrefix + "CommandScheduledTasks"
	llCommandName   = "task"
	llSummary       = "Shows information about all scheduled tasks or invokes/suspends a specified task"
	llUsage         = "task [ID] [invoke|suspend|resume]"
	llHelp          = "With no qualifier, this command shows a list of scheduled tasks defined for this service."
	llHelpTwo       = "If a single qualifier is specified, that is assumed to be the ID of a task and more detailed information is shown for that task."
	llHelpThree     = "If a task ID is specified followed by invoke, that task is will be scheduled to run immediately (but task configuration is respected, so multiple conncurrent invocations might be forbidden."
	llHelpFour      = "If a task ID is specified followed by suspend, that task will not be executed until resumed."
	llHelpFive      = "If a task ID is specified followed by resumed, that task will be allowed to run again, if it is currently suspended"
)

type taskCommand struct {
	FrameworkLogger    logging.Logger
	FrameworkManager   *logging.ComponentLoggerManager
	ApplicationManager *logging.ComponentLoggerManager
}
