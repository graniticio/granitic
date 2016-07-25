package rdbms

import (
	"database/sql"
	"github.com/graniticio/granitic/facility/querymanager"
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
}

func (drcm *DefaultRdbmsClientManager) Client() *RdbmsClient {
	return newRdbmsClient(drcm.db, drcm.QueryManager)
}

func (drcm *DefaultRdbmsClientManager) ClientFromContext(context interface{}) *RdbmsClient {
	return drcm.Client()
}

func (drcm *DefaultRdbmsClientManager) StartComponent() error {

	db, err := drcm.Provider.Database()

	if err != nil {
		return err

	} else {
		drcm.db = db
		return nil
	}

}

func (drcm *DefaultRdbmsClientManager) PrepareToStop() {

}

func (drcm *DefaultRdbmsClientManager) ReadyToStop() (bool, error) {
	return true, nil
}

func (drcm *DefaultRdbmsClientManager) Stop() error {

	db := drcm.db

	if db != nil {
		return db.Close()
	} else {
		return nil
	}
}
