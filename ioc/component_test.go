package ioc

import (
	"github.com/graniticio/granitic/v2/test"
	"testing"
)

type dummyComp struct {
}

func TestProtoCreation(t *testing.T) {

	pc := CreateProtoComponent(new(dummyComp), "TEST")

	test.ExpectString(t, "TEST", pc.Component.Name)

	if _, found := pc.Component.Instance.(*dummyComp); !found {
		t.Errorf("Instance could not be converted to *dummyComp")
		t.FailNow()
	}

}

func TestAddDependencyAndPromises(t *testing.T) {

	pc := CreateProtoComponent(new(dummyComp), "TEST")

	pc.AddDependency("F1", "OTHER")
	pc.AddConfigPromise("F2", "a.b.c")

	test.ExpectString(t, pc.Dependencies["F1"], "OTHER")
	test.ExpectString(t, pc.ConfigPromises["F2"], "a.b.c")

}
