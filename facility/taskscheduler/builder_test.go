package taskscheduler

import "testing"

func TestFacilityNaming(t *testing.T) {

	fb := new(FacilityBuilder)

	if fb.FacilityName() != "TaskScheduler" {
		t.Errorf("Unexpected facility name %s", fb.FacilityName())
	}

}
