package test

import (
	"io/ioutil"
	"testing"
)

func TestDefaultConfig(t *testing.T) {

	p, err := FindFacilityConfigFromWD()

	if err != nil {
		t.Fatalf(err.Error())
	}

	files, err := ioutil.ReadDir(p)

	if err != nil {
		t.Fatalf(err.Error())
	}

	seen := false

	for _, f := range files {

		if f.Name() == "active.json" {
			seen = true
			break
		}

	}

	if seen == false {
		t.Fatalf("Could not find file active.json in directory %s", p)
	}

}
