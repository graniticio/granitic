package json

import (
	"encoding/json"
	"github.com/graniticio/granitic/ws"
	"net/http"
)

type JSONMarshalingWriter struct {
	PrettyPrint  bool
	IndentString string
	PrefixString string
}

func (mw *JSONMarshalingWriter) MarshalAndWrite(data interface{}, w http.ResponseWriter) error {

	var b []byte
	var err error

	if mw.PrettyPrint {
		b, err = json.MarshalIndent(data, mw.PrefixString, mw.IndentString)
	} else {
		b, err = json.Marshal(data)
	}

	if err != nil {
		return err
	}

	_, err = w.Write(b)

	return err

}

type errorWrapper struct {
	Code    string
	Message string
}

type StandardJSONResponseWrapper struct {
	ErrorsFieldName string
	BodyFieldName   string
}

func (rw *StandardJSONResponseWrapper) WrapResponse(body interface{}, errors interface{}) interface{} {
	f := make(map[string]interface{})

	if errors != nil {
		f[rw.ErrorsFieldName] = errors
	}

	if body != nil {
		f[rw.BodyFieldName] = body
	}

	return f
}

type StandardJSONErrorFormatter struct{}

func (ef *StandardJSONErrorFormatter) FormatErrors(errors *ws.ServiceErrors) interface{} {

	if errors == nil || !errors.HasErrors() {
		return nil
	}

	f := make(map[string]interface{})

	generalErrors := make([]errorWrapper, 0)
	fieldErrors := make(map[string][]errorWrapper, 0)

	for _, error := range errors.Errors {

		c := ws.CategoryToCode(error.Category)
		displayCode := c + "-" + error.Code

		field := error.Field

		if field == "" {
			generalErrors = append(generalErrors, errorWrapper{displayCode, error.Message})
		} else {

			fe := fieldErrors[field]

			if fe == nil {
				fe = make([]errorWrapper, 0)

			}

			fe = append(fe, errorWrapper{displayCode, error.Message})
			fieldErrors[field] = fe

		}
	}

	if len(generalErrors) > 0 {
		f["General"] = generalErrors
	}

	if len(fieldErrors) > 0 {
		f["ByField"] = fieldErrors
	}

	return f
}
