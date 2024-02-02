// Copyright 2016-2022 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package ws

import (
	"errors"
	"fmt"
	config_access "github.com/graniticio/config-access"
	"github.com/graniticio/granitic/v3/instance"
	"github.com/graniticio/granitic/v3/ioc"
	"github.com/graniticio/granitic/v3/logging"
	"github.com/graniticio/granitic/v3/ws"
	"github.com/graniticio/granitic/v3/ws/json"
)

const jsonResponseWriterComponentName = instance.FrameworkPrefix + "JSONResponseWriter"
const jsonUnmarshallerComponentName = instance.FrameworkPrefix + "JSONUnmarshaller"

const modeWrap = "WRAP"
const modeBody = "BODY"

// JSONFacilityBuilder creates the components required to support the JSONWs facility and adds them the IoC container.
type JSONFacilityBuilder struct {
}

// BuildAndRegister implements FacilityBuilder.BuildAndRegister
func (fb *JSONFacilityBuilder) BuildAndRegister(lm *logging.ComponentLoggerManager, ca config_access.Selector, cn *ioc.ComponentContainer) error {

	wc, err := buildAndRegisterWsCommon(lm, ca, cn)

	if err != nil {
		return err
	}

	um := new(json.Unmarshaller)
	cn.WrapAndAddProto(jsonUnmarshallerComponentName, um)

	rw := new(ws.MarshallingResponseWriter)
	config_access.Populate("JSONWs.ResponseWriter", rw, ca.Config())
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

			config_access.Populate("JSONWs.ResponseWrapper", wrap, ca.Config())
			rw.ResponseWrapper = wrap
		} else {
			return err
		}

	}

	if !cn.ModifierExists(jsonResponseWriterComponentName, "MarshalingWriter") {

		mw := new(json.MarshalingWriter)
		config_access.Populate("JSONWs.Marshal", mw, ca.Config())
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
