// Copyright 2018-2020 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.
package instrument

import (
	"context"
	"github.com/graniticio/granitic/v2/test"
	"testing"
)

func TestNoInstrumentorInContext(t *testing.T) {

	ctx := context.Background()

	i := InstrumentorFromContext(ctx)

	test.ExpectNil(t, i)

	defer Event(ctx, "?")()
	defer Method(ctx)()

}

func TestInstrumentorInContext(t *testing.T) {

	ctx := context.Background()

	i := new(testInstrumentor)
	ctx = AddInstrumentorToContext(ctx, i)

	x := InstrumentorFromContext(ctx)

	test.ExpectNotNil(t, x)

	defer Event(ctx, "?")()
	defer Method(ctx)()

}

func TestAmendCalled(t *testing.T) {

	ctx := context.Background()

	i := new(testInstrumentor)
	ctx = AddInstrumentorToContext(ctx, i)

	var d interface{}

	d = 5

	Amend(ctx, d)

	if !i.amendCalled || i.amendFlag != Custom || i.amendData != d {
		t.Fail()
	}

}

type testInstrumentor struct {
	StartCalled     bool
	endCalled       bool
	forkCalled      bool
	integrateCalled bool
	amendCalled     bool
	amendData       interface{}
	amendFlag       Additional
}

func (ni *testInstrumentor) StartEvent(id string, metadata ...interface{}) EndEvent {
	return ni.endEvent
}

func (ni *testInstrumentor) endEvent() {

	return

}

func (ni *testInstrumentor) Fork(ctx context.Context) (context.Context, Instrumentor) {
	return ctx, ni
}

func (ni *testInstrumentor) Integrate(instrumentor Instrumentor) {
	return
}

func (ni *testInstrumentor) Amend(additional Additional, value interface{}) {

	ni.amendCalled = true
	ni.amendFlag = additional
	ni.amendData = value

	return
}
