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

type ExternalFloat64Validator interface {
	ValidFloat64(float64) bool
}

const FloatRuleCode = "FLOAT"

const (
	FloatOpIsRequiredCode = commonOpRequired
	FloatOpIsStopAllCode  = commonOpStopAll
	FloatOpInCode         = commonOpIn
	FloatOpBreakCode      = commonOpBreak
	FloatOpExtCode        = commonOpExt
	FloatOpRangeCode      = "RANGE"
	FloatOpMExCode        = commonOpMex
)

type floatValidationOperation uint

const (
	FloatOpUnsupported = iota
	FloatOpRequired
	FloatOpStopAll
	FloatOpIn
	FloatOpBreak
	FloatOpExt
	FloatOpRange
	FloatOpMex
)

func NewFloatValidator(field, defaultErrorCode string) *FloatValidator {
	fv := new(FloatValidator)
	fv.defaultErrorCode = defaultErrorCode
	fv.field = field
	fv.codesInUse = types.NewOrderedStringSet([]string{})
	fv.dependsFields = determinePathFields(field)
	fv.operations = make([]*floatOperation, 0)

	fv.codesInUse.Add(fv.defaultErrorCode)

	return fv
}

type FloatValidator struct {
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

func (fv *FloatValidator) IsSet(field string, subject interface{}) (bool, error) {

	nf, err := fv.extractValue(field, subject)

	if err != nil {
		return false, err
	}

	if nf == nil || !nf.IsSet() {
		return false, nil
	} else {
		return true, nil
	}
}

func (fv *FloatValidator) Validate(vc *ValidationContext) (result *ValidationResult, unexpected error) {

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

func (fv *FloatValidator) extractValue(f string, s interface{}) (*types.NilableFloat64, error) {

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

	return fv.ToFloat64(f, v.Interface())
}

func (fv *FloatValidator) ToFloat64(f string, i interface{}) (*types.NilableFloat64, error) {
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

func (fv *FloatValidator) runOperations(field string, i float64, vc *ValidationContext, r *ValidationResult) error {

	ec := types.NewEmptyOrderedStringSet()

OpLoop:
	for _, op := range fv.operations {

		switch op.OpType {
		case FloatOpIn:
			if !fv.checkIn(i, op) {
				ec.Add(op.ErrCode)
			}

		case FloatOpBreak:
			if ec.Size() > 0 {
				break OpLoop
			}

		case FloatOpExt:
			if !op.External.ValidFloat64(i) {
				ec.Add(op.ErrCode)
			}

		case FloatOpRange:
			if !fv.inRange(i, op) {
				ec.Add(op.ErrCode)
			}
		case FloatOpMex:
			checkMExFields(op.MExFields, vc, ec, op.ErrCode)
		}

	}

	r.AddForField(field, ec.Contents())

	return nil

}

func (fv *FloatValidator) inRange(i float64, o *floatOperation) bool {

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

func (fv *FloatValidator) checkIn(i float64, o *floatOperation) bool {
	return o.InSet[i]
}

func (fv *FloatValidator) MEx(fields types.StringSet, code ...string) *FloatValidator {
	op := new(floatOperation)
	op.ErrCode = fv.chooseErrorCode(code)
	op.OpType = FloatOpMex
	op.MExFields = fields

	fv.addOperation(op)

	return fv
}

func (fv *FloatValidator) Break() *FloatValidator {

	o := new(floatOperation)
	o.OpType = FloatOpBreak

	fv.addOperation(o)

	return fv

}

func (fv *FloatValidator) addOperation(o *floatOperation) {
	fv.operations = append(fv.operations, o)
	fv.codesInUse.Add(o.ErrCode)
}

func (fv *FloatValidator) StopAllOnFail() bool {
	return fv.stopAll
}

func (fv *FloatValidator) CodesInUse() types.StringSet {
	return fv.codesInUse
}

func (fv *FloatValidator) DependsOnFields() types.StringSet {

	return fv.dependsFields
}

func (fv *FloatValidator) StopAll() *FloatValidator {

	fv.stopAll = true

	return fv
}

func (fv *FloatValidator) Required(code ...string) *FloatValidator {

	fv.required = true
	fv.missingRequiredCode = fv.chooseErrorCode(code)

	return fv
}

func (fv *FloatValidator) Range(checkMin, checkMax bool, min, max float64, code ...string) *FloatValidator {

	fv.checkMin = checkMin
	fv.checkMax = checkMax
	fv.minAllowed = min
	fv.maxAllowed = max

	ec := fv.chooseErrorCode(code)

	o := new(floatOperation)
	o.OpType = FloatOpRange
	o.ErrCode = ec

	fv.addOperation(o)

	return fv
}

func (fv *FloatValidator) In(set []float64, code ...string) *FloatValidator {

	ec := fv.chooseErrorCode(code)

	o := new(floatOperation)
	o.OpType = FloatOpIn
	o.ErrCode = ec

	fm := make(map[float64]bool)

	for _, m := range set {
		fm[m] = true
	}

	o.InSet = fm

	fv.addOperation(o)

	return fv

}

func (fv *FloatValidator) ExternalValidation(v ExternalFloat64Validator, code ...string) *FloatValidator {
	ec := fv.chooseErrorCode(code)

	o := new(floatOperation)
	o.OpType = FloatOpExt
	o.ErrCode = ec
	o.External = v

	fv.addOperation(o)

	return fv
}

func (fv *FloatValidator) chooseErrorCode(v []string) string {

	if len(v) > 0 {
		fv.codesInUse.Add(v[0])
		return v[0]
	} else {
		return fv.defaultErrorCode
	}

}

func (fv *FloatValidator) Operation(c string) (boolValidationOperation, error) {
	switch c {
	case FloatOpIsRequiredCode:
		return FloatOpRequired, nil
	case FloatOpIsStopAllCode:
		return FloatOpStopAll, nil
	case FloatOpInCode:
		return FloatOpIn, nil
	case FloatOpBreakCode:
		return FloatOpBreak, nil
	case FloatOpExtCode:
		return FloatOpExt, nil
	case FloatOpRangeCode:
		return FloatOpRange, nil
	case FloatOpMExCode:
		return FloatOpMex, nil
	}

	m := fmt.Sprintf("Unsupported int validation operation %s", c)
	return FloatOpUnsupported, errors.New(m)

}

func NewFloatValidatorBuilder(ec string, cf ioc.ComponentByNameFinder) *FloatValidatorBuilder {
	fv := new(FloatValidatorBuilder)
	fv.componentFinder = cf
	fv.defaultErrorCode = ec
	fv.rangeRegex = regexp.MustCompile("^(.*)\\|(.*)$")
	return fv
}

type FloatValidatorBuilder struct {
	defaultErrorCode string
	componentFinder  ioc.ComponentByNameFinder
	rangeRegex       *regexp.Regexp
}

func (vb *FloatValidatorBuilder) parseRule(field string, rule []string) (Validator, error) {

	defaultErrorcode := DetermineDefaultErrorCode(FloatRuleCode, rule, vb.defaultErrorCode)
	bv := NewFloatValidator(field, defaultErrorcode)

	for _, v := range rule {

		ops := DecomposeOperation(v)
		opCode := ops[0]

		if IsTypeIndicator(FloatRuleCode, opCode) {
			continue
		}

		op, err := bv.Operation(opCode)

		if err != nil {
			return nil, err
		}

		switch op {
		case FloatOpRequired:
			err = vb.markRequired(field, ops, bv)
		case FloatOpIn:
			err = vb.addFloatInOperation(field, ops, bv)
		case FloatOpStopAll:
			bv.StopAll()
		case FloatOpBreak:
			bv.Break()
		case FloatOpExt:
			err = vb.addFloatExternalOperation(field, ops, bv)
		case FloatOpRange:
			err = vb.addFloatRangeOperation(field, ops, bv)
		case FloatOpMex:
			err = vb.captureExclusiveFields(field, ops, bv)
		}

		if err != nil {

			return nil, err
		}

	}

	return bv, nil

}

func (vb *FloatValidatorBuilder) captureExclusiveFields(field string, ops []string, fv *FloatValidator) error {
	_, err := paramCount(ops, "MEX", field, 2, 3)

	if err != nil {
		return err
	}

	members := strings.SplitN(ops[1], setMemberSep, -1)
	fields := types.NewOrderedStringSet(members)

	fv.MEx(fields, extractVargs(ops, 3)...)

	return nil

}

func (vb *FloatValidatorBuilder) markRequired(field string, ops []string, fv *FloatValidator) error {

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

func (vb *FloatValidatorBuilder) addFloatRangeOperation(field string, ops []string, fv *FloatValidator) error {

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

func (vb *FloatValidatorBuilder) addFloatExternalOperation(field string, ops []string, fv *FloatValidator) error {

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

func (vb *FloatValidatorBuilder) addFloatInOperation(field string, ops []string, sv *FloatValidator) error {

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
