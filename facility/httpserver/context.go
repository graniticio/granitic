// Copyright 2018 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.
package httpserver

import (
	"context"
	"github.com/graniticio/granitic/ioc"
	"net/http"
)

// Components implementing this interface are called by the HttpServer facility before the request is matched to a handler.
// It is an opportunity to extract information from the HTTP request to add an ID for this request to the context that
// will be passed to the handler.
//
// It is the responsibility of the implementor to control the uniqueness of the allocated ID
//
// It is recommended that only the request meta data (headers, path, parameters) are accessed by implementations, as
// loading the request body will interfere with later phases of the request processing.
type IdentifiedRequestContextBuilder interface {
	// Identify uses information in the supplied request to assign an ID to this request
	Identify(ctx context.Context, req *http.Request) (IdentifiedContext, error)
}

// Interface implemented by contexts that contain the ID of a web service request
// Contexts are free to store the ID internally as a complex type, but must provide a means of interpreting that ID
// as a string
type IdentifiedContext interface {
	Id() string
}

// Injects a component whose instance is an implementation of IdentifiedRequestContextBuilder into the HTTP Server
type contextBuilderDecorator struct {
	Server HttpServer
}

func (cd *contextBuilderDecorator) OfInterest(subject *ioc.Component) bool {
	result := false

	switch subject.Instance.(type) {
	case IdentifiedRequestContextBuilder:
		result = true
	}

	return result
}

func (cd *contextBuilderDecorator) DecorateComponent(subject *ioc.Component, cc *ioc.ComponentContainer) {

	idBuilder := subject.Instance.(IdentifiedRequestContextBuilder)
	cd.Server.IdContextBuilder = idBuilder

}
