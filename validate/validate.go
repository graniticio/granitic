package validate

import (
	"errors"
	"fmt"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/types"
	"strings"
)

type ValidationRuleType uint

type parseAndBuild func(string, []string) (Validator, error)

const (
	UnknownRuleType = iota
	StringRule
)

const commandSep = ":"
const escapedCommandSep = "::"
const escapedCommandReplace = "||ESC||"
const StringRuleCode = "STR"
const RuleRefCode = "RULE"

type SubjectContext struct {
	Subject        interface{}
	KnownSetFields types.StringSet
}

type validationContext struct {
	Field          string
	Subject        interface{}
	KnownSetFields types.StringSet
}

type Validator interface {
	Validate(vc *validationContext) (errorCodes []string, unexpected error)
	StopAllOnFail() bool
	CodesInUse() types.StringSet
}

type validatorLink struct {
	validator Validator
	field     string
}

type UnparsedRuleManager struct {
	Rules map[string][]string
}

func (rm *UnparsedRuleManager) Exists(ref string) bool {
	return rm.Rules[ref] != nil
}

func (rm *UnparsedRuleManager) Rule(ref string) []string {
	return rm.Rules[ref]
}

type FieldErrors struct {
	Field      string
	ErrorCodes []string
}

type ObjectValidator struct {
	jsonConfig       interface{}
	RuleManager      *UnparsedRuleManager
	stringBuilder    *stringValidatorBuilder
	DefaultErrorCode string
	Rules            [][]string
	ComponentFinder  ioc.ComponentByNameFinder
	validatorChain   []*validatorLink
	componentName    string
	codesInUse       types.StringSet
}

func (ov *ObjectValidator) Container(container *ioc.ComponentContainer) {
	ov.ComponentFinder = container
}

func (ov *ObjectValidator) ComponentName() string {
	return ov.componentName
}

func (ov *ObjectValidator) SetComponentName(name string) {
	ov.componentName = name
}

func (ov *ObjectValidator) ErrorCodesInUse() (codes types.StringSet, sourceName string) {
	return ov.codesInUse, ov.componentName
}

func (ov *ObjectValidator) Validate(subject *SubjectContext) ([]*FieldErrors, error) {

	fieldErrors := make([]*FieldErrors, 0)

	for _, vl := range ov.validatorChain {

		vc := new(validationContext)
		vc.Field = vl.field
		vc.Subject = subject.Subject
		vc.KnownSetFields = subject.KnownSetFields

		ec, err := vl.validator.Validate(vc)

		if err != nil {
			return nil, err
		}

		if ec != nil && len(ec) > 0 {
			fe := new(FieldErrors)
			fe.Field = vl.field
			fe.ErrorCodes = ec

			fieldErrors = append(fieldErrors, fe)

			if vl.validator.StopAllOnFail() {
				break
			}

		}

	}

	return fieldErrors, nil

}

func (ov *ObjectValidator) StartComponent() error {

	if ov.Rules == nil {
		return errors.New("No Rules specified for validator.")
	}

	ov.codesInUse = types.NewUnorderedStringSet([]string{})

	if ov.DefaultErrorCode != "" {
		ov.codesInUse.Add(ov.DefaultErrorCode)
	}

	ov.stringBuilder = newStringValidatorBuilder(ov.DefaultErrorCode)
	ov.stringBuilder.componentFinder = ov.ComponentFinder

	ov.validatorChain = make([]*validatorLink, 0)

	return ov.parseRules()

}

func (ov *ObjectValidator) parseRules() error {

	var err error

	for _, rule := range ov.Rules {

		var ruleToParse []string

		if len(rule) < 2 {
			m := fmt.Sprintf("Rule is invlaid (must have at least an identifier and a type). Supplied rule is: %q", rule)
			return errors.New(m)
		}

		field := rule[0]
		ruleType := rule[1]

		if ov.isRuleRef(ruleType) {
			ruleToParse, err = ov.findRule(field, ruleType)

			if err != nil {
				break
			}

		} else {
			ruleToParse = rule[1:]
		}

		err = ov.parseRule(field, ruleToParse)

		if err != nil {
			break
		}

	}

	return err
}

func (ov *ObjectValidator) addValidator(field string, v Validator) {

	vl := new(validatorLink)
	vl.field = field
	vl.validator = v

	ov.validatorChain = append(ov.validatorChain, vl)

	c := v.CodesInUse()

	if c != nil {
		ov.codesInUse.AddAll(c)
	}

}

func (ov *ObjectValidator) isRuleRef(op string) bool {

	s := strings.SplitN(op, commandSep, -1)

	return len(s) == 2 && s[0] == RuleRefCode

}

func (ov *ObjectValidator) findRule(field, op string) ([]string, error) {

	ref := strings.SplitN(op, commandSep, -1)[1]

	rf := ov.RuleManager

	if rf == nil {
		m := fmt.Sprintf("Field %s has its rule specified as a reference to an external rule %s, but RuleManager is not set.\n", field, ref)
		return nil, errors.New(m)

	}

	if !rf.Exists(ref) {
		m := fmt.Sprintf("Field %s has its rule specified as a reference to an external rule %s, but no rule with that reference exists.\n", field, ref)
		return nil, errors.New(m)
	}

	return rf.Rule(ref), nil
}

func (ov *ObjectValidator) parseRule(field string, rule []string) error {

	rt, err := ov.extractType(field, rule)

	if err != nil {
		return err
	}

	switch rt {
	case StringRule:
		err = ov.parseAndAdd(field, rule, ov.stringBuilder.parseStringRule)

	default:
		m := fmt.Sprintf("Unsupported rule type for field %s\n", field)
		return errors.New(m)
	}

	return err

}

func (ov *ObjectValidator) parseAndAdd(field string, rule []string, pf parseAndBuild) error {
	v, err := pf(field, rule)

	if err != nil {
		return err
	} else {
		ov.addValidator(field, v)
		return nil
	}
}

func (ov *ObjectValidator) extractType(field string, rule []string) (ValidationRuleType, error) {

	for _, v := range rule {

		f := DecomposeOperation(v)

		if f[0] == StringRuleCode {
			return StringRule, nil
		}

	}

	m := fmt.Sprintf("Unable to determine the type of rule from the rule definition for field %s: %v/n", field, rule)

	return UnknownRuleType, errors.New(m)
}

func IsTypeIndicator(vType, op string) bool {

	return DecomposeOperation(op)[0] == vType

}

func DetermineDefaultErrorCode(vt string, rule []string, defaultCode string) string {
	for _, v := range rule {

		f := DecomposeOperation(v)

		if f[0] == vt {
			if len(f) > 1 {
				//Error code must be second component of type
				return f[1]
			}
		}

	}

	return defaultCode
}

func DecomposeOperation(r string) []string {

	removeEscaped := strings.Replace(r, escapedCommandSep, escapedCommandReplace, -1)
	split := strings.SplitN(removeEscaped, commandSep, -1)

	decomposed := make([]string, len(split))

	for i, v := range split {
		decomposed[i] = strings.Replace(v, escapedCommandReplace, commandSep, -1)
	}

	return decomposed

}
