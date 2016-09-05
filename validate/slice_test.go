package validate

import (
	"github.com/graniticio/granitic/test"
	"github.com/graniticio/granitic/types"
	"testing"
)

func TestSliceSet(t *testing.T) {

	vb := NewSliceValidatorBuilder("DEF", nil, nil)

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
	sb := NewSliceValidatorBuilder("DEF", nil, nil)

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

func TestSliceElemValidation(t *testing.T) {

	rv := new(RuleValidator)

	rm := new(UnparsedRuleManager)

	rules := make(map[string][]string)
	rm.Rules = rules

	rules["lenCheck"] = []string{"STR", "LEN:-5:TOOLONG", "HARDTRIM"}
	rules["intRange"] = []string{"INT", "RANGE:4|5"}
	rules["objCheck"] = []string{"OBJ"}

	rv.RuleManager = rm

	rv.stringBuilder = NewStringValidatorBuilder("DEFSTR")
	rv.objectValidatorBuilder = NewObjectValidatorBuilder("DEFOBJ", nil)
	rv.intValidatorBuilder = NewIntValidatorBuilder("DEFINT", nil)
	rv.floatValidatorBuilder = NewFloatValidatorBuilder("DEFFLT", nil)

	vb := NewSliceValidatorBuilder("DEF", nil, rv)

	field := "S"

	_, err := vb.parseRule(field, []string{"ELEM:notExist"})
	test.ExpectNotNil(t, err)

	_, err = vb.parseRule(field, []string{"ELEM:objCheck"})
	test.ExpectNotNil(t, err)

	sv, err := vb.parseRule(field, []string{"ELEM:lenCheck:LEN"})
	test.ExpectNil(t, err)

	sub := new(SliceTest)
	sub.S = []string{"A", "B", "C"}

	vc := new(ValidationContext)
	vc.Subject = sub

	r, err := sv.Validate(vc)
	c := r.ErrorCodes

	test.ExpectInt(t, len(c), 0)

	sub.S = []string{"A", "B12345", "C12345"}

	r, err = sv.Validate(vc)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c), 2)

	sub.S = []string{"   A   ", " B2345", "C1234 "}

	r, err = sv.Validate(vc)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c), 0)

	field = "NS"

	sub.NS = []*types.NilableString{types.NewNilableString("  B  ")}
	sv, err = vb.parseRule(field, []string{"ELEM:lenCheck:LEN"})

	r, err = sv.Validate(vc)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c), 0)

	field = "I"

	sub.I = []int{1, 2, 3, 4, 5}
	sv, err = vb.parseRule(field, []string{"ELEM:intRange:INTSIZE"})
	test.ExpectNil(t, err)

	r, err = sv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c), 3)

	field = "NI"

	sub.NI = []*types.NilableInt64{types.NewNilableInt64(1), types.NewNilableInt64(5)}
	sv, err = vb.parseRule(field, []string{"ELEM:intRange:INTSIZE"})
	test.ExpectNil(t, err)

	r, err = sv.Validate(vc)
	test.ExpectNil(t, err)
	c = r.ErrorCodes

	test.ExpectInt(t, len(c), 1)

}

func TestSliceMExFieldDetection(t *testing.T) {
	vb := NewSliceValidatorBuilder("DEF", nil, nil)

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
	S  []string
	NS []*types.NilableString
	I  []int
	NI []*types.NilableInt64
}
