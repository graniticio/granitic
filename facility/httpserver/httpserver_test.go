package httpserver

import (
	"context"
	"fmt"
	"github.com/graniticio/granitic/v2/httpendpoint"
	"github.com/graniticio/granitic/v2/instance"
	"github.com/graniticio/granitic/v2/instrument"
	"github.com/graniticio/granitic/v2/ioc"
	"github.com/graniticio/granitic/v2/logging"
	"github.com/graniticio/granitic/v2/test"
	"github.com/graniticio/granitic/v2/ws"
	"net"
	"net/http"
	"testing"
)

func TestServerStart(t *testing.T) {

	s := new(HTTPServer)
	s.FrameworkLogger = new(logging.ConsoleErrorLogger)

	s.SetProvidersManually(map[string]httpendpoint.Provider{})

	s.AbnormalStatusWriter = new(mockAsw)

	if err := s.StartComponent(); err != nil {
		t.Fatalf(err.Error())
	}

	s.Suspend()
	s.Resume()

	s.PrepareToStop()

	s.ReadyToStop()
	s.Stop()

}

func TestInvalidStateDetection(t *testing.T) {

	s := new(HTTPServer)
	s.FrameworkLogger = new(logging.ConsoleErrorLogger)

	s.SetProvidersManually(map[string]httpendpoint.Provider{})

	s.AbnormalStatusWriter = new(mockAsw)

	if err := s.StartComponent(); err != nil {
		t.Fatalf(err.Error())
	}

	if err := s.StartComponent(); err == nil {
		t.Errorf("Allowed start to be called twice")
	}

}

func TestManualProviderInjection(t *testing.T) {
	p := newMockProvider("GET", ".*")

	s := buildDefaultConfigServer(t, []httpendpoint.Provider{})
	s.AutoFindHandlers = false

	m := map[string]httpendpoint.Provider{"GET": p}

	s.SetProvidersManually(m)
	s.AbnormalStatusWriter = new(mockAsw)

	if err := s.StartComponent(); err != nil {
		t.Fatalf(err.Error())
	}
}

func TestServerFailsWithUnparseableProviderRegex(t *testing.T) {
	p := []httpendpoint.Provider{
		newMockProvider("GET", "\\")}

	s := buildDefaultConfigServer(t, p)

	s.AbnormalStatusWriter = new(mockAsw)

	if err := s.StartComponent(); err == nil {
		t.Fatalf("Expected failure with illegal regex")
	}
}

func TestUnrecognised404(t *testing.T) {

	p := []httpendpoint.Provider{}

	s := buildDefaultConfigServer(t, p)

	defer s.Stop()

	s.AbnormalStatusWriter = &mockAsw{code: 404}

	if err := s.StartComponent(); err != nil {
		t.Fatalf(err.Error())
	}

	if err := s.AllowAccess(); err != nil {
		t.Errorf("Failed to allow access %s", err.Error())
	}

	uri := fmt.Sprintf("http://localhost:%d/nomatch", s.Port)

	_, err := http.Get(uri)

	if err != nil {
		t.Errorf(err.Error())
	}

	/*if r.StatusCode != 404 {
		t.Error("Expected 404")
		fmt.Println(r.StatusCode)
	}*/

}

func TestMatchedRequest(t *testing.T) {

	mp := newMockProvider("GET", ".*")
	p := []httpendpoint.Provider{mp}

	s := buildDefaultConfigServer(t, p)

	defer s.Stop()

	s.AbnormalStatusWriter = &mockAsw{code: 404}

	if err := s.StartComponent(); err != nil {
		t.Fatalf(err.Error())
	}

	if err := s.AllowAccess(); err != nil {
		t.Errorf("Failed to allow access %s", err.Error())
	}

	uri := fmt.Sprintf("http://localhost:%d/match", s.Port)

	_, err := http.Get(uri)

	if err != nil {
		t.Errorf(err.Error())
	}

	if !mp.called {
		t.Errorf("Expected provider to have been called")
	}

}

func TestVersionMatchedRequest(t *testing.T) {

	mp := newMockProvider("GET", ".*")
	mp.supported = []string{"1.0.0"}
	mp.versionEnabled = true
	p := []httpendpoint.Provider{mp}

	s := buildDefaultConfigServer(t, p)

	defer s.Stop()

	ve := new(mockRequestedVersionExtractor)

	s.AbnormalStatusWriter = &mockAsw{code: 404}
	s.VersionExtractor = ve
	s.AllowEarlyInstrumentation = true

	if err := s.StartComponent(); err != nil {
		t.Fatalf(err.Error())
	}

	if err := s.AllowAccess(); err != nil {
		t.Errorf("Failed to allow access %s", err.Error())
	}

	uri := fmt.Sprintf("http://localhost:%d/match", s.Port)

	client := &http.Client{}
	req, _ := http.NewRequest("GET", uri, nil)

	req.Header.Set("version", "1.0.0")

	_, err := client.Do(req)

	if err != nil {
		t.Errorf(err.Error())
	}

	if !mp.called {
		t.Errorf("Expected provider to have been called")
	}

	if !ve.called {
		t.Errorf("Expected version extractor to have been called")
	}

}

func BenchmarkMinimalRequest(b *testing.B) {

	mp := newMockProvider("GET", ".*")
	p := []httpendpoint.Provider{mp}

	s := buildDefaultConfigServer(new(testing.T), p)
	s.AbnormalStatusWriter = &mockAsw{code: 404}

	defer s.Stop()

	if err := s.StartComponent(); err != nil {
		fmt.Println(err.Error())
	}

	if err := s.AllowAccess(); err != nil {
		fmt.Println(err.Error())
	}

	s.AbnormalStatusWriter = &mockAsw{code: 404}

	uri := fmt.Sprintf("http://localhost:%d/match", s.Port)

	req, _ := http.NewRequest("GET", uri, nil)

	req.Header.Set("version", "1.0.0")

	rw := new(mockResponseWriter)

	for i := 0; i < b.N; i++ {
		s.handleAll(rw, req)
	}
}

func TestIDContextExtraction(t *testing.T) {

	mp := newMockProvider("GET", ".*")
	p := []httpendpoint.Provider{mp}

	icb := new(mockIDContextBuilder)
	icb.id = "ID"

	s := buildDefaultConfigServer(t, p)
	s.IDContextBuilder = icb

	defer s.Stop()

	s.AbnormalStatusWriter = &mockAsw{code: 404}

	if err := s.StartComponent(); err != nil {
		t.Fatalf(err.Error())
	}

	if err := s.AllowAccess(); err != nil {
		t.Errorf("Failed to allow access %s", err.Error())
	}

	uri := fmt.Sprintf("http://localhost:%d/match", s.Port)

	_, err := http.Get(uri)

	if err != nil {
		t.Errorf(err.Error())
	}

	if !icb.called {
		t.Errorf("Expected context builder to have been called")
	}

	icb.fail = true

	uri = fmt.Sprintf("http://localhost:%d/match", s.Port)

	_, err = http.Get(uri)

	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestInstrumentationHooks(t *testing.T) {

	mp := newMockProvider("GET", ".*")
	p := []httpendpoint.Provider{mp}

	icb := new(mockIDContextBuilder)
	icb.id = "ID"

	rim := new(mockRequestInstrumentationManager)

	s := buildDefaultConfigServer(t, p)
	s.IDContextBuilder = icb
	s.InstrumentationManager = rim

	defer s.Stop()

	s.AbnormalStatusWriter = &mockAsw{code: 404}

	if err := s.StartComponent(); err != nil {
		t.Fatalf(err.Error())
	}

	if err := s.AllowAccess(); err != nil {
		t.Errorf("Failed to allow access %s", err.Error())
	}

	uri := fmt.Sprintf("http://localhost:%d/match", s.Port)

	_, err := http.Get(uri)

	if err != nil {
		t.Errorf(err.Error())
	}

	inst := rim.i

	if !inst.started {
		t.Errorf("Expected instrumentation event to have been started")
	}

	if !inst.ended {
		t.Errorf("Expected instrumentation event to have been ended")
	}

	if !inst.Called(instrument.RequestID) {
		t.Errorf("Expected request ID to have been passed")
	}

}

func TestServerStartDefaultConfig(t *testing.T) {

	p := []httpendpoint.Provider{
		newMockProvider("GET", "/test"),
		newMockProvider("GET", "/new")}

	s := buildDefaultConfigServer(t, p)

	if err := s.Suspend(); err != nil {
		t.Errorf("Failed to suspend %s", err.Error())
	}

	if err := s.Resume(); err != nil {
		t.Errorf("Failed to resume %s", err.Error())
	}

	s.AbnormalStatusWriter = new(mockAsw)

	if err := s.StartComponent(); err != nil {
		t.Fatalf(err.Error())
	}

	if err := s.AllowAccess(); err != nil {
		t.Errorf("Failed to allow access %s", err.Error())
	}

	if err := s.Suspend(); err != nil {
		t.Errorf("Failed to suspend %s", err.Error())
	}

	s.PrepareToStop()

	s.ReadyToStop()

	if err := s.Stop(); err != nil {
		t.Errorf("Failed to stop %s", err.Error())
	}

}

func setUsableListen(t *testing.T, s *HTTPServer) {

	for attempts := 0; attempts < 5; attempts++ {

		p := s.Port + attempts

		listenAddress := fmt.Sprintf("%s:%d", s.Address, p)

		if ln, err := net.Listen("tcp", listenAddress); err == nil {

			ln.Close()
			s.Port = p

			return

		}

	}

	t.Errorf("Unable to find a port for test HTTPServer to listen on tried %d-%d", s.Port, s.Port+4)
	t.FailNow()

	return
}

func TestServerStartMissingAbnormal(t *testing.T) {

	s := buildDefaultConfigServer(t, []httpendpoint.Provider{})

	// Expect fail due to no abnormal status writer
	if err := s.StartComponent(); err == nil {
		t.Fatalf("Expected startup fail due to missing AbnormalStatusWriter")
	}

}

func buildDefaultConfigServer(t *testing.T, providers []httpendpoint.Provider, extra ...string) *HTTPServer {
	lm := logging.CreateComponentLoggerManager(logging.Fatal, make(map[string]interface{}), []logging.LogWriter{}, logging.NewFrameworkLogMessageFormatter(), false)

	ca, err := configAccessor(lm, test.FilePath("accesslog.json"))

	if err != nil {
		t.Fatalf(err.Error())
	}

	fb := new(FacilityBuilder)

	system := new(instance.System)

	//Create the IoC container
	cc := ioc.NewComponentContainer(lm, ca, system)

	for i, p := range providers {

		cc.WrapAndAddProto(fmt.Sprintf("provider%d", i), p)

	}

	err = fb.BuildAndRegister(lm, ca, cc)

	if err != nil {
		t.Fatalf(err.Error())
	}

	if err = cc.Populate(); err != nil {
		t.Fatalf(err.Error())
	}

	s := cc.ComponentByName(HTTPServerComponentName).Instance.(*HTTPServer)
	s.FrameworkLogger = lm.CreateLogger("testHTTPServer")

	setUsableListen(t, s)

	return s
}

type mockAsw struct {
	code int
}

func (a *mockAsw) WriteAbnormalStatus(ctx context.Context, state *ws.ProcessState) error {

	state.HTTPResponseWriter.Status = a.code
	return nil
}

func newMockProvider(method string, pattern string) *mockProvider {
	mp := new(mockProvider)

	mp.methods = []string{method}
	mp.pattern = pattern

	return mp
}

type mockProvider struct {
	methods        []string
	pattern        string
	called         bool
	versionEnabled bool
	supported      []string
}

func (mp *mockProvider) SupportedHTTPMethods() []string {
	return mp.methods
}

func (mp *mockProvider) RegexPattern() string {
	return mp.pattern
}

func (mp *mockProvider) ServeHTTP(ctx context.Context, w *httpendpoint.HTTPResponseWriter, req *http.Request) context.Context {

	mp.called = true

	return ctx
}

func (mp *mockProvider) VersionAware() bool {
	return mp.versionEnabled
}

func (mp *mockProvider) SupportsVersion(version httpendpoint.RequiredVersion) bool {

	v := version["v"]

	if v == "" {
		return false
	}

	for _, sv := range mp.supported {
		if sv == v {
			return true
		}
	}

	return false
}

func (mp *mockProvider) AutoWireable() bool {
	return true
}

type mockIDContextBuilder struct {
	fail   bool
	called bool
	id     string
}

type mockIDKeyType string

var mockKey mockIDKeyType = "mockkey"

func (cb *mockIDContextBuilder) WithIdentity(ctx context.Context, req *http.Request) (context.Context, error) {

	if cb.fail {
		return ctx, fmt.Errorf("Forced identity error")
	}

	nctx := context.WithValue(ctx, mockKey, cb.id)
	cb.called = true

	return nctx, nil

}

func (cb *mockIDContextBuilder) ID(ctx context.Context) string {
	i := ctx.Value(mockKey)

	if i == nil {
		return ""
	} else {
		return i.(string)
	}
}

type mockRequestInstrumentationManager struct {
	i *mockRequestInstrumentor
}

func (nm *mockRequestInstrumentationManager) Begin(ctx context.Context, res http.ResponseWriter, req *http.Request) (context.Context, instrument.Instrumentor, func()) {

	ri := new(mockRequestInstrumentor)
	ri.amendCalls = make(map[instrument.Additional]bool)
	nm.i = ri
	nc := instrument.AddInstrumentorToContext(ctx, ri)

	return nc, ri, ri.StartEvent("mock")
}

// A default implementation of instrument.Instrumentor that does nothing
type mockRequestInstrumentor struct {
	amendCalls map[instrument.Additional]bool
	started    bool
	ended      bool
}

func (ni *mockRequestInstrumentor) StartEvent(id string, metadata ...interface{}) instrument.EndEvent {

	ni.started = true

	return func() { ni.ended = true }
}

func (ni *mockRequestInstrumentor) Fork(ctx context.Context) (context.Context, instrument.Instrumentor) {
	return ctx, ni
}

func (ni *mockRequestInstrumentor) Integrate(instrumentor instrument.Instrumentor) {
	return
}

func (ni *mockRequestInstrumentor) Amend(additional instrument.Additional, value interface{}) {
	ni.amendCalls[additional] = true
}

func (ni *mockRequestInstrumentor) Called(additional instrument.Additional) bool {
	return ni.amendCalls[additional]
}

type mockRequestedVersionExtractor struct {
	called bool
}

func (ve *mockRequestedVersionExtractor) Extract(r *http.Request) httpendpoint.RequiredVersion {

	ve.called = true

	v := r.Header.Get("version")

	rv := make(httpendpoint.RequiredVersion)

	rv["v"] = v

	return rv

}

type mockResponseWriter struct {
}

func (rw *mockResponseWriter) Header() http.Header {
	return make(http.Header)
}

func (rw *mockResponseWriter) Write(b []byte) (int, error) {
	return len(b), nil
}

func (rw *mockResponseWriter) WriteHeader(statusCode int) {

}
