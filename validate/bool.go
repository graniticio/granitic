// Copyright 2016 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package validate

import (
	"errors"
	"fmt"
	"github.com/graniticio/granitic/ioc"
	rt "github.com/graniticio/granitic/reflecttools"
	"github.com/graniticio/granitic/types"
	"strconv"
	"strings"
)

const boolRuleCode = "BOOL"

const (
	boolOpRequiredCode = commonOpRequired
	boolOpStopAllCode  = commonOpStopAll
	boolOpIsCode       = "IS"
	boolOpMexCode      = commonOpMex
)

type boolValidationOperation uint

const (
	boolOpUnsupported = iota
	boolOpRequired
	boolOpStopAll
	boolOpIs
	boolOpMex
)

// NewBoolValidationRule creates a new BoolValidationRule to check the named field and the supplied default error code.
func NewBoolValidationRule(field, defaultErrorCode string) *BoolValidationRule {
	bv := new(BoolValidationRule)
	bv.defaultErrorCode = defaultErrorCode
	bv.field = field
	bv.codesInUse = types.NewOrderedStringSet([]string{})
	bv.dependsFields = determinePathFields(field)
	bv.operations = make([]*boolOperation, 0)
	bv.codesInUse.Add(bv.defaultErrorCode)

	return bv
}

// A ValidationRule for checking a bool or NilableBool field on an object. See the method definitions on this type for
// the supported operations.
type BoolValidationRule struct {
	stopAll             bool
	codesInUse          types.StringSet
	dependsFields       types.StringSet
	defaultErrorCode    string
	field               string
	missingRequiredCode string
	required            bool
	requiredValue       *types.NilableBool
	requiredValueCode   string
	operations          []*boolOperation
}

type boolOperation struct {
	OpType    boolValidationOperation
	ErrCode   string
	MExFields types.StringSet
}

// IsSet returns true if the field to be validation is a bool or is a NilableBool which has been explicitly set.
func (bv *BoolValidationRule) IsSet(field string, subject interface{}) (bool, error) {

	value, err := bv.extractValue(field, subject)

	if err != nil {
		return false, err
	}

	if value == nil || !value.IsSet() {
		return false, nil
	} else {
		return true, nil
	}
}

// See ValidationRule.Validate
func (bv *BoolValidationRule) Validate(vc *ValidationContext) (result *ValidationResult, unexpected error) {

	f := bv.field

	if vc.OverrideField != "" {
		f = vc.OverrideField
	}

	var value *types.NilableBool

	sub := vc.Subject
	r := NewValidationResult()

	if vc.DirectSubject {

		b, found := sub.(*types.NilableBool)

		if !found {
			m := fmt.Sprintf("Direct validation requested for %s but supplied value is not a *NilableBool", f)
			return nil, errors.New(m)
		}

		value = b

	} else {

		set, err := bv.IsSet(f, sub)

		if err != nil {
			return nil, err

		} else if !set {
			r.Unset = true

			if bv.required {
				r.AddForField(f, []string{bv.missingRequiredCode})
			}

			return r, nil
		}

		//Ignoring error as called previously during IsSet
		value, _ = bv.extractValue(f, sub)
	}

	if bv.requiredValue != nil && value.Bool() != bv.requiredValue.Bool() {

		r.AddForField(f, []string{bv.requiredValueCode})
	}

	err := bv.runOperations(f, value.Bool(), vc, r)

	return r, err
}

func (bv *BoolValidationRule) runOperations(field string, b bool, vc *ValidationContext, r *ValidationResult) error {

	ec := types.NewEmptyOrderedStringSet()

	for _, op := range bv.operations {

		switch op.OpType {
		case boolOpMex:
			checkMExFields(op.MExFields, vc, ec, op.ErrCode)
		}
	}

	r.AddForField(field, ec.Contents())

	return nil

}

func (bv *BoolValidationRule) extractValue(f string, s interface{}) (*types.NilableBool, error) {

	v, err := rt.FindNestedField(rt.ExtractDotPath(f), s)

	if err != nil {
		return nil, err
	}

	if rt.NilPointer(v) {
		return nil, nil
	}

	b, found := v.Interface().(bool)

	if found {
		return types.NewNilableBool(b), nil
	}

	nsb, found := v.Interface().(*types.NilableBool)

	if found {
		return nsb, nil
	}

	m := fmt.Sprintf("%s is not a bool or *NilableBool.", f)

	return nil, errors.New(m)

}

func (bv *BoolValidationRule) StopAllOnFail() bool {
	return bv.stopAll
}

// See ValidationRule.CodesInUse
func (bv *BoolValidationRule) CodesInUse() types.StringSet {
	return bv.codesInUse
}

// See ValidationRule.DependsOnFields
func (bv *BoolValidationRule) DependsOnFields() types.StringSet {

	return bv.dependsFields
}

// StopAll adds a check to halt validation of this rule and all other rules if
// the previous check failed.
func (bv *BoolValidationRule) StopAll() *BoolValidationRule {

	bv.stopAll = true

	return bv
}

// Required adds a check see if the field under validation has been set.
func (bv *BoolValidationRule) Required(code ...string) *BoolValidationRule {

	bv.required = true
	bv.missingRequiredCode = bv.chooseErrorCode(code)

	return bv
}

// Is adds a check to see if the field is set to the supplied value.
func (bv *BoolValidationRule) Is(v bool, code ...string) *BoolValidationRule {

	bv.requiredValue = types.NewNilableBool(v)

	bv.requiredValueCode = bv.chooseErrorCode(code)

	return bv
}

// MEx adds a check to see if any other of the fields with which this field is mutually exclusive have been set.
func (bv *BoolValidationRule) MEx(fields types.StringSet, code ...string) *BoolValidationRule {
	op := new(boolOperation)
	op.ErrCode = bv.chooseErrorCode(code)
	op.OpType = boolOpMex
	op.MExFields = fields

	bv.addOperation(op)

	return bv
}

func (bv *BoolValidationRule) addOperation(o *boolOperation) {
	bv.operations = append(bv.operations, o)
}

func (bv *BoolValidationRule) chooseErrorCode(v []string) string {

	if len(v) > 0 {
		bv.codesInUse.Add(v[0])
		return v[0]
	} else {
		return bv.defaultErrorCode
	}

}

func (bv *BoolValidationRule) operation(c string) (boolValidationOperation, error) {
	switch c {
	case boolOpRequiredCode:
		return boolOpRequired, nil
	case boolOpStopAllCode:
		return boolOpStopAll, nil
	case boolOpIsCode:
		return boolOpIs, nil
	case boolOpMexCode:
		return boolOpMex, nil
	}

	m := fmt.Sprintf("Unsupported bool validation operation %s", c)
	return boolOpUnsupported, errors.New(m)

}

func newBoolValidationRuleBuilder(ec string, cf ioc.ComponentByNameFinder) *boolValidationRuleBuilder {
	bv := new(boolValidationRuleBuilder)
	bv.componentFinder = cf
	bv.defaultErrorCode = ec

	return bv
}

type boolValidationRuleBuilder struct {
	defaultErrorCode string
	componentFinder  ioc.ComponentByNameFinder
}

func (vb *boolValidationRuleBuilder) parseRule(field string, rule []string) (ValidationRule, error) {

	defaultErrorcode := determineDefaultErrorCode(boolRuleCode, rule, vb.defaultErrorCode)
	bv := NewBoolValidationRule(field, defaultErrorcode)

	for _, v := range rule {

		ops := decomposeOperation(v)
		opCode := ops[0]

		if isTypeIndicator(boolRuleCode, opCode) {
			continue
		}

		op, err := bv.operation(opCode)

		if err != nil {
			return nil, err
		}

		switch op {
		case boolOpRequired:
			err = vb.markRequired(field, ops, bv)
		case boolOpStopAll:
			bv.StopAll()
		case boolOpIs:
			err = vb.captureRequiredValue(field, ops, bv)
		case boolOpMex:
			err = vb.captureExclusiveFields(field, ops, bv)
		}

		if err != nil {

			return nil, err
		}

	}

	return bv, nil

}

func (vb *boolValidationRuleBuilder) captureExclusiveFields(field string, ops []string, bv *BoolValidationRule) error {
	_, err := paramCount(ops, "MEX", field, 2, 3)

	if err != nil {
		return err
	}

	members := strings.SplitN(ops[1], setMemberSep, -1)
	fields := types.NewOrderedStringSet(members)

	bv.MEx(fields, extractVargs(ops, 3)...)

	return nil

}

func (vb *boolValidationRuleBuilder) captureRequiredValue(field string, ops []string, bv *BoolValidationRule) error {
	_, err := paramCount(ops, "Is", field, 2, 3)

	if err != nil {
		return err
	}

	b, err := strconv.ParseBool(ops[1])

	if err != nil {
		m := fmt.Sprintf("Value %s prbvided as part of a BOOL/IS operation could not be interpreted as a bool\n", ops[1])
		return errors.New(m)
	}

	bv.Is(b, extractVargs(ops, 3)...)

	return nil
}

func (vb *boolValidationRuleBuilder) markRequired(field string, ops []string, bv *BoolValidationRule) error {

	_, err := paramCount(ops, "Required", field, 1, 2)

	if err != nil {
		return err
	}

	bv.Required(extractVargs(ops, 2)...)

	return nil
}
