// Copyright 2016-2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

/*
Package types provides versions of Go's built-in types that better support web-services.

Serialising data into and out of some of Go's native types can be problematic. This is because the zero value of numbers,
bools and strings are not nil, but 0, false and "" respectively. This means the intent of the original data can be lost. Did the caller provider 'false' or just forget to specify a value?

A similar problem is solved with Go's sql.NullXXX types for handling null values in and out of databases, but those types
are not suitable for use with web services.

Grantic defines a set of four 'nilable' types for handling int64, float64, bool and string values that might not always
have a value associated with them. There is deep support for these types throughout Granitic including JSON and XML
marhsalling/unmarshalling, path and query parameter binding, validation, query templating and RDBMS access. Developers
are strongly encouraged to use nilable types instead of native types wherever possible.

This package also defines a number of simple implmentations of a 'set'. Caution should be used when using these types in
your own application as they are not goroutine safe or intended to store large numbers of strings.

*/
package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

// Implemented by a type that acts as a wrapper round a native type to track whether a value has actually been set.
type Nilable interface {

	// Convert the contained value to JSON or nil if no value is set.
	MarshalJSON() ([]byte, error)

	// Populate the type with the supplied JSON value or ignore if value is JSON null
	UnmarshalJSON(b []byte) error

	// Whether or not the value in this type was explicitly set
	IsSet() bool
}

// Create a new NilableString with the supplied value.
func NewNilableString(v string) *NilableString {
	ns := new(NilableString)
	ns.Set(v)

	return ns
}

// Create a new NilableBool with the supplied value.
func NewNilableBool(b bool) *NilableBool {
	nb := new(NilableBool)
	nb.Set(b)

	return nb
}

// Create a new NilableInt64 with the supplied value.
func NewNilableInt64(i int64) *NilableInt64 {
	ni := new(NilableInt64)
	ni.Set(i)

	return ni
}

// Create a new NilableFloat64 with the supplied value.
func NewNilableFloat64(f float64) *NilableFloat64 {
	nf := new(NilableFloat64)
	nf.Set(f)

	return nf
}

// A string where it can be determined if "" is an explicitly set value, or just the default zero value
type NilableString struct {
	val string
	set bool
}

// See Nilable.MarshalJSON
func (ns *NilableString) MarshalJSON() ([]byte, error) {

	if ns.set {
		return json.Marshal(ns.val)
	}

	return nil, nil
}

// See Nilable.UnmarshalJSON
func (ns *NilableString) UnmarshalJSON(b []byte) error {
	json.Unmarshal(b, &ns.val)
	ns.set = true

	return nil
}

// Set sets the contained value to the supplied value and makes IsSet true even if the supplied value is the empty
// string.
func (ns *NilableString) Set(v string) {
	ns.val = v
	ns.set = true
}

// The currently stored value (whether or not it has been explicitly set).
func (ns *NilableString) String() string {
	return ns.val
}

// See Nilable.IsSet
func (ns *NilableString) IsSet() bool {
	return ns.set
}

// A bool where it can be determined if false is an explicitly set value, or just the default zero value.
type NilableBool struct {
	val bool
	set bool
}

// See Nilable.MarshalJSON
func (nb *NilableBool) MarshalJSON() ([]byte, error) {

	if nb.set {
		return []byte(strconv.FormatBool(nb.val)), nil
	}

	return nil, nil
}

// See Nilable.UnmarshalJSON
func (nb *NilableBool) UnmarshalJSON(b []byte) error {
	s := string(b)

	if s == "true" {
		nb.val = true
	} else if s == "false" {
		nb.val = false
	} else {
		m := fmt.Sprintf("%s is not a JSON bool value (true or false)", s)
		return errors.New(m)
	}

	nb.set = true

	return nil
}

// Set sets the contained value to the supplied value and makes IsSet true even if the supplied value is false.
func (nb *NilableBool) Set(v bool) {
	nb.val = v
	nb.set = true
}

// See Nilable.IsSet
func (nb *NilableBool) IsSet() bool {
	return nb.set
}

// The currently stored value (whether or not it has been explicitly set).
func (nb *NilableBool) Bool() bool {
	return nb.val
}

// An int64 where it can be determined if 0 is an explicitly set value, or just the default zero value.
type NilableInt64 struct {
	val int64
	set bool
}

// See Nilable.IsSet
func (ni *NilableInt64) IsSet() bool {
	return ni.set
}

// See Nilable.MarshalJSON
func (ni *NilableInt64) MarshalJSON() ([]byte, error) {

	if ni.set {
		return []byte(strconv.FormatInt(ni.val, 10)), nil
	}

	return nil, nil
}

// See Nilable.UnmarshalJSON
func (ni *NilableInt64) UnmarshalJSON(b []byte) error {
	s := string(b)

	v, err := strconv.ParseInt(s, 10, 64)

	if err != nil {
		m := fmt.Sprintf("%s cannot be parsed as an int64", s)
		return errors.New(m)
	}

	ni.val = v
	ni.set = true

	return nil
}

// Set sets the contained value to the supplied value and makes IsSet true even if the supplied value is 0.
func (ni *NilableInt64) Set(v int64) {
	ni.val = v
	ni.set = true
}

// The currently stored value (whether or not it has been explicitly set).
func (ni *NilableInt64) Int64() int64 {
	return ni.val
}

// An float64 where it can be determined if 0 is an explicitly set value, or just the default zero value.
type NilableFloat64 struct {
	val float64
	set bool
}

// See Nilable.IsSet
func (nf *NilableFloat64) IsSet() bool {
	return nf.set
}

// See Nilable.MarshalJSON
func (nf *NilableFloat64) MarshalJSON() ([]byte, error) {

	if nf.set {
		return json.Marshal(nf.val)
	}

	return nil, nil
}

// See Nilable.UnmarshalJSON
func (nf *NilableFloat64) UnmarshalJSON(b []byte) error {
	s := string(b)
	v, err := strconv.ParseFloat(s, 64)

	if err != nil {
		m := fmt.Sprintf("%s cannot be parsed as a float64", s)

		return errors.New(m)
	}

	nf.val = v
	nf.set = true

	return nil
}

// Set sets the contained value to the supplied value and makes IsSet true even if the supplied value is 0.
func (nf *NilableFloat64) Set(v float64) {
	nf.val = v
	nf.set = true
}

// The currently stored value (whether or not it has been explicitly set).
func (nf *NilableFloat64) Float64() float64 {
	return nf.val
}
