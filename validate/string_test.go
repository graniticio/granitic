package validate

import (
	"github.com/graniticio/granitic/test"
	"github.com/graniticio/granitic/ws/nillable"
	"testing"
)

func TestMissingRequiredStringField(t *testing.T) {

	sb := newStringValidatorBuilder("DEF")

	sv, err := sb.parseStringRule("S", []string{"REQ:MISSING", "LEN:5-10:SHORT"})

	test.ExpectNil(t, err)

	sub := new(StringTest)
	vc := new(validationContext)
	vc.Subject = sub

	c, err := sv.Validate(vc)

	test.ExpectNil(t, err)
	test.ExpectInt(t, len(c), 1)
	test.ExpectString(t, c[0], "MISSING")

}

type StringTest struct {
	S string
}

type NillableStringTest struct {
	S *nillable.NillableString
}
