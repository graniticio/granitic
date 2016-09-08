package ws

import (
	"fmt"
	"github.com/graniticio/granitic/config"
	"github.com/graniticio/granitic/instance"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/logging"
	"github.com/graniticio/granitic/ws/xml"
)

const xmlResponseWriterName = instance.FrameworkPrefix + "XmlResponseWriter"
const xmlUnmarshallerName = instance.FrameworkPrefix + "XmlUnmarshaller"

type XMLWsFacilityBuilder struct {
}

func (fb *XMLWsFacilityBuilder) BuildAndRegister(lm *logging.ComponentLoggerManager, ca *config.ConfigAccessor, cc *ioc.ComponentContainer) error {

	fmt.Println("Constructing XML facility")

	wc := BuildAndRegisterWsCommon(lm, ca, cc)

	um := new(xml.StandardXmlUnmarshaller)
	cc.WrapAndAddProto(xmlUnmarshallerName, um)

	rw := new(xml.StandardXMLResponseWriter)
	ca.Populate("XmlWs.ResponseWriter", rw)
	cc.WrapAndAddProto(xmlResponseWriterName, rw)

	rw.FrameworkErrors = wc.FrameworkErrors
	rw.StatusDeterminer = wc.StatusDeterminer

	BuildRegisterWsDecorator(cc, rw, um, wc, lm)
	OfferAbnormalStatusWriter(rw, cc, xmlResponseWriterName)

	return nil
}

func (fb *XMLWsFacilityBuilder) FacilityName() string {
	return "XmlWs"
}

func (fb *XMLWsFacilityBuilder) DependsOnFacilities() []string {
	return []string{}
}
