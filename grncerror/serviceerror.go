// Copyright 2016-2022 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

/*
Package grncerror defines error-message management types and error handling functions.

The primary type in this package is ServiceErrorManager, which allows an application to manage error definitions (messages and categories)
in a single location and have them looked up and referred to by error codes throughout the application.

A ServiceErrorManager is made available to applications by enabling the ServiceErrorManager facility. This facility is
documented in detail here: https://granitic.io/ref/service-error-management the package documentation for facility/serviceerror
gives a brief example of how to define errors in your application.
*/
package grncerror

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/graniticio/granitic/v3/logging"
	"github.com/graniticio/granitic/v3/types"
	"github.com/graniticio/granitic/v3/ws"
	"strings"
)

// ErrorCodeUser is implemented by components that want to decare that they use error codes, so that all codes they
// use can be validated to make sure that they have corresponding definitions.
type ErrorCodeUser interface {
	// ErrorCodesInUse returns the set of error codes that this component relies on and the component's name
	ErrorCodesInUse() (codes types.StringSet, component string)

	// ValidateMissing returns false if the component does not want the codes it uses checked.
	ValidateMissing() bool
}

// ServiceErrorManager contains a map between an error code and a ws.CategorisedError.
type ServiceErrorManager struct {
	errors map[string]*ws.CategorisedError

	// Logger used by Granitic framework components. Automatically injected.
	FrameworkLogger logging.Logger

	// Determines whether or not a panic should be triggered if a method on this type is called with
	// an error code that is not stored in the map of codes to errors.
	PanicOnMissing   bool
	errorCodeSources []ErrorCodeUser
	componentName    string
}

// ComponentName implements ioc.ComponentNamer.ComponentName
func (sem *ServiceErrorManager) ComponentName() string {
	return sem.componentName
}

// SetComponentName ioc.ComponentNamer.SetComponentName
func (sem *ServiceErrorManager) SetComponentName(name string) {
	sem.componentName = name
}

// Find returns the CategorisedError associated with the supplied code. If the code does not exist and PanicOnMissing
// is false, nil is returned. If PanicOnMissing is true the goroutine panics.
func (sem *ServiceErrorManager) Find(code string) *ws.CategorisedError {
	e := sem.errors[code]

	if e == nil {
		message := fmt.Sprintf("%s could not find error with code %s", sem.componentName, code)

		if sem.PanicOnMissing {
			panic(message)

		} else {
			sem.FrameworkLogger.LogWarnf(message)

		}

	}

	return e

}

// LoadErrors parses error definitions from the supplied definitions which will be cast from []interface to [][]string
// Each element of the sub-array is expected to be a []string with three elements.
func (sem *ServiceErrorManager) LoadErrors(definitions []interface{}) {

	l := sem.FrameworkLogger
	sem.errors = make(map[string]*ws.CategorisedError)

	for i, d := range definitions {

		e := d.([]interface{})

		category, err := ws.CodeToCategory(e[0].(string))

		if err != nil {
			l.LogWarnf("Error index %d: %s", i, err.Error())
			continue
		}

		code := e[1].(string)

		if len(strings.TrimSpace(code)) == 0 {
			l.LogWarnf("Error index %d: No code supplied", i)
			continue

		} else if sem.errors[code] != nil {
			l.LogWarnf("Error index %d: Duplicate code", i)
			continue
		}

		message := e[2].(string)

		if len(strings.TrimSpace(message)) == 0 {
			l.LogWarnf("Error index %d: No message supplied", i)
			continue
		}

		ce := ws.NewCategorisedError(category, code, message)

		sem.errors[code] = ce

	}
}

// RegisterCodeUser accepts a reference to a component ErrorCodeUser so that the set of error codes actually in use
// can be monitored.
func (sem *ServiceErrorManager) RegisterCodeUser(ecu ErrorCodeUser) {
	if sem.errorCodeSources == nil {
		sem.errorCodeSources = make([]ErrorCodeUser, 0)
	}

	sem.errorCodeSources = append(sem.errorCodeSources, ecu)
}

// AllowAccess is called by the IoC container after all components have been configured and started. At this point
// all of the error codes that have been declared to be in use can be compared with the available error code definitions.
//
// If there are any codes that are in use, but do not have a corresponding error definition an error will be returned.
func (sem *ServiceErrorManager) AllowAccess() error {

	failed := make(map[string][]string)

	for _, es := range sem.errorCodeSources {

		c, n := es.ErrorCodesInUse()

		for _, ec := range c.Contents() {

			if sem.errors[ec] == nil && es.ValidateMissing() {
				addMissingCode(n, ec, failed)
			}

		}

	}

	if len(failed) > 0 {

		var m bytes.Buffer

		m.WriteString(fmt.Sprintf("Some components are using error codes that do not have a corresponding error message: \n"))

		for k, v := range failed {

			m.WriteString(fmt.Sprintf("%s: %q\n", k, v))
		}

		return errors.New(m.String())

	}

	return nil
}

func addMissingCode(source, code string, failed map[string][]string) {

	fs := failed[source]

	if fs == nil {
		fs = make([]string, 0)
	}

	fs = append(fs, code)

	failed[source] = fs

}
