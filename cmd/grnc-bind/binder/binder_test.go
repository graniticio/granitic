package binder

import "testing"

func TestBinderContains(t *testing.T) {

	b := new(Binder)

	if !b.contains([]string{"A", "B", "C"}, "B") {
		t.Fail()
	}

	if b.contains([]string{"A", "B", "C"}, "D") {
		t.Fail()
	}
}
