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
	ov := new(boolValidator)
	ov.defaultErrorCode = defaultErrorCode
	ov.field = field
	ov.codesInUse = types.NewOrderedStringSet([]string{})
	ov.dependsFields = determinePathFields(field)
	ov.operations = make([]*boolOperation, 0)
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
	operations          []*boolOperation
}

type boolOperation struct {
	OpType    boolValidationOperation
	ErrCode   string
	MExFields types.StringSet
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

	return ov.runOperations(value.Bool(), vc)
}

func (iv *boolValidator) runOperations(b bool, vc *validationContext) (*ValidationResult, error) {

	ec := new(types.OrderedStringSet)

	for _, op := range iv.operations {

		switch op.OpType {
		case BoolOpMex:
			iv.checkMExFields(op, vc, ec)
		}
	}

	r := new(ValidationResult)
	r.ErrorCodes = ec.Contents()

	return r, nil

}

func (ov *boolValidator) checkMExFields(op *boolOperation, vc *validationContext, ec types.StringSet) {

	if vc.KnownSetFields == nil || vc.KnownSetFields.Size() == 0 {
		return
	}

	for _, s := range op.MExFields.Contents() {

		if vc.KnownSetFields.Contains(s) {
			ec.Add(op.ErrCode)
			break
		}
	}

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
	ov.missingRequiredCode = ov.chooseErrorCode(code)

	return ov
}

func (ov *boolValidator) Is(v bool, code ...string) *boolValidator {

	ov.requiredValue = types.NewNilableBool(v)

	ov.requiredValueCode = ov.chooseErrorCode(code)

	return ov
}

func (ov *boolValidator) MEx(fields types.StringSet, code ...string) *boolValidator {
	op := new(boolOperation)
	op.ErrCode = ov.chooseErrorCode(code)
	op.OpType = BoolOpMex
	op.MExFields = fields

	ov.addOperation(op)

	return ov
}

func (ov *boolValidator) addOperation(o *boolOperation) {
	ov.operations = append(ov.operations, o)
}

func (ov *boolValidator) chooseErrorCode(v []string) string {

	if len(v) > 0 {
		ov.codesInUse.Add(v[0])
		return v[0]
	} else {
		return ov.defaultErrorCode
	}

}

func (ov *boolValidator) Operation(c string) (boolValidationOperation, error) {
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
		case BoolOpMex:
			err = vb.captureExclusiveFields(field, ops, bv)
		}

		if err != nil {

			return nil, err
		}

	}

	return bv, nil

}

func (vb *boolValidatorBuilder) captureExclusiveFields(field string, ops []string, ov *boolValidator) error {
	pCount, err := paramCount(ops, "MEX", field, 2, 3)

	if err != nil {
		return err
	}

	members := strings.SplitN(ops[1], setMemberSep, -1)
	fields := types.NewOrderedStringSet(members)

	if pCount == 2 {
		ov.MEx(fields)
	} else {
		ov.MEx(fields, ops[2])
	}

	return nil

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
