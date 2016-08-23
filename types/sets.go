package types

type StringSet struct {
	members map[string]bool
}

func (ss *StringSet) Contains(m string) bool {
	return ss.members[m]
}

func NewStringSet(m []string) *StringSet {
	ss := new(StringSet)
	ss.members = make(map[string]bool)
	for _, v := range m {
		ss.members[v] = true
	}

	return ss
}
