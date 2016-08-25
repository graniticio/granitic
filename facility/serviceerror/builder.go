package serviceerror

import (
	"errors"
	"github.com/graniticio/granitic/config"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/logging"
)

const (
	serviceErrorManagerComponentName      = ioc.FrameworkPrefix + "ServiceErrorManager"
	serviceErrorDecoratorComponentName    = ioc.FrameworkPrefix + "ServiceErrorSourceDecorator"
	errorCodeSourceDecoratorComponentName = ioc.FrameworkPrefix + "ErrorCodeSourceDecorator"
)

type ServiceErrorManagerFacilityBuilder struct {
}

func (fb *ServiceErrorManagerFacilityBuilder) BuildAndRegister(lm *logging.ComponentLoggerManager, ca *config.ConfigAccessor, cn *ioc.ComponentContainer) error {

	manager := new(ServiceErrorManager)
	manager.FrameworkLogger = lm.CreateLogger(serviceErrorManagerComponentName)

	panicOnMissing, err := ca.BoolVal("ServiceErrorManager.PanicOnMissing")

	if err != nil {
		return errors.New("Unable to build service error manager " + err.Error())
	}

	manager.PanicOnMissing = panicOnMissing

	cn.WrapAndAddProto(serviceErrorManagerComponentName, manager)

	errorDecorator := new(ServiceErrorConsumerDecorator)
	errorDecorator.ErrorSource = manager
	cn.WrapAndAddProto(serviceErrorDecoratorComponentName, errorDecorator)

	codeDecorator := new(ErrorCodeSourceDecorator)
	codeDecorator.ErrorSource = manager
	cn.WrapAndAddProto(errorCodeSourceDecoratorComponentName, codeDecorator)

	definitionsPath, err := ca.StringVal("ServiceErrorManager.ErrorDefinitions")

	if err != nil {
		return errors.New("Unable to load service error messages from configuration: " + err.Error())
	}

	messages := ca.Array(definitionsPath)

	if messages == nil {
		manager.FrameworkLogger.LogWarnf("No error definitions found at config path %s", definitionsPath)
	} else {
		manager.LoadErrors(messages)
	}

	return nil
}

func (fb *ServiceErrorManagerFacilityBuilder) FacilityName() string {
	return "ServiceErrorManager"
}

func (fb *ServiceErrorManagerFacilityBuilder) DependsOnFacilities() []string {
	return []string{}
}
