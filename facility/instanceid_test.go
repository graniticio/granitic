package facility

import (
	"github.com/graniticio/granitic/v3/instance"
	"github.com/graniticio/granitic/v3/ioc"
	"testing"
)

func TestInstanceIDDecoration(t *testing.T) {

	i := instance.NewIdentifier("my-id")

	idd := new(InstanceIDDecorator)
	idd.InstanceID = i
	ir := new(idReceiver)

	c := new(ioc.Component)
	c.Instance = ir

	if !idd.OfInterest(c) {
		t.FailNow()
	}

	idd.DecorateComponent(c, nil)

	if ir.id != "my-id" {
		t.Fail()
	}

}

type idReceiver struct {
	id string
}

func (ir *idReceiver) RegisterInstanceID(i *instance.Identifier) {
	ir.id = i.ID
}
