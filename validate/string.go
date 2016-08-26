package validate

import (
	"errors"
	"fmt"
	"github.com/graniticio/granitic/ioc"
	rt "github.com/graniticio/granitic/reflecttools"
	"github.com/graniticio/granitic/types"
	"github.com/graniticio/granitic/ws/nillable"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

const NoLimit = -1
const setMemberSep = ","

const (
	stringOpTrimCode     = "TRIM"
	stringOpHardTrimCode = "HARDTRIM"
	stringOpLenCode      = "LEN"
	stringOpInCode       = "IN"
	stringOpExtCode      = "EXT"
	stringOpRequiredCode = "REQ"
	stringOpBreakCode    = "BREAK"
	stringOpRegCode      = "REG"
	stringOpStopAllCode  = "STOPALL"
)

type StringValidationOperation uint

const (
	StringOpUnsupported = iota
	StringOpTrim
	StringOpHardTrim
	StringOpLen
	StringOpIn
	StringOpExt
	StringOpRequired
	StringOpBreak
	StringOpReg
	StringOpStopAll
)

type ExternalStringValidator interface {
	ValidString(string) bool
}

type trimMode uint

const (
	noTrim   = 0
	softTrim = 1
	hardTrim = 2
)

type StringValidator struct {
	defaultErrorcode    string
	missingRequiredCode string
	field               string
	operations          []*stringOperation
	minLen              int
	maxLen              int
	trim                trimMode
	required            bool
	stopAll             bool
	codesInUse          types.StringSet
	dependsFields       types.StringSet
}

func NewStringValidator(field, defaultErrorCode string) *StringValidator {
	sv := new(StringValidator)
	sv.defaultErrorcode = defaultErrorCode
	sv.field = field
	sv.codesInUse = types.NewOrderedStringSet([]string{})
	sv.dependsFields = determinePathFields(field)

	return sv
}

func (sv *StringValidator) DependsOnFields() types.StringSet {
	return sv.dependsFields
}

func (sv *StringValidator) CodesInUse() types.StringSet {
	return sv.codesInUse
}

func (sv *StringValidator) Validate(vc *validationContext) (result *ValidationResult, unexpected error) {

	f := sv.field
	sub := vc.Subject

	fv, err := rt.FindNestedField(rt.ExtractDotPath(f), sub)

	if err != nil {
		m := fmt.Sprintf("Problem trying to find value of %s: %s\n", f, err)
		return nil, errors.New(m)
	}

	if !fv.IsValid() {
		m := fmt.Sprintf("Field %s is not a usable type\n", f)
		return nil, errors.New(m)
	}

	ns, found := fv.Interface().(*nillable.NillableString)

	if found {
		return sv.validateNillable(vc, fv, ns)
	}

	s, found := fv.Interface().(string)

	if found {
		return sv.validateStandard(vc, fv, s, f)
	} else {
		m := fmt.Sprintf("%s is not a string or a NillableString\n", f, err)
		return nil, errors.New(m)
	}

}

func (sv *StringValidator) validateNillable(vc *validationContext, rv reflect.Value, ns *nillable.NillableString) (result *ValidationResult, unexpected error) {

	if ns == nil || !ns.IsSet() {
		r := new(ValidationResult)

		if sv.required {
			r.ErrorCodes = []string{sv.missingRequiredCode}
		} else {
			r.ErrorCodes = []string{}
		}

		r.Unset = true

		return r, nil
	}

	toValidate := ns.String()

	if sv.trim == hardTrim || sv.trim == softTrim {
		toValidate = strings.TrimSpace(toValidate)

		if sv.trim == hardTrim {
			ns.Set(toValidate)
		}
	}

	return sv.runOperations(toValidate)
}

func (sv *StringValidator) validateStandard(vc *validationContext, rv reflect.Value, s string, field string) (result *ValidationResult, unexpected error) {

	if !sv.wasStringSet(s, field, vc.KnownSetFields) {

		r := new(ValidationResult)

		if sv.required {
			r.ErrorCodes = []string{sv.missingRequiredCode}
		} else {
			r.ErrorCodes = []string{}
		}

		r.Unset = true

		return r, nil

	}

	toValidate := s

	if sv.trim == hardTrim || sv.trim == softTrim {
		toValidate = strings.TrimSpace(toValidate)

		if sv.trim == hardTrim {
			rv.SetString(toValidate)
		}
	}

	return sv.runOperations(toValidate)
}

func (sv *StringValidator) runOperations(s string) (*ValidationResult, error) {

	ec := new(types.OrderedStringSet)

OpLoop:
	for _, op := range sv.operations {

		switch op.OpType {
		case StringOpLen:
			if !sv.lengthOkay(s) {
				ec.Add(op.ErrCode)
			}
		case StringOpIn:
			if !op.InSet.Contains(s) {
				ec.Add(op.ErrCode)
			}

		case StringOpExt:
			if !op.External.ValidString(s) {

				ec.Add(op.ErrCode)
			}

		case StringOpBreak:

			if ec.Size() > 0 {
				break OpLoop
			}

		case StringOpReg:
			if !op.Regex.MatchString(s) {
				ec.Add(op.ErrCode)
			}
		}

	}

	r := new(ValidationResult)
	r.ErrorCodes = ec.Contents()

	return r, nil

}

func (sv *StringValidator) lengthOkay(s string) bool {

	if sv.minLen == NoLimit && sv.maxLen == NoLimit {
		return true
	}

	sl := len(s)

	minOkay := sv.minLen == NoLimit || sl >= sv.minLen
	maxOkay := sv.maxLen == NoLimit || sl <= sv.maxLen

	return minOkay && maxOkay

}

func (sv *StringValidator) wasStringSet(s string, field string, knownSet types.StringSet) bool {

	l := len(s)

	if l > 0 {
		return true
	} else if knownSet != nil && knownSet.Contains(field) {
		return true
	} else if sv.minLen <= 0 {
		return true
	}

	return false

}

func (sv *StringValidator) StopAllOnFail() bool {
	return sv.stopAll
}

func (sv *StringValidator) Break() *StringValidator {

	o := new(stringOperation)
	o.OpType = StringOpBreak

	sv.addOperation(o)

	return sv

}

func (sv *StringValidator) Length(min, max int, code ...string) *StringValidator {

	sv.minLen = min
	sv.maxLen = max

	ec := sv.chooseErrorCode(code)

	o := new(stringOperation)
	o.OpType = StringOpLen
	o.ErrCode = ec

	sv.addOperation(o)

	return sv

}

func (sv *StringValidator) In(set []string, code ...string) *StringValidator {

	ss := types.NewUnorderedStringSet(set)

	ec := sv.chooseErrorCode(code)

	o := new(stringOperation)
	o.OpType = StringOpIn
	o.ErrCode = ec
	o.InSet = ss

	sv.addOperation(o)

	return sv

}

func (sv *StringValidator) HardTrim() *StringValidator {

	sv.trim = hardTrim

	return sv
}

func (sv *StringValidator) Trim() *StringValidator {

	sv.trim = softTrim

	return sv
}

func (sv *StringValidator) StopAll() *StringValidator {

	sv.stopAll = true

	return sv
}

func (sv *StringValidator) Required(code ...string) *StringValidator {

	sv.required = true

	if code != nil {
		sv.missingRequiredCode = code[0]
	} else {
		sv.missingRequiredCode = sv.defaultErrorcode
	}

	return sv
}

func (sv *StringValidator) ExternalValidation(v ExternalStringValidator, code ...string) *StringValidator {
	ec := sv.chooseErrorCode(code)

	o := new(stringOperation)
	o.OpType = StringOpExt
	o.ErrCode = ec
	o.External = v

	sv.addOperation(o)

	return sv
}

func (sv *StringValidator) Regex(r *regexp.Regexp, code ...string) *StringValidator {
	ec := sv.chooseErrorCode(code)

	o := new(stringOperation)
	o.OpType = StringOpReg
	o.ErrCode = ec
	o.Regex = r

	sv.addOperation(o)

	return sv
}

func (sv *StringValidator) addOperation(o *stringOperation) {
	if sv.operations == nil {
		sv.operations = make([]*stringOperation, 0)
	}

	if sv.codesInUse == nil {
		sv.codesInUse = types.NewUnorderedStringSet([]string{})
	}

	sv.operations = append(sv.operations, o)
	sv.codesInUse.Add(o.ErrCode)
}

func (sv *StringValidator) Operation(c string) (StringValidationOperation, error) {
	switch c {
	case stringOpTrimCode:
		return StringOpTrim, nil
	case stringOpHardTrimCode:
		return StringOpHardTrim, nil
	case stringOpLenCode:
		return StringOpLen, nil
	case stringOpInCode:
		return StringOpIn, nil
	case stringOpExtCode:
		return StringOpExt, nil
	case stringOpRequiredCode:
		return StringOpRequired, nil
	case stringOpBreakCode:
		return StringOpBreak, nil
	case stringOpRegCode:
		return StringOpReg, nil
	case stringOpStopAllCode:
		return StringOpStopAll, nil
	}

	m := fmt.Sprintf("Unsupported string validation operation %s", c)
	return StringOpUnsupported, errors.New(m)

}

func (sv *StringValidator) chooseErrorCode(v []string) string {

	if len(v) > 0 {
		return v[0]
	} else {
		return sv.defaultErrorcode
	}

}

type stringOperation struct {
	OpType   StringValidationOperation
	ErrCode  string
	InSet    *types.UnorderedStringSet
	External ExternalStringValidator
	Regex    *regexp.Regexp
}

type stringValidatorBuilder struct {
	strLenRegex      *regexp.Regexp
	defaultErrorCode string
	componentFinder  ioc.ComponentByNameFinder
}

func (vb *stringValidatorBuilder) parseStringRule(field string, rule []string) (Validator, error) {

	defaultErrorcode := DetermineDefaultErrorCode(StringRuleCode, rule, vb.defaultErrorCode)
	sv := NewStringValidator(field, defaultErrorcode)

	for _, v := range rule {

		ops := DecomposeOperation(v)
		opCode := ops[0]

		if IsTypeIndicator(StringRuleCode, opCode) {
			continue
		}

		op, err := sv.Operation(opCode)

		if err != nil {
			return nil, err
		}

		switch op {
		case StringOpBreak:
			sv.Break()
		case StringOpLen:
			err = vb.addStringLenOperation(field, ops, sv)
		case StringOpReg:
			err = vb.addStringRegexOperation(field, ops, sv)
		case StringOpExt:
			err = vb.addStringExternalOperation(field, ops, sv)
		case StringOpIn:
			err = vb.addStringInOperation(field, ops, sv)
		case StringOpHardTrim:
			sv.HardTrim()
		case StringOpTrim:
			sv.Trim()
		case StringOpRequired:
			vb.markRequired(field, ops, sv)
		case StringOpStopAll:
			sv.StopAll()
		}

		if err != nil {

			return nil, err
		}

	}

	return sv, nil

}

func (vb *stringValidatorBuilder) markRequired(field string, ops []string, sv *StringValidator) error {
	opParams := len(ops)

	if opParams < 1 || opParams > 2 {
		m := fmt.Sprintf("Required marked for field %s is invalid (too few or too many parameters)", field)
		return errors.New(m)
	}

	if opParams == 1 {
		sv.Required()
	} else {
		sv.Required(ops[1])
	}

	return nil
}

func (vb *stringValidatorBuilder) addStringInOperation(field string, ops []string, sv *StringValidator) error {
	opParams := len(ops)

	if opParams < 2 || opParams > 3 {
		m := fmt.Sprintf("In operation for field %s is invalid (too few or too many parameters)", field)
		return errors.New(m)
	}

	members := strings.SplitN(ops[1], setMemberSep, -1)

	if opParams == 2 {
		sv.In(members)
	} else {
		sv.In(members, ops[2])
	}

	return nil

}

func (vb *stringValidatorBuilder) addStringRegexOperation(field string, ops []string, sv *StringValidator) error {

	opParams := len(ops)

	if opParams < 2 || opParams > 3 {
		m := fmt.Sprintf("Regex operation for field %s is invalid (too few or too many parameters)", field)
		return errors.New(m)
	}

	pattern := ops[1]

	r, err := regexp.Compile(pattern)

	if err != nil {
		m := fmt.Sprintf("Regex for field %s could not be compiled. Pattern provided was ", field, pattern)
		return errors.New(m)
	}

	if opParams == 2 {
		sv.Regex(r)
	} else {
		sv.Regex(r, ops[2])
	}

	return nil

}

func (vb *stringValidatorBuilder) addStringExternalOperation(field string, ops []string, sv *StringValidator) error {

	cf := vb.componentFinder

	if cf == nil {
		m := fmt.Sprintf("Field %s relies on an external component to validate, but no ioc.ComponentByNameFinder is available.", field)
		return errors.New(m)
	}

	opParams := len(ops)

	if opParams < 2 || opParams > 3 {
		m := fmt.Sprintf("External operation for field %s is invalid (too few or too many parameters)", field)
		return errors.New(m)
	}

	ref := ops[1]
	component := cf.ComponentByName(ref)

	if component == nil {
		m := fmt.Sprintf("No external component named %s available to validate field %s", ref, field)
		return errors.New(m)
	}

	ev, found := component.Instance.(ExternalStringValidator)

	if !found {
		m := fmt.Sprintf("Component %s to validate field %s does not implement ExternalStringValidator", ref, field)
		return errors.New(m)
	}

	if opParams == 2 {
		sv.ExternalValidation(ev)
	} else {
		sv.ExternalValidation(ev, ops[2])
	}

	return nil

}

func (vb *stringValidatorBuilder) addStringLenOperation(field string, ops []string, sv *StringValidator) error {

	opParams := len(ops)

	if opParams < 2 || opParams > 3 {
		m := fmt.Sprintf("Length operation for field %s is invalid (too few or too many parameters)", field)
		return errors.New(m)
	}

	vals := ops[1]

	if !vb.strLenRegex.MatchString(vals) {
		m := fmt.Sprintf("Length parameters for field %s are invalid. Values provided: %s", field, vals)
		return errors.New(m)
	}

	min := NoLimit
	max := NoLimit

	groups := vb.strLenRegex.FindStringSubmatch(vals)

	if groups[1] != "" {
		min, _ = strconv.Atoi(groups[1])
	}

	if groups[2] != "" {
		max, _ = strconv.Atoi(groups[2])
	}

	if opParams == 2 {
		sv.Length(min, max)
	} else {
		sv.Length(min, max, ops[2])
	}

	return nil

}

func newStringValidatorBuilder(defaultErrorCode string) *stringValidatorBuilder {
	vb := new(stringValidatorBuilder)
	vb.strLenRegex = regexp.MustCompile("^(\\d*)-(\\d*)$")
	vb.defaultErrorCode = defaultErrorCode

	return vb
}
