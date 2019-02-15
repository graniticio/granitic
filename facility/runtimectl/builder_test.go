package runtimectl

import "testing"

func TestFacilityNaming(t *testing.T) {

	fb := new(FacilityBuilder)

	if fb.FacilityName() != "RuntimeCtl" {
		t.Errorf("Unexpected facility name %s", fb.FacilityName())
	}

}
