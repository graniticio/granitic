package serviceerror

import (
	"errors"
	"fmt"
	"github.com/graniticio/granitic/config"
	ge "github.com/graniticio/granitic/grncerror"
	"github.com/graniticio/granitic/instance"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/logging"
)

const (
	serviceErrorManagerComponentName      = instance.FrameworkPrefix + "ServiceErrorManager"
	serviceErrorDecoratorComponentName    = instance.FrameworkPrefix + "ServiceErrorSourceDecorator"
	errorCodeSourceDecoratorComponentName = instance.FrameworkPrefix + "ErrorCodeSourceDecorator"
)

type ServiceErrorManagerFacilityBuilder struct {
}

func (fb *ServiceErrorManagerFacilityBuilder) BuildAndRegister(lm *logging.ComponentLoggerManager, ca *config.ConfigAccessor, cn *ioc.ComponentContainer) error {

	manager := new(ge.ServiceErrorManager)
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

	if messages, err := fb.loadMessagesFromConfig(definitionsPath, ca); err != nil {
		return err
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

func (fb *ServiceErrorManagerFacilityBuilder) loadMessagesFromConfig(dPath string, ca *config.ConfigAccessor) ([]interface{}, error) {

	if !ca.PathExists(dPath) {
		m := fmt.Sprintf("No error definitions found at config path %s", dPath)
		return nil, errors.New(m)
	}

	i := ca.Value(dPath)

	if config.JsonType(i) != config.JsonArray {
		m := fmt.Sprintf("Couldn't load error messages from config path %s. Make sure the path exists in your configuration and that %s is an array of string arrays ([][]string)", dPath, dPath)
		return nil, errors.New(m)
	}

	if v, err := ca.Array(dPath); err == nil {
		return v, nil
	} else {
		return nil, err
	}

}
