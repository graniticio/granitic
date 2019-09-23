// Copyright 2016-2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

/*
Package serviceerror provides the ServiceErrorManager facility which provides error message management.

This facility is documented in detail at https://granitic.io/ref/service-error-management

Error definitions

The purpose of the ServiceErrorManager is to allow error messages to be defined in a single place rather than be embedded in code.
When application code needs to raise or work with an error, a short code is used to refer to the error and is used to lookup its related message.

Error definitions are stored in your application's configuration files and will look something like:

	{
	  "serviceErrors": [
		["C", "CREATE_RECORD", "Cannot create a record with the information provided."],
		["C", "RECORD_NAME", "Record names must be 1-128 characters long."],
		["C", "ARTIST_NAME", "Artist names must be 1-64 characters long."]
	  ]
	}

This example defines three errors with three parts that correspond to the fields in a ws.CategorisedError. The first element
of each definition is the short-hand form of the error category (see ws.ServiceErrorCategory), the second is a unique code for the error
and the third is a (normally human readable) error message.

Changing config location

By default, the ServiceErrorManager expects definitions to be available at the config path

	serviceErrors

as in the example above. This can be overridden by modifying the ServiceErrorManager facility configuration like:

	{
	  "ServiceErrorManager":{
		"ErrorDefinitions": "my.errors.path"
	  }
	}

Panic on missing

If your application code uses an error code that has no corresponding definition, you can choose how the ServiceErrorManager will
react. By default, the ServiceErrorManager will panic -  this should be a development/test phase failure. If for some reason your
application might reach production with error codes with missing definitions, you can set:

	{
	  "ServiceErrorManager":{
		"PanicOnMissing": false
	  }
	}

In this case, ServiceErrorManager will return nil when asked for the definition of an unknown code.

*/
package serviceerror

import (
	"github.com/graniticio/granitic/v2/grncerror"
	"github.com/graniticio/granitic/v2/ioc"
	"github.com/graniticio/granitic/v2/ws"
)

type consumerDecorator struct {
	ErrorSource *grncerror.ServiceErrorManager
}

func (secd *consumerDecorator) OfInterest(component *ioc.Component) bool {
	_, found := component.Instance.(ws.ServiceErrorConsumer)

	return found
}

func (secd *consumerDecorator) DecorateComponent(component *ioc.Component, container *ioc.ComponentContainer) {
	c := component.Instance.(ws.ServiceErrorConsumer)
	c.ProvideErrorFinder(secd.ErrorSource)
}

type errorCodeSourceDecorator struct {
	ErrorSource *grncerror.ServiceErrorManager
}

func (ecs *errorCodeSourceDecorator) OfInterest(component *ioc.Component) bool {
	s, found := component.Instance.(grncerror.ErrorCodeUser)

	if found {
		return s.ValidateMissing()
	}

	return found
}

func (ecs *errorCodeSourceDecorator) DecorateComponent(component *ioc.Component, container *ioc.ComponentContainer) {
	c := component.Instance.(grncerror.ErrorCodeUser)

	ecs.ErrorSource.RegisterCodeUser(c)
}
