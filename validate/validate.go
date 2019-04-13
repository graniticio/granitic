// Copyright 2016-2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

/*
Package validate provides a declarative, rules-based validation framework for validating user-supplied data.

The types in this package are designed to be used in conjunction with the handler.WsHandler type which your application
will use to represent web service endpoints (see the package documentation for ws, ws/handler and http://granitic.io/ref/web-service-handlers )

The purpose of Granitic's validation framework is to automate as much of the 'boiler-plate' validation from validated the data
supplied with the a web service call. Simple checks for ensuring a field is present, well-formed and within an allowed range
can clutter application code.

Granitic's validation framework and patterns are covered in detail at http://granitic.io/ref/validation but a brief overview of
the key types and concepts follows.

RuleValidator

Each instance of handler.WsHandler in your application has an optional field

		RuleValidator *validate.RuleValidator

If an AutoValidator is provided, its Validate method will be invoked before your handler's Logic.Process method is called. If errors
are detected by the RuleValidator, processing stops and an error response will be served to the web service caller.

A RuleValidator is declared in your component definition file in manner similar to:

	{
	  "createRecordHandler": {
		"type": "handler.WsHandler",
		"HTTPMethod": "POST",
		"Logic": "ref:createRecordLogic",
		"PathPattern": "^/record$",
		"AutoValidator": "ref:createRecordValidator"
	  },

	  "createRecordValidator": {
		"type": "validate.RuleValidator",
		"DefaultErrorCode": "CREATE_RECORD",
		"Rules": "conf:createRecordRules"
	  }
	}

Each RuleValidator requires a set of rules, which must be defined in your application's configuration file like:

	{
	  "createRecordRules": [
		["CatalogRef",  "STR",               "REQ:CATALOG_REF_MISSING", "HARDTRIM",        "BREAK",     "REG:^[A-Z]{3}-[\\d]{6}$:CATALOG_REF"],
		["Name",        "STR:RECORD_NAME",   "REQ",                     "HARDTRIM",        "LEN:1-128"],
		["Artist",      "STR:ARTIST_NAME",   "REQ",                     "HARDTRIM",        "LEN:1-64"],
		["Tracks",      "SLICE:TRACK_COUNT", "LEN:1-100",               "ELEM:trackName"]
	  ]
	}

(The spacing in the example above is to illustrate the components of a rule and has no effect on behaviour.)

Rule structure

Rules consist of three components: a field name, type and one or more operations.

The field name is a field in the WsRequest.Body object that is to be validated.

The type is a shorthand for the the type of the field to be validated (see http://granitic.io/ref/validation)

The operations are either checks that should be performed against the field (length checks, regexs etc), processing
instructions (break processing if the previous check failed) or manipulations of the data to be validated (trim a string, etc). See
http://granitic.io/ref/validation for more detail.

For checks and processing instructions, the order in which they appear in the rule is significant as checks are made from left to right.

Error codes

Error codes determine what error is sent to a web service caller if a check fails. Error codes can be defined in three
levels of granularity - on an operation, on a rule or on a RuleValidator. The most specific error available is always used.

Using the validation framework requires the ServiceErrorManager facility to be enabled (see http://granitic.io/ref/service-error-management)

Sharing rules

Sometimes it is useful for a rule to be defined once and re-used by multiple RuleValidators. This is also required
to use some advanced techniques for deep validation of the elements of a slice. This technique is described in detail at
http://granitic.io/ref/validation rule manager.

Decomposing the application of a rule

The first rule in the example above is:

	["CatalogRef",  "STR",  "REQ:CATALOG_REF_MISSING",  "HARDTRIM", "BREAK", "REG:^[A-Z]{3}-[\\d]{6}$:CATALOG_REF"]

It is a very typical example of a string validation rule and breaks down as follows.

1. The field CatalogRef on the web service's WsRequest.Body will be validated.

2. The field will be treated as a string. Note, no error code is defined with the type so the RuleValidator's DefaultErrorCode will be used.

3. The field is REQuired. If the field is not set, the error CATALOG_REF_MISSING will be included in the eventual response to the web service call.

4. The field will be HARDTRIMmed - the actual value of CatalogRef will be permanently modified to remove leading and trailing spaces before
further validation checks are applied (an alternative TRIM will mean validation occurs on a trimmed copy of the string, but the underlying data
is not permanently modified.

5. If previous checks, in this case the REQ check, failed, processing will BREAK and the next validation rule will be processed.

6. The value of CatalogRef is compared to the regex ^[A-Z]{3}-[\\d]{6}$ If there is no match, the error CATALOG_REF will be included
in the eventual response to the web service call.

Advanced techniques

The Granitic validation framework is deep and flexible and you are encouraged to read the reference at http://granitic.io/ref/validation
Advanced techniques include cross field mutual exclusivity, deep validation of slice elements and cross-field dependencies.

Programmatic creation of rules

It is possible to define rules in your application code. Each type of rule supports a fluent-style interface to make application code more readable in this case. The rule
above could be expressed as

	sv := NewStringValidator("CatalogRef", "CREATE_RECORD").
		Required("CATALOG_REF_MISSING").
		HardTrim().
		Break().
		Regex("^[A-Z]{3}-[\\d]{6}$", "CATALOG_REF")
*/
package validate

import (
	"context"
	"errors"
	"fmt"
	"github.com/graniticio/granitic/v2/ioc"
	"github.com/graniticio/granitic/v2/logging"
	"github.com/graniticio/granitic/v2/types"
	"regexp"
	"strconv"
	"strings"
)

type validationRuleType uint

type parseAndBuild func(string, []string) (ValidationRule, error)

const (
	unknownRuleType = iota
	stringRuleType
	objectRuleType
	intRuleType
	boolRuleType
	floatRuleType
	sliceRuleType
)

const commandSep = ":"
const escapedCommandSep = "::"
const escapedCommandReplace = "||ESC||"
const ruleRefCode = "RULE"

const commonOpRequired = "REQ"
const commonOpStopAll = "STOPALL"
const commonOpIn = "IN"
const commonOpBreak = "BREAK"
const commonOpExt = "EXT"
const commonOpMex = "MEX"
const commonOpLen = "LEN"

const lengthPattern = "^(\\d*)-(\\d*)$"

// SubjectContext is a wrapper for an object (the subject) to be validated
type SubjectContext struct {
	//An instance of a object to be validated.
	Subject interface{}
}

// ValidationContext is a wrapper for, and meta-data about, a field on an object to be validated.
type ValidationContext struct {
	// The object containing the field to be validated OR (if DirectSubject is true) the value to be validated.
	Subject interface{}

	// Other fields on the object that are known to have been set (for mutual exclusivity/dependency checking)
	KnownSetFields types.StringSet

	// Use this field name, rather than the one associated with the rule that this context will be passed to.
	OverrideField string

	//Indicate that the Subject in this context IS the value to be validated, rather than the container of a field.
	DirectSubject bool
}

// ValidationResult contains the result of applying a rule to a field on an object.
type ValidationResult struct {
	// A map of field names to the errors encountered on this field. In the case of non-slice types, there will
	// be only one entry in the map. For slices, there will be an entry for every index into the slice where problems
	// were found.
	ErrorCodes map[string][]string

	// If the field that was to be validated was 'unset' (definition varies by type)
	Unset bool
}

// AddForField captures the name of a field or slice index and the codes of all errors found for that field/index or
// if the field/index previously had errors records, adds these new errors.
func (vr *ValidationResult) AddForField(field string, codes []string) {

	if codes == nil || len(codes) == 0 {
		return
	}

	existing := vr.ErrorCodes[field]

	if existing == nil {
		vr.ErrorCodes[field] = codes
	} else {
		vr.ErrorCodes[field] = append(existing, codes...)
	}
}

// ErrorCount returns the total number of errors recorded in this result (NOT the number of unique error codes encountered).
func (vr *ValidationResult) ErrorCount() int {
	c := 0

	for _, v := range vr.ErrorCodes {

		c += len(v)

	}

	return c
}

// NewValidationResult creates an empty ValidationResult ready to record errors.
func NewValidationResult() *ValidationResult {
	vr := new(ValidationResult)
	vr.ErrorCodes = make(map[string][]string)

	return vr
}

// NewPopulatedValidationResult with the supplied errors recorded against the supplied field name
func NewPopulatedValidationResult(field string, codes []string) *ValidationResult {
	vr := NewValidationResult()
	vr.AddForField(field, codes)

	return vr
}

// ValidationRule is a validation rule associated with a named field on an object able to be applied in a generic fashion. Implementations exist for each type (bool, string etc)
// supported by the validation framework.
type ValidationRule interface {
	// Check the subject provided in the supplied context and return any found errors.
	Validate(vc *ValidationContext) (result *ValidationResult, unexpected error)

	// Tell the coordinating object that no other rules should be procesed if errors are found by this rule.
	StopAllOnFail() bool

	// Summarise all of the unique error codes referenced by checks in this rule.
	CodesInUse() types.StringSet

	// Declare other fields that must be set in order for this field to be valid.
	DependsOnFields() types.StringSet

	// Check whether or not the supplied object is 'set' according to this rule's definition of set.
	IsSet(string, interface{}) (bool, error)
}

type validatorLink struct {
	validationRule ValidationRule
	field          string
}

// UnparsedRuleManager is a container for rules that are shared between multiple RuleValidator instances. The rules
// are stored in their raw and unparsed JSON representation.
type UnparsedRuleManager struct {
	// A map between a name for a rule and the rule's unparsed definition.
	Rules map[string][]string
}

// Exists returns true if a rule with the supplied name exists.
func (rm *UnparsedRuleManager) Exists(ref string) bool {
	return rm.Rules[ref] != nil
}

// Rule returns the unparsed representation of the rule with the supplied name.
func (rm *UnparsedRuleManager) Rule(ref string) []string {
	return rm.Rules[ref]
}

// FieldErrors is a summary of all the errors found while validating an object
type FieldErrors struct {

	// The name of a field, or field[x] where x is a slice index if the field's type was slice
	Field string

	// The errors found on that field.
	ErrorCodes []string
}

// RuleValidator coordinates the parsing and application of rules to validate a specific object. Normally
// an instance of this type is unique to an instance of ws.WsHandler.
type RuleValidator struct {
	// A component able to look up ioc components by their name (normally the container itself)
	ComponentFinder ioc.ComponentLookup

	// The error code used to lookup error definitions if no error code is defined on a rule or rule operation.
	DefaultErrorCode string

	// Do not check to see if there are error definitions for all of the error codes referenced by the RuleValidator and its rules.
	DisableCodeValidation bool

	//Inject by the Granitic framework (an application Logger, not a framework Logger).
	Log logging.Logger

	//A source for rules that are shared across multiple RuleValidators
	RuleManager *UnparsedRuleManager

	//The text representation of rules in the order in which they should be applied.
	Rules [][]string

	jsonConfig             interface{}
	stringBuilder          *stringValidationRuleBuilder
	objectValidatorBuilder *objectValidationRuleBuilder
	boolValidatorBuilder   *boolValidationRuleBuilder
	intValidatorBuilder    *intValidationRuleBuilder
	floatValidatorBuilder  *floatValidationRuleBuilder
	sliceValidatorBuilder  *sliceValidationRuleBuilder
	validatorChain         []*validatorLink
	componentName          string
	codesInUse             types.StringSet
	state                  ioc.ComponentState
}

// ValidateMissing returns true if all error codes reference by this RuleValidator must have corresponding definitions
func (ov *RuleValidator) ValidateMissing() bool {
	return !ov.DisableCodeValidation
}

// Container accepts a reference to the IoC container.
func (ov *RuleValidator) Container(container *ioc.ComponentContainer) {
	ov.ComponentFinder = container
}

// ComponentName implements ComponentNamer.ComponentName
func (ov *RuleValidator) ComponentName() string {
	return ov.componentName
}

// SetComponentName implements ComponentNamer.SetComponentName
func (ov *RuleValidator) SetComponentName(name string) {
	ov.componentName = name
}

// ErrorCodesInUse returns all of the unique error codes used by this validator, its rules and their operations.
func (ov *RuleValidator) ErrorCodesInUse() (codes types.StringSet, sourceName string) {
	return ov.codesInUse, ov.componentName
}

// Validate the object in the supplied  subject according to the rules defined on this RuleValidator. Returns
// a summary of any problems found.
func (ov *RuleValidator) Validate(ctx context.Context, subject *SubjectContext) ([]*FieldErrors, error) {

	log := ov.Log

	fes := make([]*FieldErrors, 0)
	fieldsWithProblems := types.NewOrderedStringSet([]string{})
	unsetFields := types.NewOrderedStringSet([]string{})
	setFields := types.NewOrderedStringSet([]string{})

	for _, vl := range ov.validatorChain {
		f := vl.field
		v := vl.validationRule
		log.LogDebugf("Checking field %s set", f)

		if !ov.parentsOkay(v, fieldsWithProblems, unsetFields) {
			log.LogDebugf("Skipping set check on field %s as one or more parent objects invalid", f)
			continue
		}

		set, err := v.IsSet(f, subject.Subject)

		if err != nil {
			return nil, err
		}

		if set {
			setFields.Add(f)
		} else {
			unsetFields.Add(f)
		}

	}

Rules:
	for _, vl := range ov.validatorChain {

		f := vl.field

		log.LogDebugf("Validating field %s", f)

		vc := new(ValidationContext)
		vc.Subject = subject.Subject
		vc.KnownSetFields = setFields

		v := vl.validationRule

		if !ov.parentsOkay(v, fieldsWithProblems, unsetFields) {
			log.LogDebugf("Skipping field %s as one or more parent objects invalid", f)
			continue
		}

		r, err := vl.validationRule.Validate(vc)

		if err != nil {
			return nil, err
		}

		ec := r.ErrorCodes

		if r.Unset {
			log.LogDebugf("%s is unset", f)
			unsetFields.Add(f)
		}

		l := r.ErrorCount()

		if ec != nil && l > 0 {

			for k, v := range ec {

				fieldsWithProblems.Add(k)
				log.LogDebugf("%s has %d errors", k, l)

				fe := new(FieldErrors)
				fe.Field = k
				fe.ErrorCodes = v

				fes = append(fes, fe)

				if vl.validationRule.StopAllOnFail() {
					log.LogDebugf("Stopping all after problem found with %s", f)
					break Rules
				}
			}

		}

	}

	return fes, nil

}

func (ov *RuleValidator) parentsOkay(v ValidationRule, fieldsWithProblems types.StringSet, unsetFields types.StringSet) bool {

	log := ov.Log

	d := v.DependsOnFields()

	if d == nil || d.Size() == 0 {
		return true
	}

	for _, f := range d.Contents() {

		log.LogTracef("Depends on %s", f)

		if fieldsWithProblems.Contains(f) || unsetFields.Contains(f) {

			log.LogTracef("%s is not okay", f)
			return false
		}

	}

	return true
}

// StartComponent is called by the IoC container. Parses the rules into ValidationRule objects.
func (ov *RuleValidator) StartComponent() error {

	if ov.state != ioc.StoppedState {
		return nil
	}

	ov.state = ioc.StartingState

	if ov.Rules == nil {
		return errors.New("no Rules specified for validator")
	}

	ov.codesInUse = types.NewUnorderedStringSet([]string{})

	if ov.DefaultErrorCode != "" {
		ov.codesInUse.Add(ov.DefaultErrorCode)
	}

	ov.stringBuilder = newStringValidationRuleBuilder(ov.DefaultErrorCode)
	ov.stringBuilder.componentFinder = ov.ComponentFinder

	ov.objectValidatorBuilder = newObjectValidationRuleBuilder(ov.DefaultErrorCode, ov.ComponentFinder)
	ov.boolValidatorBuilder = newBoolValidationRuleBuilder(ov.DefaultErrorCode, ov.ComponentFinder)
	ov.validatorChain = make([]*validatorLink, 0)

	ov.intValidatorBuilder = newIntValidationRuleBuilder(ov.DefaultErrorCode, ov.ComponentFinder)
	ov.floatValidatorBuilder = newFloatValidationRuleBuilder(ov.DefaultErrorCode, ov.ComponentFinder)

	ov.sliceValidatorBuilder = newSliceValidationRuleBuilder(ov.DefaultErrorCode, ov.ComponentFinder, ov)

	return ov.parseRules()

}

func (ov *RuleValidator) parseRules() error {

	var err error

	for _, rule := range ov.Rules {

		var ruleToParse []string

		if len(rule) < 2 {
			m := fmt.Sprintf("Rule is invalid (must have at least an identifier and a type). Supplied rule is: %q", rule)
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

		v, err := ov.parseRule(field, ruleToParse)

		if err == nil {
			ov.addValidator(field, v)
		}

		if err != nil {
			return err
		}

	}

	ov.state = ioc.RunningState

	return err
}

func (ov *RuleValidator) addValidator(field string, v ValidationRule) {

	vl := new(validatorLink)
	vl.field = field
	vl.validationRule = v

	ov.validatorChain = append(ov.validatorChain, vl)

	c := v.CodesInUse()

	if c != nil {
		ov.codesInUse.AddAll(c)
	}

}

func (ov *RuleValidator) isRuleRef(op string) bool {

	s := strings.SplitN(op, commandSep, -1)

	return len(s) == 2 && s[0] == ruleRefCode

}

func (ov *RuleValidator) findRule(field, op string) ([]string, error) {

	ref := strings.SplitN(op, commandSep, -1)[1]

	rf := ov.RuleManager

	if rf == nil {
		m := fmt.Sprintf("Field %s has its rule specified as a reference to an external rule %s, but RuleManager is not set.", field, ref)
		return nil, errors.New(m)

	}

	if !rf.Exists(ref) {
		m := fmt.Sprintf("Field %s has its rule specified as a reference to an external rule %s, but no rule with that reference exists.", field, ref)
		return nil, errors.New(m)
	}

	return rf.Rule(ref), nil
}

func (ov *RuleValidator) parseRule(field string, rule []string) (ValidationRule, error) {

	rt, err := ov.extractType(field, rule)

	if err != nil {
		return nil, err
	}

	var v ValidationRule

	switch rt {
	case stringRuleType:
		v, err = ov.parse(field, rule, ov.stringBuilder.parseRule)
	case objectRuleType:
		v, err = ov.parse(field, rule, ov.objectValidatorBuilder.parseRule)
	case boolRuleType:
		v, err = ov.parse(field, rule, ov.boolValidatorBuilder.parseRule)
	case intRuleType:
		v, err = ov.parse(field, rule, ov.intValidatorBuilder.parseRule)
	case floatRuleType:
		v, err = ov.parse(field, rule, ov.floatValidatorBuilder.parseRule)
	case sliceRuleType:
		v, err = ov.parse(field, rule, ov.sliceValidatorBuilder.parseRule)

	default:
		m := fmt.Sprintf("Unsupported rule type for field %s\n", field)
		return nil, errors.New(m)
	}

	return v, err

}

func (ov *RuleValidator) parse(field string, rule []string, pf parseAndBuild) (ValidationRule, error) {
	v, err := pf(field, rule)

	if err != nil {
		return nil, err
	}

	return v, nil
}

func (ov *RuleValidator) extractType(field string, rule []string) (validationRuleType, error) {

	for _, v := range rule {

		f := decomposeOperation(v)

		switch f[0] {
		case stringRuleCode:
			return stringRuleType, nil
		case objectRuleCode:
			return objectRuleType, nil
		case boolRuleCode:
			return boolRuleType, nil
		case intRuleCode:
			return intRuleType, nil
		case floatRuleCode:
			return floatRuleType, nil
		case sliceRuleCode:
			return sliceRuleType, nil
		}
	}

	m := fmt.Sprintf("Unable to determine the type of rule from the rule definition for field %s: %v", field, rule)

	return unknownRuleType, errors.New(m)
}

func isTypeIndicator(vType, op string) bool {

	return decomposeOperation(op)[0] == vType

}

func determineDefaultErrorCode(vt string, rule []string, defaultCode string) string {
	for _, v := range rule {

		f := decomposeOperation(v)

		if f[0] == vt {
			if len(f) > 1 {
				//Error code must be second component of type
				return f[1]
			}
		}

	}

	return defaultCode
}

func decomposeOperation(r string) []string {

	removeEscaped := strings.Replace(r, escapedCommandSep, escapedCommandReplace, -1)
	split := strings.SplitN(removeEscaped, commandSep, -1)

	decomposed := make([]string, len(split))

	for i, v := range split {
		decomposed[i] = strings.Replace(v, escapedCommandReplace, commandSep, -1)
	}

	return decomposed

}

func determinePathFields(path string) types.StringSet {

	set := types.NewOrderedStringSet([]string{})

	split := strings.SplitN(path, ".", -1)

	l := len(split)

	if l > 1 {

		for i := 1; i < l; i++ {

			set.Add(strings.Join(split[0:i], "."))
		}

	}

	return set
}

func validateExternalOperation(cf ioc.ComponentLookup, field string, ops []string) (int, *ioc.Component, error) {

	if cf == nil {
		m := fmt.Sprintf("Field %s relies on an external component to validate, but no ioc.ComponentLookup is available.", field)
		return 0, nil, errors.New(m)
	}

	pCount, err := paramCount(ops, "External", field, 2, 3)

	if err != nil {
		return pCount, nil, err
	}

	ref := ops[1]
	component := cf.ComponentByName(ref)

	if component == nil {
		m := fmt.Sprintf("No external component named %s available to validate field %s", ref, field)
		return 0, nil, errors.New(m)
	}

	return pCount, component, nil
}

func checkMExFields(mf types.StringSet, vc *ValidationContext, ec types.StringSet, code string) {

	if vc.KnownSetFields == nil || vc.KnownSetFields.Size() == 0 {
		return
	}

	for _, s := range mf.Contents() {

		if vc.KnownSetFields.Contains(s) {
			ec.Add(code)
			break
		}
	}

}

func extractVargs(ops []string, l int) []string {

	if len(ops) == l {
		return []string{ops[l-1]}
	}

	return []string{}
}

func extractLengthParams(field string, vals string, pattern *regexp.Regexp) (min, max int, err error) {

	min = noBound
	max = noBound

	if !pattern.MatchString(vals) {
		m := fmt.Sprintf("Length parameters for field %s are invalid. Values provided: %s", field, vals)
		return min, max, errors.New(m)
	}

	groups := pattern.FindStringSubmatch(vals)

	if groups[1] != "" {
		min, _ = strconv.Atoi(groups[1])
	}

	if groups[2] != "" {
		max, _ = strconv.Atoi(groups[2])
	}

	return min, max, nil
}
