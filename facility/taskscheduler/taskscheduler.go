package taskscheduler

import (
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/logging"
)

type TaskScheduler struct {
	componentContainer *ioc.ComponentContainer
	state              ioc.ComponentState
	// Logger used by Granitic framework components. Automatically injected.
	FrameworkLogger logging.Logger
}

// Implements ioc.ContainerAccessor
func (ts *TaskScheduler) Container(container *ioc.ComponentContainer) {
	ts.componentContainer = container
}

// StartComponent Finds any schedules, parses them and verifies the component they reference implements schedule.RunnableTask
func (ts *TaskScheduler) StartComponent() error {

	if ts.state != ioc.StoppedState {
		return nil
	}

	ts.FrameworkLogger.LogInfof("Starting")

	return nil
}
