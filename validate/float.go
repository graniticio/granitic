package validate

import (
	"errors"
	"fmt"
	"github.com/graniticio/granitic/ioc"
	rt "github.com/graniticio/granitic/reflecttools"
	"github.com/graniticio/granitic/types"
	"reflect"
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
)

func NewFloatValidator(field, defaultErrorCode string) *FloatValidator {
	iv := new(FloatValidator)
	iv.defaultErrorCode = defaultErrorCode
	iv.field = field
	iv.codesInUse = types.NewOrderedStringSet([]string{})
	iv.dependsFields = determinePathFields(field)
	iv.operations = make([]*floatOperation, 0)

	iv.codesInUse.Add(iv.defaultErrorCode)

	return iv
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
	OpType   floatValidationOperation
	ErrCode  string
	InSet    map[float64]bool
	External ExternalFloat64Validator
}

func (iv *FloatValidator) Validate(vc *validationContext) (result *ValidationResult, unexpected error) {

	f := iv.field

	if vc.OverrideField != "" {
		f = vc.OverrideField
	}

	sub := vc.Subject

	fv, err := rt.FindNestedField(rt.ExtractDotPath(f), sub)

	if err != nil {
		m := fmt.Sprintf("Problem trying to find value of %s: %s\n", f, err)
		return nil, errors.New(m)
	}

	if !fv.IsValid() {
		m := fmt.Sprintf("Field %s is not a usable type\n", f)
		return nil, errors.New(m)
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

		return r, nil
	}

	return iv.runOperations(value.Float64())
}

func (iv *FloatValidator) runOperations(i float64) (*ValidationResult, error) {

	ec := new(types.OrderedStringSet)

OpLoop:
	for _, op := range iv.operations {

		switch op.OpType {
		case FloatOpIn:
			if !iv.checkIn(i, op) {
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
			if !iv.inRange(i, op) {
				ec.Add(op.ErrCode)
			}

		}

	}

	r := new(ValidationResult)
	r.ErrorCodes = ec.Contents()

	return r, nil

}

func (iv *FloatValidator) inRange(i float64, o *floatOperation) bool {

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

func (iv *FloatValidator) checkIn(i float64, o *floatOperation) bool {
	return o.InSet[i]
}

func (iv *FloatValidator) Break() *FloatValidator {

	o := new(floatOperation)
	o.OpType = FloatOpBreak

	iv.addOperation(o)

	return iv

}

func (iv *FloatValidator) addOperation(o *floatOperation) {
	iv.operations = append(iv.operations, o)
	iv.codesInUse.Add(o.ErrCode)
}

func (iv *FloatValidator) extractValue(v reflect.Value, f string) (*types.NilableFloat64, error) {

	if rt.NilPointer(v) {
		return nil, nil
	}

	var ex float64

	switch i := v.Interface().(type) {
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

func (iv *FloatValidator) StopAllOnFail() bool {
	return iv.stopAll
}

func (iv *FloatValidator) CodesInUse() types.StringSet {
	return iv.codesInUse
}

func (iv *FloatValidator) DependsOnFields() types.StringSet {

	return iv.dependsFields
}

func (iv *FloatValidator) StopAll() *FloatValidator {

	iv.stopAll = true

	return iv
}

func (iv *FloatValidator) Required(code ...string) *FloatValidator {

	iv.required = true
	iv.missingRequiredCode = iv.chooseErrorCode(code)

	return iv
}

func (iv *FloatValidator) Range(checkMin, checkMax bool, min, max float64, code ...string) *FloatValidator {

	iv.checkMin = checkMin
	iv.checkMax = checkMax
	iv.minAllowed = min
	iv.maxAllowed = max

	ec := iv.chooseErrorCode(code)

	o := new(floatOperation)
	o.OpType = FloatOpRange
	o.ErrCode = ec

	iv.addOperation(o)

	return iv
}

func (iv *FloatValidator) In(set []float64, code ...string) *FloatValidator {

	ec := iv.chooseErrorCode(code)

	o := new(floatOperation)
	o.OpType = FloatOpIn
	o.ErrCode = ec

	fm := make(map[float64]bool)

	for _, m := range set {
		fm[m] = true
	}

	o.InSet = fm

	iv.addOperation(o)

	return iv

}

func (iv *FloatValidator) ExternalValidation(v ExternalFloat64Validator, code ...string) *FloatValidator {
	ec := iv.chooseErrorCode(code)

	o := new(floatOperation)
	o.OpType = FloatOpExt
	o.ErrCode = ec
	o.External = v

	iv.addOperation(o)

	return iv
}

func (iv *FloatValidator) chooseErrorCode(v []string) string {

	if len(v) > 0 {
		iv.codesInUse.Add(v[0])
		return v[0]
	} else {
		return iv.defaultErrorCode
	}

}

func (iv *FloatValidator) Operation(c string) (boolValidationOperation, error) {
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
	}

	m := fmt.Sprintf("Unsupported int validation operation %s", c)
	return FloatOpUnsupported, errors.New(m)

}

func NewFloatValidatorBuilder(ec string, cf ioc.ComponentByNameFinder) *floatValidatorBuilder {
	iv := new(floatValidatorBuilder)
	iv.componentFinder = cf
	iv.defaultErrorCode = ec
	iv.rangeRegex = regexp.MustCompile("^(.*)\\|(.*)$")
	return iv
}

type floatValidatorBuilder struct {
	defaultErrorCode string
	componentFinder  ioc.ComponentByNameFinder
	rangeRegex       *regexp.Regexp
}

func (vb *floatValidatorBuilder) parseRule(field string, rule []string) (Validator, error) {

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
		}

		if err != nil {

			return nil, err
		}

	}

	return bv, nil

}

func (vb *floatValidatorBuilder) markRequired(field string, ops []string, iv *FloatValidator) error {

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

func (vb *floatValidatorBuilder) addFloatRangeOperation(field string, ops []string, iv *FloatValidator) error {

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
		iv.Range(checkMin, checkMax, float64(min), float64(max))
	} else {
		iv.Range(checkMin, checkMax, float64(min), float64(max), ops[2])
	}

	return nil
}

func (vb *floatValidatorBuilder) addFloatExternalOperation(field string, ops []string, iv *FloatValidator) error {

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
		iv.ExternalValidation(ev)
	} else {
		iv.ExternalValidation(ev, ops[2])
	}

	return nil

}

func (vb *floatValidatorBuilder) addFloatInOperation(field string, ops []string, sv *FloatValidator) error {

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
