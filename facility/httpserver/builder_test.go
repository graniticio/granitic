package httpserver

import "testing"

func TestFacilityNaming(t *testing.T) {

	fb := new(FacilityBuilder)

	if fb.FacilityName() != "HTTPServer" {
		t.Errorf("Unexpected facility name %s", fb.FacilityName())
	}

}
