package logger

import (
	"errors"
	"github.com/graniticio/granitic/config"
	"github.com/graniticio/granitic/facility/decorator"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/logging"
)

const applicationLoggingDecoratorName = ioc.FrameworkPrefix + "ApplicationLoggingDecorator"
const applicationLoggingManagerName = ioc.FrameworkPrefix + "ApplicationLoggingManager"

type ApplicationLoggingFacilityBuilder struct {
}

func (alfb *ApplicationLoggingFacilityBuilder) BuildAndRegister(lm *logging.ComponentLoggerManager, ca *config.ConfigAccessor, cn *ioc.ComponentContainer) error {
	defaultLogLevelLabel, err := ca.StringVal("ApplicationLogger.DefaultLogLevel")

	if err != nil {
		return alfb.error(err.Error())
	}

	defaultLogLevel, err := logging.LogLevelFromLabel(defaultLogLevelLabel)

	if err != nil {
		return alfb.error(err.Error())
	}

	initialLogLevelsByComponent := ca.ObjectVal("ApplicationLogger.ComponentLogLevels")

	applicationLoggingManager := logging.CreateComponentLoggerManager(defaultLogLevel, initialLogLevelsByComponent)
	cn.WrapAndAddProto(applicationLoggingManagerName, applicationLoggingManager)

	applicationLoggingDecorator := new(decorator.ApplicationLogDecorator)
	applicationLoggingDecorator.LoggerManager = applicationLoggingManager
	applicationLoggingDecorator.FrameworkLogger = lm.CreateLogger(applicationLoggingDecoratorName)

	cn.WrapAndAddProto(applicationLoggingDecoratorName, applicationLoggingDecorator)

	return nil
}

func (alfb *ApplicationLoggingFacilityBuilder) error(suffix string) error {

	return errors.New("Unable to initialise application logging: " + suffix)

}

func (alfb *ApplicationLoggingFacilityBuilder) FacilityName() string {
	return "ApplicationLogging"
}

func (alfb *ApplicationLoggingFacilityBuilder) DependsOnFacilities() []string {
	return []string{}
}
