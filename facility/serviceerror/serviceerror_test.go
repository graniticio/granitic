package serviceerror

import (
	"github.com/graniticio/granitic/v2/grncerror"
	"github.com/graniticio/granitic/v2/ioc"
	"github.com/graniticio/granitic/v2/ws"
	"testing"
)

func TestErrorSourceDecorator(t *testing.T) {

	cd := new(consumerDecorator)
	cd.ErrorSource = new(grncerror.ServiceErrorManager)

	m := new(mockTar)
	c := ioc.NewComponent("", m)

	if !cd.OfInterest(c) {
		t.FailNow()
	}

	cd.DecorateComponent(c, nil)

	if m.f == nil {
		t.FailNow()
	}

}

type mockTar struct {
	f ws.ServiceErrorFinder
}

func (mt *mockTar) ProvideErrorFinder(finder ws.ServiceErrorFinder) {
	mt.f = finder
}
