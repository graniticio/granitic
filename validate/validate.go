package validate

import (
	"errors"
	"fmt"
	"strings"
)

type ValidationRuleType uint

const (
	UnknownRuleType = iota
	StringRule
)

const commandSep = ":"
const StringRuleCode = "STR"
const RuleRefCode = "RULE"

type Validator interface {
	Validate(field string, object interface{}) (errorCodes []string, unexpected error)
}

type UnparsedRuleRuleManager struct {
	Rules map[string][]string
}

func (rm *UnparsedRuleRuleManager) Exists(ref string) bool {
	return rm.Rules[ref] != nil
}

func (rm *UnparsedRuleRuleManager) Rule(ref string) []string {
	return rm.Rules[ref]
}

type ObjectValidator struct {
	jsonConfig       interface{}
	RuleManager      *UnparsedRuleRuleManager
	stringBuilder    *stringValidatorBuilder
	DefaultErrorCode string
	Rules            [][]string
}

func (ov *ObjectValidator) StartComponent() error {

	ov.stringBuilder = newStringValidatorBuilder(ov.DefaultErrorCode)

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
		err = ov.stringBuilder.parseStringRule(field, rule)
	default:
		m := fmt.Sprintf("Unsupported rule type for field %s\n", field)
		return errors.New(m)
	}

	return err

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
	return strings.SplitN(r, commandSep, -1)
}
