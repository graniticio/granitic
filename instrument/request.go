// Copyright 2018 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package instrument

import (
	"context"
	"net/http"
)

// RequestInstrumentationManager is implemented by components that can instrument a web service request. This often involves recording
// timing data at various points in a request's lifecycle. Implementations can be attached to an instance of the Granitic HttpServer.
type RequestInstrumentationManager interface {

	// Begin starts instrumentation and returns a RequestInstrumentor that is able to instrument sub/child events of the request.
	// It is expected that most implementation will also store the RequestInstrumentor in the context so it can be easily recovered
	// at any point in the request using the function RequestInstrumentorFromContext.
	Begin(ctx context.Context, res http.ResponseWriter, req *http.Request) (context.Context, RequestInstrumentor)

	// End ends instrumentation of the request.
	End(ctx context.Context)
}

// Additional is used as a flag to indicate what additional data being passed to the RequestInstrumentor represents. These are
// used by Granitic to pass additional data about a request into a RequestInstrumentor that is not known at the point instrumentation starts
type Additional uint

const (
	REQUEST_ID      Additional = iota //string representation of a unique ID for the request
	REQUEST_VERSION                   //instance of ws.RequiredVersion
	USER_IDENTITY                     //instance of iam.ClientIdentity
	HANDLER                           //The handler that is processing the request (*ws.Handler)
)

// RequestInstrumentor is implemented by types that can add additional information to a request that is being instrumented in
// the form of sub/child events that are instrumented separately and additional framework data that was not available when instrumentation
// began.
//
// Interfaces are not expected to be explicitly goroutine safe - the Fork and Integrate methods are intended for use when the
// request under instrumentation spawns new goroutines
type RequestInstrumentor interface {
	// StartEvent indicates that a new instrumentable activity has begun with the supplied ID. Implementation specific additional
	// information about the event can be supplied via the metadata varg
	StartEvent(id string, metadata ...interface{})

	// EndEvent is called when an instrumentable activity is complete. Implementations are expected to return an error if StartEvent has not been called
	EndEvent() error

	// Fork creates a new context and RequestInstrumentor suitable for passing to a child goroutine
	Fork(ctx context.Context) (context.Context, RequestInstrumentor)

	//Integrate incorporates the data from a forked RequestInstrumentor that was passed to a goroutine
	Integrate(instrumentor RequestInstrumentor)

	//Amend allows Granitic to provide additional information about the request that was not available when instrumentation started
	Amend(additional Additional, value interface{})
}

type ctxKey int

const requestInstrumentorKey ctxKey = 0

// RequestInstrumentorFromContext returns a RequestInstrumentor from the supplied context, or nil if no RequestInstrumentor
// is present
func RequestInstrumentorFromContext(ctx context.Context) RequestInstrumentor {

	v := ctx.Value(requestInstrumentorKey)

	if ri, found := v.(RequestInstrumentor); found {
		return ri
	} else {
		return nil
	}
}

// AddRequestInstrumentorToContext stores the supplied RequestInstrumentor in a new context, derivied from the supplied context.
func AddRequestInstrumentorToContext(ctx context.Context, ri RequestInstrumentor) context.Context {
	return context.WithValue(ctx, requestInstrumentorKey, ri)
}
