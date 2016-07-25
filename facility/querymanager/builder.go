package querymanager

import (
	"github.com/graniticio/granitic/config"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/logging"
)

const QueryManagerComponentName = ioc.FrameworkPrefix + "QueryManager"
const QueryManagerFacilityName = "QueryManager"

type QueryManagerFacilityBuilder struct {
}

func (qmfb *QueryManagerFacilityBuilder) BuildAndRegister(lm *logging.ComponentLoggerManager, ca *config.ConfigAccessor, cn *ioc.ComponentContainer) {

	queryManager := new(QueryManager)
	ca.Populate("QueryManager", queryManager)

	cn.WrapAndAddProto(QueryManagerComponentName, queryManager)

}

func (qmfb *QueryManagerFacilityBuilder) FacilityName() string {
	return QueryManagerFacilityName
}

func (qmfb *QueryManagerFacilityBuilder) DependsOnFacilities() []string {
	return []string{}
}
