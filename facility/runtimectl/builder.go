package runtimectl

import (
	"github.com/graniticio/granitic/config"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/logging"
)

type RuntimeCtlFacilityBuilder struct {
}

func (fb *RuntimeCtlFacilityBuilder) BuildAndRegister(lm *logging.ComponentLoggerManager, ca *config.ConfigAccessor, cn *ioc.ComponentContainer) error {
	return nil
}

func (fb *RuntimeCtlFacilityBuilder) FacilityName() string {
	return "RuntimeCtl"
}

func (fb *RuntimeCtlFacilityBuilder) DependsOnFacilities() []string {
	return []string{}
}
