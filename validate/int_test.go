// Copyright 2016-2022 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package validate

import (
	"github.com/graniticio/granitic/v3/test"
	"github.com/graniticio/granitic/v3/types"
	"testing"
)

func TestIntTypeSupportDetection(t *testing.T) {

	iv := newIntValidationRuleBuilder("DEF", nil)

	sub := new(IntsTarget)

	sub.I = 1
	sub.I8 = 8
	sub.I16 = 16
	sub.I32 = 32
	sub.I64 = 64
	sub.NI = types.NewNilableInt64(128)
	sub.S = "NAN"

	vc := new(ValidationContext)
	vc.Subject = sub

	checkIntTypeSupport(t, "I64", vc, iv)
	checkIntTypeSupport(t, "NI", vc, iv)
	checkIntTypeSupport(t, "I", vc, iv)
	checkIntTypeSupport(t, "I8", vc, iv)
	checkIntTypeSupport(t, "I16", vc, iv)
	checkIntTypeSupport(t, "I32", vc, iv)

	bv, err := iv.parseRule("S", []string{"REQ:MISSING"})
	test.ExpectNil(t, err)

	_, err = bv.Validate(vc)

	test.ExpectNotNil(t, err)
}

func TestIntInSet(t *testing.T) {

	iv := newIntValidationRuleBuilder("DEF", nil)

	sub := new(IntsTarget)

	sub.I = 1
	sub.I8 = 0

	vc := new(ValidationContext)
	vc.Subject = sub

	field := "I"

	bv, err := iv.parseRule(field, []string{"REQ:MISSING", "IN:1,2,3,4,X"})
	test.ExpectNotNil(t, err)

	bv, err = iv.parseRule(field, []string{"REQ:MISSING", "IN:1,2,3,4.2"})
	test.ExpectNotNil(t, err)

	bv, err = iv.parseRule(field, []string{"REQ:MISSING", "IN:1,2,3,4:NOT_IN"})
	test.ExpectNil(t, err)

	r, err := bv.Validate(vc)
	test.ExpectNil(t, err)
	c := r.ErrorCodes

	test.ExpectInt(t, len(c[field]), 0)

	sub.I = 0

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c[field]), 1)
	test.ExpectString(t, c[field][0], "NOT_IN")

}

func TestIntBreakOnError(t *testing.T) {

	iv := newIntValidationRuleBuilder("DEF", new(CompFinder))

	sub := new(IntsTarget)

	sub.I = 3
	sub.I8 = 0

	vc := new(ValidationContext)
	vc.Subject = sub

	field := "I"

	bv, err := iv.parseRule(field, []string{"REQ:MISSING", "BREAK"})
	test.ExpectNil(t, err)

	r, err := bv.Validate(vc)
	test.ExpectNil(t, err)
	c := r.ErrorCodes

	test.ExpectInt(t, len(c[field]), 0)

	bv, err = iv.parseRule(field, []string{"REQ:MISSING", "IN:1,2:NOTIN", "BREAK", "EXT:extInt64Checker:EXTFAIL"})
	test.ExpectNil(t, err)

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c[field]), 1)

	test.ExpectString(t, c[field][0], "NOTIN")
}

func TestIntRange(t *testing.T) {

	iv := newIntValidationRuleBuilder("DEF", nil)

	sub := new(IntsTarget)

	sub.I = 3

	vc := new(ValidationContext)
	vc.Subject = sub

	field := "I"

	bv, err := iv.parseRule(field, []string{"REQ:MISSING", "RANGE:1|5"})
	test.ExpectNil(t, err)

	r, err := bv.Validate(vc)
	test.ExpectNil(t, err)
	c := r.ErrorCodes

	test.ExpectInt(t, len(c[field]), 0)

	sub.I = 1

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c[field]), 0)

	sub.I = 5

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c[field]), 0)

	sub.I = -1

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c[field]), 1)

	sub.I = 6

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c[field]), 1)

	bv, err = iv.parseRule(field, []string{"REQ:MISSING", "RANGE:|+5"})
	test.ExpectNil(t, err)
	sub.I = -20

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c[field]), 0)

	sub.I = 5

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c[field]), 0)

	sub.I = 6

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c[field]), 1)

	bv, err = iv.parseRule(field, []string{"REQ:MISSING", "RANGE:5|"})
	test.ExpectNil(t, err)
	sub.I = -20

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c[field]), 1)

	sub.I = 5

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c[field]), 0)

	sub.I = 6

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c[field]), 0)

	bv, err = iv.parseRule(field, []string{"REQ:MISSING", "RANGE:-10|-1"})
	test.ExpectNil(t, err)
	sub.I = -20

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c[field]), 1)

	sub.I = -1

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c[field]), 0)

	bv, err = iv.parseRule(field, []string{"REQ:MISSING", "RANGE:-1|-10"})
	test.ExpectNotNil(t, err)

}

func TestIntRequiredAndSetDetection(t *testing.T) {

	iv := newIntValidationRuleBuilder("DEF", nil)

	sub := new(IntsTarget)

	sub.I = 1
	sub.I8 = 0

	vc := new(ValidationContext)
	vc.Subject = sub

	bv, err := iv.parseRule("I", []string{"REQ:MISSING"})
	test.ExpectNil(t, err)

	r, err := bv.Validate(vc)
	test.ExpectNil(t, err)
	c := r.ErrorCodes

	test.ExpectInt(t, len(c["I"]), 0)
	test.ExpectBool(t, false, r.Unset)

	bv, err = iv.parseRule("I8", []string{"REQ:MISSING"})
	test.ExpectNil(t, err)

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c["I8"]), 0)
	test.ExpectBool(t, false, r.Unset)

	bv, err = iv.parseRule("NI", []string{"REQ:MISSING"})
	test.ExpectNil(t, err)

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c["NI"]), 1)
	test.ExpectBool(t, true, r.Unset)
	test.ExpectString(t, c["NI"][0], "MISSING")

	sub.NI = new(types.NilableInt64)

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c["NI"]), 1)
	test.ExpectBool(t, true, r.Unset)
	test.ExpectString(t, c["NI"][0], "MISSING")

	sub.NI = types.NewNilableInt64(0)

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c["NI"]), 0)
	test.ExpectBool(t, false, r.Unset)
}

func TestIntExternal(t *testing.T) {
	ivb := newIntValidationRuleBuilder("DEF", new(CompFinder))

	_, err := ivb.parseRule("I", []string{"EXT:extComp"})

	test.ExpectNotNil(t, err)

	_, err = ivb.parseRule("I", []string{"EXT:unknown"})

	test.ExpectNotNil(t, err)

	iv, err := ivb.parseRule("I", []string{"EXT:extInt64Checker:EXTFAIL"})

	test.ExpectNil(t, err)

	sub := new(IntsTarget)
	sub.I = 12

	vc := new(ValidationContext)
	vc.Subject = sub

	r, err := iv.Validate(vc)
	c := r.ErrorCodes

	test.ExpectNil(t, err)
	test.ExpectInt(t, len(c["I"]), 1)
	test.ExpectString(t, c["I"][0], "EXTFAIL")

	sub.I = 64
	r, err = iv.Validate(vc)
	c = r.ErrorCodes

	test.ExpectNil(t, err)
	test.ExpectInt(t, len(c["I"]), 0)
}

func TestIntMExFieldDetection(t *testing.T) {
	vb := newIntValidationRuleBuilder("DEF", nil)

	bv, err := vb.parseRule("I32", []string{"MEX:setField1,setField2:BAD_MEX"})

	test.ExpectNil(t, err)

	sub := new(IntsTarget)
	vc := new(ValidationContext)
	vc.Subject = sub
	vc.KnownSetFields = types.NewOrderedStringSet([]string{})

	sub.I32 = 32

	r, err := bv.Validate(vc)
	test.ExpectNil(t, err)
	c := r.ErrorCodes

	test.ExpectInt(t, len(c["I32"]), 0)

	vc.KnownSetFields.Add("ignoreField")

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c["I32"]), 0)

	vc.KnownSetFields.Add("setField1")

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c["I32"]), 1)
	test.ExpectString(t, c["I32"][0], "BAD_MEX")

	vc.KnownSetFields = types.NewOrderedStringSet([]string{})
	vc.KnownSetFields.Add("setField2")

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c["I32"]), 1)
	test.ExpectString(t, c["I32"][0], "BAD_MEX")

}

func checkIntTypeSupport(t *testing.T, it string, vc *ValidationContext, iv *intValidationRuleBuilder) {
	bv, err := iv.parseRule(it, []string{"REQ:MISSING"})
	test.ExpectNil(t, err)

	r, err := bv.Validate(vc)
	test.ExpectNil(t, err)
	c := r.ErrorCodes

	test.ExpectInt(t, len(c[it]), 0)
}

type IntsTarget struct {
	I   int
	I8  int8
	I16 int16
	I32 int32
	I64 int64
	NI  *types.NilableInt64
	S   string
}
