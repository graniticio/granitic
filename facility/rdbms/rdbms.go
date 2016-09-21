package rdbms

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/graniticio/granitic/facility/querymanager"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/logging"
	"golang.org/x/net/context"
)

type DatabaseProvider interface {
	Database() (*sql.DB, error)
	DatabaseFromContext(ctx context.Context) (*sql.DB, error)
}

type RdbmsClientManager interface {
	Client() (*RdbmsClient, error)
	ClientFromContext(ctx context.Context) (*RdbmsClient, error)
}

type ProviderComponentReceiver interface {
	RegisterProvider(p *ioc.Component)
}

type DefaultRdbmsClientManager struct {
	DisableAutoInjection bool
	InjectFieldNames     []string
	Provider             DatabaseProvider
	ProviderName         string
	QueryManager         *querymanager.QueryManager
	db                   *sql.DB
	FrameworkLogger      logging.Logger
	state                ioc.ComponentState
	candidateProviders   []*ioc.Component
}

func (cm *DefaultRdbmsClientManager) RegisterProvider(p *ioc.Component) {

	if cm.candidateProviders == nil {
		cm.candidateProviders = []*ioc.Component{p}
	} else {
		cm.candidateProviders = append(cm.candidateProviders, p)
	}
}

func (cm *DefaultRdbmsClientManager) Client() (*RdbmsClient, error) {

	var db *sql.DB
	var err error

	if db, err = cm.Provider.Database(); err != nil {
		return nil, err
	}

	return newRdbmsClient(db, cm.QueryManager), nil
}

func (cm *DefaultRdbmsClientManager) ClientFromContext(ctx context.Context) (*RdbmsClient, error) {
	var db *sql.DB
	var err error

	if db, err = cm.Provider.DatabaseFromContext(ctx); err != nil {
		return nil, err
	}

	return newRdbmsClient(db, cm.QueryManager), nil
}

func (cm *DefaultRdbmsClientManager) StartComponent() error {

	if cm.state != ioc.StoppedState {
		return nil
	}

	cm.state = ioc.StartingState

	if err := cm.selectProvider(); err != nil {
		return err
	}

	cm.state = ioc.RunningState

	return nil
}

func (cm *DefaultRdbmsClientManager) selectProvider() error {

	if cm.Provider != nil {
		return nil
	}

	cp := cm.candidateProviders
	l := len(cp)

	if l == 0 {
		return errors.New("No components implementing rdbms.DatabaseProvider are available.")
	} else if l == 1 {
		cm.Provider = cp[0].Instance.(DatabaseProvider)

	} else {

		if cm.ProviderName == "" {
			return errors.New("Multiple components implementing rdbms.DatabaseProvider are available, but no ProviderName provided to allow one to be chosen.")
		}

		cm.Provider = cm.findProviderByName()

		if cm.Provider == nil {

			m := fmt.Sprintf("No component called %s and implementing rdbms.DatabaseProvider is available", cm.ProviderName)
			return errors.New(m)

		}

	}

	return nil

}

func (cm *DefaultRdbmsClientManager) findProviderByName() DatabaseProvider {

	for _, c := range cm.candidateProviders {

		if c.Name == cm.ProviderName {
			return c.Instance.(DatabaseProvider)
		}

	}

	return nil
}

func (cm *DefaultRdbmsClientManager) PrepareToStop() {
	cm.state = ioc.StoppingState

}

func (cm *DefaultRdbmsClientManager) ReadyToStop() (bool, error) {
	return true, nil
}

func (cm *DefaultRdbmsClientManager) Stop() error {

	db := cm.db

	cm.state = ioc.StoppedState

	if db != nil {
		return db.Close()
	} else {
		return nil
	}
}
