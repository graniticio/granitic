// Copyright 2016-2019 Granitic. All rights reserved.
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

const mode_wrap = "WRAP"
const mode_body = "BODY"

// Creates the components required to support the JsonWs facility and adds them the IoC container.
type JsonFacilityBuilder struct {
}

// See FacilityBuilder.BuildAndRegister
func (fb *JsonFacilityBuilder) BuildAndRegister(lm *logging.ComponentLoggerManager, ca *config.Accessor, cn *ioc.ComponentContainer) error {

	wc := buildAndRegisterWsCommon(lm, ca, cn)

	um := new(json.Unmarshaller)
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

		// User hasn't defined their own wrapper for JSON responses, use one of the defaults
		if mode, err := ca.StringVal("JsonWs.WrapMode"); err != nil {
			return err
		} else {

			var wrap ws.ResponseWrapper

			switch mode {
			case mode_body:
				wrap = new(json.BodyOrErrorWrapper)
			case mode_wrap:
				wrap = new(json.GraniticJSONResponseWrapper)
			default:
				m := fmt.Sprintf("JsonWs.WrapMode must be either %s or %s", mode_wrap, mode_body)

				return errors.New(m)
			}

			ca.Populate("JsonWs.ResponseWrapper", wrap)
			rw.ResponseWrapper = wrap

		}

	}

	if !cn.ModifierExists(jsonResponseWriterComponentName, "MarshalingWriter") {

		mw := new(json.MarshalingWriter)
		ca.Populate("JsonWs.Marshal", mw)
		rw.MarshalingWriter = mw
	}

	offerAbnormalStatusWriter(rw, cn, jsonResponseWriterComponentName)

	return nil
}

// See FacilityBuilder.FacilityName
func (fb *JsonFacilityBuilder) FacilityName() string {
	return "JsonWs"
}

// See FacilityBuilder.DependsOnFacilities
func (fb *JsonFacilityBuilder) DependsOnFacilities() []string {
	return []string{}
}
