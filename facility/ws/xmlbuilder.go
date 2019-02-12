// Copyright 2016-2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package ws

import (
	"errors"
	"github.com/graniticio/granitic/v2/config"
	"github.com/graniticio/granitic/v2/instance"
	"github.com/graniticio/granitic/v2/ioc"
	"github.com/graniticio/granitic/v2/logging"
	"github.com/graniticio/granitic/v2/ws"
	"github.com/graniticio/granitic/v2/ws/xml"
)

const (
	xmlResponseWriterName = instance.FrameworkPrefix + "XMLResponseWriter"
	xmlUnmarshallerName   = instance.FrameworkPrefix + "XMLUnmarshaller"
	templateMode          = "TEMPLATE"
	marshalMode           = "MARSHAL"
)

// XMLFacilityBuilder creates the components required to support the XMLWs facility and adds them the IoC container.
type XMLFacilityBuilder struct {
}

// BuildAndRegister implements FacilityBuilder.BuildAndRegister
func (fb *XMLFacilityBuilder) BuildAndRegister(lm *logging.ComponentLoggerManager, ca *config.Accessor, cc *ioc.ComponentContainer) error {

	wc := buildAndRegisterWsCommon(lm, ca, cc)

	um := new(xml.Unmarshaller)
	cc.WrapAndAddProto(xmlUnmarshallerName, um)

	mode, _ := ca.StringVal("XMLWs.ResponseMode")

	var rw ws.ResponseWriter

	switch mode {
	case templateMode:
		rw = fb.createTemplateComponents(ca, cc, wc)
	case marshalMode:
		rw = fb.createMarshalComponents(ca, cc, wc)
	default:
		return errors.New("XMLWs.ResponseMode must be set to either TEMPLATE or MARSHAL")
	}

	buildRegisterWsDecorator(cc, rw, um, wc, lm)
	offerAbnormalStatusWriter(rw.(ws.AbnormalStatusWriter), cc, xmlResponseWriterName)

	return nil
}

func (fb *XMLFacilityBuilder) createTemplateComponents(ca *config.Accessor, cc *ioc.ComponentContainer, wc *wsCommon) ws.ResponseWriter {

	rw := new(xml.TemplatedXMLResponseWriter)
	ca.Populate("XMLWs.ResponseWriter", rw)
	cc.WrapAndAddProto(xmlResponseWriterName, rw)

	rw.FrameworkErrors = wc.FrameworkErrors
	rw.StatusDeterminer = wc.StatusDeterminer

	return rw

}

func (fb *XMLFacilityBuilder) createMarshalComponents(ca *config.Accessor, cc *ioc.ComponentContainer, wc *wsCommon) ws.ResponseWriter {

	rw := new(ws.MarshallingResponseWriter)
	ca.Populate("XMLWs.ResponseWriter", rw)
	cc.WrapAndAddProto(xmlResponseWriterName, rw)

	rw.StatusDeterminer = wc.StatusDeterminer
	rw.FrameworkErrors = wc.FrameworkErrors

	if !cc.ModifierExists(xmlResponseWriterName, "ErrorFormatter") {
		rw.ErrorFormatter = new(xml.GraniticXMLErrorFormatter)
	}

	if !cc.ModifierExists(xmlResponseWriterName, "ResponseWrapper") {
		wrap := new(xml.GraniticXMLResponseWrapper)
		rw.ResponseWrapper = wrap
	}

	if !cc.ModifierExists(xmlResponseWriterName, "MarshalingWriter") {

		mw := new(xml.MarshalingWriter)
		ca.Populate("XMLWs.Marshal", mw)
		rw.MarshalingWriter = mw
	}

	return rw

}

// FacilityName implements FacilityBuilder.FacilityName
func (fb *XMLFacilityBuilder) FacilityName() string {
	return "XMLWs"
}

// DependsOnFacilities implements FacilityBuilder.DependsOnFacilities
func (fb *XMLFacilityBuilder) DependsOnFacilities() []string {
	return []string{}
}
