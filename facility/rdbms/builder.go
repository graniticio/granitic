// Copyright 2016-2023 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

/*
Package rdbms provides the RdbmsAccess facility which gives application code access to an RDBMS (SQL database).

The RdbmsAccess facility is described in detail at https://granitic.io/ref/relational-databases and the programmatic
interface that applications will use for executing SQL is described in the rdbms package documentation.

The purpose of this facility is to create an rdbms.ClientManager that will be injected into your application
components. In turn, the rdbms.ClientManager will be used by your application to create instances of rdbms.RDBMSClient
which provide the interface for executing SQL queries and managing transactions.
*/
package rdbms

import (
	"errors"
	"fmt"
	"github.com/graniticio/granitic/v2/config"
	"github.com/graniticio/granitic/v2/facility/querymanager"
	"github.com/graniticio/granitic/v2/instance"
	"github.com/graniticio/granitic/v2/ioc"
	"github.com/graniticio/granitic/v2/logging"
	"github.com/graniticio/granitic/v2/rdbms"
	"github.com/graniticio/granitic/v2/types"
)

const rdbmsClientManagerConfigName = instance.FrameworkPrefix + "ClientManagerConfig"

const managerDecorator = instance.FrameworkPrefix + "DbClientManagerDecorator"

// FacilityBuilder creates an instance of rdbms.RDBMSClientManager that can be injected into your application components.
type FacilityBuilder struct {
	Log logging.Logger
}

// BuildAndRegister implements FacilityBuilder.BuildAndRegister
func (rafb *FacilityBuilder) BuildAndRegister(lm *logging.ComponentLoggerManager, ca *config.Accessor, cn *ioc.ComponentContainer) error {

	log := lm.CreateLogger(instance.FrameworkPrefix + "FacilityBuilder")
	rafb.Log = log

	log.LogTracef("Configuring RDBMS client managers")

	//Find the names of components that implement DatabaseProvider
	pn := rafb.findProviders(cn)

	if len(pn) == 0 {
		return errors.New("You must define a component that implements rdbms.DatabaseProvider if you want to use the RdbmsAccess facility")
	}

	managerConfigs := make(map[string]*rdbms.ClientManagerConfig)

	//See if client manager configs have been explicitly defined
	rafb.findConfigurations(cn, managerConfigs)

	if len(managerConfigs) == 0 {

		log.LogTracef("Provider found but no explicit rdbms.ClientManagerConfig components. Creating default configuration")

		//We have a provider
		providerName := pn[0]

		// Create config for a default ClientManager
		mc := new(rdbms.ClientManagerConfig)
		ca.Populate("RdbmsAccess.Default", mc)

		proto := ioc.CreateProtoComponent(mc, rdbmsClientManagerConfigName)

		proto.AddDependency("Provider", providerName)
		cn.AddProto(proto)

		managerConfigs[rdbmsClientManagerConfigName] = mc

	}

	return rafb.createManagers(cn, managerConfigs, lm)

}

func (rafb *FacilityBuilder) createManagers(cn *ioc.ComponentContainer, conf map[string]*rdbms.ClientManagerConfig, lm *logging.ComponentLoggerManager) error {

	mn := types.NewEmptyUnorderedStringSet()

	for _, v := range conf {
		for _, method := range v.InjectFieldNames {

			if mn.Contains(method) {
				return fmt.Errorf("more than one rdbms.ClientManagerConfig component is configured to inject into the field name %s", method)
			}

			mn.Add(method)
		}
	}

	fieldsToManager := make(map[string]rdbms.ClientManager)

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
	md.log = lm.CreateLogger(instance.FrameworkPrefix + "ClientManagerDecorator")

	cn.WrapAndAddProto(managerDecorator, md)

	return nil

}

func (rafb *FacilityBuilder) findConfigurations(cn *ioc.ComponentContainer, c map[string]*rdbms.ClientManagerConfig) {

	matcher := func(i interface{}) (okay bool) {
		_, okay = i.(*rdbms.ClientManagerConfig)
		return
	}

	for _, comp := range cn.ProtoComponentsByType(matcher) {
		name := comp.Component.Name

		rafb.Log.LogTracef("Found instance of dbms.ClientManagerConfig: %s", name)

		config, _ := comp.Component.Instance.(*rdbms.ClientManagerConfig)

		if config.ClientName == "" {
			config.ClientName = name + "ManagedClient"
		}

		rafb.Log.LogTracef("ManagedClient name will be: %s", config.ClientName)

		if config.ManagerName == "" {
			config.ManagerName = config.ClientName + "Manager"
		}

		rafb.Log.LogTracef("ClientManager name will be: %s", config.ManagerName)

		c[name] = config

	}

}

func (rafb *FacilityBuilder) findProviders(cn *ioc.ComponentContainer) []string {

	p := make([]string, 0)

	matcher := func(i interface{}) (okay bool) {
		_, okay = i.(rdbms.DatabaseProvider)
		return
	}

	for _, comp := range cn.ProtoComponentsByType(matcher) {
		p = append(p, comp.Component.Name)

	}

	return p

}

// FacilityName implements FacilityBuilder.FacilityName
func (rafb *FacilityBuilder) FacilityName() string {
	return "RdbmsAccess"
}

// DependsOnFacilities returns the other faclities that must be enabled in order to use the RdbmsAccess facility. You must
// enable the QueryManager facility.
func (rafb *FacilityBuilder) DependsOnFacilities() []string {
	return []string{querymanager.QueryManagerFacilityName}
}
