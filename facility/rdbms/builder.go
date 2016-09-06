package rdbms

import (
	"github.com/graniticio/granitic/config"
	"github.com/graniticio/granitic/facility/querymanager"
	"github.com/graniticio/granitic/instance"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/logging"
)

const rdbmsClientManagerName = instance.FrameworkPrefix + "RdbmsClientManager"

type RdbmsAccessFacilityBuilder struct {
}

func (rafb *RdbmsAccessFacilityBuilder) BuildAndRegister(lm *logging.ComponentLoggerManager, ca *config.ConfigAccessor, cn *ioc.ComponentContainer) error {

	manager := new(DefaultRdbmsClientManager)
	ca.Populate("RdbmsAccess", manager)

	proto := ioc.CreateProtoComponent(manager, rdbmsClientManagerName)

	proto.AddDependency("Provider", manager.DatabaseProviderComponentName)
	proto.AddDependency("QueryManager", querymanager.QueryManagerComponentName)

	cn.AddProto(proto)

	return nil

}

func (rafb *RdbmsAccessFacilityBuilder) FacilityName() string {
	return "RdbmsAccess"
}

func (rafb *RdbmsAccessFacilityBuilder) DependsOnFacilities() []string {
	return []string{querymanager.QueryManagerFacilityName}
}
