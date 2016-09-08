/*
Package jsonws builds the components required to support JSON-based web-services.
*/
package ws

import (
	"github.com/graniticio/granitic/config"
	"github.com/graniticio/granitic/instance"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/logging"
	"github.com/graniticio/granitic/ws/json"
)

const jsonResponseWriterComponentName = instance.FrameworkPrefix + "JsonResponseWriter"
const jsonUnmarshallerComponentName = instance.FrameworkPrefix + "JsonUnmarshaller"

type JSONWsFacilityBuilder struct {
}

func (fb *JSONWsFacilityBuilder) BuildAndRegister(lm *logging.ComponentLoggerManager, ca *config.ConfigAccessor, cn *ioc.ComponentContainer) error {

	wc := BuildAndRegisterWsCommon(lm, ca, cn)

	um := new(json.StandardJSONUnmarshaller)
	cn.WrapAndAddProto(jsonUnmarshallerComponentName, um)

	rw := new(json.StandardJSONResponseWriter)
	ca.Populate("JsonWs.ResponseWriter", rw)
	cn.WrapAndAddProto(jsonResponseWriterComponentName, rw)

	rw.StatusDeterminer = wc.StatusDeterminer
	rw.FrameworkErrors = wc.FrameworkErrors

	BuildRegisterWsDecorator(cn, rw, um, wc, lm)

	if !cn.ModifierExists(jsonResponseWriterComponentName, "ErrorFormatter") {
		rw.ErrorFormatter = new(json.StandardJSONErrorFormatter)
	}

	if !cn.ModifierExists(jsonResponseWriterComponentName, "ResponseWrapper") {
		wrap := new(json.StandardJSONResponseWrapper)
		ca.Populate("JsonWs.ResponseWrapper", wrap)
		rw.ResponseWrapper = wrap
	}

	OfferAbnormalStatusWriter(rw, cn, jsonResponseWriterComponentName)

	return nil
}

func (fb *JSONWsFacilityBuilder) FacilityName() string {
	return "JsonWs"
}

func (fb *JSONWsFacilityBuilder) DependsOnFacilities() []string {
	return []string{}
}
