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

func NewBoolValidator(field, defaultErrorCode string) *boolValidator {
	bv := new(boolValidator)
	bv.defaultErrorCode = defaultErrorCode
	bv.field = field
	bv.codesInUse = types.NewOrderedStringSet([]string{})
	bv.dependsFields = determinePathFields(field)
	bv.operations = make([]*boolOperation, 0)
	bv.codesInUse.Add(bv.defaultErrorCode)

	return bv
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
	operations          []*boolOperation
}

type boolOperation struct {
	OpType    boolValidationOperation
	ErrCode   string
	MExFields types.StringSet
}

func (bv *boolValidator) IsSet(field string, subject interface{}) (bool, error) {

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

func (bv *boolValidator) Validate(vc *validationContext) (result *ValidationResult, unexpected error) {

	f := bv.field

	if vc.OverrideField != "" {
		f = vc.OverrideField
	}

	sub := vc.Subject

	r := new(ValidationResult)
	set, err := bv.IsSet(f, sub)

	if err != nil {
		return nil, err

	} else if !set {
		r.Unset = true

		if bv.required {
			r.ErrorCodes = []string{bv.missingRequiredCode}
		} else {
			r.ErrorCodes = []string{}
		}

		return r, nil
	}

	//Ignoring error as called previously during IsSet
	value, _ := bv.extractValue(f, sub)

	if bv.requiredValue != nil && value.Bool() != bv.requiredValue.Bool() {

		r.ErrorCodes = []string{bv.requiredValueCode}
	}

	return bv.runOperations(value.Bool(), vc, r.ErrorCodes)
}

func (bv *boolValidator) runOperations(b bool, vc *validationContext, errors []string) (*ValidationResult, error) {

	if errors == nil {
		errors = []string{}
	}

	ec := types.NewOrderedStringSet(errors)

	for _, op := range bv.operations {

		switch op.OpType {
		case BoolOpMex:
			checkMExFields(op.MExFields, vc, ec, op.ErrCode)
		}
	}

	r := new(ValidationResult)
	r.ErrorCodes = ec.Contents()

	return r, nil

}

func (bv *boolValidator) extractValue(f string, s interface{}) (*types.NilableBool, error) {

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

func (bv *boolValidator) StopAllOnFail() bool {
	return bv.stopAll
}

func (bv *boolValidator) CodesInUse() types.StringSet {
	return bv.codesInUse
}

func (bv *boolValidator) DependsOnFields() types.StringSet {

	return bv.dependsFields
}

func (bv *boolValidator) StopAll() *boolValidator {

	bv.stopAll = true

	return bv
}

func (bv *boolValidator) Required(code ...string) *boolValidator {

	bv.required = true
	bv.missingRequiredCode = bv.chooseErrorCode(code)

	return bv
}

func (bv *boolValidator) Is(v bool, code ...string) *boolValidator {

	bv.requiredValue = types.NewNilableBool(v)

	bv.requiredValueCode = bv.chooseErrorCode(code)

	return bv
}

func (bv *boolValidator) MEx(fields types.StringSet, code ...string) *boolValidator {
	op := new(boolOperation)
	op.ErrCode = bv.chooseErrorCode(code)
	op.OpType = BoolOpMex
	op.MExFields = fields

	bv.addOperation(op)

	return bv
}

func (bv *boolValidator) addOperation(o *boolOperation) {
	bv.operations = append(bv.operations, o)
}

func (bv *boolValidator) chooseErrorCode(v []string) string {

	if len(v) > 0 {
		bv.codesInUse.Add(v[0])
		return v[0]
	} else {
		return bv.defaultErrorCode
	}

}

func (bv *boolValidator) Operation(c string) (boolValidationOperation, error) {
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

func NewBoolValidatorBuilder(ec string, cf ioc.ComponentByNameFinder) *boolValidatorBuilder {
	bv := new(boolValidatorBuilder)
	bv.componentFinder = cf
	bv.defaultErrorCode = ec

	return bv
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
		case BoolOpMex:
			err = vb.captureExclusiveFields(field, ops, bv)
		}

		if err != nil {

			return nil, err
		}

	}

	return bv, nil

}

func (vb *boolValidatorBuilder) captureExclusiveFields(field string, ops []string, bv *boolValidator) error {
	_, err := paramCount(ops, "MEX", field, 2, 3)

	if err != nil {
		return err
	}

	members := strings.SplitN(ops[1], setMemberSep, -1)
	fields := types.NewOrderedStringSet(members)

	bv.MEx(fields, extractVargs(ops, 3)...)

	return nil

}

func (vb *boolValidatorBuilder) captureRequiredValue(field string, ops []string, bv *boolValidator) error {
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

func (vb *boolValidatorBuilder) markRequired(field string, ops []string, bv *boolValidator) error {

	_, err := paramCount(ops, "Required", field, 1, 2)

	if err != nil {
		return err
	}

	bv.Required(extractVargs(ops, 2)...)

	return nil
}
