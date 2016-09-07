package logger

import (
	"errors"
	"github.com/graniticio/granitic/config"
	"github.com/graniticio/granitic/facility/decorator"
	"github.com/graniticio/granitic/instance"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/logging"
)

const applicationLoggingDecoratorName = instance.FrameworkPrefix + "ApplicationLoggingDecorator"
const applicationLoggingManagerName = instance.FrameworkPrefix + "ApplicationLoggingManager"

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

	writers, err := alfb.buildWriters(ca)
	formatter := new(logging.LogMessageFormatter)

	//Update the bootstrapped framework logger with the newly configured writers
	lm.UpdateWritersAndFormatter(writers, formatter)

	if err != nil {
		return alfb.error(err.Error())
	}

	alm := logging.CreateComponentLoggerManager(defaultLogLevel, initialLogLevelsByComponent, writers, formatter)
	cn.WrapAndAddProto(applicationLoggingManagerName, alm)

	ald := new(decorator.ApplicationLogDecorator)
	ald.LoggerManager = alm
	ald.FrameworkLogger = lm.CreateLogger(applicationLoggingDecoratorName)

	cn.WrapAndAddProto(applicationLoggingDecoratorName, ald)

	return nil
}

func (alfb *ApplicationLoggingFacilityBuilder) buildWriters(ca *config.ConfigAccessor) ([]logging.LogWriter, error) {
	writers := make([]logging.LogWriter, 0)

	console, err := ca.BoolVal("LogWriting.EnableConsoleLogging")

	if err != nil {
		return nil, err
	}

	if console {
		writers = append(writers, new(logging.ConsoleWriter))
	}

	file, err := ca.BoolVal("LogWriting.EnableFileLogging")

	if err != nil {
		return nil, err
	}

	if file {
		fileWriter := new(logging.AsynchFileWriter)

		err := ca.Populate("LogWriting.File", fileWriter)

		if err != nil {
			return nil, err
		}

		err = fileWriter.Init()

		if err != nil {
			return nil, err
		}

		writers = append(writers, fileWriter)
	}

	return writers, nil
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
