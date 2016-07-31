package facility

import (
	"github.com/graniticio/granitic/config"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/logging"
)

type FacilityBuilder interface {
	BuildAndRegister(lm *logging.ComponentLoggerManager, ca *config.ConfigAccessor, cn *ioc.ComponentContainer) error
	FacilityName() string
	DependsOnFacilities() []string
}
