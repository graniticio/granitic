package validate

import (
	"errors"
	"fmt"
	"github.com/graniticio/granitic/ioc"
	rt "github.com/graniticio/granitic/reflecttools"
	"github.com/graniticio/granitic/types"
	"reflect"
)

const ObjectRuleCode = "OBJ"

const (
	objOpRequiredCode = commonOpRequired
	objOpStopAllCode  = commonOpStopAll
)

type ObjectValidationOperation uint

const (
	ObjOpUnsupported = iota
	ObjOpRequired
	ObjOpStopAll
)

func NewObjectValidator(field, defaultErrorCode string) *ObjectValidator {
	ov := new(ObjectValidator)
	ov.defaultErrorCode = defaultErrorCode
	ov.field = field
	ov.codesInUse = types.NewOrderedStringSet([]string{})
	ov.dependsFields = determinePathFields(field)

	ov.codesInUse.Add(ov.defaultErrorCode)

	return ov
}

type ObjectValidator struct {
	stopAll             bool
	codesInUse          types.StringSet
	dependsFields       types.StringSet
	defaultErrorCode    string
	field               string
	missingRequiredCode string
	required            bool
}

type objectOperation struct {
	OpType  ObjectValidationOperation
	ErrCode string
}

func (ov *ObjectValidator) IsSet(field string, subject interface{}) (bool, error) {
	fv, err := rt.FindNestedField(rt.ExtractDotPath(field), subject)

	if err != nil {
		return false, err
	}

	if rt.NilPointer(fv) || rt.NilMap(fv) {
		return false, nil
	} else {

		k := fv.Kind()

		if !rt.IsPointerToStruct(fv.Interface()) && k != reflect.Map && k != reflect.Struct {
			m := fmt.Sprintf("Field %s is not a pointer to a struct, a struct or a map.", field)
			return false, errors.New(m)
		}

		return true, nil
	}
}

func (ov *ObjectValidator) Validate(vc *validationContext) (result *ValidationResult, unexpected error) {

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

	}

	return r, nil
}

func (ov *ObjectValidator) StopAllOnFail() bool {
	return ov.stopAll
}

func (ov *ObjectValidator) CodesInUse() types.StringSet {
	return ov.codesInUse
}

func (ov *ObjectValidator) DependsOnFields() types.StringSet {

	return ov.dependsFields
}

func (ov *ObjectValidator) StopAll() *ObjectValidator {

	ov.stopAll = true

	return ov
}

func (ov *ObjectValidator) Required(code ...string) *ObjectValidator {

	ov.required = true

	if code != nil {
		ov.missingRequiredCode = code[0]
	} else {
		ov.missingRequiredCode = ov.defaultErrorCode
	}

	ov.codesInUse.Add(ov.missingRequiredCode)

	return ov
}

func (ov *ObjectValidator) Operation(c string) (ObjectValidationOperation, error) {
	switch c {
	case objOpRequiredCode:
		return ObjOpRequired, nil
	case objOpStopAllCode:
		return ObjOpStopAll, nil
	}

	m := fmt.Sprintf("Unsupported object validation operation %s", c)
	return ObjOpUnsupported, errors.New(m)

}

func NewObjectValidatorBuilder(ec string, cf ioc.ComponentByNameFinder) *objectValidatorBuilder {
	ov := new(objectValidatorBuilder)
	ov.componentFinder = cf
	ov.defaultErrorCode = ec

	return ov
}

type objectValidatorBuilder struct {
	defaultErrorCode string
	componentFinder  ioc.ComponentByNameFinder
}

func (vb *objectValidatorBuilder) parseRule(field string, rule []string) (Validator, error) {

	defaultErrorcode := DetermineDefaultErrorCode(ObjectRuleCode, rule, vb.defaultErrorCode)
	ov := NewObjectValidator(field, defaultErrorcode)

	for _, v := range rule {

		ops := DecomposeOperation(v)
		opCode := ops[0]

		if IsTypeIndicator(ObjectRuleCode, opCode) {
			continue
		}

		op, err := ov.Operation(opCode)

		if err != nil {
			return nil, err
		}

		switch op {
		case ObjOpRequired:
			err = vb.markRequired(field, ops, ov)
		case ObjOpStopAll:
			ov.StopAll()
		}

		if err != nil {

			return nil, err
		}

	}

	return ov, nil

}

func (vb *objectValidatorBuilder) markRequired(field string, ops []string, ov *ObjectValidator) error {

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
