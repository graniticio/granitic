// Copyright 2018 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.
package instrument

import (
	"context"
	"runtime"
	"strings"
)

// Additional is used as a flag to indicate what additional data being passed to the Instrumentor represents. These are
// used by Granitic to pass additional data about a request into a Instrumentor that is not known at the point instrumentation starts
type Additional uint

const (
	REQUEST_ID      Additional = iota //string representation of a unique ID for the request
	REQUEST_VERSION                   //instance of ws.RequiredVersion
	USER_IDENTITY                     //instance of iam.ClientIdentity
	HANDLER                           //The handler that is processing the request (*ws.Handler)
)

// Instrumentor is implemented by types that can add additional information to a request that is being instrumented in
// the form of sub/child events that are instrumented separately and additional framework data that was not available when instrumentation
// began.
//
// Interfaces are not expected to be explicitly goroutine safe - the Fork and Integrate methods are intended for use when the
// request under instrumentation spawns new goroutines
type Instrumentor interface {
	// StartEvent indicates that a new instrumentable activity has begun with the supplied ID. Implementation specific additional
	// information about the event can be supplied via the metadata varg
	//
	// The function returned by this method should be called when the event ends. This facilitates a pattern like defer StartEvent(id)()
	StartEvent(id string, metadata ...interface{}) EndEvent

	// Fork creates a new context and Instrumentor suitable for passing to a child goroutine
	Fork(ctx context.Context) (context.Context, Instrumentor)

	//Integrate incorporates the data from a forked Instrumentor that was passed to a goroutine
	Integrate(instrumentor Instrumentor)

	//Amend allows Granitic to provide additional information about the request that was not available when instrumentation started
	Amend(additional Additional, value interface{})
}

type ctxKey int

const instrumentorKey ctxKey = 0

// InstrumentorFromContext returns a Instrumentor from the supplied context, or nil if no Instrumentor
// is present
func InstrumentorFromContext(ctx context.Context) Instrumentor {

	v := ctx.Value(instrumentorKey)

	if ri, found := v.(Instrumentor); found {
		return ri
	} else {
		return nil
	}
}

// AddInstrumentorToContext stores the supplied Instrumentor in a new context, derived from the supplied context.
func AddInstrumentorToContext(ctx context.Context, ri Instrumentor) context.Context {
	return context.WithValue(ctx, instrumentorKey, ri)
}

type EndEvent func()

// Event is convenience function that calls InstrumentorFromContext then StartEvent. This function
// fails silently if the result of InstrumentorFromContext is nil (e.g there is no Instrumentor in the context)
func Event(ctx context.Context, id string, metadata ...interface{}) EndEvent {
	if ri := InstrumentorFromContext(ctx); ri == nil {
		return func() {}
	} else {
		return ri.StartEvent(id, metadata...)
	}
}

// Method is a convenience function that calls Event with the name of the calling function as the ID.
// The format of the method name will be /path/to/package.(type).FunctionName
//
// This function fails silently if the result of InstrumentorFromContext is nil (e.g there is no Instrumentor in the context)
func Method(ctx context.Context, metadata ...interface{}) EndEvent {
	pc := make([]uintptr, 1)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()

	c := strings.Split(frame.Function, "/")

	return Event(ctx, c[len(c)-1], metadata...)

}
