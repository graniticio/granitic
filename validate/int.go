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

type ExternalInt64Validator interface {
	ValidInt64(int64) bool
}

const IntRuleCode = "INT"

const (
	IntOpIsRequiredCode = commonOpRequired
	IntOpIsStopAllCode  = commonOpStopAll
	IntOpInCode         = commonOpIn
	IntOpBreakCode      = commonOpBreak
	IntOpExtCode        = commonOpExt
	IntOpRangeCode      = "RANGE"
	IntOpMExCode        = commonOpMex
)

type intValidationOperation uint

const (
	IntOpUnsupported = iota
	IntOpRequired
	IntOpStopAll
	IntOpIn
	IntOpBreak
	IntOpExt
	IntOpRange
	IntOpMEx
)

func NewIntValidator(field, defaultErrorCode string) *IntValidator {
	iv := new(IntValidator)
	iv.defaultErrorCode = defaultErrorCode
	iv.field = field
	iv.codesInUse = types.NewOrderedStringSet([]string{})
	iv.dependsFields = determinePathFields(field)
	iv.operations = make([]*intOperation, 0)

	iv.codesInUse.Add(iv.defaultErrorCode)

	return iv
}

type IntValidator struct {
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

func (iv *IntValidator) IsSet(field string, subject interface{}) (bool, error) {
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

func (iv *IntValidator) Validate(vc *ValidationContext) (result *ValidationResult, unexpected error) {

	f := iv.field

	if vc.OverrideField != "" {
		f = vc.OverrideField
	}

	sub := vc.Subject

	r := NewValidationResult()

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
	value, _ := iv.extractValue(f, sub)

	err = iv.runOperations(f, value.Int64(), vc, r)

	return r, err
}

func (iv *IntValidator) runOperations(field string, i int64, vc *ValidationContext, r *ValidationResult) error {

	ec := types.NewEmptyOrderedStringSet()

OpLoop:
	for _, op := range iv.operations {

		switch op.OpType {
		case IntOpIn:
			if !iv.checkIn(i, op) {
				ec.Add(op.ErrCode)
			}

		case IntOpBreak:
			if ec.Size() > 0 {
				break OpLoop
			}

		case IntOpExt:
			if !op.External.ValidInt64(i) {
				ec.Add(op.ErrCode)
			}

		case IntOpRange:
			if !iv.inRange(i, op) {
				ec.Add(op.ErrCode)
			}
		case IntOpMEx:
			checkMExFields(op.MExFields, vc, ec, op.ErrCode)
		}

	}

	r.AddForField(field, ec.Contents())

	return nil

}

func (iv *IntValidator) inRange(i int64, o *intOperation) bool {

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

func (iv *IntValidator) checkIn(i int64, o *intOperation) bool {
	s := strconv.FormatInt(i, 10)

	return o.InSet.Contains(s)
}

func (iv *IntValidator) MEx(fields types.StringSet, code ...string) *IntValidator {
	op := new(intOperation)
	op.ErrCode = iv.chooseErrorCode(code)
	op.OpType = IntOpMEx
	op.MExFields = fields

	iv.addOperation(op)

	return iv
}

func (iv *IntValidator) Break() *IntValidator {

	o := new(intOperation)
	o.OpType = IntOpBreak

	iv.addOperation(o)

	return iv

}

func (iv *IntValidator) addOperation(o *intOperation) {
	iv.operations = append(iv.operations, o)
	iv.codesInUse.Add(o.ErrCode)
}

func (iv *IntValidator) extractValue(f string, s interface{}) (*types.NilableInt64, error) {

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

func (iv *IntValidator) StopAllOnFail() bool {
	return iv.stopAll
}

func (iv *IntValidator) CodesInUse() types.StringSet {
	return iv.codesInUse
}

func (iv *IntValidator) DependsOnFields() types.StringSet {

	return iv.dependsFields
}

func (iv *IntValidator) StopAll() *IntValidator {

	iv.stopAll = true

	return iv
}

func (iv *IntValidator) Required(code ...string) *IntValidator {

	iv.required = true
	iv.missingRequiredCode = iv.chooseErrorCode(code)

	return iv
}

func (iv *IntValidator) Range(checkMin, checkMax bool, min, max int64, code ...string) *IntValidator {

	iv.checkMin = checkMin
	iv.checkMax = checkMax
	iv.minAllowed = min
	iv.maxAllowed = max

	ec := iv.chooseErrorCode(code)

	o := new(intOperation)
	o.OpType = IntOpRange
	o.ErrCode = ec

	iv.addOperation(o)

	return iv
}

func (iv *IntValidator) In(set []string, code ...string) *IntValidator {

	ss := types.NewUnorderedStringSet(set)

	ec := iv.chooseErrorCode(code)

	o := new(intOperation)
	o.OpType = IntOpIn
	o.ErrCode = ec
	o.InSet = ss

	iv.addOperation(o)

	return iv

}

func (iv *IntValidator) ExternalValidation(v ExternalInt64Validator, code ...string) *IntValidator {
	ec := iv.chooseErrorCode(code)

	o := new(intOperation)
	o.OpType = IntOpExt
	o.ErrCode = ec
	o.External = v

	iv.addOperation(o)

	return iv
}

func (iv *IntValidator) chooseErrorCode(v []string) string {

	if len(v) > 0 {
		iv.codesInUse.Add(v[0])
		return v[0]
	} else {
		return iv.defaultErrorCode
	}

}

func (iv *IntValidator) Operation(c string) (boolValidationOperation, error) {
	switch c {
	case IntOpIsRequiredCode:
		return IntOpRequired, nil
	case IntOpIsStopAllCode:
		return IntOpStopAll, nil
	case IntOpInCode:
		return IntOpIn, nil
	case IntOpBreakCode:
		return IntOpBreak, nil
	case IntOpExtCode:
		return IntOpExt, nil
	case IntOpRangeCode:
		return IntOpRange, nil
	case IntOpMExCode:
		return IntOpMEx, nil
	}

	m := fmt.Sprintf("Unsupported int validation operation %s", c)
	return IntOpUnsupported, errors.New(m)

}

func NewIntValidatorBuilder(ec string, cf ioc.ComponentByNameFinder) *IntValidatorBuilder {
	iv := new(IntValidatorBuilder)
	iv.componentFinder = cf
	iv.defaultErrorCode = ec
	iv.rangeRegex = regexp.MustCompile("^([-+]{0,1}\\d*)\\|([-+]{0,1}\\d*)$")
	return iv
}

type IntValidatorBuilder struct {
	defaultErrorCode string
	componentFinder  ioc.ComponentByNameFinder
	rangeRegex       *regexp.Regexp
}

func (vb *IntValidatorBuilder) parseRule(field string, rule []string) (Validator, error) {

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
		case IntOpIn:
			err = vb.addIntInOperation(field, ops, bv)
		case IntOpStopAll:
			bv.StopAll()
		case IntOpBreak:
			bv.Break()
		case IntOpExt:
			err = vb.addIntExternalOperation(field, ops, bv)
		case IntOpRange:
			err = vb.addIntRangeOperation(field, ops, bv)
		case IntOpMEx:
			err = vb.captureExclusiveFields(field, ops, bv)
		}

		if err != nil {

			return nil, err
		}

	}

	return bv, nil

}

func (vb *IntValidatorBuilder) captureExclusiveFields(field string, ops []string, iv *IntValidator) error {
	_, err := paramCount(ops, "MEX", field, 2, 3)

	if err != nil {
		return err
	}

	members := strings.SplitN(ops[1], setMemberSep, -1)
	fields := types.NewOrderedStringSet(members)

	iv.MEx(fields, extractVargs(ops, 3)...)

	return nil

}

func (vb *IntValidatorBuilder) markRequired(field string, ops []string, iv *IntValidator) error {

	_, err := paramCount(ops, "Required", field, 1, 2)

	if err != nil {
		return err
	}

	iv.Required(extractVargs(ops, 2)...)

	return nil
}

func (vb *IntValidatorBuilder) addIntRangeOperation(field string, ops []string, iv *IntValidator) error {

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

func (vb *IntValidatorBuilder) addIntExternalOperation(field string, ops []string, iv *IntValidator) error {

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

func (vb *IntValidatorBuilder) addIntInOperation(field string, ops []string, sv *IntValidator) error {

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
