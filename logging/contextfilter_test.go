package logging

import (
	"context"
	"testing"
)

func TestInterfaceImplementation(t *testing.T) {

	f := []ContextFilter{new(PrioritisedContextFilter)}

	if len(f) != 1 {
		t.Fail()
	}

}

func TestMerging(t *testing.T) {

	a := make(FilteredContextData)
	b := make(FilteredContextData)
	c := make(FilteredContextData)

	a["A"] = "A"
	a["AB"] = "A"
	a["ABC"] = "A"

	b["B"] = "B"
	b["AB"] = "B"
	b["BC"] = "B"
	b["ABC"] = "B"

	c["C"] = "C"
	c["ABC"] = "C"
	c["BC"] = "C"

	acf := namedPrioritisedFilter{a, -1, "A"}
	bcf := namedPrioritisedFilter{b, 0, "B"}
	ccf := namedPrioritisedFilter{c, 1, "C"}

	pcf := new(PrioritisedContextFilter)

	pcf.Add(acf)
	pcf.Add(bcf)
	pcf.Add(ccf)

	merged := pcf.Extract(context.Background())

	if merged["A"] != "A" {
		t.Fatalf("Expected A got %s", merged["A"])
	}

	if merged["AB"] != "B" {
		t.Fatalf("Expected B got %s", merged["AB"])
	}

	if merged["ABC"] != "C" {
		t.Fatalf("Expected C got %s", merged["ABC"])
	}

	if merged["C"] != "C" {
		t.Fatalf("Expected C got %s", merged["C"])
	}

	if merged["BC"] != "C" {
		t.Fatalf("Expected C got %s", merged["BC"])
	}
}

func TestSorting(t *testing.T) {

	a := namedPrioritisedFilter{nil, 2, "A"}
	b := namedNonPrioritisedFilter{nil, "B"}
	c := namedPrioritisedFilter{nil, -4, "C"}
	d := namedPrioritisedFilter{nil, 8, "D"}

	pcf := new(PrioritisedContextFilter)

	pcf.Add(c)
	pcf.Add(b)
	pcf.Add(d)
	pcf.Add(a)

	if len(pcf.filters) != 4 {
		t.Fatalf("Unexpected filter count ")
	}

	if exD, okay := pcf.filters[0].(namedPrioritisedFilter); !okay {
		t.Fatalf("Expected a namedPrioritisedFilter a position 0")
	} else {

		if exD.name != "D" {
			t.Fatalf("Expected D at position 0 was %s", exD.name)
		}

	}

	if exA, okay := pcf.filters[1].(namedPrioritisedFilter); !okay {
		t.Fatalf("Expected a namedPrioritisedFilter a position 1")
	} else {

		if exA.name != "A" {
			t.Fatalf("Expected A at position 0 was %s", exA.name)
		}

	}

	if exB, okay := pcf.filters[2].(zeroPriorityWrapper); !okay {
		t.Fatalf("Expected a zeroPriorityWrapper a position 2")
	} else {

		if exB.wrapped.(namedNonPrioritisedFilter).name != "B" {
			t.Fatalf("Expected B at position 0 was %s", exB.wrapped.(namedNonPrioritisedFilter).name)
		}

	}

	if exC, okay := pcf.filters[3].(namedPrioritisedFilter); !okay {
		t.Fatalf("Expected a namedPrioritisedFilter a position 3")
	} else {

		if exC.name != "C" {
			t.Fatalf("Expected C at position 0 was %s", exC.name)
		}

	}
}

type namedPrioritisedFilter struct {
	contents FilteredContextData
	priority int64
	name     string
}

func (npf namedPrioritisedFilter) Priority() int64 {
	return npf.priority
}

func (npf namedPrioritisedFilter) Extract(ctx context.Context) FilteredContextData {

	return npf.contents
}

type namedNonPrioritisedFilter struct {
	contents FilteredContextData
	name     string
}

func (npf namedNonPrioritisedFilter) Extract(ctx context.Context) FilteredContextData {

	return npf.contents
}
