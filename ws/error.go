// Copyright 2016-2023 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package ws

import (
	"errors"
	"fmt"
)

// ServiceErrorCategory indicates the broad 'type' of a service error, used to determine the correct HTTP status code to use.
type ServiceErrorCategory int

const (
	// Unexpected is an unhandled error that will generally result in an HTTP 500 status code being set.
	Unexpected = iota

	// Client is a problem that the calling client has caused or could have foreseen, generally resulting in an HTTP 400 status code.
	Client

	// Logic is a problem that the calling client could not be expected to have foreseen (email address in use, for example) resulting in an HTTP 409.
	Logic

	// Security is an access or authentication error that might result in an HTTP 401 or 403.
	Security

	// HTTP is an error that forces a specific HTTP status code.
	HTTP
)

// CategorisedError is a service error with a concept of the general 'type' of error it is.
type CategorisedError struct {
	// The broad type of error, which influences the eventual HTTP status code set on the response.
	Category ServiceErrorCategory

	// A unique code that a caller can rely on to identify a specific error or that can be used to lookup an error message.
	Code string

	// A message suitable for displaying to the caller.
	Message string

	//If this error relates to a specific field or parameter in a web service request, this field is set to the name of that field.
	Field string
}

// NewCategorisedError creates a new CategorisedError with every field expect 'Field' set.
func NewCategorisedError(category ServiceErrorCategory, code string, message string) *CategorisedError {
	ce := new(CategorisedError)

	ce.Category = category
	ce.Code = code
	ce.Message = message

	return ce
}

// ServiceErrorFinder is implemented by a component that is able to find a message and error category given the code for an error
type ServiceErrorFinder interface {
	//Find takes a code and returns the message and category for that error. Behaviour undefined if code is not
	// recognised.
	Find(code string) *CategorisedError
}

// ServiceErrorConsumer is implemented by components that require a ServiceErrorFinder to be injected into them
type ServiceErrorConsumer interface {
	// ProvideErrorFinder receives a ServiceErrorFinder
	ProvideErrorFinder(finder ServiceErrorFinder)
}

// ServiceErrors is a structure that records each of the errors found during the processing of a request.
type ServiceErrors struct {
	// All services found, in the order in which they occurred.
	Errors []CategorisedError

	// An externally computed HTTP status code that reflects the mix of errors in this structure.
	HTTPStatus int

	// A component able to find additional information about error from that error's unique code.
	ErrorFinder ServiceErrorFinder
}

// AddNewError creates a new CategorisedError from the supplied information and captures it.
func (se *ServiceErrors) AddNewError(category ServiceErrorCategory, label string, message string) {

	error := CategorisedError{category, label, message, ""}

	se.Errors = append(se.Errors, error)

}

// AddError records the supplied error.
func (se *ServiceErrors) AddError(e *CategorisedError) {

	se.Errors = append(se.Errors, *e)

}

// AddPredefinedError creates a CategorisedError by looking up the supplied code and records that error. If the variadic field
// parameter is supplied, the created error will be associated with that field name.
func (se *ServiceErrors) AddPredefinedError(code string, field ...string) error {

	if se.ErrorFinder == nil {
		panic("No source of errors defined")
	}

	e := se.ErrorFinder.Find(code)

	if len(field) > 0 {
		e.Field = field[0]
	}

	if e == nil {
		message := fmt.Sprintf("An error occurred with code %s, but no error message is available", code)
		e = NewCategorisedError(Unexpected, code, message)

	}

	se.Errors = append(se.Errors, *e)

	return nil
}

// HasErrors returns true if one or more errors have been encountered and recorded.
func (se *ServiceErrors) HasErrors() bool {
	return len(se.Errors) != 0
}

// CodeToCategory takes the short form of a category's name (its first letter, capitialised) an maps
// that to a ServiceErrorCategory
func CodeToCategory(c string) (ServiceErrorCategory, error) {

	switch c {
	case "U":
		return Unexpected, nil
	case "C":
		return Client, nil
	case "L":
		return Logic, nil
	case "S":
		return Security, nil
	default:
		message := fmt.Sprintf("Unknown error category %s", c)
		return -1, errors.New(message)
	}

}

// CategoryToCode maps a ServiceErrorCategory to the category's name's first letter. For example, Security maps to
// 'S'
func CategoryToCode(c ServiceErrorCategory) string {
	switch c {
	default:
		return "?"
	case Unexpected:
		return "U"
	case Security:
		return "S"
	case Logic:
		return "L"
	case Client:
		return "C"
	case HTTP:
		return "H"
	}
}

// CategoryToName maps a ServiceErrorCategory to the category's name's first letter. For example, Security maps to
// 'Security'
func CategoryToName(c ServiceErrorCategory) string {
	switch c {
	default:
		return "Unknown"
	case Unexpected:
		return "Unexpected"
	case Security:
		return "Security"
	case Logic:
		return "Logic"
	case Client:
		return "Client"
	case HTTP:
		return "HTTP"
	}
}
