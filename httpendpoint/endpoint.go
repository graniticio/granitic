// Copyright 2016 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

/*
Package httpendpoint provides types that allow web-service handlers to be registered with an HTTP server.

Types in this package represent the interface between the HttpServer facility (which is a thin layer over Go's http.Server) and
the Granitic ws and handler packages that define web-services.

In most cases, user applications will not need to interact with the types in this package. Instead they will define
components of type handler.WsHandler (which already implements the key HttpEndpointProvider interface below) and the
framework will automatically register them with the HttpServer facility.

*/
package httpendpoint

import (
	"golang.org/x/net/context"
	"net/http"
)

// HttpEndPoint associates HTTP methods (GET, POST etc) and path-matching regular expressions with a handler.
type HttpEndPoint struct {
	// A map of HTTP method names to a regular expression pattern (eg GET=^/health-check$)
	MethodPatterns map[string]string

	// A handler implementing Go's built-in http.Handler interface
	Handler http.Handler
}

// Implemented by a component that is able to support a web-service request with a particular path.
type HttpEndpointProvider interface {
	//SupportedHttpMethods returns the HTTP methods (GET, POST, PUT etc) that the endpoint supports.
	SupportedHttpMethods() []string

	// RegexPattern returns an un-compiled regular expression that will be applied to the path element (e.g excluding scheme, domain and query parameters)
	// to potentially match the request to this endpoint.
	RegexPattern() string

	// ServeHTTP handles an HTTP request, including writing normal and abnormal responses. Returns a context that may have
	// been modified.
	ServeHTTP(ctx context.Context, w *HTTPResponseWriter, req *http.Request) context.Context

	// VersionAware returns true if this endpoint supports request version matching.
	VersionAware() bool

	// SupportsVersion returns true if this endpoint supports the version functionality required by the requester. Behaviour is undefined if
	// VersionAware is false. Version matching is application-specific and not defined by Granitic.
	SupportsVersion(version RequiredVersion) bool

	// AutoWireable returns false if this endpoint should not automatically be registered with HTTP servers.
	AutoWireable() bool
}

type RequiredVersion map[string]interface{}

type RequestedVersionExtractor interface {
	Extract(*http.Request) RequiredVersion
}
