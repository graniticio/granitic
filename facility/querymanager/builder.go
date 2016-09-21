package querymanager

import (
	"github.com/graniticio/granitic/config"
	"github.com/graniticio/granitic/dbquery"
	"github.com/graniticio/granitic/instance"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/logging"
)

const QueryManagerComponentName = instance.FrameworkPrefix + "QueryManager"
const QueryManagerFacilityName = "QueryManager"

type QueryManagerFacilityBuilder struct {
}

func (qmfb *QueryManagerFacilityBuilder) BuildAndRegister(lm *logging.ComponentLoggerManager, ca *config.ConfigAccessor, cn *ioc.ComponentContainer) error {

	queryManager := new(querymanager.TemplatedQueryManager)
	ca.Populate("QueryManager", queryManager)

	cn.WrapAndAddProto(QueryManagerComponentName, queryManager)

	return nil
}

func (qmfb *QueryManagerFacilityBuilder) FacilityName() string {
	return QueryManagerFacilityName
}

func (qmfb *QueryManagerFacilityBuilder) DependsOnFacilities() []string {
	return []string{}
}
