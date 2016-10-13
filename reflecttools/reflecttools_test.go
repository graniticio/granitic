// Copyright 2016 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package reflecttools

import (
	"github.com/graniticio/granitic/test"
	"reflect"
	"testing"
)

func TestNestedFieldFind(t *testing.T) {

	path := "C1.GC1.S"
	subject := nested()

	v, err := FindNestedField(ExtractDotPath(path), subject)

	test.ExpectNil(t, err)

	test.ExpectBool(t, v.Kind() == reflect.String, true)

	path = "C1.S.S"

	v, err = FindNestedField(ExtractDotPath(path), subject)
	test.ExpectNotNil(t, err)

	path = "C3.S"

	v, err = FindNestedField(ExtractDotPath(path), subject)

	test.ExpectNil(t, err)
	test.ExpectBool(t, v.Kind() == reflect.String, true)
	test.ExpectString(t, v.Interface().(string), "Not ptr")

}

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
	unexported  *Concrete
	Interface   InterfaceTest
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

func nested() *NestedParent {
	np := new(NestedParent)
	c1 := new(NestedChild)
	c2 := new(NestedChild)

	np.C1 = c1
	np.C2 = c2
	np.C3 = NestedChild{S: "Not ptr"}

	gc1 := new(NestedGrandChild)
	gc2 := new(NestedGrandChild)
	gc3 := new(NestedGrandChild)
	gc4 := new(NestedGrandChild)

	c1.GC1 = gc1
	c1.GC2 = gc2

	c2.GC1 = gc3
	c2.GC2 = gc4

	gc1.B = true
	gc1.I = 1
	gc1.S = "GC1"

	gc2.B = false
	gc2.I = 2
	gc2.S = "GC2"

	gc3.B = true
	gc3.I = 3
	gc3.S = "GC3"

	gc4.B = false
	gc4.I = 4
	gc4.S = "GC4"

	return np
}

type NestedParent struct {
	C1 *NestedChild
	C2 *NestedChild
	C3 NestedChild
}

type NestedChild struct {
	S   string
	GC1 *NestedGrandChild
	GC2 *NestedGrandChild
}

type NestedGrandChild struct {
	S string
	B bool
	I int64
}
