// Copyright 2016-2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package ws

import (
	"fmt"
	"github.com/graniticio/granitic/v2/logging"
	"strconv"
)

// FrameworkPhase is the phase of the request processing during which an error was encountered.
type FrameworkPhase int

const (
	// Unmarshall indicates an error was encountered while trying to parse an HTTP request body into a struct
	Unmarshall = iota

	// QueryBind indicates an error was encountered while mapping HTTP query parameters to fields on a struct
	QueryBind

	//PathBind indicates an error was encountered while mapping elements of an HTTP request's path to fields on a struct
	PathBind
)

// FieldAssociatedError is implemented by types that can record which field on a struct caused a problem
type FieldAssociatedError interface {
	// RecordField captures the field that was involved in the error
	RecordField(string)
}

// FrameworkError an error encountered in early phases of request processing, before application code is invoked.
type FrameworkError struct {
	// The phase of the request processing during which an error was encountered.
	Phase FrameworkPhase

	// The name of the field or parameter in the HTTP request with a problem
	ClientField string

	// The name of the field on the response body struct that was being written to
	TargetField string

	// A system generated message relating to the error.
	Message string

	// The value of the field/parameter that caused the error.
	Value string

	// For array parameters, the position in the array that caused the error.
	Position int

	// A system generated code for the error.
	Code string
}

// RecordField implements FieldAssociatedError
func (fe *FrameworkError) RecordField(f string) {
	fe.TargetField = f
}

// Error returns a string representation of a FrameworkError
func (fe *FrameworkError) Error() string {
	return fmt.Sprintf("%v %s=%s: %s", fe.Phase, fe.TargetField, fe.Value, fe.Message)
}

// NewUnmarshallFrameworkError creates a FrameworkError with fields set appropriate for an error
// encountered during parsing of the HTTP request body.
func NewUnmarshallFrameworkError(message, code string) *FrameworkError {
	f := new(FrameworkError)
	f.Phase = Unmarshall
	f.Message = message
	f.Code = code

	return f
}

// NewQueryBindFrameworkError creates a FrameworkError with fields set appropriate for an error
// encountered during mapping of HTTP query parameters to fields on a Request's Body
func NewQueryBindFrameworkError(message, code, param, target string) *FrameworkError {
	f := new(FrameworkError)
	f.Phase = QueryBind
	f.Message = message
	f.ClientField = param
	f.TargetField = target
	f.Code = code

	return f
}

// NewPathBindFrameworkError creates a FrameworkError with fields set appropriate for an error
// encountered during mapping elements of the HTTP request's path to fields on a Request's Body.
func NewPathBindFrameworkError(message, code, target string) *FrameworkError {
	f := new(FrameworkError)
	f.Phase = PathBind
	f.Message = message
	f.TargetField = target
	f.Code = code

	return f
}

// FrameworkErrorEvent uniquely identifies a 'handled' failure during the parsing and binding phases
type FrameworkErrorEvent string

const (
	// UnableToParseRequest indicates that the HTTP request could not be parsed
	UnableToParseRequest = "UnableToParseRequest"
	// QueryTargetNotArray indicates that a query parameter whose value is a list has been bound to a target field that is not an array
	QueryTargetNotArray = "QueryTargetNotArray"

	// QueryWrongType indicates that a query parameter is not compatible with the type of field to which it is bound
	QueryWrongType = "QueryWrongType"

	// PathWrongType indicates that a path element is not compatible with the type of field to which it is bound
	PathWrongType = "PathWrongType"

	// QueryNoTargetField indicates that no field on the target can be matched to the a named query parameter
	QueryNoTargetField = "QueryNoTargetField"
)

// A FrameworkErrorGenerator can create error messages for errors that occur outside of application code and messages
// that should be displayed when generic HTTP status codes (404, 500, 503 etc) are set.
type FrameworkErrorGenerator struct {
	Messages        map[FrameworkErrorEvent][]string
	HTTPMessages    map[string]string
	FrameworkLogger logging.Logger
}

// HTTPError generates a message to be displayed to a caller when a generic HTTP status (404 etc) is encountered. If
// an error message is not defined for the supplied status, the message "HTTP (code)" is returned, e.g. "HTTP 101"
func (feg *FrameworkErrorGenerator) HTTPError(status int, a ...interface{}) *CategorisedError {

	s := strconv.Itoa(status)

	m := feg.HTTPMessages[s]

	if m == "" {
		m = "HTTP " + s
	} else {
		m = fmt.Sprintf(m, a...)
	}

	ce := new(CategorisedError)

	ce.Category = HTTP
	ce.Code = s
	ce.Message = m

	return ce

}

// Error creates a service error given a framework error.
func (feg *FrameworkErrorGenerator) Error(e FrameworkErrorEvent, c ServiceErrorCategory, a ...interface{}) *CategorisedError {

	l := feg.FrameworkLogger
	mc := feg.Messages[e]

	if mc == nil || len(mc) < 2 {
		l.LogWarnf("No framework error message defined for '%s'. Returning a default message.")
		return NewCategorisedError(c, "UNKNOWN", "No error message defined for this error")
	}

	cd := mc[0]
	t := mc[1]

	fm := fmt.Sprintf(t, a...)

	return NewCategorisedError(c, cd, fm)

}

// MessageCode returns a message and code for a Framework error event (leaving the caller to create a CategorisedError)
func (feg *FrameworkErrorGenerator) MessageCode(e FrameworkErrorEvent, a ...interface{}) (message string, code string) {

	l := feg.FrameworkLogger
	mc := feg.Messages[e]

	if mc == nil || len(mc) < 2 {
		l.LogWarnf("No framework error message defined for '%s'. Returning a default message.")
		return "No error message defined for this error", "UNKNOWN"
	}

	t := mc[1]

	return fmt.Sprintf(t, a...), mc[0]

}
