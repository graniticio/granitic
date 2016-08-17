package reflecttools

import (
	"testing"
	"github.com/graniticio/granitic/test"
)

func TestInvalidDependencySet(t *testing.T) {

	r := new(Receiver)

	err := SetPtrToStruct(r, "", 123)
	test.ExpectNotNil(t, err)

	err = SetPtrToStruct(r, "", "")
	test.ExpectNotNil(t, err)

	err = SetPtrToStruct(r, "", nil)
	test.ExpectNotNil(t, err)

	var a int

	err = SetPtrToStruct(r, "", &a)
	test.ExpectNotNil(t, err)

	cp := new(Concrete)
	err = SetPtrToStruct(r, "AA", cp)
	test.ExpectNotNil(t, err)

	err = SetPtrToStruct(r, "unexported", cp)
	test.ExpectNotNil(t, err)

}

func TestValidDependencySet(t *testing.T) {

	r := new(Receiver)
	cp := new(Concrete)

	err := SetPtrToStruct(r, "ConcretePtr", cp)
	test.ExpectNil(t, err)

	test.ExpectNotNil(t, r.ConcretePtr)

}

func TestValidInterfaceDependencySet(t *testing.T) {


	r := new(Receiver)
	it := new(InterfaceImpl)

	err := SetPtrToStruct(r, "Interface", it)
	test.ExpectNil(t, err)

	test.ExpectNotNil(t, r.Interface)


}

func TestInvalidInterfaceDependencySet(t *testing.T) {


	r := new(Receiver)
	it := new(Concrete)

	err := SetPtrToStruct(r, "Interface", it)
	test.ExpectNotNil(t, err)


}


type Receiver struct {
	ConcreteVal Concrete
	ConcretePtr *Concrete
	unexported *Concrete
	Interface InterfaceTest
}

type InterfaceTest interface {
	Name() string
}

type InterfaceImpl struct {

}

func (ii *InterfaceImpl) Name() string {
	return "name"
}

type Concrete struct {

}