package facility

import (
	"errors"
	"fmt"
	"github.com/graniticio/granitic/config"
	"github.com/graniticio/granitic/facility/decorator"
	"github.com/graniticio/granitic/facility/httpserver"
	"github.com/graniticio/granitic/facility/jsonws"
	"github.com/graniticio/granitic/facility/logger"
	"github.com/graniticio/granitic/facility/querymanager"
	"github.com/graniticio/granitic/facility/rdbms"
	"github.com/graniticio/granitic/facility/serviceerror"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/logging"
)

const frameworkLoggingManagerName = ioc.FrameworkPrefix + "FrameworkLoggingManager"
const frameworkLoggerDecoratorName = ioc.FrameworkPrefix + "FrameworkLoggingDecorator"
const configErrorPrefix = "Unable to configure framework logging: "

type FacilitiesInitialisor struct {
	ConfigAccessor          *config.ConfigAccessor
	FrameworkLoggingManager *logging.ComponentLoggerManager
	Logger                  logging.Logger
	container               *ioc.ComponentContainer
	facilities              []FacilityBuilder
	facilityStatus          map[string]interface{}
}

func NewFacilitiesInitialisor(cc *ioc.ComponentContainer, flm *logging.ComponentLoggerManager) *FacilitiesInitialisor {
	fi := new(FacilitiesInitialisor)
	fi.container = cc
	fi.FrameworkLoggingManager = flm

	return fi
}

func BootstrapFrameworkLogging(bootStrapLogLevel logging.LogLevel) (*logging.ComponentLoggerManager, *ioc.ProtoComponent) {

	flm := logging.CreateComponentLoggerManager(bootStrapLogLevel, nil)
	proto := ioc.CreateProtoComponent(flm, frameworkLoggingManagerName)

	return flm, proto

}

func (fi *FacilitiesInitialisor) AddFacility(f FacilityBuilder) {
	fi.facilities = append(fi.facilities, f)
}

func (fi *FacilitiesInitialisor) buildEnabledFacilities() error {

	for _, fb := range fi.facilities {

		name := fb.FacilityName()

		if fi.facilityStatus[name] == nil {

			fi.Logger.LogWarnf("No setting for facility %s in the Facilities configuration object - will not enable this facility", name)
			continue

		}

		if fi.facilityStatus[name].(bool) {

			for _, dep := range fb.DependsOnFacilities() {

				if fi.facilityStatus[dep] == nil || fi.facilityStatus[dep].(bool) == false {
					message := fmt.Sprintf("Facility %s depends on facility %s, but %s is not enabled in configuration.", name, dep, dep)
					return errors.New(message)
				}

			}

			err := fb.BuildAndRegister(fi.FrameworkLoggingManager, fi.ConfigAccessor, fi.container)

			if err != nil {
				return err
			}
		}
	}

	return nil

}

func (fi *FacilitiesInitialisor) Initialise(ca *config.ConfigAccessor) error {
	fi.ConfigAccessor = ca

	fc := ca.ObjectVal("Facilities")
	fi.facilityStatus = fc
	fi.updateFrameworkLogLevel()

	if fc["ApplicationLogging"].(bool) {
		fi.AddFacility(new(logger.ApplicationLoggingFacilityBuilder))
	}

	fi.AddFacility(new(querymanager.QueryManagerFacilityBuilder))
	fi.AddFacility(new(httpserver.HttpServerFacilityBuilder))
	fi.AddFacility(new(jsonws.JsonWsFacilityBuilder))
	fi.AddFacility(new(serviceerror.ServiceErrorManagerFacilityBuilder))
	fi.AddFacility(new(rdbms.RdbmsAccessFacilityBuilder))

	err := fi.buildEnabledFacilities()

	return err
}

func (fi *FacilitiesInitialisor) updateFrameworkLogLevel() error {

	flm := fi.FrameworkLoggingManager

	defaultLogLevelLabel, err := fi.ConfigAccessor.StringVal("FrameworkLogger.DefaultLogLevel")

	if err != nil {
		return errors.New(configErrorPrefix + err.Error())
	}

	defaultLogLevel, err := logging.LogLevelFromLabel(defaultLogLevelLabel)

	if err != nil {
		return errors.New(configErrorPrefix + err.Error())
	}

	initialLogLevelsByComponent := fi.ConfigAccessor.ObjectVal("FrameworkLogger.ComponentLogLevels")

	flm.InitalComponentLogLevels = initialLogLevelsByComponent
	flm.UpdateGlobalThreshold(defaultLogLevel)
	flm.UpdateLocalThreshold(defaultLogLevel)

	fld := new(decorator.FrameworkLogDecorator)
	fld.FrameworkLogger = flm.CreateLogger(frameworkLoggerDecoratorName)
	fld.LoggerManager = flm

	fi.container.WrapAndAddProto(frameworkLoggerDecoratorName, fld)

	return nil

}
