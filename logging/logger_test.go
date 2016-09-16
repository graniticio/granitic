package logging

import (
	"github.com/graniticio/granitic/test"
	"testing"
)

func TestThresholdDetection(t *testing.T) {

	g := new(globalLogSource)

	lal := new(LevelAwareLogger)
	lal.global = g

	g.level = All
	lal.localLogThreshhold = All

	test.ExpectBool(t, lal.IsLevelEnabled(Debug), true)

	g.level = Error
	test.ExpectBool(t, lal.IsLevelEnabled(Debug), false)

	lal.localLogThreshhold = Debug

	test.ExpectBool(t, lal.IsLevelEnabled(Debug), true)

	g.level = Trace
	test.ExpectBool(t, lal.IsLevelEnabled(Trace), false)

	lal.localLogThreshhold = All
	test.ExpectBool(t, lal.IsLevelEnabled(Trace), true)

}
