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
	lm := logging.CreateComponentLoggerManager(logging.Fatal, make(map[string]interface{}), []logging.LogWriter{}, logging.NewFrameworkLogMessageFormatter(), false)

	ca, err := configAccessor(lm)

	if err != nil {
		t.Fatalf(err.Error())
	}

	fb := new(FacilityBuilder)

	s := new(instance.System)

	//Create the IoC container
	cc := ioc.NewComponentContainer(lm, ca, s)

	err = fb.BuildAndRegister(lm, ca, cc)

	if err != nil {
		t.Fatalf(err.Error())
	}

	if err = cc.Populate(); err != nil {
		t.Fatalf(err.Error())
	}

	mc := cc.ComponentByName(applicationLoggingManagerName).Instance

	if err = mc.(*logging.ComponentLoggerManager).StartComponent(); err != nil {
		t.Fatalf(err.Error())
	}
}

func TestBuilderWithJSONLoggingConfig(t *testing.T) {
	lm := logging.CreateComponentLoggerManager(logging.Fatal, make(map[string]interface{}), []logging.LogWriter{}, logging.NewFrameworkLogMessageFormatter(), false)

	ca, err := configAccessor(lm, test.FilePath("structured.json"))

	if err != nil {
		t.Fatalf(err.Error())
	}

	fb := new(FacilityBuilder)

	s := new(instance.System)

	//Create the IoC container
	cc := ioc.NewComponentContainer(lm, ca, s)

	err = fb.BuildAndRegister(lm, ca, cc)

	if err != nil {
		t.Fatalf(err.Error())
	}

	if err = cc.Populate(); err != nil {
		t.Fatalf(err.Error())
	}

	mc := cc.ComponentByName(applicationLoggingManagerName).Instance

	if err = mc.(*logging.ComponentLoggerManager).StartComponent(); err != nil {
		t.Fatalf(err.Error())
	}

}

func TestDefaultJSONFieldConfig(t *testing.T) {
	lm := logging.CreateComponentLoggerManager(logging.Fatal, make(map[string]interface{}), []logging.LogWriter{}, logging.NewFrameworkLogMessageFormatter(), false)

	ca, err := configAccessor(lm)

	if err != nil {
		t.Fatalf(err.Error())
	}

	cfg := new(logging.JSONConfig)

	ca.Populate("LogWriting.Format.JSON", cfg)

	if len(cfg.Fields) != 1 {
		t.Fatalf("Unexpected number of JSON fields in default configuration %d", len(cfg.Fields))
	}

	if cfg.Prefix != "" {
		t.Fatalf("Unexpected prefix value %s", cfg.Prefix)
	}

	if cfg.Suffix != "\n" {
		t.Fatalf("Unexpected suffix value %s", cfg.Suffix)
	}

}

func configAccessor(lm *logging.ComponentLoggerManager, additionalFiles ...string) (*config.Accessor, error) {

	jm := config.NewJSONMergerWithManagedLogging(lm, new(config.JSONContentParser))

	configLoc, err := test.FindFacilityConfigFromWD()

	if err != nil {
		return nil, err
	}

	jf, err := config.FindJSONFilesInDir(configLoc)

	for _, f := range additionalFiles {

		jf = append(jf, f)

	}

	if err != nil {
		return nil, err
	}

	mergedJSON, err := jm.LoadAndMergeConfigWithBase(make(map[string]interface{}), jf)

	if err != nil {
		return nil, err
	}

	caLogger := lm.CreateLogger("ca")
	return &config.Accessor{JSONData: mergedJSON, FrameworkLogger: caLogger}, nil

}
