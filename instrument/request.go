// Copyright 2018-2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

/*
	Package instrument provides patterns and helper functions for instrumenting some flow of control (typically a web service request or scheduled activity).

	This instrumentation is often the capture of timing data for the purposes of monitoring performance, but the type of instrumentation depends on the implementation of the
	interfaces in this package.

	Instrumentor

	The key concept in this package is that of the Instrumentor, which is a common interface to instrumentation that is shared by your code and the
	Granitic framework. It is expected that an instance of Instrumentator will be stored in the context.Context that is passed through your code. This
	allows your code to use the helper functions in this package to either gain access to the Instrumentor or use one-line method calls to initiate instrumentation of a method

	Web service instrumentation

	Grantic's HttpServer has support for instrumenting your web service requests. See the facility/httpserver package documentation for more details.
*/
package instrument

import (
	"context"
	"net/http"
)

// RequestInstrumentationManager is implemented by components that can instrument a web service request. This often involves recording
// timing data at various points in a request's lifecycle. Implementations can be attached to an instance of the Granitic HttpServer.
type RequestInstrumentationManager interface {

	// Begin starts instrumentation and returns a Instrumentor that is able to instrument sub/child events of the request.
	// It is expected that most implementation will also store the Instrumentor in the context so it can be easily recovered
	// at any point in the request using the function InstrumentorFromContext.
	Begin(ctx context.Context, res http.ResponseWriter, req *http.Request) (context.Context, Instrumentor, func())
}
