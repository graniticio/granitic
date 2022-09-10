package rdbms

import (
	"context"
	"database/sql"
	"github.com/graniticio/granitic/v3/ioc"
	"github.com/graniticio/granitic/v3/logging"
	"testing"
)

func TestImplementsInterface(t *testing.T) {

	var cm ClientManager

	cm = new(GraniticRdbmsClientManager)

	_ = cm

}

func TestGraniticRdbmsClientManager_ClientFromContext(t *testing.T) {

	m := new(GraniticRdbmsClientManager)

	m.FrameworkLogger = new(logging.ConsoleErrorLogger)

	conf := &ClientManagerConfig{
		Provider: new(mockProvider),
	}

	m.Configuration = conf
	m.state = ioc.RunningState

	_, err := m.ClientFromContext(context.Background())

	if err != nil {
		t.Fatalf("%v", err)
	}

}

type mockProvider struct{}

func (mp *mockProvider) Database() (*sql.DB, error) {

	return nil, nil
}
