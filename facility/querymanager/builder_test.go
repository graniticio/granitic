package querymanager

import "testing"

func TestFacilityNaming(t *testing.T) {

	fb := new(FacilityBuilder)

	if fb.FacilityName() != "QueryManager" {
		t.Errorf("Unexpected facility name %s", fb.FacilityName())
	}

}
