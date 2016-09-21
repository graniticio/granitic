package rdbms

import (
	"github.com/graniticio/granitic/config"
	"github.com/graniticio/granitic/facility/querymanager"
	"github.com/graniticio/granitic/instance"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/logging"
)

const rdbmsClientManagerName = instance.FrameworkPrefix + "RdbmsClientManager"
const providerDecorator = instance.FrameworkPrefix + "DbProviderDecorator"
const managerDecorator = instance.FrameworkPrefix + "DbClientManagerDecorator"

type RdbmsAccessFacilityBuilder struct {
}

func (rafb *RdbmsAccessFacilityBuilder) BuildAndRegister(lm *logging.ComponentLoggerManager, ca *config.ConfigAccessor, cn *ioc.ComponentContainer) error {

	manager := new(DefaultRdbmsClientManager)
	ca.Populate("RdbmsAccess", manager)

	proto := ioc.CreateProtoComponent(manager, rdbmsClientManagerName)

	proto.AddDependency("QueryManager", querymanager.QueryManagerComponentName)

	cn.AddProto(proto)

	pd := new(DatabaseProviderDecorator)
	pd.receiver = manager

	cn.WrapAndAddProto(providerDecorator, pd)

	if manager.DisableAutoInjection {
		return nil
	}

	md := new(ClientManagerDecorator)
	md.fieldNames = manager.InjectFieldNames
	md.manager = manager

	cn.WrapAndAddProto(managerDecorator, md)

	return nil

}

func (rafb *RdbmsAccessFacilityBuilder) FacilityName() string {
	return "RdbmsAccess"
}

func (rafb *RdbmsAccessFacilityBuilder) DependsOnFacilities() []string {
	return []string{querymanager.QueryManagerFacilityName}
}

type DatabaseProviderDecorator struct {
	receiver ProviderComponentReceiver
}

func (dpd *DatabaseProviderDecorator) OfInterest(component *ioc.Component) bool {

	_, found := component.Instance.(DatabaseProvider)

	return found
}

func (dpd *DatabaseProviderDecorator) DecorateComponent(c *ioc.Component, cc *ioc.ComponentContainer) {
	dpd.receiver.RegisterProvider(c)

}
