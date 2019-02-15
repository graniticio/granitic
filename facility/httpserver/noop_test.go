package httpserver

import (
	"context"
	"github.com/graniticio/granitic/v2/instrument"
	"testing"
)

func TestNoopRequestInstrumentationManager(t *testing.T) {

	var im instrument.RequestInstrumentationManager

	im = new(noopRequestInstrumentationManager)

	ctx, i, end := im.Begin(context.Background(), nil, nil)

	i.StartEvent("ID")
	i.Fork(ctx)
	i.Amend(instrument.RequestID, "MYID")
	i.Integrate(i)

	end()

}
