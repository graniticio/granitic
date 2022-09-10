package httpserver

import (
	"context"
	"github.com/graniticio/granitic/v3/ioc"
	"net/http"
	"testing"
)

func TestIdentifyContextDecoration(t *testing.T) {

	cdb := new(contextBuilderDecorator)
	s := new(HTTPServer)

	cdb.Server = s

	c := ioc.NewComponent("mockIrb", new(mockIrb))

	if !cdb.OfInterest(c) {
		t.FailNow()
	}

	cdb.DecorateComponent(c, nil)

	if s.IDContextBuilder == nil {
		t.FailNow()
	}

}

type mockIrb struct{}

func (mi *mockIrb) WithIdentity(ctx context.Context, req *http.Request) (context.Context, error) {
	return ctx, nil
}

func (mi *mockIrb) ID(ctx context.Context) string {
	return "ID"
}
