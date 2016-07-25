package logger

import (
	"github.com/graniticio/granitic/config"
	"github.com/graniticio/granitic/facility/decorator"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/logging"
)

const applicationLoggingDecoratorName = ioc.FrameworkPrefix + "ApplicationLoggingDecorator"
const applicationLoggingManagerName = ioc.FrameworkPrefix + "ApplicationLoggingManager"

type ApplicationLoggingFacilityBuilder struct {
}

func (alfb *ApplicationLoggingFacilityBuilder) BuildAndRegister(lm *logging.ComponentLoggerManager, ca *config.ConfigAccessor, cn *ioc.ComponentContainer) {
	defaultLogLevelLabel := ca.StringVal("ApplicationLogger.DefaultLogLevel")
	defaultLogLevel := logging.LogLevelFromLabel(defaultLogLevelLabel)

	initialLogLevelsByComponent := ca.ObjectVal("ApplicationLogger.ComponentLogLevels")

	applicationLoggingManager := logging.CreateComponentLoggerManager(defaultLogLevel, initialLogLevelsByComponent)
	cn.WrapAndAddProto(applicationLoggingManagerName, applicationLoggingManager)

	applicationLoggingDecorator := new(decorator.ApplicationLogDecorator)
	applicationLoggingDecorator.LoggerManager = applicationLoggingManager
	applicationLoggingDecorator.FrameworkLogger = lm.CreateLogger(applicationLoggingDecoratorName)

	cn.WrapAndAddProto(applicationLoggingDecoratorName, applicationLoggingDecorator)
}

func (alfb *ApplicationLoggingFacilityBuilder) FacilityName() string {
	return "ApplicationLogging"
}

func (alfb *ApplicationLoggingFacilityBuilder) DependsOnFacilities() []string {
	return []string{}
}
