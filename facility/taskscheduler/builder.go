// Copyright 2018-2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package taskscheduler

import (
	"github.com/graniticio/granitic/v2/config"
	"github.com/graniticio/granitic/v2/instance"
	"github.com/graniticio/granitic/v2/ioc"
	"github.com/graniticio/granitic/v2/logging"
	"github.com/graniticio/granitic/v2/schedule"
)

// The name of the TaskScheduler component as stored in the IoC framework.
const facilityName = "TaskScheduler"
const TaskSchedulerComponentName = instance.FrameworkPrefix + facilityName

// Creates the components that make up the TaskScheduler facility
type TaskSchedulerFacilityBuilder struct {
}

// See FacilityBuilder.BuildAndRegister
func (fb *TaskSchedulerFacilityBuilder) BuildAndRegister(lm *logging.ComponentLoggerManager, ca *config.ConfigAccessor, cn *ioc.ComponentContainer) error {

	ts := new(schedule.TaskScheduler)
	ts.FrameworkLogManager = lm
	ts.State = ioc.StoppedState

	//Inject JSON config
	ca.Populate(facilityName, ts)

	cn.WrapAndAddProto(TaskSchedulerComponentName, ts)

	return nil
}

// See FacilityBuilder.FacilityName
func (fb *TaskSchedulerFacilityBuilder) FacilityName() string {
	return facilityName
}

// See FacilityBuilder.DependsOnFacilities
func (fb *TaskSchedulerFacilityBuilder) DependsOnFacilities() []string {
	return []string{}
}
