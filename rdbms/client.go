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

func newRdbmsClient(database *sql.DB, querymanager dsquery.QueryManager, insertFunc InsertWithReturnedId, logger logging.Logger) *Client {
	rc := new(Client)
	rc.db = database
	rc.queryManager = querymanager
	rc.lastId = insertFunc
	rc.emptyParams = make(map[string]interface{})
	rc.binder = new(RowBinder)
	rc.tempQueries = make(map[string]string)

	rc.FrameworkLogger = logger

	return rc
}

// The interface application code should use to execute SQL against a database. See the package overview for the rdbms
// package for usage.
//
// Client is stateful and MUST NOT be shared across goroutines
type Client struct {
	db              *sql.DB
	queryManager    dsquery.QueryManager
	tx              *sql.Tx
	lastId          InsertWithReturnedId
	tempQueries     map[string]string
	emptyParams     map[string]interface{}
	binder          *RowBinder
	ctx             context.Context
	FrameworkLogger logging.Logger
}

// FindFragment returns a partial query from the underlying QueryManager. Fragments are no
// different that ordinary template queries, except they are not expected to contain any variable placeholders.
func (rc *Client) FindFragment(qid string) (string, error) {
	return rc.queryManager.FragmentFromId(qid)
}

// BuildQueryFromQIdParams returns a populated SQL query that can be manually executed later.
func (rc *Client) BuildQueryFromQIdParams(qid string, p ...interface{}) (string, error) {

	if pm, err := ParamsFromFieldsOrTags(p...); err != nil {
		return "", err
	} else {
		return rc.queryManager.BuildQueryFromId(qid, pm)
	}
}

// DeleteQIdParams executes the supplied query with the expectation that it is a 'DELETE' query.
func (rc *Client) DeleteQIdParams(qid string, params ...interface{}) (sql.Result, error) {

	return rc.execQIdParams(qid, params...)

}

// DeleteQIdParam executes the supplied query with the expectation that it is a 'DELETE' query.
func (rc *Client) DeleteQIdParam(qid string, name string, value interface{}) (sql.Result, error) {

	p := make(map[string]interface{})
	p[name] = value

	return rc.execQIdParams(qid, p)

}

// RegisterTempQuery stores the supplied query in the Client so that it can be used with methods that expect a QID.
// Note that the query is NOT stored in the underlying QueryManager.
func (rc *Client) RegisterTempQuery(qid string, query string) {
	rc.tempQueries[qid] = query
}

// ExistingIdOrInsertParams finds the ID of record or if the record does not exist, inserts a new record and retrieves the newly assigned ID
func (rc *Client) ExistingIdOrInsertParams(checkQueryId, insertQueryId string, idTarget *int64, p ...interface{}) error {

	if found, err := rc.SelectBindSingleQIdParams(checkQueryId, idTarget, p...); err != nil {
		return err
	} else if found {
		return nil
	} else {

		if err = rc.InsertCaptureQIdParams(insertQueryId, idTarget, p...); err != nil {
			return err
		}

	}

	return nil
}

// InsertQIdParams executes the supplied query with the expectation that it is an 'INSERT' query.
func (rc *Client) InsertQIdParams(qid string, params ...interface{}) (sql.Result, error) {

	return rc.execQIdParams(qid, params...)

}

// InsertCaptureQIdParams executes the supplied query with the expectation that it is an 'INSERT' query and captures
// the new row's server generated ID in the target int64
func (rc *Client) InsertCaptureQIdParams(qid string, target *int64, params ...interface{}) error {

	if query, err := rc.buildQuery(qid, params...); err != nil {
		return err
	} else {

		return rc.lastId(query, rc, target)
	}

}

// SelectBindSingleQId executes the supplied query with the expectation that it is a 'SELECT' query that returns 0 or 1 rows.
// Results of the query are bound into the target struct. Returns false if no rows were found.
func (rc *Client) SelectBindSingleQId(qid string, target interface{}) (bool, error) {
	return rc.SelectBindSingleQIdParams(qid, target, rc.emptyParams)
}

// SelectBindSingleQIdParam executes the supplied query with the expectation that it is a 'SELECT' query that returns 0 or 1 rows.
// Results of the query are bound into the target struct. Returns false if no rows were found.
func (rc *Client) SelectBindSingleQIdParam(qid string, name string, value interface{}, target interface{}) (bool, error) {
	p := make(map[string]interface{})
	p[name] = value

	return rc.SelectBindSingleQIdParams(qid, target, p)
}

// SelectBindSingleQIdParams executes the supplied query with the expectation that it is a 'SELECT' query that returns 0 or 1 rows.
// Results of the query are bound into the target struct. Returns false if no rows were found.
func (rc *Client) SelectBindSingleQIdParams(qid string, target interface{}, params ...interface{}) (bool, error) {

	var r *sql.Rows
	var err error

	if r, err = rc.SelectQIdParams(qid, params...); err != nil {
		return false, err
	}

	defer r.Close()

	return rc.binder.BindRow(r, target)

}

// SelectBindQId executes the supplied query with the expectation that it is a 'SELECT' query. Results of the query
// are returned in a slice of the same type as the supplied template struct.
func (rc *Client) SelectBindQId(qid string, template interface{}) ([]interface{}, error) {
	return rc.SelectBindQIdParams(qid, template, rc.emptyParams)
}

// SelectBindQIdParam executes the supplied query with the expectation that it is a 'SELECT' query. Results of the query
// are returned in a slice of the same type as the supplied template struct.
func (rc *Client) SelectBindQIdParam(qid string, name string, value interface{}, template interface{}) ([]interface{}, error) {
	p := make(map[string]interface{})
	p[name] = value

	return rc.SelectBindQIdParams(qid, template, p)
}

// SelectBindQIdParams executes the supplied query with the expectation that it is a 'SELECT' query. Results of the query
// are returned in a slice of the same type as the supplied template struct.
func (rc *Client) SelectBindQIdParams(qid string, template interface{}, params ...interface{}) ([]interface{}, error) {

	if r, err := rc.SelectQIdParams(qid, params...); err != nil {
		return nil, err
	} else {

		defer r.Close()

		return rc.binder.BindRows(r, template)

	}

}

// SelectQId executes the supplied query with the expectation that it is a 'SELECT' query.
func (rc *Client) SelectQId(qid string) (*sql.Rows, error) {
	return rc.SelectQIdParams(qid, rc.emptyParams)
}

// SelectQIdParam executes the supplied query with the expectation that it is a 'SELECT' query.
func (rc *Client) SelectQIdParam(qid string, name string, value interface{}) (*sql.Rows, error) {
	p := make(map[string]interface{})
	p[name] = value

	return rc.SelectQIdParams(qid, p)
}

// SelectQIdParams executes the supplied query with the expectation that it is a 'SELECT' query.
func (rc *Client) SelectQIdParams(qid string, params ...interface{}) (*sql.Rows, error) {

	query, err := rc.buildQuery(qid, params...)

	if err != nil {
		return nil, err
	}

	return rc.Query(query)

}

// UpdateQIdParams executes the supplied query with the expectation that it is an 'UPDATE' query.
func (rc *Client) UpdateQIdParams(qid string, params ...interface{}) (sql.Result, error) {

	return rc.execQIdParams(qid, params...)

}

// UpdateQIdParam executes the supplied query with the expectation that it is an 'UPDATE' query.
func (rc *Client) UpdateQIdParam(qid string, name string, value interface{}) (sql.Result, error) {

	p := make(map[string]interface{})
	p[name] = value

	return rc.execQIdParams(qid, p)

}

func (rc *Client) execQIdParams(qid string, params ...interface{}) (sql.Result, error) {

	if query, err := rc.buildQuery(qid, params...); err != nil {
		return nil, err
	} else {

		return rc.Exec(query)
	}

}

func (rc *Client) buildQuery(qid string, p ...interface{}) (string, error) {

	tq := rc.tempQueries[qid]

	if tq != "" {
		return tq, nil
	} else {

		if pm, err := ParamsFromFieldsOrTags(p...); err != nil {
			return "", err
		} else {

			if rc.FrameworkLogger.IsLevelEnabled(logging.Trace) {
				//Log the parameters to be injected into the query
				rc.FrameworkLogger.LogTracef("Parameters: %v", pm)
			}

			return rc.queryManager.BuildQueryFromId(qid, pm)
		}
	}

}

// StartTransaction opens a transaction on the underlying sql.DB object and re-maps all calls to non-transactional
// methods to their transactional equivalents.
func (rc *Client) StartTransaction() error {

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
func (rc *Client) StartTransactionWithOptions(opts *sql.TxOptions) error {

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
func (rc *Client) Rollback() {

	if rc.tx != nil {
		rc.tx.Rollback()
	}
}

// CommitTransaction commits the open transaction - does nothing if no transaction is open.
func (rc *Client) CommitTransaction() error {

	if rc.tx == nil {
		return errors.New("No open transaction to commit")
	} else {

		err := rc.tx.Commit()
		rc.tx = nil

		return err
	}
}

// Exec is a pass-through to its sql.DB equivalent (or sql.Tx equivalent is a transaction is open)
func (rc *Client) Exec(query string, args ...interface{}) (sql.Result, error) {

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
func (rc *Client) Query(query string, args ...interface{}) (*sql.Rows, error) {
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
func (rc *Client) QueryRow(query string, args ...interface{}) *sql.Row {
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

func (rc *Client) contextAware() bool {
	return rc.ctx != nil
}
