// Copyright 2016-2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

/*
Package json defines types that are specific to handling web service requests and responses as JSON. Components
implementing this type will be created when you enable the JSONWs facility. For more information on JSON web services
in Granitic, see http://granitic.io/ref/json-web-services

Marshalling and unmarshalling

The response writer and unmarshaller defined in this package are thin wrappers over the Go's built-in json handling
types. See https://golang.org/pkg/encoding/json

Response wrapping

By default, any data serialised to JSON will first be wrapped with a containing data structure by an instance of GraniticJSONResponseWrapper. This
means that all responses share a common top level structure for finding the body of the response or errors if they exist.
For more information on this behaviour (and how to override it) see: http://granitic.io/ref/json-web-services

Error formatting

Any service errors found in a response are formatted by GraniticJSONErrorFormatter before being serialised to JSON.
For more information on this behaviour (and how to override it) see: http://granitic.io/ref/json-web-services

Compatibility with existing service APIs

A hurdle to migrating existing Java and .NET services to Go is that those languages allow JSON frameworks to write and
read from member variables that start with lowercase characters. Go's rules for json decoding will map a JSON field with a
lowercase first letter into a struct field with an uppercase letter (e.g. name wil be parsed into Name).

No such logic exists for forcing Name to be serialised as name other than defining tags on your JSON struct. The CamelCase
method defined below can take an entire Go struct and create a copy of the object with all capitalised field names
replaced with lowercase equivalents.

The method must be explicitly called in your handler's logic like:

	wsResponse.Body = json.CamelCase(body)

This feature should be considered experimental.

*/
package json

import (
	"encoding/json"
	"github.com/graniticio/granitic/v2/ws"
	"net/http"
)

// MarshalingWriter is Component wrapper over Go's json.Marshalxx functions. Serialises a struct to JSON and writes it to the HTTP response
// output stream.
type MarshalingWriter struct {
	// Format generated JSON in a human readable form.
	PrettyPrint bool

	// The characters (generally tabs or spaces) to indent child elements in pretty-printed JSON.
	IndentString string

	// A prefix for each line of generated JSON.
	PrefixString string
}

// MarshalAndWrite serialises the supplied interface to JSON and writes it to the HTTP response output stream.
func (mw *MarshalingWriter) MarshalAndWrite(data interface{}, w http.ResponseWriter) error {

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

// BodyOrErrorWrapper is an implementation of ResponseWrapper that just returns the body object if not nil or the errors object if not nil
type BodyOrErrorWrapper struct {
}

// WrapResponse returns body if not nil or errors if not nil. Otherwise returns nil
func (rw *BodyOrErrorWrapper) WrapResponse(body interface{}, errors interface{}) interface{} {

	if body != nil {
		return body
	}

	if errors != nil {
		return errors
	}

	return nil
}

// GraniticJSONResponseWrapper is a component for wrapping response data before it is serialised. The wrapping structure is a map[string]string
type GraniticJSONResponseWrapper struct {
	ErrorsFieldName string
	BodyFieldName   string
}

// WrapResponse creates a map[string]string to wrap the supplied response body and errors.
func (rw *GraniticJSONResponseWrapper) WrapResponse(body interface{}, errors interface{}) interface{} {
	f := make(map[string]interface{})

	if errors != nil {
		f[rw.ErrorsFieldName] = errors
	}

	if body != nil {
		f[rw.BodyFieldName] = body
	}

	return f
}

// GraniticJSONErrorFormatter converts service errors into a data structure for consistent serialisation to JSON.
type GraniticJSONErrorFormatter struct{}

// FormatErrors converts all of the errors present in the supplied objects into a structure suitable for serialisation.
func (ef *GraniticJSONErrorFormatter) FormatErrors(errors *ws.ServiceErrors) interface{} {

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
