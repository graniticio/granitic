// Copyright 2016 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

/*
	Package rdms provides types for interacting with relational (SQL) databases.

	A full explanation of Grantic's patterns for relational database access management system (RDBMS) access can be found at http://granitic.io/1.0/ref/rdbms-access A brief
	explanation follows:

	Principles

	Granitic's RDBMS access types and components adhere to a number of principles:

	1. Application developers should be given the option to keep query definitions separate from application code.

	2. Boilerplate code for generating queries and mapping results should be minimised.

	3. Transactional and non-transactional DB access should not require different coding patterns.

	4. Code that is executing queries against a database should be readable and have obvious intent.

	Facilities

	To use Granitic's RDBMS access, your application will need to enable both the QueryManager and RdbmsAccess facilities.
	See http://granitic.io/1.0/ref/facilities for more details.

	Components and types

	Enabling these facilities creates the QueryManager and RdbmsClientManager components. Your application must provide
	a DatabaseProvider (see below)


		DatabaseProvider    A component able to provide connections to an RDBMS by creating
		                    and managing Go sql.DB objects.

		QueryManager        A component that loads files containing template queries from a filesystem
		                    and can populate than with provided variables to create a complete SQL query.

		RdbmsClient         A type providing methods to execute templated queries against an RDBMS.

		RdbmsClientManager  A component able to create RdbmsClient objects.

		RowBinder           A type able to map SQL query results into Go structs.


	DatabaseProvider

	Because of the way Go applications are built (statically linked), drivers for individual RDBMSs cannot be dynamically
	loaded at runtime. To avoid the Granitic framework importing large numbers of third-party libraries,
	Granitic application developers must create a component that implements rdbms.DatabaseProvider and imports the driver
	for the database in use.

	A simple implementation of DatabaseProvider for a MySQL database can be found in the granitic-examples repo on GitHub
	in the recordstore/database/provider.go file.

	Once you have an implementation, you will need to create a component for it in your application's component definition
	file similar to:

		{
		  "dbProvider": {
		    "type": "database.DBProvider"
		  }
		}

	As long as you only have one implementation of DatabaseProvider registered as a component, it will automatically
	be injected into the RdbmsClientManager component.

	QueryManager

	Refer to the dsquery package for more details on how QueryManagers work.5

	RdbmsClientManager

	Granitic applications are discouraged from directly interacting with sql.DB objects (although of course they are
	free to do so). Instead, they use instances of RdbmsClient. RdbmsClient objects are not reusable across goroutines,
	instead your application will need to ask for a new one to be created for each new goroutine (e.g. for each request in
	a web services application).

	The component that is able to provide these clients is RdbmsClientManager.

	Auto-injection of an RdbmsClientManager

	Any component that needs an RdbmsClient should have a field:

		DbClientManager rdbms.RdbmsClientManager

	The name DbClientManager is a default. You can change the field that Granitic looks for by setting the following in
	your application configuration.

		{
		  "RdbmsAccess":{
		    "InjectFieldNames": ["DbClientManager", "MyAlternateFieldName"]
		  }
		}

	Your code then obtains an RdbmsClient in a manner similar to:

		if rc, err := id.DBClientManager.Client(); err != nil {
		  return err
		}

	RdbmsClient

	Application code executes SQL (either directly or via a templated query) and interacts with transactions via an
	instance of RdbmsClient. Refer to the GoDoc for RdbmsClient for information on the methods available, but the general pattern
	for the methods available on RdbmsClient is:

		SQLVerb[BindingType]QID[ParameterSource]

	Where

		SQLVerb           Is Select, Delete, Update or Insert
		BindingType       Is optional and can be Bind or BindSingle
		ParameterSource   Is optional and can be either Param or Params

	QID

	A QID is the ID of query stored in the QueryManager.

	Parameter sources

	Parameters to populate template queries can either be supplied via a pair of values (Param), a map[string]interface{} (Params) or a struct whose
	fields are optionally annotated with the `dbparam` tag. If the dbparam tag is not present on a field, the field's name is used a the parameter key.

	Binding

	RdbmsClient provides a mechanism for automatically copying result data into structs or slices of structs. If the
	RdbmsClient method name contains BindSingle, you will pass a pointer to a struct into the method and its fields will be populated:

	ad := new(ArtistDetail)

		if found, err := rc.SelectBindSingleQIdParams("ARTIST_DETAIL", rid, ad); found {
		  return ad, err
		} else {
		  return nil, err
		}

	If the method contains the word Bind (not BindSingle), you will supply an example 'template' instance of a struct and the method will return a slice of that type:

		ar := new(ArtistSearchResult)

		if r, err := rc.SelectBindQIdParams("ARTIST_SEARCH_BASE", make(map[string]interface{}), ar); err != nil {
		  return nil, err
		} else {
		  return id.artistResults(r), nil
		}


	Transactions

	To call start a transaction, invoke the StartTransaction method on the RDBMSCLient like:

		db.StartTransaction()
		defer db.Rollback()

	and end your method with:

		db.CommitTransaction()


	The deferred Rollback call will do nothing if the transaction has previously been commited.


	Direct access to Go DB methods

	RdbmsClient provides pass-through access to sql.DB's Exec, Query and QueryRow methods. Note that these methods are compatible
	with Granitic's transaction pattern as described above.


	Multiple databases

	This iteration of Granitic is optimised for the most common use-case for RDBMS access, where a particular Granitic
	application will access a single logical database. It is fully acknowledged that there are many situations where an application
	needs to access mutiple logical databases.

	Facility support for that use-case will be added in later versions of Granitic, but for now you have two options:

	Option 1: use this facility to provide support for your application's 'main' database and manually add components of type rdbms.DefaultRDBMSClientManager to
	your component definition file to support your other database.

	Option 2: disable this facility and manually add components of type rdbms.DefaultRDBMSClientManager to
	your component definition file to support all of your databases.
*/
package rdbms

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/graniticio/granitic/dsquery"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/logging"
)

/*
Implemented by an object able to create a sql.DB object to connect to an instance of an RDBMS. The
implementation is expected to manage connection pooling and failover as required
*/
type DatabaseProvider interface {
	// Database returns a Go sql.DB object
	Database() (*sql.DB, error)
}

// Optional interface for DatabaseProvider implementations when the prepared statement->exec->insert pattern does not yield
// the last inserted ID as part of its result.
type NonStandardInsertProvider interface {
	InsertIDFunc() InsertWithReturnedId
}


/*
 Implemented by DatabaseProvider implementations that need to be given a context when establishing a database connection
*/
type ContextAwareDatabaseProvider interface {
	DatabaseFromContext(context.Context) (*sql.DB, error)
}

/*
Implemented by a component that can create RdbmsClient objects that application code will use to execute SQL statements.
*/
type RdbmsClientManager interface {
	// Client returns an RdbmsClient that is ready to use.
	Client() (*RdbmsClient, error)

	// ClientFromContext returns an RdbmsClient that is ready to use. Providing a context allows the underlying DatabaseProvider
	// to modify the connection to the RDBMS.
	ClientFromContext(ctx context.Context) (*RdbmsClient, error)
}

// Implemented by components that are interested in having visibility of all DatabaseProvider implementations available
// to an application.
type ProviderComponentReceiver interface {
	RegisterProvider(p *ioc.Component)
}

/*
	Granitic's default implementation of RdbmsClientManager. An instance of this will be created when you enable the

	RdbmsAccess access facility and will be injected into any component that needs database access - see the package
	documentation for facilty/rdbms for more details.
*/
type GraniticRdbmsClientManager struct {
	// Set to true if you are creating an instance of GraniticRdbmsClientManager manually
	DisableAutoInjection bool

	// The names of fields on a component (of type rdbms.RdbmsClientManager) that should have a reference to this component
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

	SharedLog logging.Logger

	state              ioc.ComponentState
	candidateProviders []*ioc.Component
}

// BlockAccess returns true if BlockUntilConnected is set to true and a connection to the underlying RDBMS
// has not yet been established.
func (cm *GraniticRdbmsClientManager) BlockAccess() (bool, error) {

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
func (cm *GraniticRdbmsClientManager) RegisterProvider(p *ioc.Component) {

	if cm.candidateProviders == nil {
		cm.candidateProviders = []*ioc.Component{p}
	} else {
		cm.candidateProviders = append(cm.candidateProviders, p)
	}
}

// See RdbmsClientManager.Client
func (cm *GraniticRdbmsClientManager) Client() (*RdbmsClient, error) {

	if cm.state != ioc.RunningState {
		return nil, errors.New("No Client will be created because ClientManager is not running. Application shutting down?")
	}

	var db *sql.DB
	var err error

	if db, err = cm.Provider.Database(); err != nil {
		return nil, err
	}

	return newRdbmsClient(db, cm.QueryManager, cm.chooseInsertFunction(), cm.SharedLog), nil
}

// See RdbmsClientManager.ClientFromContext
func (cm *GraniticRdbmsClientManager) ClientFromContext(ctx context.Context) (*RdbmsClient, error) {

	if cm.state != ioc.RunningState {
		return nil, errors.New("No Client will be created because ClientManager is not running. Application shutting down?")
	}

	var db *sql.DB
	var err error

	if cdp, found := cm.Provider.(ContextAwareDatabaseProvider); found {

		if db, err = cdp.DatabaseFromContext(ctx); err != nil {
			return nil, err
		}

	} else {
		if db, err = cm.Provider.Database(); err != nil {
			return nil, err
		}
	}

	rc := newRdbmsClient(db, cm.QueryManager, cm.chooseInsertFunction(), cm.SharedLog)
	rc.ctx = ctx

	return rc, nil
}

func (cm *GraniticRdbmsClientManager) chooseInsertFunction() InsertWithReturnedId {

	if iwi, found := cm.Provider.(NonStandardInsertProvider); found{
		return iwi.InsertIDFunc()
	} else {
		return DefaultInsertWithReturnedId
	}

}

// StartComponent selects a DatabaseProvider to use
func (cm *GraniticRdbmsClientManager) StartComponent() error {

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

func (cm *GraniticRdbmsClientManager) selectProvider() error {

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

func (cm *GraniticRdbmsClientManager) findProviderByName() DatabaseProvider {

	for _, c := range cm.candidateProviders {

		if c.Name == cm.ProviderName {
			return c.Instance.(DatabaseProvider)
		}

	}

	return nil
}

// PrepareToStop transitions component to stopping state, prevent new Client objects from being created.
func (cm *GraniticRdbmsClientManager) PrepareToStop() {
	cm.state = ioc.StoppingState
}

// ReadyToStop always returns true, nil
func (cm *GraniticRdbmsClientManager) ReadyToStop() (bool, error) {
	return true, nil
}

// Stop always returns nil
func (cm *GraniticRdbmsClientManager) Stop() error {
	return nil
}
