package validate

import (
	"github.com/graniticio/granitic/test"
	"github.com/graniticio/granitic/types"
	"testing"
)

func TestFloatTypeSupportDetection(t *testing.T) {

	fv := NewFloatValidatorBuilder("DEF", nil)

	sub := new(FloatsTarget)

	sub.F32 = 32.00
	sub.F64 = 64.1
	sub.NF = types.NewNilableFloat64(128E10)
	sub.S = "NAN"

	vc := new(validationContext)
	vc.Subject = sub

	checkFloatTypeSupport(t, "F64", vc, fv)
	checkFloatTypeSupport(t, "NF", vc, fv)
	checkFloatTypeSupport(t, "F32", vc, fv)

	bv, err := fv.parseRule("S", []string{"REQ:MISSING"})
	test.ExpectNil(t, err)

	_, err = bv.Validate(vc)

	test.ExpectNotNil(t, err)
}

func checkFloatTypeSupport(t *testing.T, it string, vc *validationContext, fvb *floatValidatorBuilder) {
	bv, err := fvb.parseRule(it, []string{"REQ:MISSING"})
	test.ExpectNil(t, err)

	r, err := bv.Validate(vc)
	test.ExpectNil(t, err)
	c := r.ErrorCodes

	test.ExpectInt(t, len(c), 0)
}

type FloatsTarget struct {
	F32 float32
	F64 float64
	NF  *types.NillableFloat64
	S   string
}
