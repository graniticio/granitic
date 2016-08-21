package nillable

import (
	"errors"
	"fmt"
	"strconv"
)

type NillableString struct {
	val string
	set bool
}

func (ns *NillableString) MarshalJSON() ([]byte, error) {

	if ns.set {
		return []byte(ns.val), nil
	} else {
		return nil, nil
	}

}

func (ns *NillableString) UnmarshalJSON(b []byte) error {
	ns.val = string(b)
	ns.set = true

	return nil
}

func (ns *NillableString) Set(v string) {
	ns.val = v
	ns.set = true
}

func NewNillableString(v string) *NillableString {
	ns := new(NillableString)
	ns.Set(v)

	return ns
}

type NillableBool struct {
	val bool
	set bool
}

func (nb *NillableBool) MarshalJSON() ([]byte, error) {

	if nb.set {
		return []byte(strconv.FormatBool(nb.val)), nil
	} else {
		return nil, nil
	}

}

func (nb *NillableBool) UnmarshalJSON(b []byte) error {
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

func (nb *NillableBool) Set(v bool) {
	nb.val = v
	nb.set = true
}

func NewNillableBool(b bool) *NillableBool {
	nb := new(NillableBool)
	nb.Set(b)

	return nb
}

type NillableInt64 struct {
	val int64
	set bool
}

func (ni *NillableInt64) MarshalJSON() ([]byte, error) {

	if ni.set {
		return []byte(strconv.FormatInt(ni.val, 10)), nil
	} else {
		return nil, nil
	}

}

func (ni *NillableInt64) UnmarshalJSON(b []byte) error {
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

func (ni *NillableInt64) Set(v int64) {
	ni.val = v
	ni.set = true
}

func NewNillableInt64(i int64) *NillableInt64 {
	ni := new(NillableInt64)
	ni.Set(i)

	return ni
}

type NillableFloat64 struct {
	val float64
	set bool
}

func (nf *NillableFloat64) MarshalJSON() ([]byte, error) {

	if nf.set {
		return []byte(strconv.FormatFloat(nf.val, 'E', -1, 64)), nil
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

func NewNillableFloat64(f float64) *NillableFloat64 {
	nf := new(NillableFloat64)
	nf.Set(f)

	return nf
}
