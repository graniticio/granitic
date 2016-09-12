package xml

import (
	"encoding/xml"
	"github.com/graniticio/granitic/ws"
	"net/http"
)

type XMLMarshalingWriter struct {
	PrettyPrint  bool
	IndentString string
	PrefixString string
}

func (mw *XMLMarshalingWriter) MarshalAndWrite(data interface{}, w http.ResponseWriter) error {

	var b []byte
	var err error

	if mw.PrettyPrint {
		b, err = xml.MarshalIndent(data, mw.PrefixString, mw.IndentString)
	} else {
		b, err = xml.Marshal(data)
	}

	if err != nil {
		return err
	}

	_, err = w.Write(b)

	return err

}

type StandardXMLResponseWrapper struct {
}

func (rw *StandardXMLResponseWrapper) WrapResponse(body interface{}, errors interface{}) interface{} {

	w := new(xmlWrapper)

	w.XMLName = xml.Name{"", "response"}
	w.Body = body
	w.Errors = errors

	return w

}

type xmlWrapper struct {
	XMLName xml.Name
	Errors  interface{}
	Body    interface{} `xml:"body"`
}

type StandardXMLErrorFormatter struct{}

func (ef *StandardXMLErrorFormatter) FormatErrors(errors *ws.ServiceErrors) interface{} {

	if errors == nil || !errors.HasErrors() {
		return nil
	}

	es := new(xmlErrors)
	es.XMLName = xml.Name{"", "errors"}

	fe := make([]*xmlError, len(errors.Errors))

	for i, se := range errors.Errors {

		e := new(xmlError)
		e.XMLName = xml.Name{"", "error"}

		fe[i] = e
		e.Error = se.Message
		e.Field = se.Field
		e.Category = ws.CategoryToName(se.Category)
		e.Code = se.Label

	}

	es.Errors = fe

	return es
}

type xmlErrors struct {
	XMLName xml.Name
	Errors  interface{}
}

type xmlError struct {
	XMLName  xml.Name
	Error    string `xml:",chardata"`
	Field    string `xml:"field,attr,omitempty"`
	Code     string `xml:"code,attr,omitempty"`
	Category string `xml:"category,attr,omitempty"`
}
