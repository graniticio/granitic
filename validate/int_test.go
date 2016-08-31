package validate

import (
	"github.com/graniticio/granitic/test"
	"github.com/graniticio/granitic/types"
	"testing"
)

func TestIntTypeSupportDetection(t *testing.T) {

	iv := NewIntValidatorBuilder("DEF", nil)

	sub := new(IntsTarget)

	sub.I = 1
	sub.I8 = 8
	sub.I16 = 16
	sub.I32 = 32
	sub.I64 = 64
	sub.NI = types.NewNilableInt64(128)
	sub.S = "NAN"

	vc := new(validationContext)
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

	iv := NewIntValidatorBuilder("DEF", nil)

	sub := new(IntsTarget)

	sub.I = 1
	sub.I8 = 0

	vc := new(validationContext)
	vc.Subject = sub

	bv, err := iv.parseRule("I", []string{"REQ:MISSING", "IN:1,2,3,4,X"})
	test.ExpectNotNil(t, err)

	bv, err = iv.parseRule("I", []string{"REQ:MISSING", "IN:1,2,3,4.2"})
	test.ExpectNotNil(t, err)

	bv, err = iv.parseRule("I", []string{"REQ:MISSING", "IN:1,2,3,4:NOT_IN"})
	test.ExpectNil(t, err)

	r, err := bv.Validate(vc)
	test.ExpectNil(t, err)
	c := r.ErrorCodes

	test.ExpectInt(t, len(c), 0)

	sub.I = 0

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c), 1)
	test.ExpectString(t, c[0], "NOT_IN")

}

func TestIntBreakOnError(t *testing.T) {

	iv := NewIntValidatorBuilder("DEF", nil)

	sub := new(IntsTarget)

	sub.I = 1
	sub.I8 = 0

	vc := new(validationContext)
	vc.Subject = sub

	bv, err := iv.parseRule("I", []string{"REQ:MISSING", "BREAK"})
	test.ExpectNil(t, err)

	r, err := bv.Validate(vc)
	test.ExpectNil(t, err)
	c := r.ErrorCodes

	test.ExpectInt(t, len(c), 0)
}

func TestIntRequiredAndSetDetection(t *testing.T) {

	iv := NewIntValidatorBuilder("DEF", nil)

	sub := new(IntsTarget)

	sub.I = 1
	sub.I8 = 0

	vc := new(validationContext)
	vc.Subject = sub

	bv, err := iv.parseRule("I", []string{"REQ:MISSING"})
	test.ExpectNil(t, err)

	r, err := bv.Validate(vc)
	test.ExpectNil(t, err)
	c := r.ErrorCodes

	test.ExpectInt(t, len(c), 0)
	test.ExpectBool(t, false, r.Unset)

	bv, err = iv.parseRule("I8", []string{"REQ:MISSING"})
	test.ExpectNil(t, err)

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c), 0)
	test.ExpectBool(t, false, r.Unset)

	bv, err = iv.parseRule("NI", []string{"REQ:MISSING"})
	test.ExpectNil(t, err)

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c), 1)
	test.ExpectBool(t, true, r.Unset)
	test.ExpectString(t, c[0], "MISSING")

	sub.NI = new(types.NilableInt64)

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c), 1)
	test.ExpectBool(t, true, r.Unset)
	test.ExpectString(t, c[0], "MISSING")

	sub.NI = types.NewNilableInt64(0)

	r, err = bv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c), 0)
	test.ExpectBool(t, false, r.Unset)
}

func checkIntTypeSupport(t *testing.T, it string, vc *validationContext, iv *intValidatorBuilder) {
	bv, err := iv.parseRule(it, []string{"REQ:MISSING"})
	test.ExpectNil(t, err)

	r, err := bv.Validate(vc)
	test.ExpectNil(t, err)
	c := r.ErrorCodes

	test.ExpectInt(t, len(c), 0)
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
