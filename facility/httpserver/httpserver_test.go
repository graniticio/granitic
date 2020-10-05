package httpserver

import (
	"context"
	"fmt"
	"github.com/graniticio/granitic/v2/httpendpoint"
	"github.com/graniticio/granitic/v2/instance"
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

func TestServerStartDefaultConfig(t *testing.T) {

	s := buildDefaultConfigServer(t, []httpendpoint.Provider{newMockProvider("GET", "/test")})

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
}

func (a *mockAsw) WriteAbnormalStatus(ctx context.Context, state *ws.ProcessState) error {
	return nil
}

func newMockProvider(method string, pattern string) httpendpoint.Provider {
	mp := new(mockProvider)

	mp.methods = []string{method}
	mp.pattern = pattern

	return mp
}

type mockProvider struct {
	methods []string
	pattern string
	called  bool
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
	return false
}

func (mp *mockProvider) SupportsVersion(version httpendpoint.RequiredVersion) bool {

	return true
}

func (mp *mockProvider) AutoWireable() bool {
	return true
}
