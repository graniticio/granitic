// Copyright 2016-2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package validate

import (
	"errors"
	"fmt"
	"github.com/graniticio/granitic/v2/ioc"
	rt "github.com/graniticio/granitic/v2/reflecttools"
	"github.com/graniticio/granitic/v2/types"
	"reflect"
	"regexp"
	"strings"
)

const sliceRuleCode = "SLICE"

const (
	sliceOpRequiredCode = commonOpRequired
	sliceOpStopAllCode  = commonOpStopAll
	sliceOpMexCode      = commonOpMex
	sliceOpLenCode      = commonOpLen
	sliceOpElemCode     = "ELEM"
)

type sliceValidationOperation uint

const (
	sliceOpUnsupported = iota
	sliceOpRequired
	sliceOpStopAll
	sliceOpMex
	sliceOpLen
	sliceOpElem
)

type sliceOperation struct {
	OpType        sliceValidationOperation
	ErrCode       string
	MExFields     types.StringSet
	elemValidator ValidationRule
}

// NewSliceValidationRule creates a new NewSliceValidationRule to check the specified field.
func NewSliceValidationRule(field, defaultErrorCode string) *SliceValidationRule {
	bv := new(SliceValidationRule)
	bv.defaultErrorCode = defaultErrorCode
	bv.field = field
	bv.codesInUse = types.NewOrderedStringSet([]string{})
	bv.dependsFields = determinePathFields(field)
	bv.operations = make([]*sliceOperation, 0)
	bv.codesInUse.Add(bv.defaultErrorCode)
	bv.minLen = noBound
	bv.maxLen = noBound

	return bv
}

// SliceValidationRule is a ValidationRule able to validate a slice field and the individual elements of that slice. See the method definitions on this type for
// the supported operations.
type SliceValidationRule struct {
	stopAll             bool
	codesInUse          types.StringSet
	dependsFields       types.StringSet
	defaultErrorCode    string
	field               string
	missingRequiredCode string
	required            bool
	operations          []*sliceOperation
	minLen              int
	maxLen              int
}

// IsSet returns true if the field to be validated is a non-nil slice.
func (sv *SliceValidationRule) IsSet(field string, subject interface{}) (bool, error) {

	ps, err := sv.extractReflectValue(field, subject)

	if err != nil {
		return false, err
	}

	if ps == nil {
		return false, nil
	}

	return true, nil
}

// Validate implements ValidationRule.Validate
func (sv *SliceValidationRule) Validate(vc *ValidationContext) (result *ValidationResult, unexpected error) {

	f := sv.field

	if vc.OverrideField != "" {
		f = vc.OverrideField
	}

	sub := vc.Subject

	r := NewValidationResult()
	set, err := sv.IsSet(f, sub)

	if err != nil {
		return nil, err

	} else if !set {
		r.Unset = true

		if sv.required {
			r.AddForField(f, []string{sv.missingRequiredCode})
		}

		return r, nil
	}

	//Ignoring error as called previously during IsSet
	value, _ := sv.extractReflectValue(f, sub)

	err = sv.runOperations(f, value.(reflect.Value), vc, r)

	return r, err
}

func (sv *SliceValidationRule) runOperations(field string, v reflect.Value, vc *ValidationContext, r *ValidationResult) error {

	ec := types.NewEmptyOrderedStringSet()

	var err error

	for _, op := range sv.operations {

		switch op.OpType {
		case sliceOpMex:
			checkMExFields(op.MExFields, vc, ec, op.ErrCode)
		case sliceOpLen:
			if !sv.lengthOkay(v) {
				ec.Add(op.ErrCode)
			}
		case sliceOpElem:

			err = sv.checkElementContents(field, v, op.elemValidator, r, vc, op.ErrCode)
		}
	}

	r.AddForField(field, ec.Contents())

	return err

}

func (sv *SliceValidationRule) checkElementContents(field string, slice reflect.Value, v ValidationRule, r *ValidationResult, pvc *ValidationContext, overrideError string) error {

	useOverride := overrideError != sv.defaultErrorCode

	stringElement := false
	nilable := false

	sl := slice.Len()

	var err error

	for i := 0; i < sl; i++ {

		fa := fmt.Sprintf("%s[%d]", field, i)

		vc := new(ValidationContext)
		vc.OverrideField = fa
		vc.KnownSetFields = pvc.KnownSetFields
		vc.DirectSubject = true

		e := slice.Index(i)

		switch tv := v.(type) {
		case *StringValidationRule:
			vc.Subject, nilable, err = sv.stringValue(e, fa)
			stringElement = true
		case *IntValidationRule:
			vc.Subject, err = tv.toInt64(fa, e.Interface())
		case *FloatValidationRule:
			vc.Subject, err = tv.toFloat64(fa, e.Interface())
		case *BoolValidationRule:
			vc.Subject, err = sv.boolValue(e, fa)
		}

		if err != nil {
			return err
		}

		vr, err := v.Validate(vc)

		if err != nil {
			return err
		}

		ee := vr.ErrorCodes[fa]

		if useOverride && len(ee) > 0 {
			ee = []string{overrideError}
		}

		r.AddForField(fa, ee)

		if stringElement {
			sv.overwriteStringValue(e, vc.Subject.(*types.NilableString), nilable)
		}

	}

	return nil
}

// String validation is unique in that it can modify the value under consideration
func (sv *SliceValidationRule) overwriteStringValue(v reflect.Value, ns *types.NilableString, wasNilable bool) {

	if !wasNilable {

		v.Set(reflect.ValueOf(ns.String()))
	}

}

func (sv *SliceValidationRule) stringValue(v reflect.Value, fa string) (*types.NilableString, bool, error) {

	s := v.Interface()

	switch s := s.(type) {
	case *types.NilableString:
		return s, true, nil
	case string:
		return types.NewNilableString(s), false, nil
	default:
		m := fmt.Sprintf("%s is not a string or *NilableString", fa)
		return nil, false, errors.New(m)
	}

}

func (sv *SliceValidationRule) boolValue(v reflect.Value, fa string) (*types.NilableBool, error) {

	b := v.Interface()

	switch b := b.(type) {
	case *types.NilableBool:
		return b, nil
	case bool:
		return types.NewNilableBool(b), nil
	default:
		m := fmt.Sprintf("%s is not a bool or *NilableBool", fa)
		return nil, errors.New(m)
	}

}

func (sv *SliceValidationRule) extractReflectValue(f string, s interface{}) (interface{}, error) {

	v, err := rt.FindNestedField(rt.ExtractDotPath(f), s)

	if err != nil {
		return nil, err
	}

	if rt.NilPointer(v) {
		return nil, nil
	}

	if v.IsValid() && v.Kind() == reflect.Slice {

		if v.IsNil() {
			return nil, nil
		}

		return v, nil
	}

	m := fmt.Sprintf("%s is not a slice", f)

	return nil, errors.New(m)

}

// Length adds a check to see if the slice under consideration has an element count between the supplied min and max values.
func (sv *SliceValidationRule) Length(min, max int, code ...string) *SliceValidationRule {

	sv.minLen = min
	sv.maxLen = max

	ec := sv.chooseErrorCode(code)

	o := new(sliceOperation)
	o.OpType = sliceOpLen
	o.ErrCode = ec

	sv.addOperation(o)

	return sv

}

// StopAllOnFail implements ValidationRule.StopAllOnFail
func (sv *SliceValidationRule) StopAllOnFail() bool {
	return sv.stopAll
}

// CodesInUse implements ValidationRule.CodesInUse
func (sv *SliceValidationRule) CodesInUse() types.StringSet {
	return sv.codesInUse
}

// DependsOnFields implements ValidationRule.DependsOnFields
func (sv *SliceValidationRule) DependsOnFields() types.StringSet {

	return sv.dependsFields
}

// StopAll indicates that no further rules should be rule if this one fails.
func (sv *SliceValidationRule) StopAll() *SliceValidationRule {

	sv.stopAll = true

	return sv
}

// Required adds a check to see if the field under validation has been set.
func (sv *SliceValidationRule) Required(code ...string) *SliceValidationRule {

	sv.required = true
	sv.missingRequiredCode = sv.chooseErrorCode(code)

	return sv
}

// MEx adds a check to see if any other of the fields with which this field is mutually exclusive have been set.
func (sv *SliceValidationRule) MEx(fields types.StringSet, code ...string) *SliceValidationRule {
	op := new(sliceOperation)
	op.ErrCode = sv.chooseErrorCode(code)
	op.OpType = sliceOpMex
	op.MExFields = fields

	sv.addOperation(op)

	return sv
}

// Elem supplies a ValidationRule that can be used to checked the validity of the elements of the slice.
func (sv *SliceValidationRule) Elem(v ValidationRule, code ...string) *SliceValidationRule {

	op := new(sliceOperation)
	op.ErrCode = sv.chooseErrorCode(code)
	op.OpType = sliceOpElem
	op.elemValidator = v

	sv.addOperation(op)

	return sv
}

func (sv *SliceValidationRule) addOperation(o *sliceOperation) {
	sv.operations = append(sv.operations, o)
}

func (sv *SliceValidationRule) chooseErrorCode(v []string) string {

	if len(v) > 0 {
		sv.codesInUse.Add(v[0])
		return v[0]
	}

	return sv.defaultErrorCode
}

func (sv *SliceValidationRule) operation(c string) (sliceValidationOperation, error) {
	switch c {
	case sliceOpRequiredCode:
		return sliceOpRequired, nil
	case sliceOpStopAllCode:
		return sliceOpStopAll, nil
	case sliceOpMexCode:
		return sliceOpMex, nil
	case sliceOpLenCode:
		return sliceOpLen, nil
	case sliceOpElemCode:
		return sliceOpElem, nil
	}

	m := fmt.Sprintf("Unsupported slice validation operation %s", c)
	return sliceOpUnsupported, errors.New(m)

}

func (sv *SliceValidationRule) lengthOkay(r reflect.Value) bool {

	if sv.minLen == noBound && sv.maxLen == noBound {
		return true
	}

	sl := r.Len()

	minOkay := sv.minLen == noBound || sl >= sv.minLen
	maxOkay := sv.maxLen == noBound || sl <= sv.maxLen

	return minOkay && maxOkay

}

func newSliceValidationRuleBuilder(ec string, cf ioc.ComponentLookup, rv *RuleValidator) *sliceValidationRuleBuilder {
	bv := new(sliceValidationRuleBuilder)
	bv.componentFinder = cf
	bv.defaultErrorCode = ec
	bv.sliceLenRegex = regexp.MustCompile(lengthPattern)
	bv.ruleValidator = rv

	return bv
}

type sliceValidationRuleBuilder struct {
	defaultErrorCode string
	componentFinder  ioc.ComponentLookup
	sliceLenRegex    *regexp.Regexp
	ruleValidator    *RuleValidator
}

func (vb *sliceValidationRuleBuilder) parseRule(field string, rule []string) (ValidationRule, error) {

	defaultErrorcode := determineDefaultErrorCode(sliceRuleCode, rule, vb.defaultErrorCode)
	bv := NewSliceValidationRule(field, defaultErrorcode)

	for _, v := range rule {

		ops := decomposeOperation(v)
		opCode := ops[0]

		if isTypeIndicator(sliceRuleCode, opCode) {
			continue
		}

		op, err := bv.operation(opCode)

		if err != nil {
			return nil, err
		}

		switch op {
		case sliceOpRequired:
			err = vb.markRequired(field, ops, bv)
		case sliceOpStopAll:
			bv.StopAll()
		case sliceOpMex:
			err = vb.captureExclusiveFields(field, ops, bv)
		case sliceOpLen:
			err = vb.addLengthOperation(field, ops, bv)
		case sliceOpElem:
			err = vb.addElementValidationOperation(field, ops, v, bv)
		}

		if err != nil {

			return nil, err
		}

	}

	return bv, nil

}

func (vb *sliceValidationRuleBuilder) addElementValidationOperation(field string, ops []string, unparsedRule string, sv *SliceValidationRule) error {

	_, err := paramCount(ops, "Elem", field, 2, 3)

	if err != nil {
		return err
	}

	rv := vb.ruleValidator
	rule, err := rv.findRule(field, unparsedRule)

	if err != nil {
		return err
	}

	v, err := rv.parseRule(field, rule)

	if err != nil {
		return err
	}

	sv.codesInUse.AddAll(v.CodesInUse())

	switch v.(type) {
	case *StringValidationRule, *BoolValidationRule, *IntValidationRule, *FloatValidationRule:
		break
	default:
		m := fmt.Sprintf("Only %s, %s, %s and %s rules may be used to validate slice elements. Field %s is trying to use %s",
			intRuleCode, floatRuleCode, boolRuleCode, stringRuleCode, field, rule[0])
		return errors.New(m)
	}

	sv.Elem(v, extractVargs(ops, 3)...)

	return nil
}

func (vb *sliceValidationRuleBuilder) addLengthOperation(field string, ops []string, sv *SliceValidationRule) error {

	_, err := paramCount(ops, "Length", field, 2, 3)

	if err != nil {
		return err
	}

	min, max, err := extractLengthParams(field, ops[1], vb.sliceLenRegex)

	if err != nil {
		return err
	}

	sv.Length(min, max, extractVargs(ops, 3)...)

	return nil

}

func (vb *sliceValidationRuleBuilder) captureExclusiveFields(field string, ops []string, bv *SliceValidationRule) error {
	_, err := paramCount(ops, "MEX", field, 2, 3)

	if err != nil {
		return err
	}

	members := strings.SplitN(ops[1], setMemberSep, -1)
	fields := types.NewOrderedStringSet(members)

	bv.MEx(fields, extractVargs(ops, 3)...)

	return nil

}

func (vb *sliceValidationRuleBuilder) markRequired(field string, ops []string, bv *SliceValidationRule) error {

	_, err := paramCount(ops, "Required", field, 1, 2)

	if err != nil {
		return err
	}

	bv.Required(extractVargs(ops, 2)...)

	return nil
}
