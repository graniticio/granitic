// Copyright 2016-2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package validate

import (
	"errors"
	"fmt"
	"github.com/graniticio/granitic/v2/ioc"
	rt "github.com/graniticio/granitic/v2/reflecttools"
	"github.com/graniticio/granitic/v2/types"
	"regexp"
	"strconv"
	"strings"
)

// An object able to evaulate the supplied float64 to see if it meets some definition of validity.
type ExternalFloat64Validator interface {
	// ValidFloat64 returns true if the object considers the supplied float64 to be valid.
	ValidFloat64(float64) (bool, error)
}

const floatRuleCode = "FLOAT"

const (
	floatOpIsRequiredCode = commonOpRequired
	floatOpIsStopAllCode  = commonOpStopAll
	floatOpInCode         = commonOpIn
	floatOpBreakCode      = commonOpBreak
	floatOpExtCode        = commonOpExt
	floatOpRangeCode      = "RANGE"
	floatOpMExCode        = commonOpMex
)

type floatValidationOperation uint

const (
	floatOpUnsupported = iota
	floatOpRequired
	floatOpStopAll
	floatOpIn
	floatOpBreak
	floatOpExt
	floatOpRange
	floatOpMex
)

// NewFloatValidationRule creates a new FloatValidationRule to check the named field and the supplied default error code.
func NewFloatValidationRule(field, defaultErrorCode string) *FloatValidationRule {
	fv := new(FloatValidationRule)
	fv.defaultErrorCode = defaultErrorCode
	fv.field = field
	fv.codesInUse = types.NewOrderedStringSet([]string{})
	fv.dependsFields = determinePathFields(field)
	fv.operations = make([]*floatOperation, 0)

	fv.codesInUse.Add(fv.defaultErrorCode)

	return fv
}

// A ValidationRule for checking a float32, float64 or NilableFloat64 field on an object. See the method definitions on this type for
// the supported operations. Note that float32 are converted to float64 before validation.
type FloatValidationRule struct {
	stopAll             bool
	codesInUse          types.StringSet
	dependsFields       types.StringSet
	defaultErrorCode    string
	field               string
	missingRequiredCode string
	required            bool
	operations          []*floatOperation
	checkMin            bool
	checkMax            bool
	minAllowed          float64
	maxAllowed          float64
}

type floatOperation struct {
	OpType    floatValidationOperation
	ErrCode   string
	InSet     map[float64]bool
	External  ExternalFloat64Validator
	MExFields types.StringSet
}

// IsSet returns true if the field to be validated is a float32, float64 or a NilableFloat64 whose value has been explicitly set.
func (fv *FloatValidationRule) IsSet(field string, subject interface{}) (bool, error) {

	nf, err := fv.extractValue(field, subject)

	if err != nil {
		return false, err
	}

	if nf == nil || !nf.IsSet() {
		return false, nil
	}

	return true, nil
}

// See ValidationRule.Validate
func (fv *FloatValidationRule) Validate(vc *ValidationContext) (result *ValidationResult, unexpected error) {

	f := fv.field

	if vc.OverrideField != "" {
		f = vc.OverrideField
	}

	var value *types.NilableFloat64
	sub := vc.Subject
	r := NewValidationResult()

	if vc.DirectSubject {

		i, found := sub.(*types.NilableFloat64)

		if !found {
			m := fmt.Sprintf("Direct validation requested for %s but supplied value is not a *NilableFloat64", f)
			return nil, errors.New(m)
		}

		value = i

	} else {

		set, err := fv.IsSet(f, sub)

		if err != nil {
			return nil, err
		} else if !set {
			r.Unset = true

			if fv.required {
				r.AddForField(f, []string{fv.missingRequiredCode})
			}

			return r, nil
		}

		//Ignoring error as called previously during IsSet
		value, _ = fv.extractValue(f, sub)
	}

	err := fv.runOperations(f, value.Float64(), vc, r)

	return r, err
}

func (fv *FloatValidationRule) extractValue(f string, s interface{}) (*types.NilableFloat64, error) {

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

	return fv.toFloat64(f, v.Interface())
}

func (fv *FloatValidationRule) toFloat64(f string, i interface{}) (*types.NilableFloat64, error) {
	var ex float64

	switch i := i.(type) {
	case *types.NilableFloat64:
		return i, nil
	case float32:
		sc := strconv.FormatFloat(float64(i), 'f', -1, 32)
		ex, _ = strconv.ParseFloat(sc, 64)
	case float64:
		ex = i
	default:
		m := fmt.Sprintf("%s is type %T, not a float32, float64 or *NilableFloat.", f, i)
		return nil, errors.New(m)

	}

	return types.NewNilableFloat64(ex), nil
}

func (fv *FloatValidationRule) runOperations(field string, i float64, vc *ValidationContext, r *ValidationResult) error {

	ec := types.NewEmptyOrderedStringSet()

OpLoop:
	for _, op := range fv.operations {

		switch op.OpType {
		case floatOpIn:
			if !fv.checkIn(i, op) {
				ec.Add(op.ErrCode)
			}

		case floatOpBreak:
			if ec.Size() > 0 {
				break OpLoop
			}

		case floatOpExt:
			if valid, err := op.External.ValidFloat64(i); err == nil && !valid {
				ec.Add(op.ErrCode)
			} else if err != nil {
				return err
			}

		case floatOpRange:
			if !fv.inRange(i, op) {
				ec.Add(op.ErrCode)
			}
		case floatOpMex:
			checkMExFields(op.MExFields, vc, ec, op.ErrCode)
		}

	}

	r.AddForField(field, ec.Contents())

	return nil

}

func (fv *FloatValidationRule) inRange(i float64, o *floatOperation) bool {

	moreThanMin := true
	lessThanMax := true

	if fv.checkMin {
		moreThanMin = i >= fv.minAllowed
	}

	if fv.checkMax {
		lessThanMax = i <= fv.maxAllowed
	}

	return moreThanMin && lessThanMax
}

func (fv *FloatValidationRule) checkIn(i float64, o *floatOperation) bool {
	return o.InSet[i]
}

// MEx adds a check to see if any other of the fields with which this field is mutually exclusive have been set.
func (fv *FloatValidationRule) MEx(fields types.StringSet, code ...string) *FloatValidationRule {
	op := new(floatOperation)
	op.ErrCode = fv.chooseErrorCode(code)
	op.OpType = floatOpMex
	op.MExFields = fields

	fv.addOperation(op)

	return fv
}

// Break adds a check to stop processing this rule if the previous check has failed.
func (fv *FloatValidationRule) Break() *FloatValidationRule {

	o := new(floatOperation)
	o.OpType = floatOpBreak

	fv.addOperation(o)

	return fv

}

func (fv *FloatValidationRule) addOperation(o *floatOperation) {
	fv.operations = append(fv.operations, o)
	fv.codesInUse.Add(o.ErrCode)
}

// See ValidationRule.StopAllOnFail
func (fv *FloatValidationRule) StopAllOnFail() bool {
	return fv.stopAll
}

// See ValidationRule.CodesInUse
func (fv *FloatValidationRule) CodesInUse() types.StringSet {
	return fv.codesInUse
}

// See ValidationRule.DependsOnFields
func (fv *FloatValidationRule) DependsOnFields() types.StringSet {

	return fv.dependsFields
}

// StopAll indicates that no further rules should be rule if this one fails.
func (fv *FloatValidationRule) StopAll() *FloatValidationRule {

	fv.stopAll = true

	return fv
}

// Required adds a check to see if the field under validation has been set.
func (fv *FloatValidationRule) Required(code ...string) *FloatValidationRule {

	fv.required = true
	fv.missingRequiredCode = fv.chooseErrorCode(code)

	return fv
}

// Range adds a check to see if the float under validation is in the supplied range. checkMin/Max are set to false if no
// minimum or maximum bound is in effect.
func (fv *FloatValidationRule) Range(checkMin, checkMax bool, min, max float64, code ...string) *FloatValidationRule {

	fv.checkMin = checkMin
	fv.checkMax = checkMax
	fv.minAllowed = min
	fv.maxAllowed = max

	ec := fv.chooseErrorCode(code)

	o := new(floatOperation)
	o.OpType = floatOpRange
	o.ErrCode = ec

	fv.addOperation(o)

	return fv
}

// In adds a check to see if the float under validation is exactly equal to one of the float values specified.
func (fv *FloatValidationRule) In(set []float64, code ...string) *FloatValidationRule {

	ec := fv.chooseErrorCode(code)

	o := new(floatOperation)
	o.OpType = floatOpIn
	o.ErrCode = ec

	fm := make(map[float64]bool)

	for _, m := range set {
		fm[m] = true
	}

	o.InSet = fm

	fv.addOperation(o)

	return fv

}

// ExternalValidation adds a check to call the supplied object to ask it to check the validity of the float in question.
func (fv *FloatValidationRule) ExternalValidation(v ExternalFloat64Validator, code ...string) *FloatValidationRule {
	ec := fv.chooseErrorCode(code)

	o := new(floatOperation)
	o.OpType = floatOpExt
	o.ErrCode = ec
	o.External = v

	fv.addOperation(o)

	return fv
}

func (fv *FloatValidationRule) chooseErrorCode(v []string) string {

	if len(v) > 0 {
		fv.codesInUse.Add(v[0])
		return v[0]
	}

	return fv.defaultErrorCode
}

func (fv *FloatValidationRule) operation(c string) (boolValidationOperation, error) {
	switch c {
	case floatOpIsRequiredCode:
		return floatOpRequired, nil
	case floatOpIsStopAllCode:
		return floatOpStopAll, nil
	case floatOpInCode:
		return floatOpIn, nil
	case floatOpBreakCode:
		return floatOpBreak, nil
	case floatOpExtCode:
		return floatOpExt, nil
	case floatOpRangeCode:
		return floatOpRange, nil
	case floatOpMExCode:
		return floatOpMex, nil
	}

	m := fmt.Sprintf("Unsupported int validation operation %s", c)
	return floatOpUnsupported, errors.New(m)

}

func newFloatValidationRuleBuilder(ec string, cf ioc.ComponentByNameFinder) *floatValidationRuleBuilder {
	fv := new(floatValidationRuleBuilder)
	fv.componentFinder = cf
	fv.defaultErrorCode = ec
	fv.rangeRegex = regexp.MustCompile("^(.*)\\|(.*)$")
	return fv
}

type floatValidationRuleBuilder struct {
	defaultErrorCode string
	componentFinder  ioc.ComponentByNameFinder
	rangeRegex       *regexp.Regexp
}

func (vb *floatValidationRuleBuilder) parseRule(field string, rule []string) (ValidationRule, error) {

	defaultErrorcode := determineDefaultErrorCode(floatRuleCode, rule, vb.defaultErrorCode)
	bv := NewFloatValidationRule(field, defaultErrorcode)

	for _, v := range rule {

		ops := decomposeOperation(v)
		opCode := ops[0]

		if isTypeIndicator(floatRuleCode, opCode) {
			continue
		}

		op, err := bv.operation(opCode)

		if err != nil {
			return nil, err
		}

		switch op {
		case floatOpRequired:
			err = vb.markRequired(field, ops, bv)
		case floatOpIn:
			err = vb.addFloatInOperation(field, ops, bv)
		case floatOpStopAll:
			bv.StopAll()
		case floatOpBreak:
			bv.Break()
		case floatOpExt:
			err = vb.addFloatExternalOperation(field, ops, bv)
		case floatOpRange:
			err = vb.addFloatRangeOperation(field, ops, bv)
		case floatOpMex:
			err = vb.captureExclusiveFields(field, ops, bv)
		}

		if err != nil {

			return nil, err
		}

	}

	return bv, nil

}

func (vb *floatValidationRuleBuilder) captureExclusiveFields(field string, ops []string, fv *FloatValidationRule) error {
	_, err := paramCount(ops, "MEX", field, 2, 3)

	if err != nil {
		return err
	}

	members := strings.SplitN(ops[1], setMemberSep, -1)
	fields := types.NewOrderedStringSet(members)

	fv.MEx(fields, extractVargs(ops, 3)...)

	return nil

}

func (vb *floatValidationRuleBuilder) markRequired(field string, ops []string, fv *FloatValidationRule) error {

	pCount, err := paramCount(ops, "Required", field, 1, 2)

	if err != nil {
		return err
	}

	if pCount == 1 {
		fv.Required()
	} else {
		fv.Required(ops[1])
	}

	return nil
}

func (vb *floatValidationRuleBuilder) addFloatRangeOperation(field string, ops []string, fv *FloatValidationRule) error {

	pCount, err := paramCount(ops, "Range", field, 2, 3)

	if err != nil {
		return err
	}

	vals := ops[1]

	if !vb.rangeRegex.MatchString(vals) {
		m := fmt.Sprintf("Range parameters for field %s are invalid. Values provided: %s", field, vals)
		return errors.New(m)
	}

	var min float64
	var max float64

	checkMin := false
	checkMax := false

	groups := vb.rangeRegex.FindStringSubmatch(vals)

	if groups[1] != "" {
		min, err = strconv.ParseFloat(groups[1], 64)

		if err != nil {
			m := fmt.Sprintf("Range parameters for field %s are invalid (cannot parse min as float64). Values provided: %s", field, vals)
			return errors.New(m)
		}

		checkMin = true
	}

	if groups[2] != "" {
		max, err = strconv.ParseFloat(groups[2], 64)

		if err != nil {
			m := fmt.Sprintf("Range parameters for field %s are invalid (cannot parse max as float64). Values provided: %s", field, vals)
			return errors.New(m)
		}

		checkMax = true
	}

	if checkMin && checkMax && min > max {
		m := fmt.Sprintf("Range parameters for field %s are invalid (minimum greater than maximum). Values provided: %s", field, vals)
		return errors.New(m)
	}

	if pCount == 2 {
		fv.Range(checkMin, checkMax, float64(min), float64(max))
	} else {
		fv.Range(checkMin, checkMax, float64(min), float64(max), ops[2])
	}

	return nil
}

func (vb *floatValidationRuleBuilder) addFloatExternalOperation(field string, ops []string, fv *FloatValidationRule) error {

	pCount, i, err := validateExternalOperation(vb.componentFinder, field, ops)

	if err != nil {
		return err
	}

	ev, found := i.Instance.(ExternalFloat64Validator)

	if !found {
		m := fmt.Sprintf("Component %s to validate field %s does not implement ExternalFloat64Validator", i.Name, field)
		return errors.New(m)
	}

	if pCount == 2 {
		fv.ExternalValidation(ev)
	} else {
		fv.ExternalValidation(ev, ops[2])
	}

	return nil

}

func (vb *floatValidationRuleBuilder) addFloatInOperation(field string, ops []string, sv *FloatValidationRule) error {

	pCount, err := paramCount(ops, "In Set", field, 2, 3)

	if err != nil {
		return err
	}

	members := strings.SplitN(ops[1], setMemberSep, -1)
	floats := make([]float64, len(members))

	for i, m := range members {

		f, err := strconv.ParseFloat(m, 64)

		if err != nil {
			m := fmt.Sprintf("%s defined as a valid value when validating field %s cannot be parsed as a float64", m, field)
			return errors.New(m)
		}

		floats[i] = f

	}

	if pCount == 2 {
		sv.In(floats)
	} else {
		sv.In(floats, ops[2])
	}

	return nil

}
