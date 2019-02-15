package serviceerror

import "testing"

func TestFacilityNaming(t *testing.T) {

	fb := new(FacilityBuilder)

	if fb.FacilityName() != "ServiceErrorManager" {
		t.Errorf("Unexpected facility name %s", fb.FacilityName())
	}

}
