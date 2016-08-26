package types

type StringSet interface {
	Contains(m string) bool
	Add(s string)
	Contents() []string
	Size() int
	AddAll(os StringSet)
}

type UnorderedStringSet struct {
	members map[string]bool
}

func (us *UnorderedStringSet) AddAll(ss StringSet) {

	for _, m := range ss.Contents() {
		us.Add(m)
	}
}

func (ss *UnorderedStringSet) Size() int {
	return len(ss.members)
}

func (ss *UnorderedStringSet) Contents() []string {
	c := make([]string, 0)

	for k, _ := range ss.members {
		c = append(c, k)
	}

	return c

}

func (ss *UnorderedStringSet) Contains(m string) bool {
	return ss.members[m]
}

func (ss *UnorderedStringSet) Add(s string) {
	ss.members[s] = true
}

func NewUnorderedStringSet(m []string) *UnorderedStringSet {
	ss := new(UnorderedStringSet)
	ss.members = make(map[string]bool)
	for _, v := range m {
		ss.members[v] = true
	}

	return ss
}

func NewOrderedStringSet(m []string) *OrderedStringSet {
	os := new(OrderedStringSet)

	os.ordered = make([]string, 0)
	os.members = NewUnorderedStringSet(os.ordered)

	for _, v := range m {
		os.Add(v)
	}

	return os
}

type OrderedStringSet struct {
	members *UnorderedStringSet
	ordered []string
}

func (os *OrderedStringSet) AddAll(ss StringSet) {

	for _, m := range ss.Contents() {
		os.Add(m)
	}
}

func (os *OrderedStringSet) Contains(m string) bool {
	return os.members.Contains(m)
}

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

func (os *OrderedStringSet) Contents() []string {
	return os.ordered
}

func (os *OrderedStringSet) Size() int {
	return len(os.ordered)
}
