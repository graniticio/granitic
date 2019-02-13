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
	"strings"
)

const objectRuleCode = "OBJ"

const (
	objOpRequiredCode = commonOpRequired
	objOpStopAllCode  = commonOpStopAll
	objOpMExCode      = commonOpMex
)

type objectValidationOperation uint

const (
	objOpUnsupported = iota
	objOpRequired
	objOpStopAll
	objOpMEx
)

// NewObjectValidationRule creates a new ObjectValidationRule to check the specified field.
func NewObjectValidationRule(field, defaultErrorCode string) *ObjectValidationRule {
	ov := new(ObjectValidationRule)
	ov.defaultErrorCode = defaultErrorCode
	ov.field = field
	ov.codesInUse = types.NewOrderedStringSet([]string{})
	ov.dependsFields = determinePathFields(field)
	ov.operations = make([]*objectOperation, 0)
	ov.codesInUse.Add(ov.defaultErrorCode)

	return ov
}

// ObjectValidationRule is a ValidationRule for checking a struct or map field on an object. See the method definitions on this type for
// the supported operations.
type ObjectValidationRule struct {
	stopAll             bool
	codesInUse          types.StringSet
	dependsFields       types.StringSet
	defaultErrorCode    string
	field               string
	missingRequiredCode string
	required            bool
	operations          []*objectOperation
}

type objectOperation struct {
	OpType    objectValidationOperation
	ErrCode   string
	MExFields types.StringSet
}

// IsSet returns true if the field to be validated is a non-nil struct or map
func (ov *ObjectValidationRule) IsSet(field string, subject interface{}) (bool, error) {
	fv, err := rt.FindNestedField(rt.ExtractDotPath(field), subject)

	if err != nil {
		return false, err
	}

	if rt.NilPointer(fv) || rt.NilMap(fv) {
		return false, nil
	}

	k := fv.Kind()

	if k == reflect.Invalid {
		m := fmt.Sprintf("Field %s is not a valid type. Does the field exist?", field)
		return false, errors.New(m)
	}

	if !rt.IsPointerToStruct(fv.Interface()) && k != reflect.Map && k != reflect.Struct {
		m := fmt.Sprintf("Field %s is not a pointer to a struct, a struct or a map.", field)
		return false, errors.New(m)
	}

	return true, nil

}

// Validate implements ValidationRule.Validate
func (ov *ObjectValidationRule) Validate(vc *ValidationContext) (result *ValidationResult, unexpected error) {

	f := ov.field

	if vc.OverrideField != "" {
		f = vc.OverrideField
	}

	sub := vc.Subject
	r := NewValidationResult()

	set, err := ov.IsSet(f, sub)

	if err != nil {
		return nil, err

	} else if !set {

		r.Unset = true

		if ov.required {
			r.AddForField(f, []string{ov.missingRequiredCode})
		}

	}

	err = ov.runOperations(f, vc, r)

	return r, err
}

func (ov *ObjectValidationRule) runOperations(field string, vc *ValidationContext, r *ValidationResult) error {

	ec := types.NewEmptyOrderedStringSet()

	for _, op := range ov.operations {

		switch op.OpType {
		case objOpMEx:
			checkMExFields(op.MExFields, vc, ec, op.ErrCode)
		}
	}

	r.AddForField(field, ec.Contents())

	return nil

}

// StopAllOnFail implements ValidationRule.StopAllOnFail
func (ov *ObjectValidationRule) StopAllOnFail() bool {
	return ov.stopAll
}

// CodesInUse implements ValidationRule.CodesInUse
func (ov *ObjectValidationRule) CodesInUse() types.StringSet {
	return ov.codesInUse
}

// DependsOnFields implements ValidationRule.DependsOnFields
func (ov *ObjectValidationRule) DependsOnFields() types.StringSet {

	return ov.dependsFields
}

// StopAll indicates that no further rules should be rule if this one fails.
func (ov *ObjectValidationRule) StopAll() *ObjectValidationRule {

	ov.stopAll = true

	return ov
}

// Required adds a check to see if the field under validation has been set.
func (ov *ObjectValidationRule) Required(code ...string) *ObjectValidationRule {

	ov.required = true

	if code != nil {
		ov.missingRequiredCode = code[0]
	} else {
		ov.missingRequiredCode = ov.defaultErrorCode
	}

	ov.codesInUse.Add(ov.missingRequiredCode)

	return ov
}

// MEx adds a check to see if any other of the fields with which this field is mutually exclusive have been set.
func (ov *ObjectValidationRule) MEx(fields types.StringSet, code ...string) *ObjectValidationRule {
	op := new(objectOperation)
	op.ErrCode = ov.chooseErrorCode(code)
	op.OpType = objOpMEx
	op.MExFields = fields

	ov.addOperation(op)

	return ov
}

func (ov *ObjectValidationRule) chooseErrorCode(v []string) string {

	if len(v) > 0 {
		ov.codesInUse.Add(v[0])
		return v[0]
	}

	return ov.defaultErrorCode
}

func (ov *ObjectValidationRule) addOperation(o *objectOperation) {
	ov.operations = append(ov.operations, o)
}

func (ov *ObjectValidationRule) operation(c string) (objectValidationOperation, error) {
	switch c {
	case objOpRequiredCode:
		return objOpRequired, nil
	case objOpStopAllCode:
		return objOpStopAll, nil
	case objOpMExCode:
		return objOpMEx, nil
	}

	m := fmt.Sprintf("Unsupported object validation operation %s", c)
	return objOpUnsupported, errors.New(m)

}

func newObjectValidationRuleBuilder(ec string, cf ioc.ComponentLookup) *objectValidationRuleBuilder {
	ov := new(objectValidationRuleBuilder)
	ov.componentFinder = cf
	ov.defaultErrorCode = ec

	return ov
}

type objectValidationRuleBuilder struct {
	defaultErrorCode string
	componentFinder  ioc.ComponentLookup
}

func (vb *objectValidationRuleBuilder) parseRule(field string, rule []string) (ValidationRule, error) {

	defaultErrorcode := determineDefaultErrorCode(objectRuleCode, rule, vb.defaultErrorCode)
	ov := NewObjectValidationRule(field, defaultErrorcode)

	for _, v := range rule {

		ops := decomposeOperation(v)
		opCode := ops[0]

		if isTypeIndicator(objectRuleCode, opCode) {
			continue
		}

		op, err := ov.operation(opCode)

		if err != nil {
			return nil, err
		}

		switch op {
		case objOpRequired:
			err = vb.markRequired(field, ops, ov)
		case objOpStopAll:
			ov.StopAll()
		case objOpMEx:
			err = vb.captureExclusiveFields(field, ops, ov)
		}

		if err != nil {

			return nil, err
		}

	}

	return ov, nil

}

func (vb *objectValidationRuleBuilder) captureExclusiveFields(field string, ops []string, bv *ObjectValidationRule) error {
	_, err := paramCount(ops, "MEX", field, 2, 3)

	if err != nil {
		return err
	}

	members := strings.SplitN(ops[1], setMemberSep, -1)
	fields := types.NewOrderedStringSet(members)

	bv.MEx(fields, extractVargs(ops, 3)...)

	return nil

}

func (vb *objectValidationRuleBuilder) markRequired(field string, ops []string, ov *ObjectValidationRule) error {

	pCount, err := paramCount(ops, "Required", field, 1, 2)

	if err != nil {
		return err
	}

	if pCount == 1 {
		ov.Required()
	} else {
		ov.Required(ops[1])
	}

	return nil
}
