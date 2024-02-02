package facility

import (
	"context"
	config_access "github.com/graniticio/config-access"
	"github.com/graniticio/granitic/v3/config"
	"github.com/graniticio/granitic/v3/instance"
	"github.com/graniticio/granitic/v3/ioc"
	"github.com/graniticio/granitic/v3/logging"
	"github.com/graniticio/granitic/v3/test"
	"testing"
)

func TestBuilderWithDefaultConfig(t *testing.T) {
	lm := logging.CreateComponentLoggerManager(logging.Fatal, make(map[string]interface{}), []logging.LogWriter{}, logging.NewFrameworkLogMessageFormatter(), false)

	ca, err := configAccessor(lm)

	if err != nil {
		t.Fatalf(err.Error())
	}

	fb := new(ContextFilterBuilder)

	s := new(instance.System)

	//Create the IoC container
	cc := ioc.NewComponentContainer(lm, ca, s)

	fc := new(cfReceiver)

	cc.WrapAndAddProto("cfReceiver", fc)

	err = fb.BuildAndRegister(lm, ca, cc)

	if err != nil {
		t.Fatalf(err.Error())
	}

	if err = cc.Populate(); err != nil {
		t.Fatalf(err.Error())
	}

	if fc.ContextFilter != nil {
		t.Fatalf("Didn't expect context filter to be populated")
	}
}

func TestBuilderWithDefaultConfigSingleFilter(t *testing.T) {
	lm := logging.CreateComponentLoggerManager(logging.Fatal, make(map[string]interface{}), []logging.LogWriter{}, logging.NewFrameworkLogMessageFormatter(), false)

	ca, err := configAccessor(lm)

	if err != nil {
		t.Fatalf(err.Error())
	}

	fb := new(ContextFilterBuilder)

	s := new(instance.System)

	//Create the IoC container
	cc := ioc.NewComponentContainer(lm, ca, s)

	fc := new(cfReceiver)

	cc.WrapAndAddProto("cfReceiver", fc)

	cf := namedPrioritisedFilter{nil, 1, "CF"}

	cc.WrapAndAddProto("cf", &cf)
	err = fb.BuildAndRegister(lm, ca, cc)

	if err != nil {
		t.Fatalf(err.Error())
	}

	if err = cc.Populate(); err != nil {
		t.Fatalf(err.Error())
	}

	if fc.ContextFilter == nil {
		t.Fatalf("Expected context filter to be populated")
	}

	if _, okay := fc.ContextFilter.(*namedPrioritisedFilter); !okay {
		t.Fatalf("Expected filter to be a *namedPrioritisedFilter was a %T", fc.ContextFilter)
	}

}

func TestBuilderWithDefaultConfigMultiFilter(t *testing.T) {
	lm := logging.CreateComponentLoggerManager(logging.Fatal, make(map[string]interface{}), []logging.LogWriter{}, logging.NewFrameworkLogMessageFormatter(), false)

	ca, err := configAccessor(lm)

	if err != nil {
		t.Fatalf(err.Error())
	}

	fb := new(ContextFilterBuilder)

	s := new(instance.System)

	//Create the IoC container
	cc := ioc.NewComponentContainer(lm, ca, s)

	fc := new(cfReceiver)

	cc.WrapAndAddProto("cfReceiver", fc)

	cf := namedPrioritisedFilter{nil, 1, "CF"}

	cc.WrapAndAddProto("cf", &cf)

	cf2 := namedPrioritisedFilter{nil, 0, "CF2"}

	cc.WrapAndAddProto("cf2", &cf2)

	err = fb.BuildAndRegister(lm, ca, cc)

	if err != nil {
		t.Fatalf(err.Error())
	}

	if err = cc.Populate(); err != nil {
		t.Fatalf(err.Error())
	}

	if fc.ContextFilter == nil {
		t.Fatalf("Expected context filter to be populated")
	}

	if _, okay := fc.ContextFilter.(*logging.PrioritisedContextFilter); !okay {
		t.Fatalf("Expected filter to be a logging.PrioritisedContextFilter was a %T", fc.ContextFilter)
	}

}

type cfReceiver struct {
	ContextFilter logging.ContextFilter
}

type namedPrioritisedFilter struct {
	contents logging.FilteredContextData
	priority int64
	name     string
}

func (npf namedPrioritisedFilter) Priority() int64 {
	return npf.priority
}

func (npf namedPrioritisedFilter) Extract(ctx context.Context) logging.FilteredContextData {

	return npf.contents
}

func configAccessor(lm *logging.ComponentLoggerManager, additionalFiles ...string) (config_access.Selector, error) {

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

	return config_access.NewGraniticSelector(mergedJSON), nil

}
