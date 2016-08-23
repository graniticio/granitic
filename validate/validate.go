package validate

import (
	"encoding/json"
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
	DefaultErrorCode string
}

func (ov *ObjectValidator) UnmarshalJSON(b []byte) error {

	var jc interface{}
	json.Unmarshal(b, &jc)

	ov.jsonConfig = jc

	return nil
}

func (ov *ObjectValidator) StartComponent() error {

	m, found := ov.jsonConfig.(map[string]interface{})

	if !found {
		return errors.New("Unexpected config format for rules (was expecting JSON map/object)\n")
	}

	return ov.parseRules(m)

}

func (ov *ObjectValidator) parseRules(m map[string]interface{}) error {

	var err error

	for f, v := range m {

		var rule []string

		if ov.isNestedField(v) {
			continue
		}

		if ov.isRuleRef(v) {
			rule, err = ov.findRule(f, v.(string))

			if err != nil {
				break
			}

		} else {

			isRule, err := ov.isRule(v)

			if err != nil {
				break
			}

			if isRule {
				rule = ToStringArray(v.([]interface{}))
			}
		}

		if len(rule) == 0 {
			m := fmt.Sprintf("Rule for field %s is empty", f)
			return errors.New(m)
		}

		err = ov.parseRule(f, rule)

		if err != nil {
			break
		}

	}

	return err
}

func (ov *ObjectValidator) isRule(v interface{}) (bool, error) {

	a, found := v.([]interface{})

	if found {

		for _, ve := range a {
			_, foundStr := ve.(string)

			if !foundStr {
				m := fmt.Sprintf("Rule %q contains elements that are not strings\n", v)
				return false, errors.New(m)
			}

		}

	}

	return found, nil
}

func (ov *ObjectValidator) isRuleRef(v interface{}) bool {
	_, found := v.(string)

	return found
}

func (ov *ObjectValidator) isNestedField(v interface{}) bool {
	_, found := v.(map[string]interface{})

	return found
}

func (ov *ObjectValidator) findRule(field, ref string) ([]string, error) {

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
		ov.parseStringRule(field, rule)
	default:
		m := fmt.Sprintf("Unsupported rule type for field %s\n", field)
		return errors.New(m)
	}

	return nil

}

func (ov *ObjectValidator) extractType(field string, rule []string) (ValidationRuleType, error) {

	for _, v := range rule {

		f := DecomposeRule(v)

		if f[0] == StringRuleCode {
			return StringRule, nil
		}

	}

	m := fmt.Sprintf("Unable to determine the type of rule from the rule definition for field %s: %v/n", field, rule)

	return UnknownRuleType, errors.New(m)
}

func (ov *ObjectValidator) parseStringRule(field string, rule []string) error {

	sv := new(StringValidator)
	sv.DefaultErrorcode = ov.determineDefaultErrorCode(StringRuleCode, rule)

	for _, v := range rule {

		f := DecomposeRule(v)
		opCode := f[0]

		op, err := sv.Operation(opCode)

		if err != nil {
			return err
		}

		switch op {
		case StringOpBreak:
			sv.Break()
		}

	}

	return nil

}

func (ov *ObjectValidator) determineDefaultErrorCode(vt string, rule []string) string {
	for _, v := range rule {

		f := DecomposeRule(v)

		if f[0] == vt {
			if len(f) > 1 {
				//Error code must be second component of type
				return f[1]
			}
		}

	}

	return ov.DefaultErrorCode
}

func ToStringArray(v []interface{}) []string {
	sa := make([]string, len(v))

	for i, m := range v {
		sa[i] = m.(string)
	}

	return sa
}

func DecomposeRule(r string) []string {
	return strings.SplitN(r, commandSep, -1)
}
