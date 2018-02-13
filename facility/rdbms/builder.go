// Copyright 2016-2018 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

/*
	Package rdbms provides the RdbmsAccess facility which gives application code access to an RDBMS (SQL database).

	The RdbmsAccess facility is described in detail at http://granitic.io/1.0/ref/rdbms-access and the programmatic
	interface that applications will use for executing SQL is described in the rdbms package documentation.

	The purpose of this facility is to create an rdbms.RdbmsClientManager that will be injected into your application
	components. In turn, the rdbms.RdbmsClientManager will be used by your application to create instances of rdbms.RDBMSClient
	which provide the interface for executing SQL queries and managing transactions.

*/
package rdbms

import (
	"fmt"
	"github.com/graniticio/granitic/config"
	"github.com/graniticio/granitic/facility/querymanager"
	"github.com/graniticio/granitic/instance"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/logging"
	"github.com/graniticio/granitic/rdbms"
	"github.com/graniticio/granitic/types"
	"github.com/pkg/errors"
)

const rdbmsClientManagerConfigName = instance.FrameworkPrefix + "RdbmsClientManagerConfig"

const managerDecorator = instance.FrameworkPrefix + "DbClientManagerDecorator"

// Creates an instance of rdbms.RDBMSClientManager that can be injected into your application components.
type RdbmsAccessFacilityBuilder struct {
}

// See FacilityBuilder.BuildAndRegister
func (rafb *RdbmsAccessFacilityBuilder) BuildAndRegister(lm *logging.ComponentLoggerManager, ca *config.ConfigAccessor, cn *ioc.ComponentContainer) error {

	//Find the names of components that implement DatabaseProvider
	pn := rafb.findProviders(cn)

	if len(pn) == 0 {
		return errors.New("You must define a component that implements rdbms.DatabaseProvider if you want to use the RdbmsAccess facility")
	}

	managerConfigs := make(map[string]*rdbms.RdbmsClientManagerConfig)

	if len(pn) == 1 {

		providerName := pn[0]

		// Create config for a default ClientManager
		mc := new(rdbms.RdbmsClientManagerConfig)
		ca.Populate("RdbmsAccess.Default", mc)

		proto := ioc.CreateProtoComponent(mc, mc.ClientName+"ManagerConfig")

		proto.AddDependency("Provider", providerName)
		cn.AddProto(proto)

		managerConfigs[rdbmsClientManagerConfigName] = mc

	}

	return rafb.createManagers(cn, managerConfigs, lm)

}

func (rafb *RdbmsAccessFacilityBuilder) createManagers(cn *ioc.ComponentContainer, conf map[string]*rdbms.RdbmsClientManagerConfig, lm *logging.ComponentLoggerManager) error {

	mn := types.NewEmptyUnorderedStringSet()

	for _, v := range conf {
		for _, method := range v.InjectFieldNames {

			if mn.Contains(method) {
				message := fmt.Sprintf("More than one rdbms.RdbmsClientManagerConfig component is configured to inject into the field name %s", method)
				return errors.New(message)
			} else {
				mn.Add(method)
			}

		}
	}

	fieldsToManager := make(map[string]rdbms.RdbmsClientManager)

	for k, managerConf := range conf {
		manager := new(rdbms.GraniticRdbmsClientManager)
		manager.SharedLog = lm.CreateLogger(managerConf.ClientName)

		if managerConf.ManagerName == "" {
			managerConf.ManagerName = managerConf.ClientName + "Manager"
		}

		proto := ioc.CreateProtoComponent(manager, managerConf.ManagerName)

		proto.AddDependency("QueryManager", querymanager.QueryManagerComponentName)
		proto.AddDependency("Configuration", k)
		cn.AddProto(proto)

		for _, methodToInject := range managerConf.InjectFieldNames {
			fieldsToManager[methodToInject] = manager
		}

	}

	md := new(clientManagerDecorator)
	md.fieldNameManager = fieldsToManager

	cn.WrapAndAddProto(managerDecorator, md)

	return nil

}

func (rafb *RdbmsAccessFacilityBuilder) findProviders(cn *ioc.ComponentContainer) []string {

	p := make([]string, 0)

	for name, c := range cn.ProtoComponents() {

		i := c.Component.Instance

		if i != nil {

			if _, okay := i.(rdbms.DatabaseProvider); okay {
				p = append(p, name)
			}

		}

	}

	return p

}

// See FacilityBuilder.FacilityName
func (rafb *RdbmsAccessFacilityBuilder) FacilityName() string {
	return "RdbmsAccess"
}

// DependsOnFacilities returns the other faclities that must be enabled in order to use the RdbmsAccess facility. You must
// enable the QueryManager facility.
func (rafb *RdbmsAccessFacilityBuilder) DependsOnFacilities() []string {
	return []string{querymanager.QueryManagerFacilityName}
}
