package querymanager

import (
	"github.com/graniticio/granitic/v2/config"
	"github.com/graniticio/granitic/v2/dsquery"
	"github.com/graniticio/granitic/v2/instance"
	"github.com/graniticio/granitic/v2/ioc"
	"github.com/graniticio/granitic/v2/logging"
	"github.com/graniticio/granitic/v2/test"
	"testing"
)

func TestFacilityNaming(t *testing.T) {

	fb := new(FacilityBuilder)

	if fb.FacilityName() != "QueryManager" {
		t.Errorf("Unexpected facility name %s", fb.FacilityName())
	}

}

func TestBuilderWithDefaultConfig(t *testing.T) {
	lm := logging.CreateComponentLoggerManager(logging.Fatal, make(map[string]interface{}), []logging.LogWriter{}, logging.NewFrameworkLogMessageFormatter(), false)

	qf := test.FilePath("valid")

	ca, err := configAccessor(lm, qf)

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

	mc := cc.ComponentByName(QueryManagerComponentName).Instance

	tqm, okay := mc.(*dsquery.TemplatedQueryManager)

	if !okay {
		t.Fatalf("Unexpected type for %s %t", QueryManagerComponentName, mc)
	}

	_, okay = tqm.ValueProcessor.(*dsquery.ConfigurableProcessor)

	if !okay {
		t.Fatalf("Unexpected type for ValueProcessor %t", tqm.ValueProcessor)
	}

	tqm.FrameworkLogger = lm.CreateLogger(QueryManagerComponentName)

	if err = tqm.StartComponent(); err != nil {
		t.Fatalf(err.Error())
	}
}

func configAccessor(lm *logging.ComponentLoggerManager, queryFolder string, additionalFiles ...string) (*config.Accessor, error) {

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

	qmConfig := mergedJSON["QueryManager"].(map[string]interface{})

	qmConfig["TemplateLocation"] = queryFolder

	caLogger := lm.CreateLogger("ca")
	return &config.Accessor{JSONData: mergedJSON, FrameworkLogger: caLogger}, nil

}
