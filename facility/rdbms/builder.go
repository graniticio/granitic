// Copyright 2016 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

/*
	Package rdbms provides the RdbmsAccess facility which gives application code access to an RDBMS (SQL database).

	The RdbmsAccess facility is described in detail at http://granitic.io/1.0/ref/rdbms-access and the programmatic
	interface that applications will use for executing SQL is described in the rdbms package documentation.

	The purpose of this facility is to create an rdbms.RDBMSClientManager that will be injected into your application
	components. In turn, the rdbms.RDBMSClientManager will be used by your application to create instances of rdbms.RDBMSClient
	which provide the interface for executing SQL queries and managing transactions.

*/
package rdbms

import (
	"github.com/graniticio/granitic/config"
	"github.com/graniticio/granitic/facility/querymanager"
	"github.com/graniticio/granitic/instance"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/logging"
	"github.com/graniticio/granitic/rdbms"
)

const rdbmsClientManagerName = instance.FrameworkPrefix + "RdbmsClientManager"
const providerDecorator = instance.FrameworkPrefix + "DbProviderDecorator"
const managerDecorator = instance.FrameworkPrefix + "DbClientManagerDecorator"

// Creates an instance of rdbms.RDBMSClientManager that can be injected into your application components.
type RDBMSAccessFacilityBuilder struct {
}

// See FacilityBuilder.BuildAndRegister
func (rafb *RDBMSAccessFacilityBuilder) BuildAndRegister(lm *logging.ComponentLoggerManager, ca *config.ConfigAccessor, cn *ioc.ComponentContainer) error {

	manager := new(rdbms.GraniticRDBMSClientManager)
	ca.Populate("RdbmsAccess", manager)

	proto := ioc.CreateProtoComponent(manager, rdbmsClientManagerName)

	proto.AddDependency("QueryManager", querymanager.QueryManagerComponentName)

	cn.AddProto(proto)

	pd := new(databaseProviderDecorator)
	pd.receiver = manager

	cn.WrapAndAddProto(providerDecorator, pd)

	if manager.DisableAutoInjection {
		return nil
	}

	md := new(clientManagerDecorator)
	md.fieldNames = manager.InjectFieldNames
	md.manager = manager

	cn.WrapAndAddProto(managerDecorator, md)

	return nil

}

// See FacilityBuilder.FacilityName
func (rafb *RDBMSAccessFacilityBuilder) FacilityName() string {
	return "RdbmsAccess"
}

// DependsOnFacilities returns the other faclities that must be enabled in order to use the RdbmsAccess facility. You must
// enable the QueryManager facility.
func (rafb *RDBMSAccessFacilityBuilder) DependsOnFacilities() []string {
	return []string{querymanager.QueryManagerFacilityName}
}

// Finds implementations of rdbms.DatabaseProvider in the IoC container and injects them into the RDBMSClientManager
type databaseProviderDecorator struct {
	receiver rdbms.ProviderComponentReceiver
}

// OfInterest returns true if the subject component implements rdbms.DatabaseProvider
func (dpd *databaseProviderDecorator) OfInterest(subject *ioc.Component) bool {

	_, found := subject.Instance.(rdbms.DatabaseProvider)

	return found
}

// DecorateComponent injects the DatabaseProvider into the subject component.
func (dpd *databaseProviderDecorator) DecorateComponent(subject *ioc.Component, cc *ioc.ComponentContainer) {
	dpd.receiver.RegisterProvider(subject)

}
