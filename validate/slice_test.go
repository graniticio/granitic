package validate

import (
	"github.com/graniticio/granitic/test"
	"github.com/graniticio/granitic/types"
	"testing"
)

func TestSliceSet(t *testing.T) {

	vb := NewSliceValidatorBuilder("DEF", nil)

	sv, err := vb.parseRule("S", []string{"REQ:MISSING"})

	test.ExpectNil(t, err)

	st := new(SliceTest)

	set, err := sv.IsSet("S", st)
	test.ExpectNil(t, err)

	test.ExpectBool(t, set, false)

	st.S = []string{}
	set, err = sv.IsSet("S", st)
	test.ExpectNil(t, err)

	test.ExpectBool(t, set, true)

}

func TestSliceLength(t *testing.T) {
	sb := NewSliceValidatorBuilder("DEF", nil)

	field := "S"

	sv, err := sb.parseRule(field, []string{"REQ:MISSING", "LEN:2-:LENGTH"})

	test.ExpectNil(t, err)

	sub := new(SliceTest)
	sub.S = []string{"A"}

	vc := new(ValidationContext)
	vc.Subject = sub

	r, err := sv.Validate(vc)
	c := r.ErrorCodes

	test.ExpectNil(t, err)
	test.ExpectInt(t, len(c[field]), 1)
	test.ExpectString(t, c["S"][0], "LENGTH")

	sub.S = []string{"A", "B"}

	r, err = sv.Validate(vc)
	c = r.ErrorCodes

	test.ExpectNil(t, err)
	test.ExpectInt(t, len(c[field]), 0)

	sv, err = sb.parseRule(field, []string{"REQ:MISSING", "LEN:2-3:LENGTH"})

	sub.S = []string{"A", "B", "C"}

	r, err = sv.Validate(vc)
	c = r.ErrorCodes

	test.ExpectNil(t, err)
	test.ExpectInt(t, len(c[field]), 0)

	sub.S = []string{"A", "B", "C", "D"}

	r, err = sv.Validate(vc)
	c = r.ErrorCodes

	test.ExpectNil(t, err)
	test.ExpectInt(t, len(c[field]), 1)
	test.ExpectString(t, c[field][0], "LENGTH")

	sv, err = sb.parseRule(field, []string{"REQ:MISSING", "LEN:-3:LENGTH"})

	sub.S = []string{}

	r, err = sv.Validate(vc)
	c = r.ErrorCodes

	test.ExpectNil(t, err)
	test.ExpectInt(t, len(c[field]), 0)

	sub.S = []string{"A", "B", "C"}

	r, err = sv.Validate(vc)
	c = r.ErrorCodes

	test.ExpectNil(t, err)
	test.ExpectInt(t, len(c[field]), 0)

	sub.S = []string{"A", "B", "C", "D"}

	r, err = sv.Validate(vc)
	c = r.ErrorCodes

	test.ExpectNil(t, err)
	test.ExpectInt(t, len(c[field]), 1)
	test.ExpectString(t, c[field][0], "LENGTH")

}

func TestSliceMExFieldDetection(t *testing.T) {
	vb := NewSliceValidatorBuilder("DEF", nil)

	field := "S"

	bv, err := vb.parseRule(field, []string{"MEX:setField1,setField2:BAD_MEX"})

	test.ExpectNil(t, err)

	sub := new(SliceTest)
	vc := new(ValidationContext)
	vc.Subject = sub
	vc.KnownSetFields = types.NewOrderedStringSet([]string{})

	sub.S = []string{"A"}

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

type SliceTest struct {
	S []string
}
