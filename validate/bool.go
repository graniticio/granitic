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

const BoolRuleCode = "BOOL"

const (
	boolOpRequiredCode = commonOpRequired
	boolOpStopAllCode  = commonOpStopAll
	boolOpIsCode       = "IS"
	boolOpMexCode      = commonOpMex
)

type boolValidationOperation uint

const (
	BoolOpUnsupported = iota
	BoolOpRequired
	BoolOpStopAll
	BoolOpIs
	BoolOpMex
)

func NewBoolValidator(field, defaultErrorCode string) *BoolValidator {
	bv := new(BoolValidator)
	bv.defaultErrorCode = defaultErrorCode
	bv.field = field
	bv.codesInUse = types.NewOrderedStringSet([]string{})
	bv.dependsFields = determinePathFields(field)
	bv.operations = make([]*boolOperation, 0)
	bv.codesInUse.Add(bv.defaultErrorCode)

	return bv
}

type BoolValidator struct {
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

func (bv *BoolValidator) IsSet(field string, subject interface{}) (bool, error) {

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

func (bv *BoolValidator) Validate(vc *ValidationContext) (result *ValidationResult, unexpected error) {

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

func (bv *BoolValidator) runOperations(field string, b bool, vc *ValidationContext, r *ValidationResult) error {

	ec := types.NewEmptyOrderedStringSet()

	for _, op := range bv.operations {

		switch op.OpType {
		case BoolOpMex:
			checkMExFields(op.MExFields, vc, ec, op.ErrCode)
		}
	}

	r.AddForField(field, ec.Contents())

	return nil

}

func (bv *BoolValidator) extractValue(f string, s interface{}) (*types.NilableBool, error) {

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

func (bv *BoolValidator) StopAllOnFail() bool {
	return bv.stopAll
}

func (bv *BoolValidator) CodesInUse() types.StringSet {
	return bv.codesInUse
}

func (bv *BoolValidator) DependsOnFields() types.StringSet {

	return bv.dependsFields
}

func (bv *BoolValidator) StopAll() *BoolValidator {

	bv.stopAll = true

	return bv
}

func (bv *BoolValidator) Required(code ...string) *BoolValidator {

	bv.required = true
	bv.missingRequiredCode = bv.chooseErrorCode(code)

	return bv
}

func (bv *BoolValidator) Is(v bool, code ...string) *BoolValidator {

	bv.requiredValue = types.NewNilableBool(v)

	bv.requiredValueCode = bv.chooseErrorCode(code)

	return bv
}

func (bv *BoolValidator) MEx(fields types.StringSet, code ...string) *BoolValidator {
	op := new(boolOperation)
	op.ErrCode = bv.chooseErrorCode(code)
	op.OpType = BoolOpMex
	op.MExFields = fields

	bv.addOperation(op)

	return bv
}

func (bv *BoolValidator) addOperation(o *boolOperation) {
	bv.operations = append(bv.operations, o)
}

func (bv *BoolValidator) chooseErrorCode(v []string) string {

	if len(v) > 0 {
		bv.codesInUse.Add(v[0])
		return v[0]
	} else {
		return bv.defaultErrorCode
	}

}

func (bv *BoolValidator) Operation(c string) (boolValidationOperation, error) {
	switch c {
	case boolOpRequiredCode:
		return BoolOpRequired, nil
	case boolOpStopAllCode:
		return BoolOpStopAll, nil
	case boolOpIsCode:
		return BoolOpIs, nil
	case boolOpMexCode:
		return BoolOpMex, nil
	}

	m := fmt.Sprintf("Unsupported bool validation operation %s", c)
	return BoolOpUnsupported, errors.New(m)

}

func NewBoolValidatorBuilder(ec string, cf ioc.ComponentByNameFinder) *BoolValidatorBuilder {
	bv := new(BoolValidatorBuilder)
	bv.componentFinder = cf
	bv.defaultErrorCode = ec

	return bv
}

type BoolValidatorBuilder struct {
	defaultErrorCode string
	componentFinder  ioc.ComponentByNameFinder
}

func (vb *BoolValidatorBuilder) parseRule(field string, rule []string) (ValidationRule, error) {

	defaultErrorcode := determineDefaultErrorCode(BoolRuleCode, rule, vb.defaultErrorCode)
	bv := NewBoolValidator(field, defaultErrorcode)

	for _, v := range rule {

		ops := decomposeOperation(v)
		opCode := ops[0]

		if isTypeIndicator(BoolRuleCode, opCode) {
			continue
		}

		op, err := bv.Operation(opCode)

		if err != nil {
			return nil, err
		}

		switch op {
		case BoolOpRequired:
			err = vb.markRequired(field, ops, bv)
		case BoolOpStopAll:
			bv.StopAll()
		case BoolOpIs:
			err = vb.captureRequiredValue(field, ops, bv)
		case BoolOpMex:
			err = vb.captureExclusiveFields(field, ops, bv)
		}

		if err != nil {

			return nil, err
		}

	}

	return bv, nil

}

func (vb *BoolValidatorBuilder) captureExclusiveFields(field string, ops []string, bv *BoolValidator) error {
	_, err := paramCount(ops, "MEX", field, 2, 3)

	if err != nil {
		return err
	}

	members := strings.SplitN(ops[1], setMemberSep, -1)
	fields := types.NewOrderedStringSet(members)

	bv.MEx(fields, extractVargs(ops, 3)...)

	return nil

}

func (vb *BoolValidatorBuilder) captureRequiredValue(field string, ops []string, bv *BoolValidator) error {
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

func (vb *BoolValidatorBuilder) markRequired(field string, ops []string, bv *BoolValidator) error {

	_, err := paramCount(ops, "Required", field, 1, 2)

	if err != nil {
		return err
	}

	bv.Required(extractVargs(ops, 2)...)

	return nil
}
