// Copyright 2016-2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package types

// A new UnorderedStringSet seeded with the supplied strings.
func NewUnorderedStringSet(m []string) *UnorderedStringSet {
	ss := new(UnorderedStringSet)
	ss.members = make(map[string]bool)
	for _, v := range m {
		ss.members[v] = true
	}

	return ss
}

// An empty OrderedStringSet
func NewEmptyOrderedStringSet() *OrderedStringSet {
	return NewOrderedStringSet([]string{})
}

// An empty UnorderedStringSet
func NewEmptyUnorderedStringSet() *UnorderedStringSet {
	return NewUnorderedStringSet([]string{})
}

// An OrderedStringSet with the supplied strings added to the new set in the provided order.
func NewOrderedStringSet(m []string) *OrderedStringSet {
	os := new(OrderedStringSet)

	os.ordered = make([]string, 0)
	os.members = NewUnorderedStringSet(os.ordered)

	for _, v := range m {
		os.Add(v)
	}

	return os
}

// Common behaviour for an ordered or unordered set of strings.
type StringSet interface {
	// Whether or not the set contains the supplied string.
	Contains(m string) bool

	// Add the supplied string to the set. If the set already contains the supplied value, it is ignored.
	Add(s string)

	// The members of the set a string slice.
	Contents() []string

	// The number of members of the set.
	Size() int

	// Add all the members of the supplied set to this set.
	AddAll(os StringSet)
}

// A set of strings where the order in which the strings were added to the set is not recorded.
//
// This type is not goroutine safe and not recommended for the storage large number of strings.
type UnorderedStringSet struct {
	members map[string]bool
}

// See StringSet.AddAll
func (us *UnorderedStringSet) AddAll(ss StringSet) {

	for _, m := range ss.Contents() {
		us.Add(m)
	}
}

// See StringSet.Size
func (ss *UnorderedStringSet) Size() int {
	return len(ss.members)
}

// Contents returns all of the strings contained in this set in a nondeterministic order
func (ss *UnorderedStringSet) Contents() []string {
	c := make([]string, 0)

	for k, _ := range ss.members {
		c = append(c, k)
	}

	return c

}

// See StringSet.Contains
func (ss *UnorderedStringSet) Contains(m string) bool {
	return ss.members[m]
}

// See StringSet.Enqueue
func (ss *UnorderedStringSet) Add(s string) {
	ss.members[s] = true
}

// A set of strings where the order in which the strings were added to the set is preserved. Calls to Contents will
// return the strings in the same order in which they were added.

// This type is not goroutine safe and not recommended for the storage of large number of strings.
type OrderedStringSet struct {
	members *UnorderedStringSet
	ordered []string
}

// See StringSet.AddAll
func (os *OrderedStringSet) AddAll(ss StringSet) {

	for _, m := range ss.Contents() {
		os.Add(m)
	}
}

// See StringSet.Contains
func (os *OrderedStringSet) Contains(m string) bool {
	return os.members.Contains(m)
}

// See StringSet.Enqueue
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

// See StringSet.Size
func (os *OrderedStringSet) Size() int {
	return len(os.ordered)
}
