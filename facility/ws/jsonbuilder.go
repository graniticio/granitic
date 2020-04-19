// Copyright 2016-2020 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package ws

import (
	"errors"
	"fmt"
	"github.com/graniticio/granitic/v2/config"
	"github.com/graniticio/granitic/v2/instance"
	"github.com/graniticio/granitic/v2/ioc"
	"github.com/graniticio/granitic/v2/logging"
	"github.com/graniticio/granitic/v2/ws"
	"github.com/graniticio/granitic/v2/ws/json"
)

const jsonResponseWriterComponentName = instance.FrameworkPrefix + "JSONResponseWriter"
const jsonUnmarshallerComponentName = instance.FrameworkPrefix + "JSONUnmarshaller"

const modeWrap = "WRAP"
const modeBody = "BODY"

// JSONFacilityBuilder creates the components required to support the JSONWs facility and adds them the IoC container.
type JSONFacilityBuilder struct {
}

// BuildAndRegister implements FacilityBuilder.BuildAndRegister
func (fb *JSONFacilityBuilder) BuildAndRegister(lm *logging.ComponentLoggerManager, ca *config.Accessor, cn *ioc.ComponentContainer) error {

	wc, err := buildAndRegisterWsCommon(lm, ca, cn)

	if err != nil {
		return err
	}

	um := new(json.Unmarshaller)
	cn.WrapAndAddProto(jsonUnmarshallerComponentName, um)

	rw := new(ws.MarshallingResponseWriter)
	ca.Populate("JSONWs.ResponseWriter", rw)
	cn.WrapAndAddProto(jsonResponseWriterComponentName, rw)

	rw.StatusDeterminer = wc.StatusDeterminer
	rw.FrameworkErrors = wc.FrameworkErrors

	buildRegisterWsDecorator(cn, rw, um, wc, lm)

	if !cn.ModifierExists(jsonResponseWriterComponentName, "ErrorFormatter") {
		rw.ErrorFormatter = new(json.GraniticJSONErrorFormatter)
	}

	if !cn.ModifierExists(jsonResponseWriterComponentName, "ResponseWrapper") {

		// User hasn't defined their own wrapper for JSON responses, use one of the defaults
		if mode, err := ca.StringVal("JSONWs.WrapMode"); err == nil {
			var wrap ws.ResponseWrapper

			switch mode {
			case modeBody:
				wrap = new(json.BodyOrErrorWrapper)
			case modeWrap:
				wrap = new(json.GraniticJSONResponseWrapper)
			default:
				m := fmt.Sprintf("JSONWs.WrapMode must be either %s or %s", modeWrap, modeBody)

				return errors.New(m)
			}

			ca.Populate("JSONWs.ResponseWrapper", wrap)
			rw.ResponseWrapper = wrap
		} else {
			return err
		}

	}

	if !cn.ModifierExists(jsonResponseWriterComponentName, "MarshalingWriter") {

		mw := new(json.MarshalingWriter)
		ca.Populate("JSONWs.Marshal", mw)
		rw.MarshalingWriter = mw
	}

	offerAbnormalStatusWriter(rw, cn, jsonResponseWriterComponentName)

	return nil
}

// FacilityName implements FacilityBuilder.FacilityName
func (fb *JSONFacilityBuilder) FacilityName() string {
	return "JSONWs"
}

// DependsOnFacilities implements FacilityBuilder.DependsOnFacilities
func (fb *JSONFacilityBuilder) DependsOnFacilities() []string {
	return []string{}
}
