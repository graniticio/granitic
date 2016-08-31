package validate

import (
	"errors"
	"fmt"
	"github.com/graniticio/granitic/ioc"
	rt "github.com/graniticio/granitic/reflecttools"
	"github.com/graniticio/granitic/types"
	"reflect"
)

const IntRuleCode = "INT"

const (
	IntOpIsRequiredCode = commonOpRequired
	IntOpIsStopAllCode  = commonOpStopAll
)

type intValidationOperation uint

const (
	IntOpUnsupported = iota
	IntOpRequired
	IntOpStopAll
)

func NewIntValidator(field, defaultErrorCode string) *intValidator {
	iv := new(intValidator)
	iv.defaultErrorCode = defaultErrorCode
	iv.field = field
	iv.codesInUse = types.NewOrderedStringSet([]string{})
	iv.dependsFields = determinePathFields(field)

	return iv
}

type intValidator struct {
	stopAll             bool
	codesInUse          types.StringSet
	dependsFields       types.StringSet
	defaultErrorCode    string
	field               string
	missingRequiredCode string
	required            bool
}

type IntOpIseration struct {
	OpType  intValidationOperation
	ErrCode string
}

func (iv *intValidator) Validate(vc *validationContext) (result *ValidationResult, unexpected error) {

	f := iv.field

	if vc.OverrideField != "" {
		f = vc.OverrideField
	}

	sub := vc.Subject

	fv, err := rt.FindNestedField(rt.ExtractDotPath(f), sub)

	if err != nil {
		return nil, err
	}

	r := new(ValidationResult)

	value, err := iv.extractValue(fv, f)

	if err != nil {
		return nil, err
	}

	if value == nil || !value.IsSet() {

		r.Unset = true

		if iv.required {
			r.ErrorCodes = []string{iv.missingRequiredCode}
		} else {
			r.ErrorCodes = []string{}
		}

	}

	return r, nil
}

func (iv *intValidator) extractValue(v reflect.Value, f string) (*types.NilableInt64, error) {

	if rt.NilPointer(v) {
		return nil, nil
	}

	var ex int64

	switch i := v.Interface().(type) {
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

func (iv *intValidator) StopAllOnFail() bool {
	return iv.stopAll
}

func (iv *intValidator) CodesInUse() types.StringSet {
	return iv.codesInUse
}

func (iv *intValidator) DependsOnFields() types.StringSet {

	return iv.dependsFields
}

func (iv *intValidator) StopAll() *intValidator {

	iv.stopAll = true

	return iv
}

func (iv *intValidator) Required(code ...string) *intValidator {

	iv.required = true

	if code != nil {
		iv.missingRequiredCode = code[0]
	} else {
		iv.missingRequiredCode = iv.defaultErrorCode
	}

	return iv
}

func (iv *intValidator) Operation(c string) (boolValidationOperation, error) {
	switch c {
	case IntOpIsRequiredCode:
		return IntOpRequired, nil
	case IntOpIsStopAllCode:
		return IntOpStopAll, nil
	}

	m := fmt.Sprintf("Unsupported int validation operation %s", c)
	return IntOpUnsupported, errors.New(m)

}

func NewIntValidatorBuilder(ec string, cf ioc.ComponentByNameFinder) *intValidatorBuilder {
	iv := new(intValidatorBuilder)
	iv.componentFinder = cf
	iv.defaultErrorCode = ec

	return iv
}

type intValidatorBuilder struct {
	defaultErrorCode string
	componentFinder  ioc.ComponentByNameFinder
}

func (vb *intValidatorBuilder) parseRule(field string, rule []string) (Validator, error) {

	defaultErrorcode := DetermineDefaultErrorCode(IntRuleCode, rule, vb.defaultErrorCode)
	bv := NewIntValidator(field, defaultErrorcode)

	for _, v := range rule {

		ops := DecomposeOperation(v)
		opCode := ops[0]

		if IsTypeIndicator(IntRuleCode, opCode) {
			continue
		}

		op, err := bv.Operation(opCode)

		if err != nil {
			return nil, err
		}

		switch op {
		case IntOpRequired:
			err = vb.markRequired(field, ops, bv)
		case IntOpStopAll:
			bv.StopAll()
		}

		if err != nil {

			return nil, err
		}

	}

	return bv, nil

}

func (vb *intValidatorBuilder) markRequired(field string, ops []string, iv *intValidator) error {

	pCount, err := paramCount(ops, "Required", field, 1, 2)

	if err != nil {
		return err
	}

	if pCount == 1 {
		iv.Required()
	} else {
		iv.Required(ops[1])
	}

	return nil
}
