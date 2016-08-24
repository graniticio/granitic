package types

type StringSet struct {
	members map[string]bool
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

type OrderedStringSet struct {
	members *StringSet
	ordered []string
}

func (os *OrderedStringSet) Contains(m string) bool {
	return os.members.Contains(m)
}

func (os *OrderedStringSet) Add(s string) {
	if os.members.Contains(s) {
		return
	}

	if os.ordered == nil {
		os.ordered = make([]string, 0)
	}

	os.ordered = append(os.ordered, s)
	os.members.Add(s)
}

func (os *OrderedStringSet) Contents() []string {
	return os.ordered
}
