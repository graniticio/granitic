package runtimectl

import (
	"github.com/graniticio/granitic/config"
	"github.com/graniticio/granitic/facility/httpserver"
	"github.com/graniticio/granitic/instance"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/logging"
	"github.com/graniticio/granitic/ws"
)

const (
	runtimeCtlServerName = instance.FrameworkPrefix + "CtlServer"
)

type RuntimeCtlFacilityBuilder struct {
}

func (fb *RuntimeCtlFacilityBuilder) BuildAndRegister(lm *logging.ComponentLoggerManager, ca *config.ConfigAccessor, cn *ioc.ComponentContainer) error {

	sv := new(httpserver.HTTPServer)
	ca.Populate("RuntimeCtl.Server", sv)

	cn.WrapAndAddProto(runtimeCtlServerName, sv)

	rw := new(ws.MarshallingResponseWriter)
	sv.AbnormalStatusWriter = rw

	return nil
}

func (fb *RuntimeCtlFacilityBuilder) FacilityName() string {
	return "RuntimeCtl"
}

func (fb *RuntimeCtlFacilityBuilder) DependsOnFacilities() []string {
	return []string{}
}
