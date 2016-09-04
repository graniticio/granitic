package validate

import (
	"github.com/graniticio/granitic/test"
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

type SliceTest struct {
	S []string
}
