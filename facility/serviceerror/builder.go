package serviceerror

import (
	"github.com/graniticio/granitic/config"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/logging"
)

type ServiceErrorManagerFacilityBuilder struct {
}

func (fb *ServiceErrorManagerFacilityBuilder) BuildAndRegister(lm *logging.ComponentLoggerManager, ca *config.ConfigAccessor, cn *ioc.ComponentContainer) {

	manager := new(ServiceErrorManager)
	manager.PanicOnMissing = ca.BoolValue("ServiceErrorManager.PanicOnMissing")
	cn.WrapAndAddProto(serviceErrorManagerComponentName, manager)

	decorator := new(ServiceErrorConsumerDecorator)
	decorator.ErrorSource = manager
	cn.WrapAndAddProto(serviceErrorDecoratorComponentName, decorator)

	definitions := ca.StringVal("ServiceErrorManager.ErrorDefinitions")
	errors := ca.Array(definitions)

	if errors == nil {
		manager.FrameworkLogger.LogWarnf("No error definitions found at config path %s", definitions)
	} else {
		manager.LoadErrors(errors)
	}
}

func (fb *ServiceErrorManagerFacilityBuilder) FacilityName() string {
	return "ServiceErrorManager"
}

func (fb *ServiceErrorManagerFacilityBuilder) DependsOnFacilities() []string {
	return []string{}
}
