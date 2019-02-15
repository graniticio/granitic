package rdbms

import "testing"

func TestFacilityNaming(t *testing.T) {

	fb := new(FacilityBuilder)

	if fb.FacilityName() != "RdbmsAccess" {
		t.Errorf("Unexpected facility name %s", fb.FacilityName())
	}

}
