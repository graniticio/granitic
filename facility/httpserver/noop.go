// Copyright 2018 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package httpserver

import (
	"context"
	"github.com/graniticio/granitic/instrument"
	"net/http"
)

// A default implementation of RequestInstrumentationManager that does nothing
type noopRequestInstrumentationManager struct{}

func (nm *noopRequestInstrumentationManager) Begin(ctx context.Context, res http.ResponseWriter, req *http.Request) (context.Context, instrument.RequestInstrumentor) {
	ri := new(noopRequestInstrumentor)
	nc := instrument.AddRequestInstrumentorToContext(ctx, ri)

	return nc, ri
}

func (nm *noopRequestInstrumentationManager) End(context.Context) {}

// A default implementation of instrument.RequestInstrumentor that does nothing
type noopRequestInstrumentor struct {
}

func (ni *noopRequestInstrumentor) StartEvent(id string, metadata ...interface{}) {
	return
}

func (ni *noopRequestInstrumentor) EndEvent() error {
	return nil
}

func (ni *noopRequestInstrumentor) Fork(ctx context.Context) (context.Context, instrument.RequestInstrumentor) {
	return ctx, ni
}

func (ni *noopRequestInstrumentor) Integrate(instrumentor instrument.RequestInstrumentor) {
	return
}

func (ni *noopRequestInstrumentor) Amend(additional instrument.Additional, value interface{}) {
	return
}
