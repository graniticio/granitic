// Copyright 2016 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package rdbms

import (
	"database/sql"
	"errors"
	"github.com/graniticio/granitic/dsquery"
	"context"
)

func newRdbmsClient(database dbProxy, querymanager dsquery.QueryManager, insertFunc InsertWithReturnedId) *RdbmsClient {
	rc := new(RdbmsClient)
	rc.db = database
	rc.queryManager = querymanager
	rc.lastID = insertFunc
	rc.emptyParams = make(map[string]interface{})
	rc.binder = new(RowBinder)
	return rc
}

// The interface application code should use to execute SQL against a database. See the package overview for the rdbms
// package for usage.
//
// RdbmsClient is stateful and MUST NOT be shared across goroutines
type RdbmsClient struct {
	db           dbProxy
	queryManager dsquery.QueryManager
	tx           txProxy
	lastID       InsertWithReturnedId
	tempQueries  map[string]string
	emptyParams  map[string]interface{}
	binder       *RowBinder
	ctx          context.Context
}

// FindFragment returns a partial query from the underlying QueryManager. Fragments are no
// different that ordinary template queries, except they are not expected to contain any variable placeholders.
func (rc *RdbmsClient) FindFragment(qid string) (string, error) {
	return rc.queryManager.FragmentFromID(qid)
}

// BuildQueryQIDParams returns a populated SQL query that can be manually executed later.
func (rc *RdbmsClient) BuildQueryQIDParams(qid string, p ...interface{}) (string, error) {

	if pm, err := ParamsFromFieldsOrTags(p...); err != nil {
		return "", err
	} else {
		return rc.queryManager.BuildQueryFromID(qid, pm)
	}
}

// DeleteQIDParams executes the supplied query with the expectation that it is a 'DELETE' query.
func (rc *RdbmsClient) DeleteQIDParams(qid string, params ...interface{}) (sql.Result, error) {

	return rc.execQIDParams(qid, params...)

}

// DeleteQIDParam executes the supplied query with the expectation that it is a 'DELETE' query.
func (rc *RdbmsClient) DeleteQIDParam(qid string, name string, value interface{}) (sql.Result, error) {

	p := make(map[string]interface{})
	p[name] = value

	return rc.execQIDParams(qid, p)

}

// RegisterTempQuery stores the supplied query in the RdbmsClient so that it can be used with methods that expect a QID.
// Note that the query is NOT stored in the underlying QueryManager.
func (rc *RdbmsClient) RegisterTempQuery(qid string, query string) {
	rc.tempQueries[qid] = query
}

// ExistingIDOrInsertParams finds the ID of record or if the record does not exist, inserts a new record and retrieves the newly assigned ID
func (rc *RdbmsClient) ExistingIDOrInsertParams(checkQueryId, insertQueryId string, idTarget *int64, p ...interface{}) error {

	if found, err := rc.SelectBindSingleQIDParams(checkQueryId, idTarget, p...); err != nil {
		return err
	} else if found {
		return nil
	} else {

		if err = rc.InsertCaptureQIDParams(insertQueryId, idTarget, p...); err != nil {
			return err
		}

	}

	return nil
}

// InsertQIDParams executes the supplied query with the expectation that it is an 'INSERT' query.
func (rc *RdbmsClient) InsertQIDParams(qid string, params ...interface{}) (sql.Result, error) {

	return rc.execQIDParams(qid, params...)

}

// InsertCaptureQIDParams executes the supplied query with the expectation that it is an 'INSERT' query and captures
// the new row's server generated ID in the target int64
func (rc *RdbmsClient) InsertCaptureQIDParams(qid string, target *int64, params ...interface{}) error {

	if query, err := rc.buildQuery(qid, params...); err != nil {
		return err
	} else {

		return rc.lastID(query, rc, target)
	}

}

// SelectBindSingleQID executes the supplied query with the expectation that it is a 'SELECT' query that returns 0 or 1 rows.
// Results of the query are bound into the target struct. Returns false if no rows were found.
func (rc *RdbmsClient) SelectBindSingleQID(qid string, target interface{}) (bool, error) {
	return rc.SelectBindSingleQIDParams(qid, rc.emptyParams, target)
}

// SelectBindSingleQIDParam executes the supplied query with the expectation that it is a 'SELECT' query that returns 0 or 1 rows.
// Results of the query are bound into the target struct. Returns false if no rows were found.
func (rc *RdbmsClient) SelectBindSingleQIDParam(qid string, name string, value interface{}, target interface{}) (bool, error) {
	p := make(map[string]interface{})
	p[name] = value

	return rc.SelectBindSingleQIDParams(qid, p, target)
}

// SelectBindSingleQIDParams executes the supplied query with the expectation that it is a 'SELECT' query that returns 0 or 1 rows.
// Results of the query are bound into the target struct. Returns false if no rows were found.
func (rc *RdbmsClient) SelectBindSingleQIDParams(qid string, target interface{}, params ...interface{}) (bool, error) {

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
func (rc *RdbmsClient) SelectBindQID(qid string, template interface{}) ([]interface{}, error) {
	return rc.SelectBindQIDParams(qid, rc.emptyParams, template)
}

// SelectBindQIDParam executes the supplied query with the expectation that it is a 'SELECT' query. Results of the query
// are returned in a slice of the same type as the supplied template struct.
func (rc *RdbmsClient) SelectBindQIDParam(qid string, name string, value interface{}, template interface{}) ([]interface{}, error) {
	p := make(map[string]interface{})
	p[name] = value

	return rc.SelectBindQIDParams(qid, p, template)
}

// SelectBindQIDParams executes the supplied query with the expectation that it is a 'SELECT' query. Results of the query
// are returned in a slice of the same type as the supplied template struct.
func (rc *RdbmsClient) SelectBindQIDParams(qid string, template interface{}, params ...interface{}) ([]interface{}, error) {

	if r, err := rc.SelectQIDParams(qid, params...); err != nil {
		return nil, err
	} else {

		defer r.Close()

		return rc.binder.BindRows(r, template)

	}

}

// SelectQID executes the supplied query with the expectation that it is a 'SELECT' query.
func (rc *RdbmsClient) SelectQID(qid string) (*sql.Rows, error) {
	return rc.SelectQIDParams(qid, rc.emptyParams)
}

// SelectQIDParam executes the supplied query with the expectation that it is a 'SELECT' query.
func (rc *RdbmsClient) SelectQIDParam(qid string, name string, value interface{}) (*sql.Rows, error) {
	p := make(map[string]interface{})
	p[name] = value

	return rc.SelectQIDParams(qid, p)
}

// SelectQIDParams executes the supplied query with the expectation that it is a 'SELECT' query.
func (rc *RdbmsClient) SelectQIDParams(qid string, params ...interface{}) (*sql.Rows, error) {

	query, err := rc.buildQuery(qid, params...)

	if err != nil {
		return nil, err
	}

	return rc.Query(query)

}

// UpdateQIDParams executes the supplied query with the expectation that it is an 'UPDATE' query.
func (rc *RdbmsClient) UpdateQIDParams(qid string, params ...interface{}) (sql.Result, error) {

	return rc.execQIDParams(qid, params...)

}

// UpdateQIDParam executes the supplied query with the expectation that it is an 'UPDATE' query.
func (rc *RdbmsClient) UpdateQIDParam(qid string, name string, value interface{}) (sql.Result, error) {

	p := make(map[string]interface{})
	p[name] = value

	return rc.execQIDParams(qid, p)

}

func (rc *RdbmsClient) execQIDParams(qid string, params ...interface{}) (sql.Result, error) {

	if query, err := rc.buildQuery(qid, params...); err != nil {
		return nil, err
	} else {

		return rc.Exec(query)
	}

}

func (rc *RdbmsClient) buildQuery(qid string, p ...interface{}) (string, error) {

	tq := rc.tempQueries[qid]

	if tq != "" {
		return tq, nil
	} else {

		if pm, err := ParamsFromFieldsOrTags(p...); err != nil {
			return "", err
		} else {
			return rc.queryManager.BuildQueryFromID(qid, pm)
		}
	}

}

// StartTransaction opens a transaction on the underlying sql.DB object and re-maps all calls to non-transactional
// methods to their transactional equivalents.
func (rc *RdbmsClient) StartTransaction() error {

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

// StartTransactionWithOptions opens a transaction on the underlying sql.DB object and re-maps all calls to non-transactional
// methods to their transactional equivalents.
func (rc *RdbmsClient) StartTransactionWithOptions(opts *sql.TxOptions) error {

	if rc.tx != nil {
		return errors.New("Transaction already open")
	} else {

		var ctx context.Context

		if rc.contextAware() {
			ctx = rc.ctx
		} else {
			ctx = context.Background()
		}

		tx, err := rc.db.BeginTx(ctx, opts)

		if err != nil {
			return err
		} else {
			rc.tx = tx
			return nil
		}
	}
}

// Rollback rolls the open transaction back - does nothing if no transaction is open.
func (rc *RdbmsClient) Rollback() {

	if rc.tx != nil {
		rc.tx.Rollback()
	}
}

// CommitTransaction commits the open transaction - does nothing if no transaction is open.
func (rc *RdbmsClient) CommitTransaction() error {

	if rc.tx == nil {
		return errors.New("No open transaction to commit")
	} else {

		err := rc.tx.Commit()
		rc.tx = nil

		return err
	}
}

// Exec is a pass-through to its sql.DB equivalent (or sql.Tx equivalent is a transaction is open)
func (rc *RdbmsClient) Exec(query string, args ...interface{}) (sql.Result, error) {

	tx := rc.tx

	if rc.contextAware() {
		if tx != nil {
			return tx.ExecContext(rc.ctx, query, args...)
		} else {
			return rc.db.ExecContext(rc.ctx, query, args...)
		}
	} else {
		if tx != nil {
			return tx.Exec(query, args...)
		} else {
			return rc.db.Exec(query, args...)
		}
	}

}

// Query is a pass-through to its sql.DB equivalent (or sql.Tx equivalent is a transaction is open)
func (rc *RdbmsClient) Query(query string, args ...interface{}) (*sql.Rows, error) {
	tx := rc.tx

	if rc.contextAware() {
		if tx != nil {
			return tx.QueryContext(rc.ctx, query, args...)
		} else {
			return rc.db.QueryContext(rc.ctx, query, args...)
		}
	} else {
		if tx != nil {
			return tx.Query(query, args...)
		} else {
			return rc.db.Query(query, args...)
		}
	}
}

// QueryRow is a pass-through to its sql.DB equivalent (or sql.Tx equivalent is a transaction is open)
func (rc *RdbmsClient) QueryRow(query string, args ...interface{}) *sql.Row {
	tx := rc.tx

	if rc.contextAware() {
		if tx != nil {
			return tx.QueryRowContext(rc.ctx, query, args...)
		} else {
			return rc.db.QueryRowContext(rc.ctx, query, args...)
		}
	} else {
		if tx != nil {
			return tx.QueryRow(query, args...)
		} else {
			return rc.db.QueryRow(query, args...)
		}
	}



}

func (rc *RdbmsClient) contextAware() bool {
	return rc.ctx != nil
}