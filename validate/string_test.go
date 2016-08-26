package validate

import (
	"github.com/graniticio/granitic/test"
	"github.com/graniticio/granitic/ws/nillable"
	"testing"
)

func TestMissingRequiredStringField(t *testing.T) {

	sb := newStringValidatorBuilder("DEF")

	_, err := sb.parseStringRule("S", []string{"REQ:MISSING"})

	test.ExpectNil(t, err)

}

type StringTest struct {
	S  string
	NS *nillable.NillableString
}
