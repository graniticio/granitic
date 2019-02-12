package dsquery

import (
	"errors"
	"fmt"
	"github.com/graniticio/granitic/v2/types"
)

// ParamValueProcessor is implemented by components able to escape the value of a parameter to a query and handle unset parameters
type ParamValueProcessor interface {
	EscapeParamValue(v *paramValueContext)
	SubstituteUnset(v *paramValueContext) error
}

type paramValueContext struct {
	Key     string
	Value   interface{}
	QueryID string
	Escaped bool
}

// ConfigurableProcessor allows missing parameter values to be substituted for a defined value and for
// string (and nilable string) values to be wrapped with a defined character (e.g. ')
type ConfigurableProcessor struct {
	// Whether string parameters should have their contents wrapped with the string defined in StringWrapWith before
	// being injected into the query. For example, SQL RDBMSs usually require strings to be wrapped with ' or "
	WrapStrings bool

	// Whether or not strings that are set to the same value as DefaultParameterValue should be wrapped
	DisableWrapWhenDefaultParameterValue bool

	// Use the value of 'DefaultParameterValue' instead of returning an error if a parameter required for a query is missing
	UseDefaultForMissingParameter bool

	// The value to use when populating a query if the named value is missing
	DefaultParameterValue interface{}

	// If a default value has been substituted for a missing parameter, prevent further escaping
	EscapeDefaultValues bool

	// A string that will be used as a prefix and suffix to a string parameter if WrapStrings is true.
	StringWrapWith string
}

// EscapeParamValue implements ParamValueProcessor.EscapeParamValue
func (cp *ConfigurableProcessor) EscapeParamValue(v *paramValueContext) {

	switch t := v.Value.(type) {
	case string:
		cp.wrapString(v, t)
	case types.NilableString:
		cp.wrapString(v, t.String())
	case *types.NilableString:
		cp.wrapString(v, t.String())
	}

}

func (cp *ConfigurableProcessor) wrapString(v *paramValueContext, s string) {
	if cp.WrapStrings && (s != cp.DefaultParameterValue || !cp.DisableWrapWhenDefaultParameterValue) {
		v.Value = cp.StringWrapWith + s + cp.StringWrapWith
	}
}

// SubstituteUnset implements ParamValueProcessor.SubstituteUnset
func (cp *ConfigurableProcessor) SubstituteUnset(v *paramValueContext) error {

	if cp.UseDefaultForMissingParameter {

		v.Value = cp.DefaultParameterValue
		v.Escaped = !cp.EscapeDefaultValues
	} else {
		//Substitution of missing parameters not supported
		m := fmt.Sprintf("Parameter %s must be supplied for query %s", v.Key, v.QueryID)
		return errors.New(m)
	}

	return nil
}

// SQLProcessor replaces missing values with the word null, wraps strings with single quotes and
// replaces bool values with the value the BoolTrue and BoolFalse members
type SQLProcessor struct {
	BoolTrue  interface{}
	BoolFalse interface{}
}

// EscapeParamValue modifies the value in the supplied parameter + value so that is beocomes valid SQL
// (e.g. quotes strings, converts to RDBMS specific representation of booleans)
func (sp *SQLProcessor) EscapeParamValue(v *paramValueContext) {
	switch t := v.Value.(type) {
	case string:
		sp.escapeString(v, t)
	case types.NilableString:
		sp.escapeString(v, t.String())
	case *types.NilableString:
		sp.escapeString(v, t.String())
	case bool:
		sp.replaceBool(v, t)
	case types.NilableBool:
		sp.replaceBool(v, t.Bool())
	case *types.NilableBool:
		sp.replaceBool(v, t.Bool())
	}
}

func (sp *SQLProcessor) escapeString(v *paramValueContext, o string) {

	if !v.Escaped {
		v.Value = fmt.Sprintf("'%s'", o)
	}
}

func (sp *SQLProcessor) replaceBool(v *paramValueContext, o bool) {

	if o {
		v.Value = sp.BoolTrue
	} else {
		v.Value = sp.BoolFalse
	}

}

// SubstituteUnset changes the value of a unset parameter value to null
func (sp *SQLProcessor) SubstituteUnset(v *paramValueContext) error {

	v.Value = "null"
	v.Escaped = true

	return nil
}
