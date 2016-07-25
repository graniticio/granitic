package rdbms

import (
	"database/sql"
	"errors"
	"github.com/graniticio/granitic/facility/querymanager"
)

func newRdbmsClient(database *sql.DB, querymanager *querymanager.QueryManager) *RdbmsClient {
	rc := new(RdbmsClient)
	rc.db = database
	rc.queryManager = querymanager

	return rc
}

type RdbmsClient struct {
	db           *sql.DB
	queryManager *querymanager.QueryManager
	tx           *sql.Tx
}

func (rc *RdbmsClient) InsertQueryIdParamMap(queryId string, params map[string]interface{}) (sql.Result, error) {

	query, err := rc.queryManager.SubstituteMap(queryId, params)

	if err != nil {
		return nil, err
	}

	result, err := rc.Exec(query)

	return result, err
}

func (rc *RdbmsClient) InsertQueryIdParamMapReturnedId(queryId string, params map[string]interface{}) (int, error) {

	query, err := rc.queryManager.SubstituteMap(queryId, params)

	if err != nil {
		return 0, err
	}

	var id int

	err = rc.QueryRow(query).Scan(&id)

	return id, err
}

func (rc *RdbmsClient) SelectQueryIdParamMap(queryId string, params map[string]interface{}) (*sql.Rows, error) {
	query, err := rc.queryManager.SubstituteMap(queryId, params)

	if err != nil {
		return nil, err
	}

	return rc.Query(query)

}

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

func (rc *RdbmsClient) Rollback() {

	if rc.tx != nil {
		rc.tx.Rollback()
	}
}

func (rc *RdbmsClient) CommitTransaction() error {

	if rc.tx == nil {
		return errors.New("No open transaction to commit")
	} else {

		err := rc.tx.Commit()
		rc.tx = nil

		return err
	}
}

func (rc *RdbmsClient) Exec(query string, args ...interface{}) (sql.Result, error) {

	tx := rc.tx

	if tx != nil {
		return tx.Exec(query, args...)
	} else {
		return rc.db.Exec(query, args...)
	}

}

func (rc *RdbmsClient) Query(query string, args ...interface{}) (*sql.Rows, error) {
	tx := rc.tx

	if tx != nil {
		return tx.Query(query, args...)
	} else {
		return rc.db.Query(query, args...)
	}
}

func (rc *RdbmsClient) QueryRow(query string, args ...interface{}) *sql.Row {
	tx := rc.tx

	if tx != nil {
		return tx.QueryRow(query, args...)
	} else {
		return rc.db.QueryRow(query, args...)
	}
}
