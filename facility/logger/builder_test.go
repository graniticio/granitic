package logger

import "testing"

func TestFacilityNaming(t *testing.T) {

	fb := new(FacilityBuilder)

	if fb.FacilityName() != "ApplicationLogging" {
		t.Errorf("Unexpected facility name %s", fb.FacilityName())
	}

}
