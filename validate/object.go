package validate

import (
	"errors"
	"fmt"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/types"
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

func (ov *ObjectValidator) Validate(vc *validationContext) (result *ValidationResult, unexpected error) {

	r := new(ValidationResult)
	r.ErrorCodes = []string{}

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
		case StringOpRequired:
			//ov.markRequired(field, ops, sv)
		case StringOpStopAll:
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
