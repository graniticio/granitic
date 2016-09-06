// Package facility defines the high-level features that Granitic makes available to applications. Theses features are packaged into groups of components known as a
// 'facility'.
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
