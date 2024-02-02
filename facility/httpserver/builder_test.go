package httpserver

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

func TestFacilityNaming(t *testing.T) {

	fb := new(FacilityBuilder)

	if fb.FacilityName() != "HTTPServer" {
		t.Errorf("Unexpected facility name %s", fb.FacilityName())
	}

}

func TestBuilderWithDefaultConfig(t *testing.T) {
	lm := logging.CreateComponentLoggerManager(logging.Fatal, make(map[string]interface{}), []logging.LogWriter{}, logging.NewFrameworkLogMessageFormatter(), false)

	ca, err := configAccessor(lm, test.FilePath("accesslog.json"))

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

}

func TestBuilderWithJSONConfig(t *testing.T) {
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

}

/*
	func TestBuilderWithAllFieldsJSONConfig(t *testing.T) {
		lm := logging.CreateComponentLoggerManager(logging.Fatal, make(map[string]interface{}), []logging.LogWriter{}, logging.NewFrameworkLogMessageFormatter(), false)

		ca, err := configAccessor(lm, test.FilePath("allfieldvariants.json"))

		if err != nil {
			t.Fatalf(err.Error())
		}

		fb := new(FacilityBuilder)
		cxf := new(testFilter)

		cxf.m = make(logging.FilteredContextData)
		cxf.m["someKey"] = "someVal"

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

		alw := cc.ComponentByName(accessLogWriterName).Instance.(*AccessLogWriter)

		lb := alw.builder
		lb.SetContextFilter(cxf)

		if _, ok := lb.(*JSONLineBuilder); !ok {
			t.Fatalf("Unexpected type of LineBuilder %T", lb)
		}

		ctx := context.Background()

		req := new(http.Request)
		req.URL, _ = url.Parse("http://localhost/some/path?a=b")
		req.Method = "GET"
		req.Proto = "HTTPS"
		req.RequestURI = "/some/path"
		end := time.Now()

		start := end.Add(time.Second * -2)

		rw := responseWriter(true, 200)

		lb.BuildLine(ctx, req, rw, &start, &end)

}
*/
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

type testFilter struct {
	m logging.FilteredContextData
}

func (tf testFilter) Extract(ctx context.Context) logging.FilteredContextData {
	return tf.m
}
