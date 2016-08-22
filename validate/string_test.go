package validate

import (
	"github.com/graniticio/granitic/test"
	"github.com/graniticio/granitic/ws/nillable"
	"regexp"
	"testing"
)

func TestStringValidatorSanity(t *testing.T) {

	sv := new(StringValidator)

	_, err := sv.Validate(NewValidateContext(1))

	test.ExpectNotNil(t, err)

	_, err = sv.Validate(NewValidateContext("NOCODE"))

	test.ExpectNotNil(t, err)

	sv.DefaultCode("DEF")

	c, err := sv.Validate(NewValidateContext("NOCODE"))

	test.ExpectInt(t, len(c), 0)
	test.ExpectNil(t, err)

}

func TestStringValidatorAllChecksPass(t *testing.T) {
	r := regexp.MustCompile("[A-Z].*")

	sv := NewStringValidator("DEF").MinLength(1).MaxLength(10).
		ForbidPadding().Match(r).
		Code(MinLengthCheck, "MIN").Code(MaxLengthCheck, "MAX").
		Code(RegexMatch, "REG")

	c, err := sv.Validate(NewValidateContext("Abcdefghij"))

	test.ExpectNil(t, err)
	test.ExpectInt(t, len(c), 0)

}

func TestStringValidatorAllChecksFail(t *testing.T) {

	r := regexp.MustCompile("[A-Z].*")

	sv := NewStringValidator("DEF").MinLength(10).MaxLength(1).
		MinTrimmedLength(10).MaxTrimmedLength(2).ForbidPadding().Match(r).
		Code(MinLengthCheck, "MIN").Code(MaxLengthCheck, "MAX").
		Code(MinTrimmedLengthCheck, "TMIN").Code(MaxTrimmedLengthCheck, "TMAX").
		Code(RegexMatch, "REG")

	c, err := sv.Validate(NewValidateContext("  abc "))

	test.ExpectNil(t, err)
	test.ExpectInt(t, len(c), 6)

}

func TestStringValidatorBreaks(t *testing.T) {

	r := regexp.MustCompile("[A-Z].*")

	sv := NewStringValidator("DEF").MinLength(10).ForbidPadding().BreakOnFail().Match(r).
		Code(MinLengthCheck, "MIN").Code(NoPadding, "NP").Code(RegexMatch, "REG")

	c, err := sv.Validate(NewValidateContext("  abc "))

	test.ExpectNil(t, err)
	test.ExpectInt(t, len(c), 2)

	ns := nillable.NewNillableString("abcdefghij")

	c, err = sv.Validate(NewValidateContext(ns))

	test.ExpectNil(t, err)
	test.ExpectInt(t, len(c), 1)
	test.ExpectString(t, c[0], "REG")

}

func TestUnsetDetection(t *testing.T) {
	sv := NewStringValidator("UNSET").IsSet().MinLength(2).Code(StringSet, "SET").Code(MinLengthCheck, "MINL")

	vc := new(ValidateContext)
	vc.FieldName = "f1"

	vc.V = new(nillable.NillableString)

	c, err := sv.Validate(vc)

	test.ExpectNil(t, err)
	test.ExpectInt(t, len(c), 1)
	test.ExpectString(t, c[0], "SET")

	vc.V = ""
	vc.BoundFields = []string{"f2", "f3"}
	c, err = sv.Validate(vc)

	test.ExpectNil(t, err)
	test.ExpectInt(t, len(c), 1)
	test.ExpectString(t, c[0], "SET")
}

func TestSetDetection(t *testing.T) {
	sv := NewStringValidator("UNSET").IsSet().MinLength(2).Code(StringSet, "SET").Code(MinLengthCheck, "MINL")

	vc := new(ValidateContext)
	vc.FieldName = "f1"

	vc.V = nillable.NewNillableString("")

	c, err := sv.Validate(vc)

	test.ExpectNil(t, err)
	test.ExpectInt(t, len(c), 1)
	test.ExpectString(t, c[0], "MINL")

	vc.V = ""
	vc.BoundFields = []string{"f1", "f3"}
	c, err = sv.Validate(vc)

	test.ExpectNil(t, err)
	test.ExpectInt(t, len(c), 1)
	test.ExpectString(t, c[0], "MINL")

	vc.V = "A"
	vc.BoundFields = []string{}
	c, err = sv.Validate(vc)

	test.ExpectNil(t, err)
	test.ExpectInt(t, len(c), 1)
	test.ExpectString(t, c[0], "MINL")
}
