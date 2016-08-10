package ws

import (
	"errors"
	"fmt"
	"strconv"
)

type ServiceErrorCategory int

const (
	Unexpected = iota
	Client
	Logic
	Security
	HTTP
)

type CategorisedError struct {
	Category ServiceErrorCategory
	Label    string
	Message  string
}

func NewCategorisedError(category ServiceErrorCategory, label string, message string) *CategorisedError {
	ce := new(CategorisedError)

	ce.Category = category
	ce.Label = label
	ce.Message = message

	return ce
}

func NewHTTPError(status int, message string) *CategorisedError {
	ce := new(CategorisedError)

	ce.Category = HTTP
	ce.Label = strconv.Itoa(status)
	ce.Message = message

	return ce
}

type ServiceErrorFinder interface {
	Find(code string) *CategorisedError
}

type ServiceErrorConsumer interface {
	ProvideErrorFinder(finder ServiceErrorFinder)
}

type ServiceErrors struct {
	Errors      []CategorisedError
	HttpStatus  int
	ErrorFinder ServiceErrorFinder
}

func (se *ServiceErrors) AddNewError(category ServiceErrorCategory, label string, message string) {

	error := CategorisedError{category, label, message}

	se.Errors = append(se.Errors, error)

}

func (se *ServiceErrors) AddError(e *CategorisedError) {

	se.Errors = append(se.Errors, *e)

}

func (se *ServiceErrors) AddPredefinedError(code string) error {

	if se.ErrorFinder == nil {
		panic("No source of errors defined")
	}

	e := se.ErrorFinder.Find(code)

	if e == nil {
		message := fmt.Sprintf("An error occured with code %s, but no error message is available", code)
		e = NewCategorisedError(Unexpected, code, message)

	}

	se.Errors = append(se.Errors, *e)

	return nil
}

func (se *ServiceErrors) HasErrors() bool {
	return len(se.Errors) != 0
}

func CodeToCategory(code string) (ServiceErrorCategory, error) {

	switch code {
	case "U":
		return Unexpected, nil
	case "C":
		return Client, nil
	case "L":
		return Logic, nil
	case "S":
		return Security, nil
	default:
		message := fmt.Sprint("Unknown error category %s", code)
		return -1, errors.New(message)
	}

}
