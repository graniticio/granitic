package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

type NilableString struct {
	val string
	set bool
}

func (ns *NilableString) MarshalJSON() ([]byte, error) {

	if ns.set {
		return json.Marshal(ns.val)
	} else {
		return nil, nil
	}

}

func (ns *NilableString) UnmarshalJSON(b []byte) error {
	ns.val = string(b)
	ns.set = true

	return nil
}

func (ns *NilableString) Set(v string) {
	ns.val = v
	ns.set = true
}

func (ns *NilableString) String() string {
	return ns.val
}

func (ns *NilableString) IsSet() bool {
	return ns.set
}

func NewNilableString(v string) *NilableString {
	ns := new(NilableString)
	ns.Set(v)

	return ns
}

type NilableBool struct {
	val bool
	set bool
}

func (nb *NilableBool) MarshalJSON() ([]byte, error) {

	if nb.set {
		return []byte(strconv.FormatBool(nb.val)), nil
	} else {
		return nil, nil
	}

}

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

func (nb *NilableBool) Set(v bool) {
	nb.val = v
	nb.set = true
}

func (nb *NilableBool) Bool() bool {
	return nb.val
}

func NewNilableBool(b bool) *NilableBool {
	nb := new(NilableBool)
	nb.Set(b)

	return nb
}

type NilableInt64 struct {
	val int64
	set bool
}

func (ni *NilableInt64) MarshalJSON() ([]byte, error) {

	if ni.set {
		return []byte(strconv.FormatInt(ni.val, 10)), nil
	} else {
		return nil, nil
	}

}

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

func (ni *NilableInt64) Set(v int64) {
	ni.val = v
	ni.set = true
}

func (ni *NilableInt64) Int64() int64 {
	return ni.val
}

func NewNilableInt64(i int64) *NilableInt64 {
	ni := new(NilableInt64)
	ni.Set(i)

	return ni
}

type NillableFloat64 struct {
	val float64
	set bool
}

func (nf *NillableFloat64) MarshalJSON() ([]byte, error) {

	if nf.set {
		return json.Marshal(nf.val)
	} else {
		return nil, nil
	}

}

func (nf *NillableFloat64) UnmarshalJSON(b []byte) error {
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

func (nf *NillableFloat64) Set(v float64) {
	nf.val = v
	nf.set = true
}

func (ni *NillableFloat64) Float64() float64 {
	return ni.val
}

func NewNilableFloat64(f float64) *NillableFloat64 {
	nf := new(NillableFloat64)
	nf.Set(f)

	return nf
}
