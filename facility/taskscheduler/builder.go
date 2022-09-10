// Copyright 2018-2022 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package taskscheduler

import (
	"github.com/graniticio/granitic/v3/config"
	"github.com/graniticio/granitic/v3/instance"
	"github.com/graniticio/granitic/v3/ioc"
	"github.com/graniticio/granitic/v3/logging"
	"github.com/graniticio/granitic/v3/schedule"
)

const facilityName = "TaskScheduler"

// TaskSchedulerComponentName is the name of the TaskScheduler component as stored in the IoC framework.
const TaskSchedulerComponentName = instance.FrameworkPrefix + facilityName

// FacilityBuilder creates the components that make up the TaskScheduler facility
type FacilityBuilder struct {
}

// BuildAndRegister implements FacilityBuilder.BuildAndRegister
func (fb *FacilityBuilder) BuildAndRegister(lm *logging.ComponentLoggerManager, ca *config.Accessor, cn *ioc.ComponentContainer) error {

	ts := new(schedule.TaskScheduler)
	ts.FrameworkLogManager = lm
	ts.State = ioc.StoppedState

	//Inject JSON config
	ca.Populate(facilityName, ts)

	cn.WrapAndAddProto(TaskSchedulerComponentName, ts)

	return nil
}

// FacilityName implements FacilityBuilder.FacilityName
func (fb *FacilityBuilder) FacilityName() string {
	return facilityName
}

// DependsOnFacilities implements FacilityBuilder.DependsOnFacilities
func (fb *FacilityBuilder) DependsOnFacilities() []string {
	return []string{}
}
