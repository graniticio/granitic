package binder

import (
	"regexp"
	"testing"
)

func TestModFileRequireShouldMatch(t *testing.T) {

	reqRe := regexp.MustCompile(reqRegex)

	shouldMatch := "github.com/graniticio/granitic/v2 v2.2.0"

	reqMatches := reqRe.FindStringSubmatch(shouldMatch)

	if len(reqMatches) >= 3 {
		majorVersion := reqMatches[1]
		requiredVersion := reqMatches[2]

		if majorVersion != "v2" || requiredVersion != "v2.2.0" {
			t.Fail()
		}
	} else {
		t.Fail()
	}

}

func TestModFileRequireShouldNotMatch(t *testing.T) {

	reqRe := regexp.MustCompile(reqRegex)

	shouldNotMatch := "githubacom/graniticio/granitic/v2 v2.2.0"

	reqMatches := reqRe.FindStringSubmatch(shouldNotMatch)

	if len(reqMatches) > 0 {
		t.Fail()
	}

}
