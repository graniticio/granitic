// Copyright 2016-2022 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package validate

import (
	"github.com/graniticio/granitic/v3/test"
	"github.com/graniticio/granitic/v3/types"
	"testing"
)

func TestFloatTypeSupportDetection(t *testing.T) {

	fv := newFloatValidationRuleBuilder("DEF", nil)

	sub := new(FloatsTarget)

	sub.F32 = 32.00
	sub.F64 = 64.1
	sub.NF = types.NewNilableFloat64(128e10)
	sub.S = "NAN"

	vc := new(ValidationContext)
	vc.Subject = sub

	checkFloatTypeSupport(t, "F64", vc, fv)
	checkFloatTypeSupport(t, "NF", vc, fv)
	checkFloatTypeSupport(t, "F32", vc, fv)

	bv, err := fv.parseRule("S", []string{"REQ:MISSING"})
	test.ExpectNil(t, err)

	_, err = bv.Validate(vc)

	test.ExpectNotNil(t, err)
}

func checkFloatTypeSupport(t *testing.T, it string, vc *ValidationContext, fvb *floatValidationRuleBuilder) {
	bv, err := fvb.parseRule(it, []string{"REQ:MISSING"})
	test.ExpectNil(t, err)

	r, err := bv.Validate(vc)
	test.ExpectNil(t, err)
	c := r.ErrorCodes

	test.ExpectInt(t, len(c[it]), 0)
}

func TestFloatInSet(t *testing.T) {

	field := "F64"

	iv := newFloatValidationRuleBuilder("DEF", nil)

	sub := new(FloatsTarget)

	sub.F64 = 3.0

	vc := new(ValidationContext)
	vc.Subject = sub

	bv, err := iv.parseRule(field, []string{"REQ:MISSING", "IN:1,2,3,4,X"})
	test.ExpectNotNil(t, err)

	bv, err = iv.parseRule(field, []string{"REQ:MISSING", "IN:1,2E10,3,4:NOT_IN"})
	test.ExpectNil(t, err)

	r, err := bv.Validate(vc)
	test.ExpectNil(t, err)
	c := r.ErrorCodes

	test.ExpectInt(t, len(c[field]), 0)

	sub.F64 = 2.1e10

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c[field]), 1)
	test.ExpectString(t, c[field][0], "NOT_IN")

}

func TestFloatBreakOnError(t *testing.T) {

	field := "F64"

	iv := newFloatValidationRuleBuilder("DEF", new(CompFinder))

	sub := new(FloatsTarget)

	sub.F64 = 3

	vc := new(ValidationContext)
	vc.Subject = sub

	bv, err := iv.parseRule(field, []string{"REQ:MISSING", "BREAK"})
	test.ExpectNil(t, err)

	r, err := bv.Validate(vc)
	test.ExpectNil(t, err)
	c := r.ErrorCodes

	test.ExpectInt(t, len(c[field]), 0)

	bv, err = iv.parseRule(field, []string{"REQ:MISSING", "IN:1,2:NOTIN", "BREAK", "EXT:extFloat64Checker:EXTFAIL"})
	test.ExpectNil(t, err)

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c[field]), 1)

	test.ExpectString(t, c[field][0], "NOTIN")
}

func TestFloatRange(t *testing.T) {

	field := "F32"

	iv := newFloatValidationRuleBuilder("DEF", nil)

	sub := new(FloatsTarget)

	sub.F32 = 3.1

	vc := new(ValidationContext)
	vc.Subject = sub

	bv, err := iv.parseRule(field, []string{"REQ:MISSING", "RANGE:1|5"})
	test.ExpectNil(t, err)

	r, err := bv.Validate(vc)
	test.ExpectNil(t, err)
	c := r.ErrorCodes

	test.ExpectInt(t, len(c[field]), 0)

	sub.F32 = 1.0

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c[field]), 0)

	sub.F32 = 5.0

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c[field]), 0)

	sub.F32 = -1.22

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c[field]), 1)

	sub.F32 = 6

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c[field]), 1)

	bv, err = iv.parseRule(field, []string{"REQ:MISSING", "RANGE:|5"})
	sub.F32 = -20

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c[field]), 0)

	sub.F32 = 5

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c[field]), 0)

	sub.F32 = 6

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c[field]), 1)

	bv, err = iv.parseRule(field, []string{"REQ:MISSING", "RANGE:5|"})
	sub.F32 = -20

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c[field]), 1)

	sub.F32 = 5

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c[field]), 0)

	sub.F32 = 6

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c[field]), 0)

	bv, err = iv.parseRule(field, []string{"REQ:MISSING", "RANGE:@5|1"})
	test.ExpectNotNil(t, err)

	bv, err = iv.parseRule(field, []string{"REQ:MISSING", "RANGE:5|k1"})
	test.ExpectNotNil(t, err)

	bv, err = iv.parseRule(field, []string{"REQ:MISSING", "RANGE:5|1"})
	test.ExpectNotNil(t, err)

}

func TestFloatRequiredAndSetDetection(t *testing.T) {

	iv := newFloatValidationRuleBuilder("DEF", nil)

	sub := new(FloatsTarget)

	sub.F32 = 1
	sub.F64 = 0

	vc := new(ValidationContext)
	vc.Subject = sub

	bv, err := iv.parseRule("F32", []string{"REQ:MISSING"})
	test.ExpectNil(t, err)

	r, err := bv.Validate(vc)
	test.ExpectNil(t, err)
	c := r.ErrorCodes

	test.ExpectInt(t, len(c["F32"]), 0)
	test.ExpectBool(t, false, r.Unset)

	bv, err = iv.parseRule("F64", []string{"REQ:MISSING"})
	test.ExpectNil(t, err)

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c["F64"]), 0)
	test.ExpectBool(t, false, r.Unset)

	bv, err = iv.parseRule("NF", []string{"REQ:MISSING"})
	test.ExpectNil(t, err)

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c["NF"]), 1)
	test.ExpectBool(t, true, r.Unset)
	test.ExpectString(t, c["NF"][0], "MISSING")

	sub.NF = new(types.NilableFloat64)

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c["NF"]), 1)
	test.ExpectBool(t, true, r.Unset)
	test.ExpectString(t, c["NF"][0], "MISSING")

	sub.NF = types.NewNilableFloat64(0)

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c["NF"]), 0)
	test.ExpectBool(t, false, r.Unset)
}

func TestFloatExternal(t *testing.T) {
	fvb := newFloatValidationRuleBuilder("DEF", new(CompFinder))

	field := "F32"

	_, err := fvb.parseRule(field, []string{"EXT:extComp"})

	test.ExpectNotNil(t, err)

	_, err = fvb.parseRule(field, []string{"EXT:unknown"})

	test.ExpectNotNil(t, err)

	iv, err := fvb.parseRule(field, []string{"EXT:extFloat64Checker:EXTFAIL"})

	test.ExpectNil(t, err)

	sub := new(FloatsTarget)
	sub.F32 = 12

	vc := new(ValidationContext)
	vc.Subject = sub

	r, err := iv.Validate(vc)
	c := r.ErrorCodes

	test.ExpectNil(t, err)
	test.ExpectInt(t, len(c[field]), 1)
	test.ExpectString(t, c[field][0], "EXTFAIL")

	sub.F32 = 64.21019
	iv, err = fvb.parseRule(field, []string{"EXT:extFloat64Checker:EXTFAIL"})
	r, err = iv.Validate(vc)
	c = r.ErrorCodes

	test.ExpectNil(t, err)
	test.ExpectInt(t, len(c[field]), 0)
}

func TestFloatMExFieldDetection(t *testing.T) {
	vb := newFloatValidationRuleBuilder("DEF", nil)

	field := "F32"

	bv, err := vb.parseRule("F32", []string{"MEX:setField1,setField2:BAD_MEX"})

	test.ExpectNil(t, err)

	sub := new(FloatsTarget)
	vc := new(ValidationContext)
	vc.Subject = sub
	vc.KnownSetFields = types.NewOrderedStringSet([]string{})

	sub.F32 = 32

	r, err := bv.Validate(vc)
	test.ExpectNil(t, err)
	c := r.ErrorCodes

	test.ExpectInt(t, len(c[field]), 0)

	vc.KnownSetFields.Add("ignoreField")

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c[field]), 0)

	vc.KnownSetFields.Add("setField1")

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c[field]), 1)
	test.ExpectString(t, c[field][0], "BAD_MEX")

	vc.KnownSetFields = types.NewOrderedStringSet([]string{})
	vc.KnownSetFields.Add("setField2")

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c[field]), 1)
	test.ExpectString(t, c[field][0], "BAD_MEX")

}

type FloatsTarget struct {
	F32 float32
	F64 float64
	NF  *types.NilableFloat64
	S   string
}
