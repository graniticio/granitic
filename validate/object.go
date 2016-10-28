package validate

import (
	"errors"
	"fmt"
	"github.com/graniticio/granitic/ioc"
	rt "github.com/graniticio/granitic/reflecttools"
	"github.com/graniticio/granitic/types"
	"reflect"
	"strings"
)

const ObjectRuleCode = "OBJ"

const (
	objOpRequiredCode = commonOpRequired
	objOpStopAllCode  = commonOpStopAll
	objOpMExCode      = commonOpMex
)

type ObjectValidationOperation uint

const (
	ObjOpUnsupported = iota
	ObjOpRequired
	ObjOpStopAll
	ObjOpMEx
)

func NewObjectValidator(field, defaultErrorCode string) *ObjectValidator {
	ov := new(ObjectValidator)
	ov.defaultErrorCode = defaultErrorCode
	ov.field = field
	ov.codesInUse = types.NewOrderedStringSet([]string{})
	ov.dependsFields = determinePathFields(field)
	ov.operations = make([]*objectOperation, 0)
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
	operations          []*objectOperation
}

type objectOperation struct {
	OpType    ObjectValidationOperation
	ErrCode   string
	MExFields types.StringSet
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

		if k == reflect.Invalid {
			m := fmt.Sprintf("Field %s is not a valid type. Does the field exist?", field)
			return false, errors.New(m)
		}

		if !rt.IsPointerToStruct(fv.Interface()) && k != reflect.Map && k != reflect.Struct {
			m := fmt.Sprintf("Field %s is not a pointer to a struct, a struct or a map.", field)
			return false, errors.New(m)
		}

		return true, nil
	}
}

func (ov *ObjectValidator) Validate(vc *ValidationContext) (result *ValidationResult, unexpected error) {

	f := ov.field

	if vc.OverrideField != "" {
		f = vc.OverrideField
	}

	sub := vc.Subject
	r := NewValidationResult()

	set, err := ov.IsSet(f, sub)

	if err != nil {
		return nil, err

	} else if !set {

		r.Unset = true

		if ov.required {
			r.AddForField(f, []string{ov.missingRequiredCode})
		}

	}

	err = ov.runOperations(f, vc, r)

	return r, err
}

func (ov *ObjectValidator) runOperations(field string, vc *ValidationContext, r *ValidationResult) error {

	ec := types.NewEmptyOrderedStringSet()

	for _, op := range ov.operations {

		switch op.OpType {
		case ObjOpMEx:
			checkMExFields(op.MExFields, vc, ec, op.ErrCode)
		}
	}

	r.AddForField(field, ec.Contents())

	return nil

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

func (ov *ObjectValidator) MEx(fields types.StringSet, code ...string) *ObjectValidator {
	op := new(objectOperation)
	op.ErrCode = ov.chooseErrorCode(code)
	op.OpType = ObjOpMEx
	op.MExFields = fields

	ov.addOperation(op)

	return ov
}

func (ov *ObjectValidator) chooseErrorCode(v []string) string {

	if len(v) > 0 {
		ov.codesInUse.Add(v[0])
		return v[0]
	} else {
		return ov.defaultErrorCode
	}

}

func (ov *ObjectValidator) addOperation(o *objectOperation) {
	ov.operations = append(ov.operations, o)
}

func (ov *ObjectValidator) Operation(c string) (ObjectValidationOperation, error) {
	switch c {
	case objOpRequiredCode:
		return ObjOpRequired, nil
	case objOpStopAllCode:
		return ObjOpStopAll, nil
	case objOpMExCode:
		return ObjOpMEx, nil
	}

	m := fmt.Sprintf("Unsupported object validation operation %s", c)
	return ObjOpUnsupported, errors.New(m)

}

func NewObjectValidatorBuilder(ec string, cf ioc.ComponentByNameFinder) *ObjectValidatorBuilder {
	ov := new(ObjectValidatorBuilder)
	ov.componentFinder = cf
	ov.defaultErrorCode = ec

	return ov
}

type ObjectValidatorBuilder struct {
	defaultErrorCode string
	componentFinder  ioc.ComponentByNameFinder
}

func (vb *ObjectValidatorBuilder) parseRule(field string, rule []string) (ValidationRule, error) {

	defaultErrorcode := determineDefaultErrorCode(ObjectRuleCode, rule, vb.defaultErrorCode)
	ov := NewObjectValidator(field, defaultErrorcode)

	for _, v := range rule {

		ops := decomposeOperation(v)
		opCode := ops[0]

		if isTypeIndicator(ObjectRuleCode, opCode) {
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
		case ObjOpMEx:
			err = vb.captureExclusiveFields(field, ops, ov)
		}

		if err != nil {

			return nil, err
		}

	}

	return ov, nil

}

func (vb *ObjectValidatorBuilder) captureExclusiveFields(field string, ops []string, bv *ObjectValidator) error {
	_, err := paramCount(ops, "MEX", field, 2, 3)

	if err != nil {
		return err
	}

	members := strings.SplitN(ops[1], setMemberSep, -1)
	fields := types.NewOrderedStringSet(members)

	bv.MEx(fields, extractVargs(ops, 3)...)

	return nil

}

func (vb *ObjectValidatorBuilder) markRequired(field string, ops []string, ov *ObjectValidator) error {

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
