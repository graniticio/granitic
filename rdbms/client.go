// Copyright 2016-2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package rdbms

import (
	"context"
	"database/sql"
	"errors"
	"github.com/graniticio/granitic/v2/dsquery"
	"github.com/graniticio/granitic/v2/logging"
)

// Client provides access to methods for executing SQL queries and managing transactions
type Client interface {
	FindFragment(qid string) (string, error)
	BuildQueryFromQIDParams(qid string, p ...interface{}) (string, error)
	DeleteQIDParams(qid string, params ...interface{}) (sql.Result, error)
	DeleteQIDParam(qid string, name string, value interface{}) (sql.Result, error)
	RegisterTempQuery(qid string, query string)
	ExistingIDOrInsertParams(checkQueryID, insertQueryID string, idTarget *int64, p ...interface{}) error
	InsertQIDParams(qid string, params ...interface{}) (sql.Result, error)
	InsertCaptureQIDParams(qid string, target *int64, params ...interface{}) error
	SelectBindSingleQID(qid string, target interface{}) (bool, error)
	SelectBindSingleQIDParam(qid string, name string, value interface{}, target interface{}) (bool, error)
	SelectBindSingleQIDParams(qid string, target interface{}, params ...interface{}) (bool, error)
	SelectBindQID(qid string, template interface{}) ([]interface{}, error)
	SelectBindQIDParam(qid string, name string, value interface{}, template interface{}) ([]interface{}, error)
	SelectBindQIDParams(qid string, template interface{}, params ...interface{}) ([]interface{}, error)
	SelectQID(qid string) (*sql.Rows, error)
	SelectQIDParam(qid string, name string, value interface{}) (*sql.Rows, error)
	SelectQIDParams(qid string, params ...interface{}) (*sql.Rows, error)
	UpdateQIDParams(qid string, params ...interface{}) (sql.Result, error)
	UpdateQIDParam(qid string, name string, value interface{}) (sql.Result, error)
	StartTransaction() error
	StartTransactionWithOptions(opts *sql.TxOptions) error
	Rollback()
	CommitTransaction() error
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

func newRdbmsClient(database *sql.DB, querymanager dsquery.QueryManager, insertFunc InsertWithReturnedID, logger logging.Logger) *ManagedClient {
	rc := new(ManagedClient)
	rc.db = database
	rc.queryManager = querymanager
	rc.lastID = insertFunc
	rc.emptyParams = make(map[string]interface{})
	rc.binder = new(RowBinder)
	rc.tempQueries = make(map[string]string)

	rc.FrameworkLogger = logger

	return rc
}

// ManagedClient is the interface application code should use to execute SQL against a database. See the package overview for the rdbms
// package for usage.
//
// ManagedClient is stateful and MUST NOT be shared across goroutines
type ManagedClient struct {
	db              *sql.DB
	queryManager    dsquery.QueryManager
	tx              *sql.Tx
	lastID          InsertWithReturnedID
	tempQueries     map[string]string
	emptyParams     map[string]interface{}
	binder          *RowBinder
	ctx             context.Context
	FrameworkLogger logging.Logger
}

// FindFragment returns a partial query from the underlying QueryManager. Fragments are no
// different that ordinary template queries, except they are not expected to contain any variable placeholders.
func (rc *ManagedClient) FindFragment(qid string) (string, error) {
	return rc.queryManager.FragmentFromID(qid)
}

// BuildQueryFromQIDParams returns a populated SQL query that can be manually executed later.
func (rc *ManagedClient) BuildQueryFromQIDParams(qid string, p ...interface{}) (string, error) {

	var pm map[string]interface{}
	var err error

	if pm, err = ParamsFromFieldsOrTags(p...); err == nil {
		return rc.queryManager.BuildQueryFromID(qid, pm)
	}

	return "", err
}

// DeleteQIDParams executes the supplied query with the expectation that it is a 'DELETE' query.
func (rc *ManagedClient) DeleteQIDParams(qid string, params ...interface{}) (sql.Result, error) {

	return rc.execQIDParams(qid, params...)

}

// DeleteQIDParam executes the supplied query with the expectation that it is a 'DELETE' query.
func (rc *ManagedClient) DeleteQIDParam(qid string, name string, value interface{}) (sql.Result, error) {

	p := make(map[string]interface{})
	p[name] = value

	return rc.execQIDParams(qid, p)

}

// RegisterTempQuery stores the supplied query in the ManagedClient so that it can be used with methods that expect a QID.
// Note that the query is NOT stored in the underlying QueryManager.
func (rc *ManagedClient) RegisterTempQuery(qid string, query string) {
	rc.tempQueries[qid] = query
}

// ExistingIDOrInsertParams finds the ID of record or if the record does not exist, inserts a new record and retrieves the newly assigned ID
func (rc *ManagedClient) ExistingIDOrInsertParams(checkQueryID, insertQueryID string, idTarget *int64, p ...interface{}) error {

	if found, err := rc.SelectBindSingleQIDParams(checkQueryID, idTarget, p...); err != nil {
		return err
	} else if found {
		return nil
	} else {

		if err = rc.InsertCaptureQIDParams(insertQueryID, idTarget, p...); err != nil {
			return err
		}

	}

	return nil
}

// InsertQIDParams executes the supplied query with the expectation that it is an 'INSERT' query.
func (rc *ManagedClient) InsertQIDParams(qid string, params ...interface{}) (sql.Result, error) {

	return rc.execQIDParams(qid, params...)

}

// InsertCaptureQIDParams executes the supplied query with the expectation that it is an 'INSERT' query and captures
// the new row's server generated ID in the target int64
func (rc *ManagedClient) InsertCaptureQIDParams(qid string, target *int64, params ...interface{}) error {

	var query string
	var err error

	if query, err = rc.buildQuery(qid, params...); err != nil {
		return err
	}

	return rc.lastID(query, rc, target)
}

// SelectBindSingleQID executes the supplied query with the expectation that it is a 'SELECT' query that returns 0 or 1 rows.
// Results of the query are bound into the target struct. Returns false if no rows were found.
func (rc *ManagedClient) SelectBindSingleQID(qid string, target interface{}) (bool, error) {
	return rc.SelectBindSingleQIDParams(qid, target, rc.emptyParams)
}

// SelectBindSingleQIDParam executes the supplied query with the expectation that it is a 'SELECT' query that returns 0 or 1 rows.
// Results of the query are bound into the target struct. Returns false if no rows were found.
func (rc *ManagedClient) SelectBindSingleQIDParam(qid string, name string, value interface{}, target interface{}) (bool, error) {
	p := make(map[string]interface{})
	p[name] = value

	return rc.SelectBindSingleQIDParams(qid, target, p)
}

// SelectBindSingleQIDParams executes the supplied query with the expectation that it is a 'SELECT' query that returns 0 or 1 rows.
// Results of the query are bound into the target struct. Returns false if no rows were found.
func (rc *ManagedClient) SelectBindSingleQIDParams(qid string, target interface{}, params ...interface{}) (bool, error) {

	var r *sql.Rows
	var err error

	if r, err = rc.SelectQIDParams(qid, params...); err != nil {
		return false, err
	}

	defer r.Close()

	return rc.binder.BindRow(r, target)

}

// SelectBindQID executes the supplied query with the expectation that it is a 'SELECT' query. Results of the query
// are returned in a slice of the same type as the supplied template struct.
func (rc *ManagedClient) SelectBindQID(qid string, template interface{}) ([]interface{}, error) {
	return rc.SelectBindQIDParams(qid, template, rc.emptyParams)
}

// SelectBindQIDParam executes the supplied query with the expectation that it is a 'SELECT' query. Results of the query
// are returned in a slice of the same type as the supplied template struct.
func (rc *ManagedClient) SelectBindQIDParam(qid string, name string, value interface{}, template interface{}) ([]interface{}, error) {
	p := make(map[string]interface{})
	p[name] = value

	return rc.SelectBindQIDParams(qid, template, p)
}

// SelectBindQIDParams executes the supplied query with the expectation that it is a 'SELECT' query. Results of the query
// are returned in a slice of the same type as the supplied template struct.
func (rc *ManagedClient) SelectBindQIDParams(qid string, template interface{}, params ...interface{}) ([]interface{}, error) {
	var r *sql.Rows
	var err error

	if r, err = rc.SelectQIDParams(qid, params...); err != nil {
		return nil, err
	}

	defer r.Close()

	return rc.binder.BindRows(r, template)
}

// SelectQID executes the supplied query with the expectation that it is a 'SELECT' query.
func (rc *ManagedClient) SelectQID(qid string) (*sql.Rows, error) {
	return rc.SelectQIDParams(qid, rc.emptyParams)
}

// SelectQIDParam executes the supplied query with the expectation that it is a 'SELECT' query.
func (rc *ManagedClient) SelectQIDParam(qid string, name string, value interface{}) (*sql.Rows, error) {
	p := make(map[string]interface{})
	p[name] = value

	return rc.SelectQIDParams(qid, p)
}

// SelectQIDParams executes the supplied query with the expectation that it is a 'SELECT' query.
func (rc *ManagedClient) SelectQIDParams(qid string, params ...interface{}) (*sql.Rows, error) {

	query, err := rc.buildQuery(qid, params...)

	if err != nil {
		return nil, err
	}

	return rc.Query(query)

}

// UpdateQIDParams executes the supplied query with the expectation that it is an 'UPDATE' query.
func (rc *ManagedClient) UpdateQIDParams(qid string, params ...interface{}) (sql.Result, error) {

	return rc.execQIDParams(qid, params...)

}

// UpdateQIDParam executes the supplied query with the expectation that it is an 'UPDATE' query.
func (rc *ManagedClient) UpdateQIDParam(qid string, name string, value interface{}) (sql.Result, error) {

	p := make(map[string]interface{})
	p[name] = value

	return rc.execQIDParams(qid, p)

}

func (rc *ManagedClient) execQIDParams(qid string, params ...interface{}) (sql.Result, error) {

	var query string
	var err error

	if query, err = rc.buildQuery(qid, params...); err != nil {
		return nil, err
	}

	return rc.Exec(query)
}

func (rc *ManagedClient) buildQuery(qid string, p ...interface{}) (string, error) {

	tq := rc.tempQueries[qid]

	if tq != "" {
		return tq, nil
	}

	var pm map[string]interface{}
	var err error

	if pm, err = ParamsFromFieldsOrTags(p...); err != nil {
		return "", err
	}

	if rc.FrameworkLogger.IsLevelEnabled(logging.Trace) {
		//Log the parameters to be injected into the query
		rc.FrameworkLogger.LogTracef("Parameters: %v", pm)
	}

	return rc.queryManager.BuildQueryFromID(qid, pm)

}

// StartTransaction opens a transaction on the underlying sql.DB object and re-maps all calls to non-transactional
// methods to their transactional equivalents.
func (rc *ManagedClient) StartTransaction() error {

	if rc.tx != nil {
		return errors.New("Transaction already open")
	}

	tx, err := rc.db.Begin()

	if err != nil {
		return err
	}

	rc.tx = tx
	return nil
}

// StartTransactionWithOptions opens a transaction on the underlying sql.DB object and re-maps all calls to non-transactional
// methods to their transactional equivalents.
func (rc *ManagedClient) StartTransactionWithOptions(opts *sql.TxOptions) error {

	if rc.tx != nil {
		return errors.New("Transaction already open")
	}

	var ctx context.Context

	if rc.contextAware() {
		ctx = rc.ctx
	} else {
		ctx = context.Background()
	}

	tx, err := rc.db.BeginTx(ctx, opts)

	if err != nil {
		return err
	}

	rc.tx = tx
	return nil

}

// Rollback rolls the open transaction back - does nothing if no transaction is open.
func (rc *ManagedClient) Rollback() {

	if rc.tx != nil {
		rc.tx.Rollback()
	}
}

// CommitTransaction commits the open transaction - does nothing if no transaction is open.
func (rc *ManagedClient) CommitTransaction() error {

	if rc.tx == nil {
		return errors.New("No open transaction to commit")
	}

	err := rc.tx.Commit()
	rc.tx = nil

	return err

}

// Exec is a pass-through to its sql.DB equivalent (or sql.Tx equivalent is a transaction is open)
func (rc *ManagedClient) Exec(query string, args ...interface{}) (sql.Result, error) {

	tx := rc.tx

	if rc.contextAware() {
		if tx != nil {
			return tx.ExecContext(rc.ctx, query, args...)
		}

		return rc.db.ExecContext(rc.ctx, query, args...)

	}

	if tx != nil {
		return tx.Exec(query, args...)
	}

	return rc.db.Exec(query, args...)
}

// Query is a pass-through to its sql.DB equivalent (or sql.Tx equivalent is a transaction is open)
func (rc *ManagedClient) Query(query string, args ...interface{}) (*sql.Rows, error) {
	tx := rc.tx

	if rc.contextAware() {
		if tx != nil {
			return tx.QueryContext(rc.ctx, query, args...)
		}

		return rc.db.QueryContext(rc.ctx, query, args...)

	}

	if tx != nil {
		return tx.Query(query, args...)
	}

	return rc.db.Query(query, args...)
}

// QueryRow is a pass-through to its sql.DB equivalent (or sql.Tx equivalent is a transaction is open)
func (rc *ManagedClient) QueryRow(query string, args ...interface{}) *sql.Row {
	tx := rc.tx

	if rc.contextAware() {
		if tx != nil {
			return tx.QueryRowContext(rc.ctx, query, args...)
		}

		return rc.db.QueryRowContext(rc.ctx, query, args...)

	}

	if tx != nil {
		return tx.QueryRow(query, args...)
	}

	return rc.db.QueryRow(query, args...)
}

func (rc *ManagedClient) contextAware() bool {
	return rc.ctx != nil
}
