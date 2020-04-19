// Copyright 2018-2020 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package schedule

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/graniticio/granitic/v2/ioc"
	"github.com/graniticio/granitic/v2/logging"
	"time"
)

// TaskScheduler is the top level Granitic component for managing the scheduled invocation of tasks
type TaskScheduler struct {
	componentContainer ioc.ComponentLookup
	managedTasks       []*invocationManager
	State              ioc.ComponentState
	// Logger used by Granitic framework components. Automatically injected.
	FrameworkLogger     logging.Logger
	FrameworkLogManager *logging.ComponentLoggerManager
}

// Container implements ioc.ContainerAccessor.Container
func (ts *TaskScheduler) Container(container *ioc.ComponentContainer) {
	ts.componentContainer = container
}

// StartComponent Finds any schedules, parses them and verifies the component they reference implements schedule.TaskLogic
func (ts *TaskScheduler) StartComponent() error {

	ts.managedTasks = make([]*invocationManager, 0)

	if ts.State != ioc.StoppedState {
		return nil
	}

	ts.FrameworkLogger.LogDebugf("Searching for schedule.Task components")

	for _, component := range ts.componentContainer.AllComponents() {

		ts.FrameworkLogger.LogTracef("Considering %s", component.Name)

		if task, found := component.Instance.(*Task); found {
			if task.ID == "" {
				//Use the name of the component to be run as ID for the task if it isn't explicitly set
				task.ID = task.Component
			}

			ts.FrameworkLogger.LogDebugf("Found Task %s", task.FullName())

			if task.Disabled {
				ts.FrameworkLogger.LogWarnf("Task %s will never run on a schedule as it has been disabled. Can be invoked manually", task.FullName())
			}

			if err := ts.validateAndPrepare(ts.componentContainer, task); err != nil {
				return fmt.Errorf("%s: %s", component.Name, err.Error())
			}

		}
	}

	return nil
}

// AllowAccess spawns a goroutinue managing each scheduled task
func (ts *TaskScheduler) AllowAccess() error {

	for _, tm := range ts.managedTasks {

		go tm.Start()

	}

	ts.FrameworkLogger.LogDebugf("%d task manager(s) started", len(ts.managedTasks))

	return nil
}

func (ts *TaskScheduler) validateAndPrepare(cn ioc.ComponentLookup, task *Task) error {

	if err := ts.findLogic(cn, task); err != nil {
		return err
	}

	if task.Every == "" {
		m := fmt.Sprintf("You must set the 'Every' field to set an execution interval")
		return errors.New(m)
	}

	if task.MaxOverlapping < 0 {
		m := fmt.Sprintf("The 'MaxOverlapping' field cannot be a negative number")
		return errors.New(m)
	}

	if task.MaxRetries > 0 {
		if task.RetryInterval == "" {
			m := fmt.Sprintf("The 'RetryInterval' must be set if 'MaxRetries' > 0")
			return errors.New(m)

		}

		var retryWait time.Duration
		var err error

		if retryWait, err = parseNaturalToDuration(task.RetryInterval); err != nil {
			return err
		}

		task.retryWait = retryWait

	}

	tm := newInvocationManager(task)
	ts.managedTasks = append(ts.managedTasks, tm)
	tm.Log = ts.FrameworkLogManager.CreateLogger(task.Component + "TaskManager")

	if interval, err := parseEvery(task.Every); err == nil {
		tm.Interval = interval
	} else {

		return err
	}

	return nil
}

func (ts *TaskScheduler) findLogic(cn ioc.ComponentLookup, task *Task) error {
	if task.Component == "" {
		return errors.New("Missing Component (you must provide the name of the component that will execute your task")
	}

	tc := cn.ComponentByName(task.Component)

	if tc == nil {
		m := fmt.Sprintf("Component %s does not exist (no component with that name)", task.Component)
		return errors.New(m)
	}

	tl, okay := tc.Instance.(TaskLogic)

	if !okay {
		m := fmt.Sprintf("Component %s does not implement schedule.TaskLogic", task.Component)
		return errors.New(m)
	}

	task.logic = tl

	if task.StatusUpdateReceiver == "" {
		return nil
	}

	sr := cn.ComponentByName(task.StatusUpdateReceiver)

	if sr == nil {
		m := fmt.Sprintf("StatusUpdateReceiver %s does not exist (no component with that name)", task.StatusUpdateReceiver)
		return errors.New(m)
	}

	sri, okay := sr.Instance.(TaskStatusUpdateReceiver)

	if !okay {
		m := fmt.Sprintf("StatusUpdateReceiver %s does not implement schedule.TaskStatusUpdateReceiver", task.Component)
		return errors.New(m)
	}

	task.receiver = sri

	return nil
}

// PrepareToStop calls the same method of each of the managed Tasks
func (ts *TaskScheduler) PrepareToStop() {

	for _, tm := range ts.managedTasks {

		tm.PrepareToStop()

	}

}

// ReadyToStop only returns true when each of the managed tasks report that they are ReadyToStop
func (ts *TaskScheduler) ReadyToStop() (bool, error) {

	ready := true
	var buffer bytes.Buffer

	for _, tm := range ts.managedTasks {

		managerReady, err := tm.ReadyToStop()

		if !managerReady {
			ready = false

			if err != nil {
				buffer.WriteString("\n")
				buffer.WriteString(err.Error())
			}

		}
	}

	if ready {
		return true, nil
	}

	var err error

	if bc := buffer.String(); len(bc) > 0 {
		err = errors.New(bc)
	}

	return false, err
}

// Stop calls the same method on each of the managed tasks. Any errors returned when stopping
// the underlying tasks are concatenated together and returned as a single error
func (ts *TaskScheduler) Stop() error {
	var buffer bytes.Buffer

	for _, tm := range ts.managedTasks {

		err := tm.Stop()

		if err != nil {
			buffer.WriteString("\n")
			buffer.WriteString(err.Error())
		}
	}

	var err error

	if bc := buffer.String(); len(bc) > 0 {
		err = errors.New(bc)
	}

	return err
}
