package rdbms

import (
	"database/sql"
	"github.com/graniticio/granitic/facility/querymanager"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/logging"
)

type DatabaseProvider interface {
	Database() (*sql.DB, error)
}

type RdbmsClientManager interface {
	Client() *RdbmsClient
	ClientFromContext(context interface{}) *RdbmsClient
}

type DefaultRdbmsClientManager struct {
	Provider                      DatabaseProvider
	DatabaseProviderComponentName string
	QueryManager                  *querymanager.QueryManager
	db                            *sql.DB
	FrameworkLogger               logging.Logger
	state                         ioc.ComponentState
}

func (drcm *DefaultRdbmsClientManager) Client() *RdbmsClient {
	return newRdbmsClient(drcm.db, drcm.QueryManager)
}

func (drcm *DefaultRdbmsClientManager) ClientFromContext(context interface{}) *RdbmsClient {
	return drcm.Client()
}

func (drcm *DefaultRdbmsClientManager) StartComponent() error {

	if drcm.state != ioc.StoppedState {
		return nil
	}

	drcm.state = ioc.StartingState

	db, err := drcm.Provider.Database()

	if err != nil {
		return err

	} else {
		drcm.db = db

		drcm.state = ioc.RunningState

		return nil
	}

}

func (drcm *DefaultRdbmsClientManager) PrepareToStop() {
	drcm.state = ioc.StoppingState

}

func (drcm *DefaultRdbmsClientManager) ReadyToStop() (bool, error) {
	return true, nil
}

func (drcm *DefaultRdbmsClientManager) Stop() error {

	db := drcm.db

	drcm.state = ioc.StoppedState

	if db != nil {
		return db.Close()
	} else {
		return nil
	}
}
