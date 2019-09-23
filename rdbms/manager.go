// Copyright 2016-2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

/*
Package rdbms provides types for interacting with relational (SQL) databases.

A full explanation of Grantic's patterns for relational database access management system (RDBMS) access can be found at https://granitic.io/ref/relational-databases

Principles

Granitic's RDBMS access types and components adhere to a number of principles:

1. Application developers should be given the option to keep query definitions separate from application code.

2. Boilerplate code for generating queries and mapping results should be minimised.

3. Transactional and non-transactional DB access should not require different coding patterns.

4. Code that is executing queries against a database should be readable and have obvious intent.

Facilities

To use Granitic's RDBMS access, your application will need to enable both the QueryManager and RdbmsAccess facilities.
 implements https://granitic.io/ref/facilities for more details.

Components and types

Enabling these facilities creates the QueryManager and ClientManager components. Your application must provide
a DatabaseProvider ( implements below)


	DatabaseProvider    A component able to provide connections to an RDBMS by creating
						and managing Go sql.DB objects.

	QueryManager        A component that loads files containing template queries from a filesystem
						and can populate than with provided variables to create a complete SQL query.

	ManagedClient         A type providing methods to execute templated queries against an RDBMS.

	ClientManager  A component able to create ManagedClient objects.

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
be injected into the ClientManager component.

QueryManager

Refer to the dsquery package for more details on how QueryManagers work.5

ClientManager

Granitic applications are discouraged from directly interacting with sql.DB objects (although of course they are
free to do so). Instead, they use instances of ManagedClient. ManagedClient objects are not reusable across goroutines,
instead your application will need to ask for a new one to be created for each new goroutine (e.g. for each request in
a web services application).

The component that is able to provide these clients is ClientManager.

Auto-injection of an ClientManager

Any component that needs an ManagedClient should have a field:

	DbClientManager rdbms.ClientManager

The name DbClientManager is a default. You can change the field that Granitic looks for by setting the following in
your application configuration.

	{
	  "RdbmsAccess":{
		"InjectFieldNames": ["DbClientManager", "MyAlternateFieldName"]
	  }
	}

Your code then obtains an ManagedClient in a manner similar to:

	if rc, err := id.DBClientManager.ManagedClient(); err != nil {
	  return err
	}

ManagedClient

Application code executes SQL (either directly or via a templated query) and interacts with transactions via an
instance of ManagedClient. Refer to the GoDoc for ManagedClient for information on the methods available, but the general pattern
for the methods available on ManagedClient is:

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

ManagedClient provides a mechanism for automatically copying result data into structs or slices of structs. If the
ManagedClient method name contains BindSingle, you will pass a pointer to a struct into the method and its fields will be populated:

ad := new(ArtistDetail)

	if found, err := rc.SelectBindSingleQIDParams("ARTIST_DETAIL", rid, ad); found {
	  return ad, err
	} else {
	  return nil, err
	}

If the method contains the word Bind (not BindSingle), you will supply an example 'template' instance of a struct and the method will return a slice of that type:

	ar := new(ArtistSearchResult)

	if r, err := rc.SelectBindQIDParams("ARTIST_SEARCH_BASE", make(map[string]interface{}), ar); err != nil {
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

ManagedClient provides pass-through access to sql.DB's Exec, Query and QueryRow methods. Note that these methods are compatible
with Granitic's transaction pattern as described above.


Multiple databases

This iteration of Granitic is optimised for the most common use-case for RDBMS access, where a particular Granitic
application will access a single logical database. It is fully acknowledged that there are many situations where an application
needs to access mutiple logical databases.

Facility support for that use-case will be added in later versions of Granitic, but for now you have two options:

Option 1: use this facility to provide support for your application's 'main' database and manually add components of type rdbms.DefaultRDBMSClientManager to
your component definition file to support your other database.

Option 2: disable this facility and manually add components of type rdbms.GraniticRdbmsClientManager to
your component definition file to support all of your databases.
*/
package rdbms

import (
	"context"
	"database/sql"
	"errors"
	"github.com/graniticio/granitic/v2/dsquery"
	"github.com/graniticio/granitic/v2/ioc"
	"github.com/graniticio/granitic/v2/logging"
)

/*
DatabaseProvider is implemented by an object able to create a sql.DB object to connect to an instance of an RDBMS. The
implementation is expected to manage connection pooling and failover as required
*/
type DatabaseProvider interface {
	// Database returns a Go sql.DB object
	Database() (*sql.DB, error)
}

// NonStandardInsertProvider is an optional interface for DatabaseProvider implementations when the prepared statement->exec->insert pattern does not yield
// the last inserted ID as part of its result.
type NonStandardInsertProvider interface {
	InsertIDFunc() InsertWithReturnedID
}

/*
ContextAwareDatabaseProvider is implemented by DatabaseProvider implementations that need to be given a context when establishing a database connection
*/
type ContextAwareDatabaseProvider interface {
	DatabaseFromContext(context.Context) (*sql.DB, error)
}

// ClientManager is implemented by a component that can create ManagedClient objects that application code will use to execute SQL statements.
type ClientManager interface {
	// ManagedClient returns an ManagedClient that is ready to use.
	Client() (Client, error)

	// ClientFromContext returns an ManagedClient that is ready to use. Providing a context allows the underlying DatabaseProvider
	// to modify the connection to the RDBMS.
	ClientFromContext(ctx context.Context) (Client, error)
}

// ProviderComponentReceiver is implemented by components that are interested in having visibility of all DatabaseProvider implementations available
// to an application.
type ProviderComponentReceiver interface {
	RegisterProvider(p *ioc.Component)
}

// ClientManagerConfig is used to organise the various components that interact to manage a database connection when your
// application needs to connect to more that one database simultaneously.
type ClientManagerConfig struct {
	Provider DatabaseProvider

	// The names of fields on a component that should have a reference to this component's associated ClientManager
	// automatically injected into them.
	InjectFieldNames []string

	BlockUntilConnected bool

	// A name that will be shared by any instances of ManagedClient created by this manager - this is used for logging purposes
	ClientName string

	// Name that will be given to the ClientManager component that will be created. If not set, it will be set the value of ClientName + "Manager"
	ManagerName string
}

/*
GraniticRdbmsClientManager is Granitic's default implementation of ClientManager. An instance of this will be created when you enable the

RdbmsAccess access facility and will be injected into any component that needs database access -  implements the package
documentation for facilty/rdbms for more details.
*/
type GraniticRdbmsClientManager struct {
	// Set to true if you are creating an instance of GraniticRdbmsClientManager manually
	DisableAutoInjection bool

	// Auto-injected if the QueryManager facility is enabled
	QueryManager dsquery.QueryManager

	Configuration *ClientManagerConfig

	// Injected by Granitic.
	FrameworkLogger logging.Logger

	SharedLog logging.Logger

	state ioc.ComponentState
}

// BlockAccess returns true if BlockUntilConnected is set to true and a connection to the underlying RDBMS
// has not yet been established.
func (cm *GraniticRdbmsClientManager) BlockAccess() (bool, error) {

	if !cm.Configuration.BlockUntilConnected {
		return false, nil
	}

	provider := cm.Configuration.Provider

	db, err := provider.Database()

	if err != nil {
		return true, errors.New("Unable to connect to database: " + err.Error())
	}

	if err = db.Ping(); err == nil {
		return false, nil
	}

	return true, errors.New("Unable to connect to database: " + err.Error())

}

// Client implements ClientManager.Client
func (cm *GraniticRdbmsClientManager) Client() (Client, error) {

	if cm.state != ioc.RunningState {
		return nil, errors.New("No Client will be created because ClientManager is not running. Application shutting down?")
	}

	var db *sql.DB
	var err error

	provider := cm.Configuration.Provider

	if db, err = provider.Database(); err != nil {
		return nil, err
	}

	return newRdbmsClient(db, cm.QueryManager, cm.chooseInsertFunction(), cm.SharedLog), nil
}

// ClientFromContext implements ClientManager.ClientFromContext
func (cm *GraniticRdbmsClientManager) ClientFromContext(ctx context.Context) (Client, error) {

	if cm.state != ioc.RunningState {
		return nil, errors.New("No Client will be created because ClientManager is not running. Application shutting down?")
	}

	var db *sql.DB
	var err error

	provider := cm.Configuration.Provider

	if cdp, found := provider.(ContextAwareDatabaseProvider); found {

		if db, err = cdp.DatabaseFromContext(ctx); err != nil {
			return nil, err
		}

	} else {
		if db, err = provider.Database(); err != nil {
			return nil, err
		}
	}

	rc := newRdbmsClient(db, cm.QueryManager, cm.chooseInsertFunction(), cm.SharedLog)
	rc.ctx = ctx

	return rc, nil
}

func (cm *GraniticRdbmsClientManager) chooseInsertFunction() InsertWithReturnedID {

	if iwi, found := cm.Configuration.Provider.(NonStandardInsertProvider); found {
		return iwi.InsertIDFunc()
	}

	return DefaultInsertWithReturnedID
}

// StartComponent selects a DatabaseProvider to use
func (cm *GraniticRdbmsClientManager) StartComponent() error {

	if cm.state != ioc.StoppedState {
		return nil
	}

	cm.state = ioc.StartingState

	cm.state = ioc.RunningState

	return nil
}

// PrepareToStop transitions component to stopping state, prevent new ManagedClient objects from being created.
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
