package types

import "testing"

func TestNilableStringJson(t *testing.T) {

	ns := NewNilableString("test")

	b, err := ns.MarshalJSON()

	if err != nil {
		t.FailNow()
	}

	ns2 := new(NilableString)

	err = ns2.UnmarshalJSON(b)

	if err != nil {
		t.FailNow()
	}

	if ns2.String() != "test" {
		t.FailNow()
	}

}
