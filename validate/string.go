// Copyright 2016-2023 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package validate

import (
	"errors"
	"fmt"
	"github.com/graniticio/granitic/v2/ioc"
	rt "github.com/graniticio/granitic/v2/reflecttools"
	"github.com/graniticio/granitic/v2/types"
	"regexp"
	"strings"
)

const noBound = -1
const setMemberSep = ","
const stringRuleCode = "STR"

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

type stringValidationOperation uint

const (
	stringOpUnsupported = iota
	stringOpTrim
	stringOpHardTrim
	stringOpLen
	stringOpIn
	stringOpExt
	stringOpRequired
	stringOpBreak
	stringOpReg
	stringOpStopAll
	stringOpMEx
)

// An ExternalStringValidator is an object able to evaluate the supplied string to see if it meets some definition of validity.
type ExternalStringValidator interface {
	// ValidString returns true if the implementation considers the supplied string to be valid.
	ValidString(string) (bool, error)
}

type trimMode uint

const (
	noTrim   = 0
	softTrim = 1
	hardTrim = 2
)

// NewStringValidationRule creates a new NewStringValidationRule to check the specified field.
func NewStringValidationRule(field, defaultErrorCode string) *StringValidationRule {
	sv := new(StringValidationRule)
	sv.defaultErrorCode = defaultErrorCode
	sv.field = field
	sv.codesInUse = types.NewOrderedStringSet([]string{})
	sv.dependsFields = determinePathFields(field)
	sv.trim = noTrim

	return sv
}

// StringValidationRule is a ValidationRule able to validate a string or NilableString field. See the method definitions on this type for
// the supported operations.
type StringValidationRule struct {
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

// DependsOnFields implements ValidationRule.DependsOnFields
func (sv *StringValidationRule) DependsOnFields() types.StringSet {
	return sv.dependsFields
}

// CodesInUse implements ValidationRule.CodesInUse
func (sv *StringValidationRule) CodesInUse() types.StringSet {
	return sv.codesInUse
}

// IsSet returns false if the field is a string or if it is a nil or unset NilableString
func (sv *StringValidationRule) IsSet(field string, subject interface{}) (bool, error) {
	ns, err := sv.extractValue(field, subject)

	if err != nil {
		return false, err
	}

	if ns == nil || !ns.IsSet() {
		return false, nil
	}

	return true, nil
}

// Validate implements ValidationRule.Validate
func (sv *StringValidationRule) Validate(vc *ValidationContext) (result *ValidationResult, unexpected error) {

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

func (sv *StringValidationRule) applyTrimming(f string, s interface{}, ns *types.NilableString, vc *ValidationContext) string {

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

func (sv *StringValidationRule) extractValue(f string, s interface{}) (*types.NilableString, error) {

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

func (sv *StringValidationRule) runOperations(field string, s string, vc *ValidationContext, r *ValidationResult) error {

	ec := types.NewEmptyOrderedStringSet()

OpLoop:
	for _, op := range sv.operations {

		switch op.OpType {
		case stringOpLen:
			if !sv.lengthOkay(s) {
				ec.Add(op.ErrCode)
			}
		case stringOpIn:
			if !op.InSet.Contains(s) {
				ec.Add(op.ErrCode)
			}

		case stringOpExt:
			if valid, err := op.External.ValidString(s); err == nil && !valid {
				ec.Add(op.ErrCode)
			} else if err != nil {
				return err
			}

		case stringOpBreak:

			if ec.Size() > 0 {
				break OpLoop
			}

		case stringOpReg:
			if !op.Regex.MatchString(s) {
				ec.Add(op.ErrCode)
			}

		case stringOpMEx:
			checkMExFields(op.MExFields, vc, ec, op.ErrCode)
		}

	}

	r.AddForField(field, ec.Contents())

	return nil

}

func (sv *StringValidationRule) lengthOkay(s string) bool {

	if sv.minLen == noBound && sv.maxLen == noBound {
		return true
	}

	sl := len(s)

	minOkay := sv.minLen == noBound || sl >= sv.minLen
	maxOkay := sv.maxLen == noBound || sl <= sv.maxLen

	return minOkay && maxOkay

}

func (sv *StringValidationRule) wasStringSet(s string, field string, knownSet types.StringSet) bool {

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

// StopAllOnFail implements ValidationRule.StopAllOnFail
func (sv *StringValidationRule) StopAllOnFail() bool {
	return sv.stopAll
}

// MEx adds a check to see if any other of the fields with which this field is mutually exclusive have been set.
func (sv *StringValidationRule) MEx(fields types.StringSet, code ...string) *StringValidationRule {
	op := new(stringOperation)
	op.ErrCode = sv.chooseErrorCode(code)
	op.OpType = stringOpMEx
	op.MExFields = fields

	sv.addOperation(op)

	return sv
}

// Break adds a check to stop processing this rule if the previous check has failed.
func (sv *StringValidationRule) Break() *StringValidationRule {

	o := new(stringOperation)
	o.OpType = stringOpBreak

	sv.addOperation(o)

	return sv

}

// Length adds a check to see if the string has a length between the supplied min and max values.
func (sv *StringValidationRule) Length(min, max int, code ...string) *StringValidationRule {

	sv.minLen = min
	sv.maxLen = max

	ec := sv.chooseErrorCode(code)

	o := new(stringOperation)
	o.OpType = stringOpLen
	o.ErrCode = ec

	sv.addOperation(o)

	return sv

}

// In adds a check to confirm that the string exactly matches one of those in the supplied set.
func (sv *StringValidationRule) In(set []string, code ...string) *StringValidationRule {

	ss := types.NewUnorderedStringSet(set)

	ec := sv.chooseErrorCode(code)

	o := new(stringOperation)
	o.OpType = stringOpIn
	o.ErrCode = ec
	o.InSet = ss

	sv.addOperation(o)

	return sv

}

// HardTrim specifies that the string to be validated should be trimmed before validation. This affects
// the underlying value permanently. To have the same functionality without modifying the string, use
// the Trim method instead.
func (sv *StringValidationRule) HardTrim() *StringValidationRule {

	sv.trim = hardTrim

	return sv
}

// Trim specifies that all validation checks should be performed on a copy of the string with leading and
// trailing whitespace removed.
func (sv *StringValidationRule) Trim() *StringValidationRule {

	sv.trim = softTrim

	return sv
}

// StopAll indicates that no further rules should be rule if this one fails.
func (sv *StringValidationRule) StopAll() *StringValidationRule {

	sv.stopAll = true

	return sv
}

// Required adds a check to see if the field under validation has been set.
func (sv *StringValidationRule) Required(code ...string) *StringValidationRule {

	sv.required = true

	if code != nil {
		sv.missingRequiredCode = code[0]
	} else {
		sv.missingRequiredCode = sv.defaultErrorCode
	}

	return sv
}

// ExternalValidation adds a check to call the supplied object to ask it to check the validity of the string in question.
func (sv *StringValidationRule) ExternalValidation(v ExternalStringValidator, code ...string) *StringValidationRule {
	ec := sv.chooseErrorCode(code)

	o := new(stringOperation)
	o.OpType = stringOpExt
	o.ErrCode = ec
	o.External = v

	sv.addOperation(o)

	return sv
}

// Regex adds a check to confirm that the string being validated matches the supplied regular expression.
func (sv *StringValidationRule) Regex(r *regexp.Regexp, code ...string) *StringValidationRule {
	ec := sv.chooseErrorCode(code)

	o := new(stringOperation)
	o.OpType = stringOpReg
	o.ErrCode = ec
	o.Regex = r

	sv.addOperation(o)

	return sv
}

func (sv *StringValidationRule) addOperation(o *stringOperation) {
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

func (sv *StringValidationRule) operation(c string) (stringValidationOperation, error) {
	switch c {
	case stringOpTrimCode:
		return stringOpTrim, nil
	case stringOpHardTrimCode:
		return stringOpHardTrim, nil
	case stringOpLenCode:
		return stringOpLen, nil
	case stringOpInCode:
		return stringOpIn, nil
	case stringOpExtCode:
		return stringOpExt, nil
	case stringOpRequiredCode:
		return stringOpRequired, nil
	case stringOpBreakCode:
		return stringOpBreak, nil
	case stringOpRegCode:
		return stringOpReg, nil
	case stringOpStopAllCode:
		return stringOpStopAll, nil
	case stringOpMExCode:
		return stringOpMEx, nil
	}

	m := fmt.Sprintf("Unsupported string validation operation %s", c)
	return stringOpUnsupported, errors.New(m)

}

func (sv *StringValidationRule) chooseErrorCode(v []string) string {

	if len(v) > 0 {
		sv.codesInUse.Add(v[0])
		return v[0]
	}

	return sv.defaultErrorCode
}

type stringOperation struct {
	OpType    stringValidationOperation
	ErrCode   string
	InSet     *types.UnorderedStringSet
	External  ExternalStringValidator
	Regex     *regexp.Regexp
	MExFields types.StringSet
}

func newStringValidationRuleBuilder(defaultErrorCode string) *stringValidationRuleBuilder {
	vb := new(stringValidationRuleBuilder)
	vb.strLenRegex = regexp.MustCompile(lengthPattern)
	vb.defaultErrorCode = defaultErrorCode

	return vb
}

type stringValidationRuleBuilder struct {
	strLenRegex      *regexp.Regexp
	defaultErrorCode string
	componentFinder  ioc.ComponentLookup
}

func (vb *stringValidationRuleBuilder) parseRule(field string, rule []string) (ValidationRule, error) {

	defaultErrorcode := determineDefaultErrorCode(stringRuleCode, rule, vb.defaultErrorCode)
	sv := NewStringValidationRule(field, defaultErrorcode)

	for _, v := range rule {

		ops := decomposeOperation(v)
		opCode := ops[0]

		if isTypeIndicator(stringRuleCode, opCode) {
			continue
		}

		op, err := sv.operation(opCode)

		if err != nil {
			return nil, err
		}

		switch op {
		case stringOpBreak:
			sv.Break()
		case stringOpLen:
			err = vb.addStringLenOperation(field, ops, sv)
		case stringOpReg:
			err = vb.addStringRegexOperation(field, ops, sv)
		case stringOpExt:
			err = vb.addStringExternalOperation(field, ops, sv)
		case stringOpIn:
			err = vb.addStringInOperation(field, ops, sv)
		case stringOpHardTrim:
			sv.HardTrim()
		case stringOpTrim:
			sv.Trim()
		case stringOpRequired:
			err = vb.markRequired(field, ops, sv)
		case stringOpStopAll:
			sv.StopAll()
		case stringOpMEx:
			err = vb.captureExclusiveFields(field, ops, sv)
		}

		if err != nil {

			return nil, err
		}

	}

	return sv, nil

}

func (vb *stringValidationRuleBuilder) captureExclusiveFields(field string, ops []string, iv *StringValidationRule) error {
	_, err := paramCount(ops, "MEX", field, 2, 3)

	if err != nil {
		return err
	}

	members := strings.SplitN(ops[1], setMemberSep, -1)
	fields := types.NewOrderedStringSet(members)

	iv.MEx(fields, extractVargs(ops, 3)...)

	return nil

}

func (vb *stringValidationRuleBuilder) markRequired(field string, ops []string, sv *StringValidationRule) error {

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

func (vb *stringValidationRuleBuilder) addStringInOperation(field string, ops []string, sv *StringValidationRule) error {

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

func (vb *stringValidationRuleBuilder) addStringRegexOperation(field string, ops []string, sv *StringValidationRule) error {

	pCount, err := paramCount(ops, "Regex", field, 2, 3)

	if err != nil {
		return err
	}

	pattern := ops[1]

	r, err := regexp.Compile(pattern)

	if err != nil {
		m := fmt.Sprintf("Regex for field %s could not be compiled. Pattern provided was ", field)
		return errors.New(m)
	}

	if pCount == 2 {
		sv.Regex(r)
	} else {
		sv.Regex(r, ops[2])
	}

	return nil

}

func (vb *stringValidationRuleBuilder) addStringExternalOperation(field string, ops []string, sv *StringValidationRule) error {

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

func (vb *stringValidationRuleBuilder) addStringLenOperation(field string, ops []string, sv *StringValidationRule) error {

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
