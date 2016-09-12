package ws

import (
	"errors"
	"github.com/graniticio/granitic/config"
	"github.com/graniticio/granitic/instance"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/logging"
	"github.com/graniticio/granitic/ws/xml"
)

const (
	xmlResponseWriterName = instance.FrameworkPrefix + "XmlResponseWriter"
	xmlUnmarshallerName   = instance.FrameworkPrefix + "XmlUnmarshaller"
	templateMode          = "TEMPLATE"
	marshalMode           = "MARSHAL"
)

type XMLWsFacilityBuilder struct {
}

func (fb *XMLWsFacilityBuilder) BuildAndRegister(lm *logging.ComponentLoggerManager, ca *config.ConfigAccessor, cc *ioc.ComponentContainer) error {

	wc := BuildAndRegisterWsCommon(lm, ca, cc)

	um := new(xml.StandardXmlUnmarshaller)
	cc.WrapAndAddProto(xmlUnmarshallerName, um)

	mode, _ := ca.StringVal("XmlWs.ResponseMode")

	switch mode {
	case templateMode:
		fb.createTemplateComponents(ca, cc, wc, um, lm)
	case marshalMode:
	default:
		return errors.New("XmlWs.ResponseMode must be set to either TEMPLATE or MARHSAL")
	}

	return nil
}

func (fb *XMLWsFacilityBuilder) createTemplateComponents(ca *config.ConfigAccessor, cc *ioc.ComponentContainer, wc *WsCommon,
	um *xml.StandardXmlUnmarshaller, lm *logging.ComponentLoggerManager) {

	rw := new(xml.TemplatedXMLResponseWriter)
	ca.Populate("XmlWs.ResponseWriter", rw)
	cc.WrapAndAddProto(xmlResponseWriterName, rw)

	rw.FrameworkErrors = wc.FrameworkErrors
	rw.StatusDeterminer = wc.StatusDeterminer

	BuildRegisterWsDecorator(cc, rw, um, wc, lm)
	OfferAbnormalStatusWriter(rw, cc, xmlResponseWriterName)
}

func (fb *XMLWsFacilityBuilder) FacilityName() string {
	return "XmlWs"
}

func (fb *XMLWsFacilityBuilder) DependsOnFacilities() []string {
	return []string{}
}
