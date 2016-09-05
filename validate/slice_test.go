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

func TestSliceMExFieldDetection(t *testing.T) {
	vb := NewSliceValidatorBuilder("DEF", nil)

	bv, err := vb.parseRule("S", []string{"MEX:setField1,setField2:BAD_MEX"})

	test.ExpectNil(t, err)

	sub := new(SliceTest)
	vc := new(validationContext)
	vc.Subject = sub
	vc.KnownSetFields = types.NewOrderedStringSet([]string{})

	sub.S = []string{"A"}

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

type SliceTest struct {
	S []string
}
