// Copyright 2016 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package rdbms

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/graniticio/granitic/dsquery"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/logging"
	"golang.org/x/net/context"
)

/*
Implemented by an object able to create a sql.DB object to connect to an instance of an RDBMS. The
implementation is expected to manage connection pooling and failover as required
*/
type DatabaseProvider interface {
	// Database returns a Go sql.DB object
	Database() (*sql.DB, error)

	// DatabaseFromContext returns a Go sql.DB object. Information in the context may affect the returned object
	// (e.g. a context might provide an alternative DB user or DB name).
	DatabaseFromContext(ctx context.Context) (*sql.DB, error)

	// InsertIDFunc returns an implementation of the InsertWithReturnedID function appropriate for the underlying RDBMS.
	InsertIDFunc() InsertWithReturnedID
}

/*
Implemented by a component that can create RDBMSClient objects that application code will use to execute SQL statements.
*/
type RDBMSClientManager interface {
	// Client returns an RDBMSClient that is ready to use.
	Client() (*RDBMSClient, error)

	// ClientFromContext returns an RDBMSClient that is ready to use. Providing a context allows the underlying DatabaseProvider
	// to modify the connection to the RDBMS.
	ClientFromContext(ctx context.Context) (*RDBMSClient, error)
}

// Implemented by components that are interested in having visibility of all DatabaseProvider implementations available
// to an application.
type ProviderComponentReceiver interface {
	RegisterProvider(p *ioc.Component)
}

/*
	Granitic's default implementation of RDBMSClientManager. An instance of this will be created when you enable the

	RdbmsAccess access facility and will be injected into any component that needs database access - see the package
	documentation for facilty/rdbms for more details.
*/
type GraniticRDBMSClientManager struct {
	// Set to true if you are creating an instance of GraniticRDBMSClientManager manually
	DisableAutoInjection bool

	// The names of fields on a component (of type rdbms.RDBMSClientManager) that should have a reference to this component
	// automatically injected into them.
	InjectFieldNames []string

	// Directly set the DatabaseProvider to use when operating multiple RDBMSClientManagers
	Provider DatabaseProvider

	// If multiple DatabaseProviders are available, the name of the component to use.
	ProviderName string

	// Auto-injected if the QueryManager facility is enabled
	QueryManager dsquery.QueryManager

	// Do not allow the Granitic application to start become Accessible (see ioc pacakge) until a connection to the
	// underlying RDBMS is established.
	BlockUntilConnected bool

	// Injected by Granitic.
	FrameworkLogger    logging.Logger
	state              ioc.ComponentState
	candidateProviders []*ioc.Component
}

// BlockAccess returns true if BlockUntilConnected is set to true and a connection to the underlying RDBMS
// has not yet been established.
func (cm *GraniticRDBMSClientManager) BlockAccess() (bool, error) {

	if !cm.BlockUntilConnected {
		return false, nil
	}

	db, err := cm.Provider.Database()

	if err != nil {
		return true, errors.New("Unable to connect to database: " + err.Error())
	}

	if err = db.Ping(); err == nil {
		return false, nil
	} else {
		return true, errors.New("Unable to connect to database: " + err.Error())
	}

}

// See ProviderComponentReceiver.RegisterProvider
func (cm *GraniticRDBMSClientManager) RegisterProvider(p *ioc.Component) {

	if cm.candidateProviders == nil {
		cm.candidateProviders = []*ioc.Component{p}
	} else {
		cm.candidateProviders = append(cm.candidateProviders, p)
	}
}

// See RDBMSClientManager.Client
func (cm *GraniticRDBMSClientManager) Client() (*RDBMSClient, error) {

	if cm.state != ioc.RunningState {
		return nil, errors.New("No Client will be created because ClientManager is not running. Application shutting down?")
	}

	var db *sql.DB
	var err error

	if db, err = cm.Provider.Database(); err != nil {
		return nil, err
	}

	return newRDBMSClient(db, cm.QueryManager, cm.Provider.InsertIDFunc()), nil
}

// See RDBMSClientManager.ClientFromContext
func (cm *GraniticRDBMSClientManager) ClientFromContext(ctx context.Context) (*RDBMSClient, error) {

	if cm.state != ioc.RunningState {
		return nil, errors.New("No Client will be created because ClientManager is not running. Application shutting down?")
	}

	var db *sql.DB
	var err error

	if db, err = cm.Provider.DatabaseFromContext(ctx); err != nil {
		return nil, err
	}

	return newRDBMSClient(db, cm.QueryManager, cm.Provider.InsertIDFunc()), nil
}

// StartComponent selects a DatabaseProvider to use
func (cm *GraniticRDBMSClientManager) StartComponent() error {

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

func (cm *GraniticRDBMSClientManager) selectProvider() error {

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

func (cm *GraniticRDBMSClientManager) findProviderByName() DatabaseProvider {

	for _, c := range cm.candidateProviders {

		if c.Name == cm.ProviderName {
			return c.Instance.(DatabaseProvider)
		}

	}

	return nil
}

// PrepareToStop transitions component to stopping state, prevent new Client objects from being created.
func (cm *GraniticRDBMSClientManager) PrepareToStop() {
	cm.state = ioc.StoppingState
}

// ReadyToStop always returns true, nil
func (cm *GraniticRDBMSClientManager) ReadyToStop() (bool, error) {
	return true, nil
}

// Stop always returns nil
func (cm *GraniticRDBMSClientManager) Stop() error {
	return nil
}
