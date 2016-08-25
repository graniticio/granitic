package types

type Set interface {
	Contains(m string) bool
	Add(s string)
	Contents() []string
	Size() int
}

type StringSet struct {
	members map[string]bool
}

func (ss *StringSet) Size() int {
	return len(ss.members)
}

func (ss *StringSet) Contents() []string {
	c := make([]string, 0)

	for k, _ := range ss.members {
		c = append(c, k)
	}

	return c

}

func (ss *StringSet) Contains(m string) bool {
	return ss.members[m]
}

func (ss *StringSet) Add(s string) {
	ss.members[s] = true
}

func NewStringSet(m []string) *StringSet {
	ss := new(StringSet)
	ss.members = make(map[string]bool)
	for _, v := range m {
		ss.members[v] = true
	}

	return ss
}

func NewOrderedStringSet(m []string) *OrderedStringSet {
	os := new(OrderedStringSet)
	for _, v := range m {
		os.Add(v)
	}

	return os
}

type OrderedStringSet struct {
	members *StringSet
	ordered []string
}

func (os *OrderedStringSet) Contains(m string) bool {
	return os.members.Contains(m)
}

func (os *OrderedStringSet) Add(s string) {

	if os.ordered == nil {
		os.ordered = make([]string, 0)
		os.members = NewStringSet(os.ordered)
	} else if os.members.Contains(s) {
		return
	}

	os.ordered = append(os.ordered, s)
	os.members.Add(s)
}

func (os *OrderedStringSet) Contents() []string {
	return os.ordered
}

func (os *OrderedStringSet) Size() int {
	return len(os.ordered)
}
