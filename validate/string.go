package validate

import (
	"errors"
	"github.com/graniticio/granitic/ws/nillable"
	"regexp"
	"strings"
)

type StringOp uint

const (
	MinLengthCheck = iota
	MaxLengthCheck
	MinTrimmedLengthCheck
	MaxTrimmedLengthCheck
	RegexMatch
	NoPadding
	BreakOnFail
	StringSet
)

type StringValidator struct {
	minLength        int
	maxLength        int
	minTrimmedLength int
	maxTrimmedLength int
	paddingForbidden bool
	matchPattern     *regexp.Regexp
	defaultErrorCode string
	codes            map[StringOp]string
	ops              []StringOp
}

func (sv *StringValidator) Validate(vc *ValidateContext) (failcodes []string, err error) {

	var s string
	var nilType bool

	if sv.defaultErrorCode == "" {
		return nil, errors.New("No default error code set.")
	}

	ns, found := vc.V.(*nillable.NillableString)

	if found {
		nilType = true
		s = ns.String()
	} else {

		s, found = vc.V.(string)

		if !found {
			return nil, errors.New("Value to validate is not a string.")
		}
	}

	return sv.runChecks(s, vc, nilType), nil
}

func (sv *StringValidator) addOp(o StringOp) {
	sv.ops = append(sv.ops, o)
}

func (sv *StringValidator) codeForOp(o StringOp) string {

	c := sv.codes[o]

	if c == "" {
		c = sv.defaultErrorCode
	}

	return c
}

func (sv *StringValidator) addErrorCode(o StringOp, codes map[string]string) {
	c := sv.codeForOp(o)

	codes[c] = c
}

func (sv *StringValidator) runChecks(s string, vc *ValidateContext, nilType bool) []string {
	ec := make(map[string]string)

	trimmed := strings.TrimSpace(s)
	l := len(s)
	tl := len(trimmed)

OpLoop:
	for _, o := range sv.ops {

		switch o {

		case StringSet:
			sv.checkWasSet(o, ec, s, vc, nilType)
		case MinLengthCheck:
			if l < sv.minLength {
				sv.addErrorCode(o, ec)
			}

		case MaxLengthCheck:
			if l > sv.maxLength {
				sv.addErrorCode(o, ec)
			}

		case MinTrimmedLengthCheck:
			if tl < sv.minTrimmedLength {
				sv.addErrorCode(o, ec)
			}

		case MaxTrimmedLengthCheck:
			if tl > sv.maxTrimmedLength {
				sv.addErrorCode(o, ec)
			}

		case NoPadding:
			if l != tl {
				sv.addErrorCode(o, ec)
			}

		case RegexMatch:
			if !sv.matchPattern.MatchString(s) {
				sv.addErrorCode(o, ec)
			}

		case BreakOnFail:
			if (len(ec)) > 0 {
				break OpLoop
			}
		}
	}

	ecs := make([]string, 0)

	for k, _ := range ec {
		ecs = append(ecs, k)
	}

	return ecs

}

func (sv *StringValidator) checkWasSet(o StringOp, codes map[string]string, s string, vc *ValidateContext, nilType bool) {

	var set bool

	if nilType {
		ns := vc.V.(*nillable.NillableString)

		set = ns.IsSet()
	} else if vc.WasBound(vc.FieldName) {
		set = true
	} else {
		set = len(s) > 0
	}

	if !set {
		sv.addErrorCode(o, codes)
	}
}

func (sv *StringValidator) IsSet() *StringValidator {

	sv.ops = append([]StringOp{StringSet, BreakOnFail}, sv.ops...)

	return sv
}

func (sv *StringValidator) Match(r *regexp.Regexp) *StringValidator {
	sv.matchPattern = r
	sv.addOp(RegexMatch)

	return sv
}

func (sv *StringValidator) MinLength(l int) *StringValidator {
	sv.minLength = l
	sv.addOp(MinLengthCheck)

	return sv
}

func (sv *StringValidator) MaxLength(l int) *StringValidator {
	sv.maxLength = l
	sv.addOp(MaxLengthCheck)

	return sv
}

func (sv *StringValidator) MinTrimmedLength(l int) *StringValidator {
	sv.minTrimmedLength = l
	sv.addOp(MinTrimmedLengthCheck)

	return sv
}

func (sv *StringValidator) MaxTrimmedLength(l int) *StringValidator {
	sv.maxTrimmedLength = l
	sv.addOp(MaxTrimmedLengthCheck)

	return sv
}

func (sv *StringValidator) ForbidPadding() *StringValidator {
	sv.paddingForbidden = true
	sv.addOp(NoPadding)
	return sv
}

func (sv *StringValidator) DefaultCode(c string) *StringValidator {
	sv.defaultErrorCode = c

	return sv
}

func (sv *StringValidator) Code(o StringOp, c string) *StringValidator {

	if sv.codes == nil {
		sv.codes = make(map[StringOp]string)
	}

	sv.codes[o] = c

	return sv
}

func (sv *StringValidator) BreakOnFail() *StringValidator {
	sv.addOp(BreakOnFail)

	return sv
}

func NewStringValidator(defaultCode string) *StringValidator {
	sv := new(StringValidator)
	sv.DefaultCode(defaultCode)

	return sv
}
