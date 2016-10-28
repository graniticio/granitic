package validate

import (
	"errors"
	"fmt"
	"github.com/graniticio/granitic/ioc"
	rt "github.com/graniticio/granitic/reflecttools"
	"github.com/graniticio/granitic/types"
	"regexp"
	"strings"
)

const NoLimit = -1
const setMemberSep = ","
const StringRuleCode = "STR"

const (
	stringOpTrimCode     = "TRIM"
	stringOpHardTrimCode = "HARDTRIM"
	stringOpLenCode      = commonOpLen
	stringOpInCode       = commonOpIn
	stringOpExtCode      = commonOpExt
	stringOpRequiredCode = commonOpRequired
	stringOpBreakCode    = commonOpBreak
	stringOpRegCode      = "REG"
	stringOpStopAllCode  = commonOpStopAll
	stringOpMExCode      = commonOpMex
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
	StringOpMEx
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
	defaultErrorCode    string
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
	sv.defaultErrorCode = defaultErrorCode
	sv.field = field
	sv.codesInUse = types.NewOrderedStringSet([]string{})
	sv.dependsFields = determinePathFields(field)
	sv.trim = noTrim

	return sv
}

func (sv *StringValidator) DependsOnFields() types.StringSet {
	return sv.dependsFields
}

func (sv *StringValidator) CodesInUse() types.StringSet {
	return sv.codesInUse
}

func (sv *StringValidator) IsSet(field string, subject interface{}) (bool, error) {
	ns, err := sv.extractValue(field, subject)

	if err != nil {
		return false, err
	}

	if ns == nil || !ns.IsSet() {
		return false, nil
	} else {
		return true, nil
	}
}

func (sv *StringValidator) Validate(vc *ValidationContext) (result *ValidationResult, unexpected error) {

	var value *types.NilableString

	f := sv.field

	if vc.OverrideField != "" {
		f = vc.OverrideField
	}

	sub := vc.Subject
	r := NewValidationResult()

	if vc.DirectSubject {

		s, found := sub.(*types.NilableString)

		if !found {
			m := fmt.Sprintf("Direct validation requested for %s but supplied value is not a *NilableString", f)
			return nil, errors.New(m)
		}

		value = s

	} else {

		set, err := sv.IsSet(f, sub)

		if err != nil {
			return nil, err
		} else if !set {

			if sv.required {
				r.AddForField(f, []string{sv.missingRequiredCode})
			}

			r.Unset = true

			return r, nil
		}

		//Ignoring error as called previously during IsSet
		value, _ = sv.extractValue(f, sub)
	}

	toValidate := sv.applyTrimming(f, sub, value, vc)
	err := sv.runOperations(f, toValidate, vc, r)

	return r, err
}

func (sv *StringValidator) applyTrimming(f string, s interface{}, ns *types.NilableString, vc *ValidationContext) string {

	if sv.trim == hardTrim || sv.trim == softTrim {

		t := strings.TrimSpace(ns.String())

		if sv.trim == hardTrim {

			if !vc.DirectSubject {

				v, _ := rt.FindNestedField(rt.ExtractDotPath(f), s)
				_, found := v.Interface().(string)

				if found {
					v.SetString(t)
				}
			}

			ns.Set(t)
		}

		return t
	}

	return ns.String()
}

func (sv *StringValidator) extractValue(f string, s interface{}) (*types.NilableString, error) {

	v, err := rt.FindNestedField(rt.ExtractDotPath(f), s)

	if err != nil {
		m := fmt.Sprintf("Problem trying to find value of %s: %s\n", f, err)
		return nil, errors.New(m)
	}

	if !v.IsValid() {
		m := fmt.Sprintf("Field %s is not a usable type\n", f)
		return nil, errors.New(m)
	}

	if rt.NilPointer(v) {
		return nil, nil
	}

	switch i := v.Interface().(type) {
	case *types.NilableString:
		return i, nil
	case string:
		return types.NewNilableString(i), nil
	default:
		m := fmt.Sprintf("%s is type %T, not a string or *NilableString.", f, i)
		return nil, errors.New(m)
	}
}

func (sv *StringValidator) runOperations(field string, s string, vc *ValidationContext, r *ValidationResult) error {

	ec := types.NewEmptyOrderedStringSet()

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

		case StringOpMEx:
			checkMExFields(op.MExFields, vc, ec, op.ErrCode)
		}

	}

	r.AddForField(field, ec.Contents())

	return nil

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

func (sv *StringValidator) MEx(fields types.StringSet, code ...string) *StringValidator {
	op := new(stringOperation)
	op.ErrCode = sv.chooseErrorCode(code)
	op.OpType = StringOpMEx
	op.MExFields = fields

	sv.addOperation(op)

	return sv
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
		sv.missingRequiredCode = sv.defaultErrorCode
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

	if o.ErrCode != "" {
		sv.codesInUse.Add(o.ErrCode)
	}
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
	case stringOpMExCode:
		return StringOpMEx, nil
	}

	m := fmt.Sprintf("Unsupported string validation operation %s", c)
	return StringOpUnsupported, errors.New(m)

}

func (sv *StringValidator) chooseErrorCode(v []string) string {

	if len(v) > 0 {
		sv.codesInUse.Add(v[0])
		return v[0]
	} else {
		return sv.defaultErrorCode
	}

}

type stringOperation struct {
	OpType    StringValidationOperation
	ErrCode   string
	InSet     *types.UnorderedStringSet
	External  ExternalStringValidator
	Regex     *regexp.Regexp
	MExFields types.StringSet
}

type StringValidatorBuilder struct {
	strLenRegex      *regexp.Regexp
	defaultErrorCode string
	componentFinder  ioc.ComponentByNameFinder
}

func (vb *StringValidatorBuilder) parseRule(field string, rule []string) (ValidationRule, error) {

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
			err = vb.markRequired(field, ops, sv)
		case StringOpStopAll:
			sv.StopAll()
		case StringOpMEx:
			err = vb.captureExclusiveFields(field, ops, sv)
		}

		if err != nil {

			return nil, err
		}

	}

	return sv, nil

}

func (vb *StringValidatorBuilder) captureExclusiveFields(field string, ops []string, iv *StringValidator) error {
	_, err := paramCount(ops, "MEX", field, 2, 3)

	if err != nil {
		return err
	}

	members := strings.SplitN(ops[1], setMemberSep, -1)
	fields := types.NewOrderedStringSet(members)

	iv.MEx(fields, extractVargs(ops, 3)...)

	return nil

}

func (vb *StringValidatorBuilder) markRequired(field string, ops []string, sv *StringValidator) error {

	pCount, err := paramCount(ops, "Required", field, 1, 2)

	if err != nil {
		return err
	}

	if pCount == 1 {
		sv.Required()
	} else {
		sv.Required(ops[1])
	}

	return nil
}

func (vb *StringValidatorBuilder) addStringInOperation(field string, ops []string, sv *StringValidator) error {

	pCount, err := paramCount(ops, "In Set", field, 2, 3)

	if err != nil {
		return err
	}

	members := strings.SplitN(ops[1], setMemberSep, -1)

	if pCount == 2 {
		sv.In(members)
	} else {
		sv.In(members, ops[2])
	}

	return nil

}

func (vb *StringValidatorBuilder) addStringRegexOperation(field string, ops []string, sv *StringValidator) error {

	pCount, err := paramCount(ops, "Regex", field, 2, 3)

	if err != nil {
		return err
	}

	pattern := ops[1]

	r, err := regexp.Compile(pattern)

	if err != nil {
		m := fmt.Sprintf("Regex for field %s could not be compiled. Pattern provided was ", field, pattern)
		return errors.New(m)
	}

	if pCount == 2 {
		sv.Regex(r)
	} else {
		sv.Regex(r, ops[2])
	}

	return nil

}

func (vb *StringValidatorBuilder) addStringExternalOperation(field string, ops []string, sv *StringValidator) error {

	pCount, i, err := validateExternalOperation(vb.componentFinder, field, ops)

	if err != nil {
		return err
	}

	ev, found := i.Instance.(ExternalStringValidator)

	if !found {
		m := fmt.Sprintf("Component %s to validate field %s does not implement ExternalStringValidator", i.Name, field)
		return errors.New(m)
	}

	if pCount == 2 {
		sv.ExternalValidation(ev)
	} else {
		sv.ExternalValidation(ev, ops[2])
	}

	return nil

}

func (vb *StringValidatorBuilder) addStringLenOperation(field string, ops []string, sv *StringValidator) error {

	_, err := paramCount(ops, "Length", field, 2, 3)

	if err != nil {
		return err
	}

	min, max, err := extractLengthParams(field, ops[1], vb.strLenRegex)

	if err != nil {
		return err
	}

	sv.Length(min, max, extractVargs(ops, 3)...)

	return nil

}

func paramCount(opParams []string, opName, field string, min, max int) (count int, err error) {
	pCount := len(opParams)

	if pCount < min || pCount > max {
		m := fmt.Sprintf("%s operation for field %s is invalid (too few or too many parameters)", opName, field)
		return 0, errors.New(m)
	}

	return pCount, nil
}

func NewStringValidatorBuilder(defaultErrorCode string) *StringValidatorBuilder {
	vb := new(StringValidatorBuilder)
	vb.strLenRegex = regexp.MustCompile(lengthPattern)
	vb.defaultErrorCode = defaultErrorCode

	return vb
}
