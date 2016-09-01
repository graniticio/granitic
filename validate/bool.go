package validate

import (
	"errors"
	"fmt"
	"github.com/graniticio/granitic/ioc"
	rt "github.com/graniticio/granitic/reflecttools"
	"github.com/graniticio/granitic/types"
	"strconv"
)

const BoolRuleCode = "BOOL"

const (
	boolOpRequiredCode = commonOpRequired
	boolOpStopAllCode  = commonOpStopAll
	boolOpIsCode       = "IS"
)

type boolValidationOperation uint

const (
	BoolOpUnsupported = iota
	BoolOpRequired
	BoolOpStopAll
	BoolOpIs
)

func NewBoolValidator(field, defaultErrorCode string) *boolValidator {
	ov := new(boolValidator)
	ov.defaultErrorCode = defaultErrorCode
	ov.field = field
	ov.codesInUse = types.NewOrderedStringSet([]string{})
	ov.dependsFields = determinePathFields(field)

	ov.codesInUse.Add(ov.defaultErrorCode)

	return ov
}

type boolValidator struct {
	stopAll             bool
	codesInUse          types.StringSet
	dependsFields       types.StringSet
	defaultErrorCode    string
	field               string
	missingRequiredCode string
	required            bool
	requiredValue       *types.NilableBool
	requiredValueCode   string
}

type boolOperation struct {
	OpType  boolValidationOperation
	ErrCode string
}

func (ov *boolValidator) IsSet(field string, subject interface{}) (bool, error) {

	value, err := ov.extractValue(field, subject)

	if err != nil {
		return false, err
	}

	if value == nil || !value.IsSet() {
		return false, nil
	} else {
		return true, nil
	}
}

func (ov *boolValidator) Validate(vc *validationContext) (result *ValidationResult, unexpected error) {

	f := ov.field

	if vc.OverrideField != "" {
		f = vc.OverrideField
	}

	sub := vc.Subject

	r := new(ValidationResult)
	set, err := ov.IsSet(f, sub)

	if err != nil {
		return nil, err

	} else if !set {
		r.Unset = true

		if ov.required {
			r.ErrorCodes = []string{ov.missingRequiredCode}
		} else {
			r.ErrorCodes = []string{}
		}

		return r, nil
	}

	//Ignoring error as called previously during IsSet
	value, _ := ov.extractValue(f, sub)

	if ov.requiredValue != nil && value.Bool() != ov.requiredValue.Bool() {

		r.ErrorCodes = []string{ov.requiredValueCode}
	}

	return r, nil
}

func (ov *boolValidator) extractValue(f string, s interface{}) (*types.NilableBool, error) {

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

func (ov *boolValidator) StopAllOnFail() bool {
	return ov.stopAll
}

func (ov *boolValidator) CodesInUse() types.StringSet {
	return ov.codesInUse
}

func (ov *boolValidator) DependsOnFields() types.StringSet {

	return ov.dependsFields
}

func (ov *boolValidator) StopAll() *boolValidator {

	ov.stopAll = true

	return ov
}

func (ov *boolValidator) Required(code ...string) *boolValidator {

	ov.required = true

	if code != nil {
		ov.missingRequiredCode = code[0]
	} else {
		ov.missingRequiredCode = ov.defaultErrorCode
	}

	ov.codesInUse.Add(ov.missingRequiredCode)

	return ov
}

func (ov *boolValidator) Is(v bool, code ...string) *boolValidator {

	ov.requiredValue = types.NewNilableBool(v)

	if code != nil {
		ov.requiredValueCode = code[0]
	} else {
		ov.requiredValueCode = ov.defaultErrorCode
	}

	return ov
}

func (ov *boolValidator) Operation(c string) (boolValidationOperation, error) {
	switch c {
	case boolOpRequiredCode:
		return BoolOpRequired, nil
	case boolOpStopAllCode:
		return BoolOpStopAll, nil
	case boolOpIsCode:
		return BoolOpIs, nil
	}

	m := fmt.Sprintf("Unsupported bool validation operation %s", c)
	return BoolOpUnsupported, errors.New(m)

}

func NewBoolValidatorBuilder(ec string, cf ioc.ComponentByNameFinder) *boolValidatorBuilder {
	ov := new(boolValidatorBuilder)
	ov.componentFinder = cf
	ov.defaultErrorCode = ec

	return ov
}

type boolValidatorBuilder struct {
	defaultErrorCode string
	componentFinder  ioc.ComponentByNameFinder
}

func (vb *boolValidatorBuilder) parseRule(field string, rule []string) (Validator, error) {

	defaultErrorcode := DetermineDefaultErrorCode(BoolRuleCode, rule, vb.defaultErrorCode)
	bv := NewBoolValidator(field, defaultErrorcode)

	for _, v := range rule {

		ops := DecomposeOperation(v)
		opCode := ops[0]

		if IsTypeIndicator(BoolRuleCode, opCode) {
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
		}

		if err != nil {

			return nil, err
		}

	}

	return bv, nil

}

func (vb *boolValidatorBuilder) captureRequiredValue(field string, ops []string, ov *boolValidator) error {
	pCount, err := paramCount(ops, "Is", field, 2, 3)

	if err != nil {
		return err
	}

	b, err := strconv.ParseBool(ops[1])

	if err != nil {
		m := fmt.Sprintf("Value %s provided as part of a BOOL/IS operation could not be interpreted as a bool\n", ops[1])
		return errors.New(m)
	}

	if pCount == 2 {
		ov.Is(b)
	} else {
		ov.Is(b, ops[2])
	}

	return nil
}

func (vb *boolValidatorBuilder) markRequired(field string, ops []string, ov *boolValidator) error {

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
