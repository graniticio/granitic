// Copyright 2016-2022 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package json

import (
	"github.com/graniticio/granitic/v3/test"
	"testing"
)

func TestCamelCasing(t *testing.T) {

	p := new(Parent)

	c1 := new(Child)
	c2 := new(Child)

	p.ChildOne = c1
	p.ChildTwo = c2

	g1 := new(GrandChild)
	g1.Name = "GC1"

	g2 := new(GrandChild)
	g2.Name = "GC2"

	g3 := new(GrandChild)
	g3.Name = "GC3"

	g4 := new(GrandChild)
	g4.Name = "GC4"

	c1.GrandChildOne = g1
	c1.GrandChildTwo = g2

	c2.GrandChildOne = g3
	c2.GrandChildTwo = g4

	p.FamilySize = 8
	p.HasChildren = true

	v, err := CamelCase(p)

	test.ExpectNil(t, err)
	test.ExpectNotNil(t, v)

	//sm, found := v.(map[string]interface{})

}

type Parent struct {
	ChildOne      *Child
	ChildTwo      *Child
	GrandChildren []*GrandChild
	FamilySize    int
	HasChildren   bool
}

type Child struct {
	GrandChildOne *GrandChild
	GrandChildTwo *GrandChild
}

type GrandChild struct {
	Name string
}
