// Copyright 2016 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package ws

import (
	"github.com/graniticio/granitic/config"
	"github.com/graniticio/granitic/instance"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/logging"
	"github.com/graniticio/granitic/ws"
	"github.com/graniticio/granitic/ws/json"
)

const jsonResponseWriterComponentName = instance.FrameworkPrefix + "JsonResponseWriter"
const jsonUnmarshallerComponentName = instance.FrameworkPrefix + "JsonUnmarshaller"

// Creates the components required to support the JsonWs facility and adds them the IoC container.
type JSONWsFacilityBuilder struct {
}

// See FacilityBuilder.BuildAndRegister
func (fb *JSONWsFacilityBuilder) BuildAndRegister(lm *logging.ComponentLoggerManager, ca *config.ConfigAccessor, cn *ioc.ComponentContainer) error {

	wc := buildAndRegisterWsCommon(lm, ca, cn)

	um := new(json.StandardJSONUnmarshaller)
	cn.WrapAndAddProto(jsonUnmarshallerComponentName, um)

	rw := new(ws.MarshallingResponseWriter)
	ca.Populate("JsonWs.ResponseWriter", rw)
	cn.WrapAndAddProto(jsonResponseWriterComponentName, rw)

	rw.StatusDeterminer = wc.StatusDeterminer
	rw.FrameworkErrors = wc.FrameworkErrors

	buildRegisterWsDecorator(cn, rw, um, wc, lm)

	if !cn.ModifierExists(jsonResponseWriterComponentName, "ErrorFormatter") {
		rw.ErrorFormatter = new(json.GraniticJSONErrorFormatter)
	}

	if !cn.ModifierExists(jsonResponseWriterComponentName, "ResponseWrapper") {
		wrap := new(json.GraniticJSONResponseWrapper)
		ca.Populate("JsonWs.ResponseWrapper", wrap)
		rw.ResponseWrapper = wrap
	}

	if !cn.ModifierExists(jsonResponseWriterComponentName, "MarshalingWriter") {

		mw := new(json.JSONMarshalingWriter)
		ca.Populate("JsonWs.Marshal", mw)
		rw.MarshalingWriter = mw
	}

	offerAbnormalStatusWriter(rw, cn, jsonResponseWriterComponentName)

	return nil
}

// See FacilityBuilder.FacilityName
func (fb *JSONWsFacilityBuilder) FacilityName() string {
	return "JsonWs"
}

// See FacilityBuilder.DependsOnFacilities
func (fb *JSONWsFacilityBuilder) DependsOnFacilities() []string {
	return []string{}
}
