package rdbms

import (
	"github.com/graniticio/granitic/v3/ioc"
	"github.com/graniticio/granitic/v3/logging"
	"github.com/graniticio/granitic/v3/rdbms"
	"testing"
)

func TestDecorator(t *testing.T) {

	cm := new(rdbms.GraniticRdbmsClientManager)

	tar := new(mockTarget)

	c := ioc.NewComponent("", tar)

	d := new(clientManagerDecorator)
	d.fieldNameManager = map[string]rdbms.ClientManager{"Client": cm}
	d.log = new(logging.ConsoleErrorLogger)

	if !d.OfInterest(c) {
		t.FailNow()
	}

	d.DecorateComponent(c, nil)

	if tar.Client == nil {
		t.FailNow()
	}

}

type mockTarget struct {
	Client rdbms.ClientManager
}
