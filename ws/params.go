// Copyright 2016-2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package ws

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
)

// NewParamsForPath creates a Params used to store the elements of a request
// path extracted using regular expression groups.
func NewParamsForPath(targets []string, values []string) *Params {

	contents := make(url.Values)
	v := len(values)
	var names []string

	for i, k := range targets {

		if i < v {
			contents[strconv.Itoa(i)] = []string{values[i]}
			names = append(names, k)
		}

	}

	p := new(Params)
	p.values = contents
	p.paramNames = names

	return p

}

// NewParamsForQuery creates a Params storing the HTTP query parameters from a request.
func NewParamsForQuery(values url.Values) *Params {

	wp := new(Params)
	wp.values = values

	var names []string

	for k := range values {
		names = append(names, k)
	}

	wp.paramNames = names

	return wp

}

// Params is an abstraction of the HTTP query parameters or path parameters with type-safe accessors.
type Params struct {
	values     url.Values
	paramNames []string
}

// ParamNames returns the names of all of the parameters stored
func (wp *Params) ParamNames() []string {
	return wp.paramNames
}

// NotEmpty returns true if a parameter with the supplied name exists and has a non-empty string representation.
func (wp *Params) NotEmpty(key string) bool {

	if !wp.Exists(key) {
		return false
	}

	s, _ := wp.StringValue(key)

	return s != ""

}

// Exists returns true if a parameter with the supplied name exists, even if that parameter's value is an empty string.
func (wp *Params) Exists(key string) bool {
	return wp.values[key] != nil
}

// MultipleValues returns true if the parameter with the supplied name was set more than once (allowed for HTTP query parameters).
func (wp *Params) MultipleValues(key string) bool {

	value := wp.values[key]

	return value != nil && len(value) > 1

}

// StringValue returns the string representation of the specified parameter or an error if no value exists for that parameter.
func (wp *Params) StringValue(key string) (string, error) {

	s := wp.values[key]

	if s == nil {
		return "", wp.noVal(key)
	}

	return s[len(s)-1], nil

}

// BoolValue returns the bool representation of the specified parameter (using Go's bool conversion rules) or an error if no value exists for that parameter.
func (wp *Params) BoolValue(key string) (bool, error) {

	v := wp.values[key]

	if v == nil {
		return false, wp.noVal(key)
	}

	b, err := strconv.ParseBool(v[len(v)-1])

	return b, err

}

// FloatNValue returns a float representation of the specified parameter with the specified bit size, or an error if no value exists for that parameter or
// if the value could not be converted to a float.
func (wp *Params) FloatNValue(key string, bits int) (float64, error) {

	v := wp.values[key]

	if v == nil {
		return 0.0, wp.noVal(key)
	}

	i, err := strconv.ParseFloat(v[len(v)-1], bits)

	return i, err

}

// IntNValue returns a signed int representation of the specified parameter with the specified bit size, or an error if no value exists for that parameter or
// if the value could not be converted to an int.
func (wp *Params) IntNValue(key string, bits int) (int64, error) {

	v := wp.values[key]

	if v == nil {
		return 0, wp.noVal(key)
	}

	i, err := strconv.ParseInt(v[len(v)-1], 10, bits)

	return i, err

}

// UIntNValue returns an unsigned int representation of the specified parameter with the specified bit size, or an error if no value exists for that parameter or
// if the value could not be converted to an unsigned int.
func (wp *Params) UIntNValue(key string, bits int) (uint64, error) {

	v := wp.values[key]

	if v == nil {
		return 0, wp.noVal(key)
	}

	i, err := strconv.ParseUint(v[len(v)-1], 10, bits)

	return i, err

}

func (wp *Params) noVal(key string) error {
	message := fmt.Sprintf("No value available for key %s", key)
	return errors.New(message)
}
