package httpserver

import (
	"context"
	"github.com/graniticio/granitic/v2/httpendpoint"
	"github.com/graniticio/granitic/v2/logging"
	"github.com/graniticio/granitic/v2/ws"
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

type mockAsw struct {
}

func (a *mockAsw) WriteAbnormalStatus(ctx context.Context, state *ws.ProcessState) error {
	return nil
}
