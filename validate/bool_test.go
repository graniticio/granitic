// Copyright 2016 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package validate

import (
	"github.com/graniticio/granitic/test"
	"github.com/graniticio/granitic/types"
	"testing"
)

func TestUnsetBoolDetection(t *testing.T) {

	vb := newBoolValidationRuleBuilder("DEF", nil)

	bv, err := vb.parseRule("B", []string{"REQ:MISSING"})

	test.ExpectNil(t, err)

	sub := new(BoolTest)
	vc := new(ValidationContext)
	vc.Subject = sub

	sub.B = true

	r, err := bv.Validate(vc)
	test.ExpectNil(t, err)
	c := r.ErrorCodes

	test.ExpectInt(t, len(c["B"]), 0)

	bv, err = vb.parseRule("NSB", []string{"REQ:MISSING"})

	test.ExpectNil(t, err)

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c["NSB"]), 1)

	sub.NSB = new(types.NilableBool)

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c["NSB"]), 1)

	sub.NSB = nil

	bv, err = vb.parseRule("NSB", []string{})

	test.ExpectNil(t, err)

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c["B"]), 0)

	sub.NSB = new(types.NilableBool)

	bv, err = vb.parseRule("NSB", []string{})

	test.ExpectNil(t, err)

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c["B"]), 0)

}

func TestSetBoolDetection(t *testing.T) {

	vb := newBoolValidationRuleBuilder("DEF", nil)

	bv, err := vb.parseRule("B", []string{"REQ:MISSING"})

	test.ExpectNil(t, err)

	sub := new(BoolTest)
	vc := new(ValidationContext)
	vc.Subject = sub

	sub.B = true

	r, err := bv.Validate(vc)
	test.ExpectNil(t, err)
	c := r.ErrorCodes

	test.ExpectInt(t, len(c["B"]), 0)

	bv, err = vb.parseRule("NSB", []string{"REQ:MISSING"})

	test.ExpectNil(t, err)

	sub.NSB = types.NewNilableBool(true)

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c["NSB"]), 0)
}

func TestBoolMExFieldDetection(t *testing.T) {
	vb := newBoolValidationRuleBuilder("DEF", nil)

	field := "B"

	bv, err := vb.parseRule(field, []string{"MEX:setField1,setField2:BAD_MEX"})

	test.ExpectNil(t, err)

	sub := new(BoolTest)
	vc := new(ValidationContext)
	vc.Subject = sub
	vc.KnownSetFields = types.NewOrderedStringSet([]string{})

	sub.B = true

	r, err := bv.Validate(vc)
	test.ExpectNil(t, err)
	c := r.ErrorCodes

	test.ExpectInt(t, len(c["B"]), 0)

	vc.KnownSetFields.Add("ignoreField")

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c["B"]), 0)

	vc.KnownSetFields.Add("setField1")

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c["B"]), 1)
	test.ExpectString(t, c[field][0], "BAD_MEX")

	vc.KnownSetFields = types.NewOrderedStringSet([]string{})
	vc.KnownSetFields.Add("setField2")

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c["B"]), 1)
	test.ExpectString(t, c[field][0], "BAD_MEX")

}

func TestRequiredValueDetection(t *testing.T) {

	vb := newBoolValidationRuleBuilder("DEF", nil)

	field := "B"

	bv, err := vb.parseRule(field, []string{"REQ:MISSING", "IS:false:WRONG"})

	test.ExpectNil(t, err)

	sub := new(BoolTest)
	vc := new(ValidationContext)
	vc.Subject = sub

	sub.B = true

	r, err := bv.Validate(vc)
	test.ExpectNil(t, err)
	c := r.ErrorCodes

	test.ExpectInt(t, len(c["B"]), 1)
	test.ExpectString(t, c[field][0], "WRONG")

	sub.B = false

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c["B"]), 0)

	bv, err = vb.parseRule("B", []string{"REQ:MISSING", "IS:zzzz:WRONG"})

	test.ExpectNotNil(t, err)

}

type BoolTest struct {
	B   bool
	NSB *types.NilableBool
}
