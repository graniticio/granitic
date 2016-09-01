package validate

import (
	"github.com/graniticio/granitic/test"
	"github.com/graniticio/granitic/types"
	"testing"
)

func TestUnsetBoolDetection(t *testing.T) {

	vb := NewBoolValidatorBuilder("DEF", nil)

	bv, err := vb.parseRule("B", []string{"REQ:MISSING"})

	test.ExpectNil(t, err)

	sub := new(BoolTest)
	vc := new(validationContext)
	vc.Subject = sub

	sub.B = true

	r, err := bv.Validate(vc)
	test.ExpectNil(t, err)
	c := r.ErrorCodes

	test.ExpectInt(t, len(c), 0)

	bv, err = vb.parseRule("NSB", []string{"REQ:MISSING"})

	test.ExpectNil(t, err)

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c), 1)

	sub.NSB = new(types.NilableBool)

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c), 1)

	sub.NSB = nil

	bv, err = vb.parseRule("NSB", []string{})

	test.ExpectNil(t, err)

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c), 0)

	sub.NSB = new(types.NilableBool)

	bv, err = vb.parseRule("NSB", []string{})

	test.ExpectNil(t, err)

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c), 0)

}

func TestSetBoolDetection(t *testing.T) {

	vb := NewBoolValidatorBuilder("DEF", nil)

	bv, err := vb.parseRule("B", []string{"REQ:MISSING"})

	test.ExpectNil(t, err)

	sub := new(BoolTest)
	vc := new(validationContext)
	vc.Subject = sub

	sub.B = true

	r, err := bv.Validate(vc)
	test.ExpectNil(t, err)
	c := r.ErrorCodes

	test.ExpectInt(t, len(c), 0)

	bv, err = vb.parseRule("NSB", []string{"REQ:MISSING"})

	test.ExpectNil(t, err)

	sub.NSB = types.NewNilableBool(true)

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c), 0)
}

func TestBoolMExFieldDetection(t *testing.T) {
	vb := NewBoolValidatorBuilder("DEF", nil)

	bv, err := vb.parseRule("B", []string{"MEX:setField1,setField2:BAD_MEX"})

	test.ExpectNil(t, err)

	sub := new(BoolTest)
	vc := new(validationContext)
	vc.Subject = sub
	vc.KnownSetFields = types.NewOrderedStringSet([]string{})

	sub.B = true

	r, err := bv.Validate(vc)
	test.ExpectNil(t, err)
	c := r.ErrorCodes

	test.ExpectInt(t, len(c), 0)

	vc.KnownSetFields.Add("ignoreField")

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c), 0)

	vc.KnownSetFields.Add("setField1")

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c), 1)
	test.ExpectString(t, c[0], "BAD_MEX")

	vc.KnownSetFields = types.NewOrderedStringSet([]string{})
	vc.KnownSetFields.Add("setField2")

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c), 1)
	test.ExpectString(t, c[0], "BAD_MEX")

}

func TestRequiredValueDetection(t *testing.T) {

	vb := NewBoolValidatorBuilder("DEF", nil)

	bv, err := vb.parseRule("B", []string{"REQ:MISSING", "IS:false:WRONG"})

	test.ExpectNil(t, err)

	sub := new(BoolTest)
	vc := new(validationContext)
	vc.Subject = sub

	sub.B = true

	r, err := bv.Validate(vc)
	test.ExpectNil(t, err)
	c := r.ErrorCodes

	test.ExpectInt(t, len(c), 1)
	test.ExpectString(t, c[0], "WRONG")

	sub.B = false

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c), 0)

	bv, err = vb.parseRule("B", []string{"REQ:MISSING", "IS:zzzz:WRONG"})

	test.ExpectNotNil(t, err)

}

type BoolTest struct {
	B   bool
	NSB *types.NilableBool
}
