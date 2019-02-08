// Copyright 2018-2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package httpserver

import (
	"context"
	"github.com/graniticio/granitic/v2/instrument"
	"net/http"
)

// A default implementation of RequestInstrumentationManager that does nothing
type noopRequestInstrumentationManager struct{}

func (nm *noopRequestInstrumentationManager) Begin(ctx context.Context, res http.ResponseWriter, req *http.Request) (context.Context, instrument.Instrumentor, func()) {
	ri := new(noopRequestInstrumentor)
	nc := instrument.AddInstrumentorToContext(ctx, ri)

	return nc, ri, func() {}
}

// A default implementation of instrument.Instrumentor that does nothing
type noopRequestInstrumentor struct {
}

func (ni *noopRequestInstrumentor) StartEvent(id string, metadata ...interface{}) instrument.EndEvent {
	return func() {}
}

func (ni *noopRequestInstrumentor) Fork(ctx context.Context) (context.Context, instrument.Instrumentor) {
	return ctx, ni
}

func (ni *noopRequestInstrumentor) Integrate(instrumentor instrument.Instrumentor) {
	return
}

func (ni *noopRequestInstrumentor) Amend(additional instrument.Additional, value interface{}) {
	return
}
