// Copyright 2016 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package validate

import (
	"errors"
	"fmt"
	"github.com/graniticio/granitic/ioc"
	rt "github.com/graniticio/granitic/reflecttools"
	"github.com/graniticio/granitic/types"
	"regexp"
	"strconv"
	"strings"
)

// An object able to evaulate the supplied int64 to see if it meets some definition of validity.
type ExternalInt64Validator interface {
	// ValidInt64 returns true if the implementation considers the supplied int64 to be valid.
	ValidInt64(int64) bool
}

const intRuleCode = "INT"

const (
	intOpIsRequiredCode = commonOpRequired
	intOpIsStopAllCode  = commonOpStopAll
	intOpInCode         = commonOpIn
	intOpBreakCode      = commonOpBreak
	intOpExtCode        = commonOpExt
	intOpRangeCode      = "RANGE"
	intOpMExCode        = commonOpMex
)

type intValidationOperation uint

const (
	intOpUnsupported = iota
	intOpRequired
	intOpStopAll
	intOpIn
	intOpBreak
	intOpExt
	intOpRange
	untOpMEx
)

// NewIntValidationRule creates a new IntValidationRule to check the named field with the supplied default error code.
func NewIntValidationRule(field, defaultErrorCode string) *IntValidationRule {
	iv := new(IntValidationRule)
	iv.defaultErrorCode = defaultErrorCode
	iv.field = field
	iv.codesInUse = types.NewOrderedStringSet([]string{})
	iv.dependsFields = determinePathFields(field)
	iv.operations = make([]*intOperation, 0)

	iv.codesInUse.Add(iv.defaultErrorCode)

	return iv
}

// A ValidationRule for checking a native signed int type or NilableInt64 field on an object. See the method definitions on this type for
// the supported operations. Note that any native int types are converted to int64 before validation.
type IntValidationRule struct {
	stopAll             bool
	codesInUse          types.StringSet
	dependsFields       types.StringSet
	defaultErrorCode    string
	field               string
	missingRequiredCode string
	required            bool
	operations          []*intOperation
	checkMin            bool
	checkMax            bool
	minAllowed          int64
	maxAllowed          int64
}

type intOperation struct {
	OpType    intValidationOperation
	ErrCode   string
	InSet     types.StringSet
	External  ExternalInt64Validator
	MExFields types.StringSet
}

// IsSet returns true if the field to be validated is a native intxx type or a NilableInt64 whose value has been explicitly set.
func (iv *IntValidationRule) IsSet(field string, subject interface{}) (bool, error) {
	nf, err := iv.extractValue(field, subject)

	if err != nil {
		return false, err
	}

	if nf == nil || !nf.IsSet() {
		return false, nil
	} else {
		return true, nil
	}
}

// See ValidationRule.Validate
func (iv *IntValidationRule) Validate(vc *ValidationContext) (result *ValidationResult, unexpected error) {

	f := iv.field

	if vc.OverrideField != "" {
		f = vc.OverrideField
	}

	sub := vc.Subject

	r := NewValidationResult()

	var value *types.NilableInt64

	if vc.DirectSubject {

		i, found := sub.(*types.NilableInt64)

		if !found {
			m := fmt.Sprintf("Direct validation requested for %s but supplied value is not a *NilableInt64", f)
			return nil, errors.New(m)
		}

		value = i

	} else {

		set, err := iv.IsSet(f, sub)

		if err != nil {
			return nil, err
		} else if !set {
			r.Unset = true

			if iv.required {
				r.AddForField(f, []string{iv.missingRequiredCode})
			}
			return r, nil
		}

		//Ignoring error as called previously during IsSet
		value, _ = iv.extractValue(f, sub)
	}

	err := iv.runOperations(f, value.Int64(), vc, r)

	return r, err
}

func (iv *IntValidationRule) runOperations(field string, i int64, vc *ValidationContext, r *ValidationResult) error {

	ec := types.NewEmptyOrderedStringSet()

OpLoop:
	for _, op := range iv.operations {

		switch op.OpType {
		case intOpIn:
			if !iv.checkIn(i, op) {
				ec.Add(op.ErrCode)
			}

		case intOpBreak:
			if ec.Size() > 0 {
				break OpLoop
			}

		case intOpExt:
			if !op.External.ValidInt64(i) {
				ec.Add(op.ErrCode)
			}

		case intOpRange:
			if !iv.inRange(i, op) {
				ec.Add(op.ErrCode)
			}
		case untOpMEx:
			checkMExFields(op.MExFields, vc, ec, op.ErrCode)
		}

	}

	r.AddForField(field, ec.Contents())

	return nil

}

func (iv *IntValidationRule) inRange(i int64, o *intOperation) bool {

	moreThanMin := true
	lessThanMax := true

	if iv.checkMin {
		moreThanMin = i >= iv.minAllowed
	}

	if iv.checkMax {
		lessThanMax = i <= iv.maxAllowed
	}

	return moreThanMin && lessThanMax
}

func (iv *IntValidationRule) checkIn(i int64, o *intOperation) bool {
	s := strconv.FormatInt(i, 10)

	return o.InSet.Contains(s)
}

// MEx adds a check to see if any other of the fields with which this field is mutually exclusive have been set.
func (iv *IntValidationRule) MEx(fields types.StringSet, code ...string) *IntValidationRule {
	op := new(intOperation)
	op.ErrCode = iv.chooseErrorCode(code)
	op.OpType = untOpMEx
	op.MExFields = fields

	iv.addOperation(op)

	return iv
}

// Break adds a check to stop processing this rule if the previous check has failed.
func (iv *IntValidationRule) Break() *IntValidationRule {

	o := new(intOperation)
	o.OpType = intOpBreak

	iv.addOperation(o)

	return iv

}

func (iv *IntValidationRule) addOperation(o *intOperation) {
	iv.operations = append(iv.operations, o)
	iv.codesInUse.Add(o.ErrCode)
}

func (iv *IntValidationRule) extractValue(f string, s interface{}) (*types.NilableInt64, error) {

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

	return iv.toInt64(f, v.Interface())

}

func (iv *IntValidationRule) toInt64(f string, i interface{}) (*types.NilableInt64, error) {

	var ex int64

	switch i := i.(type) {
	case *types.NilableInt64:
		return i, nil
	case int:
		ex = int64(i)
	case int8:
		ex = int64(i)
	case int16:
		ex = int64(i)
	case int32:
		ex = int64(i)
	case int64:
		ex = i
	default:
		m := fmt.Sprintf("%s is type %T, not an int, int8, int16, int32, int64 or *NilableInt.", f, i)
		return nil, errors.New(m)

	}

	return types.NewNilableInt64(ex), nil

}

// See ValidationRule.StopAllOnFail
func (iv *IntValidationRule) StopAllOnFail() bool {
	return iv.stopAll
}

// See ValidationRule.CodesInUse
func (iv *IntValidationRule) CodesInUse() types.StringSet {
	return iv.codesInUse
}

// See ValidationRule.DependsOnFields
func (iv *IntValidationRule) DependsOnFields() types.StringSet {

	return iv.dependsFields
}

// StopAll indicates that no further rules should be rule if this one fails.
func (iv *IntValidationRule) StopAll() *IntValidationRule {

	iv.stopAll = true

	return iv
}

// Required adds a check to see if the field under validation has been set.
func (iv *IntValidationRule) Required(code ...string) *IntValidationRule {

	iv.required = true
	iv.missingRequiredCode = iv.chooseErrorCode(code)

	return iv
}

// Range adds a check to see if the float under validation is in the supplied range. checkMin/Max are set to false if no
// minimum or maximum bound is in effect.
func (iv *IntValidationRule) Range(checkMin, checkMax bool, min, max int64, code ...string) *IntValidationRule {

	iv.checkMin = checkMin
	iv.checkMax = checkMax
	iv.minAllowed = min
	iv.maxAllowed = max

	ec := iv.chooseErrorCode(code)

	o := new(intOperation)
	o.OpType = intOpRange
	o.ErrCode = ec

	iv.addOperation(o)

	return iv
}

// In adds a check to see if the float under validation is exactly equal to one of the int values specified.
func (iv *IntValidationRule) In(set []string, code ...string) *IntValidationRule {

	ss := types.NewUnorderedStringSet(set)

	ec := iv.chooseErrorCode(code)

	o := new(intOperation)
	o.OpType = intOpIn
	o.ErrCode = ec
	o.InSet = ss

	iv.addOperation(o)

	return iv

}

// ExternalValidation adds a check to call the supplied object to ask it to check the validity of the int in question.
func (iv *IntValidationRule) ExternalValidation(v ExternalInt64Validator, code ...string) *IntValidationRule {
	ec := iv.chooseErrorCode(code)

	o := new(intOperation)
	o.OpType = intOpExt
	o.ErrCode = ec
	o.External = v

	iv.addOperation(o)

	return iv
}

func (iv *IntValidationRule) chooseErrorCode(v []string) string {

	if len(v) > 0 {
		iv.codesInUse.Add(v[0])
		return v[0]
	} else {
		return iv.defaultErrorCode
	}

}

func (iv *IntValidationRule) operation(c string) (boolValidationOperation, error) {
	switch c {
	case intOpIsRequiredCode:
		return intOpRequired, nil
	case intOpIsStopAllCode:
		return intOpStopAll, nil
	case intOpInCode:
		return intOpIn, nil
	case intOpBreakCode:
		return intOpBreak, nil
	case intOpExtCode:
		return intOpExt, nil
	case intOpRangeCode:
		return intOpRange, nil
	case intOpMExCode:
		return untOpMEx, nil
	}

	m := fmt.Sprintf("Unsupported int validation operation %s", c)
	return intOpUnsupported, errors.New(m)

}

func newIntValidationRuleBuilder(ec string, cf ioc.ComponentByNameFinder) *intValidationRuleBuilder {
	iv := new(intValidationRuleBuilder)
	iv.componentFinder = cf
	iv.defaultErrorCode = ec
	iv.rangeRegex = regexp.MustCompile("^([-+]{0,1}\\d*)\\|([-+]{0,1}\\d*)$")
	return iv
}

type intValidationRuleBuilder struct {
	defaultErrorCode string
	componentFinder  ioc.ComponentByNameFinder
	rangeRegex       *regexp.Regexp
}

func (vb *intValidationRuleBuilder) parseRule(field string, rule []string) (ValidationRule, error) {

	defaultErrorcode := determineDefaultErrorCode(intRuleCode, rule, vb.defaultErrorCode)
	bv := NewIntValidationRule(field, defaultErrorcode)

	for _, v := range rule {

		ops := decomposeOperation(v)
		opCode := ops[0]

		if isTypeIndicator(intRuleCode, opCode) {
			continue
		}

		op, err := bv.operation(opCode)

		if err != nil {
			return nil, err
		}

		switch op {
		case intOpRequired:
			err = vb.markRequired(field, ops, bv)
		case intOpIn:
			err = vb.addIntInOperation(field, ops, bv)
		case intOpStopAll:
			bv.StopAll()
		case intOpBreak:
			bv.Break()
		case intOpExt:
			err = vb.addIntExternalOperation(field, ops, bv)
		case intOpRange:
			err = vb.addIntRangeOperation(field, ops, bv)
		case untOpMEx:
			err = vb.captureExclusiveFields(field, ops, bv)
		}

		if err != nil {

			return nil, err
		}

	}

	return bv, nil

}

func (vb *intValidationRuleBuilder) captureExclusiveFields(field string, ops []string, iv *IntValidationRule) error {
	_, err := paramCount(ops, "MEX", field, 2, 3)

	if err != nil {
		return err
	}

	members := strings.SplitN(ops[1], setMemberSep, -1)
	fields := types.NewOrderedStringSet(members)

	iv.MEx(fields, extractVargs(ops, 3)...)

	return nil

}

func (vb *intValidationRuleBuilder) markRequired(field string, ops []string, iv *IntValidationRule) error {

	_, err := paramCount(ops, "Required", field, 1, 2)

	if err != nil {
		return err
	}

	iv.Required(extractVargs(ops, 2)...)

	return nil
}

func (vb *intValidationRuleBuilder) addIntRangeOperation(field string, ops []string, iv *IntValidationRule) error {

	pCount, err := paramCount(ops, "Range", field, 2, 3)

	if err != nil {
		return err
	}

	vals := ops[1]

	if !vb.rangeRegex.MatchString(vals) {
		m := fmt.Sprintf("Range parameters for field %s are invalid. Values provided: %s", field, vals)
		return errors.New(m)
	}

	var min int
	var max int

	checkMin := false
	checkMax := false

	groups := vb.rangeRegex.FindStringSubmatch(vals)

	if groups[1] != "" {
		min, _ = strconv.Atoi(groups[1])
		checkMin = true
	}

	if groups[2] != "" {
		max, _ = strconv.Atoi(groups[2])
		checkMax = true
	}

	if checkMin && checkMax && min > max {
		m := fmt.Sprintf("Range parameters for field %s are invalid (min value greater than max). Values provided: %s", field, vals)
		return errors.New(m)
	}

	if pCount == 2 {
		iv.Range(checkMin, checkMax, int64(min), int64(max))
	} else {
		iv.Range(checkMin, checkMax, int64(min), int64(max), ops[2])
	}

	return nil
}

func (vb *intValidationRuleBuilder) addIntExternalOperation(field string, ops []string, iv *IntValidationRule) error {

	pCount, i, err := validateExternalOperation(vb.componentFinder, field, ops)

	if err != nil {
		return err
	}

	ev, found := i.Instance.(ExternalInt64Validator)

	if !found {
		m := fmt.Sprintf("Component %s to validate field %s does not implement ExternalInt64Validator", i.Name, field)
		return errors.New(m)
	}

	if pCount == 2 {
		iv.ExternalValidation(ev)
	} else {
		iv.ExternalValidation(ev, ops[2])
	}

	return nil

}

func (vb *intValidationRuleBuilder) addIntInOperation(field string, ops []string, sv *IntValidationRule) error {

	pCount, err := paramCount(ops, "In Set", field, 2, 3)

	if err != nil {
		return err
	}

	members := strings.SplitN(ops[1], setMemberSep, -1)

	for _, m := range members {

		_, err := strconv.ParseInt(m, 10, 64)

		if err != nil {
			m := fmt.Sprintf("%s defined as a valid value when validating field %s cannot be parsed as an int64", m, field)
			return errors.New(m)
		}

	}

	if pCount == 2 {
		sv.In(members)
	} else {
		sv.In(members, ops[2])
	}

	return nil

}
