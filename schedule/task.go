// Copyright 2018 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package schedule

import "fmt"

type OverlapBehaviour int

const (
	SKIP OverlapBehaviour = iota
	ALLOW
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
	// How the scheduler should handle two overlapping invocations of a task. See OverlapBehaviour const names for allowed values - default is SKIP
	Overlap string
	// If set to true, suppress warning messages being logged when a task is scheduled to run while another instance is already running
	NoWarnOnOverlap bool
	// A human-readable expression (in English) of how frequently the task should be run - see package docs
	Every string

	logic   TaskLogic
	overlap OverlapBehaviour
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

type TaskStatus int

const (
	TASK_START TaskStatus = iota
	TASK_PROGRESS_UPDATE
	TASK_END_SUCCESS
	TASK_END_FAIL
	TASK_END_PARTIAL
)

type TaskStatusUpdate struct {
	Status TaskStatus
}

type TaskLogic interface {
	ExecuteTask(c chan TaskStatusUpdate)
}
