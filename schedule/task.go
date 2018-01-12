// Copyright 2018 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package schedule

import "fmt"

// Task describes when and how frequently a scheduled task should be executed and the component that provides a
// method to actually perform the task
type Task struct {
	Name      string
	Id        string
	Component string
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
