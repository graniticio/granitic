// Copyright 2016 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package rdbms

import (
	"database/sql"
	"errors"
	"github.com/graniticio/granitic/dsquery"
)

func newRDBMSClient(database *sql.DB, querymanager dsquery.QueryManager, insertFunc InsertWithReturnedID) *RDBMSClient {
	rc := new(RDBMSClient)
	rc.db = database
	rc.queryManager = querymanager
	rc.lastID = insertFunc
	rc.emptyParams = make(map[string]interface{})
	rc.binder = new(RowBinder)
	return rc
}

// The interface application code should use to execute SQL against a database. See the package overview for the rdbms
// package for usage.
type RDBMSClient struct {
	db           *sql.DB
	queryManager dsquery.QueryManager
	tx           *sql.Tx
	lastID       InsertWithReturnedID
	tempQueries  map[string]string
	emptyParams  map[string]interface{}
	binder       *RowBinder
}

// FindFragment returns a partial query from the underlying QueryManager. Fragments are no
// different that ordinary template queries, except they are not expected to contain any variable placeholders.
func (rc *RDBMSClient) FindFragment(qid string) (string, error) {
	return rc.queryManager.FragmentFromID(qid)
}

// BuildQueryQIDParams returns a populated SQL query that can be manually executed later.
func (rc *RDBMSClient) BuildQueryQIDParams(qid string, p ...interface{}) (string, error) {

	if pm, err := ParamsFromFieldsOrTags(p); err != nil {
		return "", err
	} else {
		return rc.queryManager.BuildQueryFromID(qid, pm)
	}
}

// DeleteQIDParams executes the supplied query with the expectation that it is a 'DELETE' query.
func (rc *RDBMSClient) DeleteQIDParams(qid string, params ...interface{}) (sql.Result, error) {

	return rc.execQIDParams(qid, params)

}

// DeleteQIDParam executes the supplied query with the expectation that it is a 'DELETE' query.
func (rc *RDBMSClient) DeleteQIDParam(qid string, name string, value interface{}) (sql.Result, error) {

	p := make(map[string]interface{})
	p[name] = value

	return rc.execQIDParams(qid, p)

}

// RegisterTempQuery stores the supplied query in the RDBMSClient so that it can be used with methods that expect a QID.
// Note that the query is NOT stored in the underlying QueryManager.
func (rc *RDBMSClient) RegisterTempQuery(qid string, query string) {
	rc.tempQueries[qid] = query
}

// ExistingIDOrInsertParams finds the ID of record or if the record does not exist, inserts a new record and retrieves the newly assigned ID
func (rc *RDBMSClient) ExistingIDOrInsertParams(checkQueryId, insertQueryId string, idTarget *int64, p ...interface{}) error {

	if found, err := rc.SelectBindSingleQIDParams(checkQueryId, p, idTarget); err != nil {
		return err
	} else if found {
		return nil
	} else {

		if err = rc.InsertCaptureQIDParams(insertQueryId, idTarget, p); err != nil {
			return err
		}

	}

	return nil
}

// InsertQIDParams executes the supplied query with the expectation that it is an 'INSERT' query.
func (rc *RDBMSClient) InsertQIDParams(qid string, params ...interface{}) (sql.Result, error) {

	return rc.execQIDParams(qid, params)

}

// InsertCaptureQIDParams executes the supplied query with the expectation that it is an 'INSERT' query and captures
// the new row's server generated ID in the target int64
func (rc *RDBMSClient) InsertCaptureQIDParams(qid string, target *int64, params ...interface{}, ) error {

	if query, err := rc.buildQuery(qid, params); err != nil {
		return err
	} else {

		return rc.lastID(query, rc, target)
	}

}

// SelectBindSingleQID executes the supplied query with the expectation that it is a 'SELECT' query that returns 0 or 1 rows.
// Results of the query are bound into the target struct. Returns false if no rows were found.
func (rc *RDBMSClient) SelectBindSingleQID(qid string, target interface{}) (bool, error) {
	return rc.SelectBindSingleQIDParams(qid, rc.emptyParams, target)
}

// SelectBindSingleQIDParam executes the supplied query with the expectation that it is a 'SELECT' query that returns 0 or 1 rows.
// Results of the query are bound into the target struct. Returns false if no rows were found.
func (rc *RDBMSClient) SelectBindSingleQIDParam(qid string, name string, value interface{}, target interface{}) (bool, error) {
	p := make(map[string]interface{})
	p[name] = value

	return rc.SelectBindSingleQIDParams(qid, p, target)
}

// SelectBindSingleQIDParams executes the supplied query with the expectation that it is a 'SELECT' query that returns 0 or 1 rows.
// Results of the query are bound into the target struct. Returns false if no rows were found.
func (rc *RDBMSClient) SelectBindSingleQIDParams(qid string, target interface{}, params ...interface{}) (bool, error) {

	var r *sql.Rows
	var err error

	if r, err = rc.SelectQIDParams(qid, params); err != nil {
		return false, err
	}

	defer r.Close()

	return rc.binder.BindRow(r, target)

}

// SelectBindQID executes the supplied query with the expectation that it is a 'SELECT' query. Results of the query
// are returned in a slice of the same type as the supplied template struct.
func (rc *RDBMSClient) SelectBindQID(qid string, template interface{}) ([]interface{}, error) {
	return rc.SelectBindQIDParams(qid, rc.emptyParams, template)
}

// SelectBindQIDParam executes the supplied query with the expectation that it is a 'SELECT' query. Results of the query
// are returned in a slice of the same type as the supplied template struct.
func (rc *RDBMSClient) SelectBindQIDParam(qid string, name string, value interface{}, template interface{}) ([]interface{}, error) {
	p := make(map[string]interface{})
	p[name] = value

	return rc.SelectBindQIDParams(qid, p, template)
}

// SelectBindQIDParams executes the supplied query with the expectation that it is a 'SELECT' query. Results of the query
// are returned in a slice of the same type as the supplied template struct.
func (rc *RDBMSClient) SelectBindQIDParams(qid string, template interface{}, params ...interface{}) ([]interface{}, error) {

	if r, err := rc.SelectQIDParams(qid, params); err != nil {
		return nil, err
	} else {

		defer r.Close()

		return rc.binder.BindRows(r, template)

	}

}

// SelectQID executes the supplied query with the expectation that it is a 'SELECT' query.
func (rc *RDBMSClient) SelectQID(qid string) (*sql.Rows, error) {
	return rc.SelectQIDParams(qid, rc.emptyParams)
}

// SelectQIDParam executes the supplied query with the expectation that it is a 'SELECT' query.
func (rc *RDBMSClient) SelectQIDParam(qid string, name string, value interface{}) (*sql.Rows, error) {
	p := make(map[string]interface{})
	p[name] = value

	return rc.SelectQIDParams(qid, p)
}

// SelectQIDParams executes the supplied query with the expectation that it is a 'SELECT' query.
func (rc *RDBMSClient) SelectQIDParams(qid string, params ...interface{}) (*sql.Rows, error) {
	query, err := rc.buildQuery(qid, params)

	if err != nil {
		return nil, err
	}

	return rc.Query(query)

}

// UpdateQIDParams executes the supplied query with the expectation that it is an 'UPDATE' query.
func (rc *RDBMSClient) UpdateQIDParams(qid string, params ...interface{}) (sql.Result, error) {

	return rc.execQIDParams(qid, params)

}

// UpdateQIDParam executes the supplied query with the expectation that it is an 'UPDATE' query.
func (rc *RDBMSClient) UpdateQIDParam(qid string, name string, value interface{}) (sql.Result, error) {

	p := make(map[string]interface{})
	p[name] = value

	return rc.execQIDParams(qid, p)

}

func (rc *RDBMSClient) execQIDParams(qid string, params ...interface{}) (sql.Result, error) {

	if query, err := rc.buildQuery(qid, params); err != nil {
		return nil, err
	} else {
		return rc.Exec(query)
	}

}

func (rc *RDBMSClient) buildQuery(qid string, p ...interface{}) (string, error) {

	tq := rc.tempQueries[qid]

	if tq != "" {
		return tq, nil
	} else {

		if pm, err := ParamsFromFieldsOrTags(p); err != nil {
			return "", err
		} else {
			return rc.queryManager.BuildQueryFromID(qid, pm)
		}
	}

}

// StartTransaction opens a transaction on the underlying sql.DB object and re-maps all calls to non-transactional
// methods to their transactional equivalents.
func (rc *RDBMSClient) StartTransaction() error {

	if rc.tx != nil {
		return errors.New("Transaction already open")
	} else {

		tx, err := rc.db.Begin()

		if err != nil {
			return err
		} else {
			rc.tx = tx
			return nil
		}
	}
}

// Rollback rolls the open transaction back - does nothing if no transaction is open.
func (rc *RDBMSClient) Rollback() {

	if rc.tx != nil {
		rc.tx.Rollback()
	}
}

// CommitTransaction commits the open transaction - does nothing if no transaction is open.
func (rc *RDBMSClient) CommitTransaction() error {

	if rc.tx == nil {
		return errors.New("No open transaction to commit")
	} else {

		err := rc.tx.Commit()
		rc.tx = nil

		return err
	}
}

// Exec is a pass-through to its sql.DB equivalent (or sql.Tx equivalent is a transaction is open)
func (rc *RDBMSClient) Exec(query string, args ...interface{}) (sql.Result, error) {

	tx := rc.tx

	if tx != nil {
		return tx.Exec(query, args...)
	} else {
		return rc.db.Exec(query, args...)
	}

}

// Query is a pass-through to its sql.DB equivalent (or sql.Tx equivalent is a transaction is open)
func (rc *RDBMSClient) Query(query string, args ...interface{}) (*sql.Rows, error) {
	tx := rc.tx

	if tx != nil {
		return tx.Query(query, args...)
	} else {
		return rc.db.Query(query, args...)
	}
}

// QueryRow is a pass-through to its sql.DB equivalent (or sql.Tx equivalent is a transaction is open)
func (rc *RDBMSClient) QueryRow(query string, args ...interface{}) *sql.Row {
	tx := rc.tx

	if tx != nil {
		return tx.QueryRow(query, args...)
	} else {
		return rc.db.QueryRow(query, args...)
	}
}
