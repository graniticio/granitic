package validate

import (
	"github.com/graniticio/granitic/test"
	"testing"
)

func TestUnsetObjDetection(t *testing.T) {

	ob := NewObjectValidatorBuilder("DEF", nil)

	ov, err := ob.parseRule("CP", []string{"REQ:MISSING"})

	test.ExpectNil(t, err)

	sub := new(Parent)
	vc := new(validationContext)
	vc.Subject = sub

	r, err := ov.Validate(vc)
	test.ExpectNil(t, err)
	c := r.ErrorCodes

	test.ExpectInt(t, len(c), 1)

	ov, err = ob.parseRule("CM", []string{"REQ:MISSING"})

	test.ExpectNil(t, err)

	r, err = ov.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c), 1)

	ov, err = ob.parseRule("CV", []string{"REQ:MISSING"})

	test.ExpectNil(t, err)

	r, err = ov.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c), 0)

}

func TestSetObjDetection(t *testing.T) {

	ob := NewObjectValidatorBuilder("DEF", nil)

	ov, err := ob.parseRule("CP", []string{"REQ:MISSING"})

	test.ExpectNil(t, err)

	sub := new(Parent)
	sub.CP = new(Child)
	vc := new(validationContext)
	vc.Subject = sub

	r, err := ov.Validate(vc)
	test.ExpectNil(t, err)
	c := r.ErrorCodes

	test.ExpectInt(t, len(c), 0)

	sub.CM = make(map[string]interface{})
	ov, err = ob.parseRule("CM", []string{"REQ:MISSING"})

	test.ExpectNil(t, err)

	r, err = ov.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c), 0)

	ov, err = ob.parseRule("CV", []string{"REQ:MISSING"})

	test.ExpectNil(t, err)

	r, err = ov.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c), 0)

}

type Parent struct {
	CP *Child
	CV Child
	CM map[string]interface{}
}

type Child struct {
}
