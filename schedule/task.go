// Copyright 2018-2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package schedule

import (
	"fmt"
	"time"
)

// Task describes when and how frequently a scheduled task should be executed and the component that provides a
// method to actually perform the task
type Task struct {

	// A human-readable name for the task
	Name string
	// An optional unique ID for the task (the IoC component name for this task will be used if not specified)
	Id string
	// The name of the IoC component implementing TaskLogic that actually performs this task
	Component string
	// The maximum number of overlapping instances of the task that are allowed to run. Zero means only one instance of this task can run at a time
	MaxOverlapping int
	// If set to true, suppress warning messages being logged when a task is scheduled to run while another instance is already running
	NoWarnOnOverlap bool
	// A human-readable expression (in English) of how frequently the task should be run - see package docs
	Every string

	// If set to true, any status updates messages sent from the task to the scheduler will be logged
	LogStatusMessages bool

	// The name of a component that is interested in receiving status updates from a running task
	StatusUpdateReceiver string

	// If set to true the task will never run
	Disabled bool

	// The number of times an invocation of this task should be re-tried if the task fails with an AllowRetryError
	MaxRetries int

	// A human-readable expression (in English) of how the interval to wait between a failure and a retry (e.g. 1 minute, 20 seconds)
	// Must be set if MaxRetries > 0
	RetryInterval string

	receiver TaskStatusUpdateReceiver

	logic TaskLogic

	retryWait time.Duration
}

func (t *Task) FullName() string {

	if t.Id == "" {
		return t.Name
	} else if t.Name == "" {
		return t.Id
	} else {
		return fmt.Sprintf("%s (%s)", t.Name, t.Id)
	}

}

func StatusMessagef(format string, a ...interface{}) TaskStatusUpdate {
	message := fmt.Sprintf(format, a...)

	return TaskStatusUpdate{
		Message: message,
	}

}

type TaskStatusUpdate struct {
	Message string
	Status  interface{}
}

type TaskLogic interface {
	ExecuteTask(c chan TaskStatusUpdate) error
}

type TaskStatusUpdateReceiver interface {
	Receive(summary TaskInvocationSummary, update TaskStatusUpdate)
}

type TaskInvocationSummary struct {
	TaskName        string
	TaskId          string
	StartedAt       time.Time
	InvocationCount uint64
}
