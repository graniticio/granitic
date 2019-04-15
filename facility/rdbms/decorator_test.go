package rdbms

import (
	"github.com/graniticio/granitic/v2/ioc"
	"github.com/graniticio/granitic/v2/logging"
	"github.com/graniticio/granitic/v2/rdbms"
	"testing"
)

func TestDecorator(t *testing.T) {

	cm := new(rdbms.GraniticRdbmsClientManager)

	tar := new(mockTarget)

	c := ioc.NewComponent("", tar)

	d := new(clientManagerDecorator)
	d.fieldNameManager = map[string]rdbms.ClientManager{"sClient": cm}
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
