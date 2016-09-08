package ws

import (
	"fmt"
	"github.com/graniticio/granitic/config"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/logging"
)

type XMLWsFacilityBuilder struct {
}

func (fb *XMLWsFacilityBuilder) BuildAndRegister(lm *logging.ComponentLoggerManager, ca *config.ConfigAccessor, cn *ioc.ComponentContainer) error {

	fmt.Println("Constructing XML facility")

	//wc := BuildAndRegisterWsCommon(lm, ca, cn)

	return nil
}

func (fb *XMLWsFacilityBuilder) FacilityName() string {
	return "XmlWs"
}

func (fb *XMLWsFacilityBuilder) DependsOnFacilities() []string {
	return []string{}
}
