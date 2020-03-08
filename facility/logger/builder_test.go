package logger

import (
	"github.com/graniticio/granitic/v2/config"
	"github.com/graniticio/granitic/v2/instance"
	"github.com/graniticio/granitic/v2/ioc"
	"github.com/graniticio/granitic/v2/logging"
	"github.com/graniticio/granitic/v2/test"
	"testing"
)

func TestFacilityNaming(t *testing.T) {

	fb := new(FacilityBuilder)

	if fb.FacilityName() != "ApplicationLogging" {
		t.Errorf("Unexpected facility name %s", fb.FacilityName())
	}

}

func TestBuilderWithDefaultConfig(t *testing.T) {
	lm := logging.CreateComponentLoggerManager(logging.Fatal, make(map[string]interface{}), []logging.LogWriter{}, logging.NewFrameworkLogMessageFormatter())

	jm := config.NewJSONMergerWithManagedLogging(lm, new(config.JSONContentParser))

	configLoc, err := test.FindFacilityConfigFromWD()

	if err != nil {
		t.Fatalf(err.Error())
	}

	jf, err := config.FindJSONFilesInDir(configLoc)

	if err != nil {
		t.Fatalf(err.Error())
	}

	mergedJSON, err := jm.LoadAndMergeConfigWithBase(make(map[string]interface{}), jf)

	if err != nil {
		t.Fatalf(err.Error())
	}

	caLogger := lm.CreateLogger("ca")
	ca := &config.Accessor{JSONData: mergedJSON, FrameworkLogger: caLogger}
	fb := new(FacilityBuilder)

	s := new(instance.System)

	//Create the IoC container
	cc := ioc.NewComponentContainer(lm, ca, s)

	err = fb.BuildAndRegister(lm, ca, cc)

	if err != nil {
		t.Fatalf(err.Error())
	}

}
