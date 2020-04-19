// Copyright 2016-2020 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package types

// NewUnorderedStringSet creats a new UnorderedStringSet seeded with the supplied strings.
func NewUnorderedStringSet(m []string) *UnorderedStringSet {
	ss := new(UnorderedStringSet)
	ss.members = make(map[string]bool)
	for _, v := range m {
		ss.members[v] = true
	}

	return ss
}

// NewEmptyOrderedStringSet creates an empty OrderedStringSet
func NewEmptyOrderedStringSet() *OrderedStringSet {
	return NewOrderedStringSet([]string{})
}

// NewEmptyUnorderedStringSet creates an empty UnorderedStringSet
func NewEmptyUnorderedStringSet() *UnorderedStringSet {
	return NewUnorderedStringSet([]string{})
}

// NewOrderedStringSet creates an OrderedStringSet with the supplied strings added to the new set in the provided order.
func NewOrderedStringSet(m []string) *OrderedStringSet {
	os := new(OrderedStringSet)

	os.ordered = make([]string, 0)
	os.members = NewUnorderedStringSet(os.ordered)

	for _, v := range m {
		os.Add(v)
	}

	return os
}

// StringSet defines common behaviour for an ordered or unordered set of strings.
type StringSet interface {
	// Contains returns true if the set contains the supplied string
	Contains(m string) bool

	// Add adds the supplied string to the set. If the set already contains the supplied value, it is ignored.
	Add(s string)

	// Contents returns the members of the set as a string slice
	Contents() []string

	// Size returns the number of members of the set.
	Size() int

	// AddAll adds all the members of the supplied set to this set.
	AddAll(os StringSet)
}

// An UnorderedStringSet is a set of strings where the order in which the strings were added to the set is not recorded.
//
// This type is not goroutine safe and not recommended for the storage large number of strings.
type UnorderedStringSet struct {
	members map[string]bool
}

// AddAll implements StringSet.AddAll
func (us *UnorderedStringSet) AddAll(ss StringSet) {

	for _, m := range ss.Contents() {
		us.Add(m)
	}
}

// Size implements StringSet.Size
func (us *UnorderedStringSet) Size() int {
	return len(us.members)
}

// Contents returns all of the strings contained in this set in a nondeterministic order
func (us *UnorderedStringSet) Contents() []string {
	c := make([]string, 0)

	for k := range us.members {
		c = append(c, k)
	}

	return c

}

// Contains implements StringSet.Contains
func (us *UnorderedStringSet) Contains(m string) bool {
	return us.members[m]
}

// Add implements StringSet.Add
func (us *UnorderedStringSet) Add(s string) {
	us.members[s] = true
}

// An OrderedStringSet is a set of strings where the order in which the strings were added to the set is preserved. Calls to Contents will
// return the strings in the same order in which they were added.
// This type is not goroutine safe and not recommended for the storage of large number of strings.
type OrderedStringSet struct {
	members *UnorderedStringSet
	ordered []string
}

// AddAll implements StringSet.AddAll
func (os *OrderedStringSet) AddAll(ss StringSet) {

	for _, m := range ss.Contents() {
		os.Add(m)
	}
}

// Contains implements StringSet.Contains
func (os *OrderedStringSet) Contains(m string) bool {
	return os.members.Contains(m)
}

// Add implements StringSet.Add
func (os *OrderedStringSet) Add(s string) {

	if os.ordered == nil {
		os.ordered = make([]string, 0)
		os.members = NewUnorderedStringSet(os.ordered)
	} else if os.members.Contains(s) {
		return
	}

	os.ordered = append(os.ordered, s)
	os.members.Add(s)
}

// Contents returns the all of the strings in the set in the same order in which they were added.
func (os *OrderedStringSet) Contents() []string {
	return os.ordered
}

// Size implements StringSet.Size
func (os *OrderedStringSet) Size() int {
	return len(os.ordered)
}
