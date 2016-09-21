package rdbms

import (
	"database/sql"
	"errors"
	"github.com/graniticio/granitic/dbquery"
)

type InsertWithReturnedID func(string, *RDBMSClient) (int64, error)

func newRDBMSClient(database *sql.DB, querymanager querymanager.QueryManager, insertFunc InsertWithReturnedID) *RDBMSClient {
	rc := new(RDBMSClient)
	rc.db = database
	rc.queryManager = querymanager
	rc.lastID = insertFunc

	return rc
}

type RDBMSClient struct {
	db           *sql.DB
	queryManager querymanager.QueryManager
	tx           *sql.Tx
	lastID       InsertWithReturnedID
}

func (rc *RDBMSClient) InsertIDTags(queryId string, tagSource interface{}) (sql.Result, error) {

	if p, err := ParamsFromTags(tagSource); err != nil {
		return nil, err
	} else {

		return rc.InsertIDParams(queryId, p)
	}

}

func (rc *RDBMSClient) InsertIDParams(queryId string, params map[string]interface{}) (sql.Result, error) {

	if query, err := rc.queryManager.SubstituteMap(queryId, params); err != nil {
		return nil, err
	} else {
		return rc.Exec(query)
	}

}

func (rc *RDBMSClient) InsertIDTagsAssigned(queryId string, tagSource interface{}) (int64, error) {
	if p, err := ParamsFromTags(tagSource); err != nil {
		return 0, err
	} else {

		return rc.InsertIDParamsAssigned(queryId, p)
	}
}

func (rc *RDBMSClient) InsertIDParamsAssigned(queryId string, params map[string]interface{}) (int64, error) {

	if query, err := rc.queryManager.SubstituteMap(queryId, params); err != nil {
		return 0, err
	} else {

		return rc.lastID(query, rc)
	}

}

func (rc *RDBMSClient) SelectIDParam(queryId string, name string, value interface{}) (*sql.Rows, error) {
	p := make(map[string]interface{})
	p[name] = value

	return rc.SelectIDParams(queryId, p)
}

func (rc *RDBMSClient) SelectIDParams(queryId string, params map[string]interface{}) (*sql.Rows, error) {
	query, err := rc.queryManager.SubstituteMap(queryId, params)

	if err != nil {
		return nil, err
	}

	return rc.Query(query)

}

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

func (rc *RDBMSClient) Rollback() {

	if rc.tx != nil {
		rc.tx.Rollback()
	}
}

func (rc *RDBMSClient) CommitTransaction() error {

	if rc.tx == nil {
		return errors.New("No open transaction to commit")
	} else {

		err := rc.tx.Commit()
		rc.tx = nil

		return err
	}
}

func (rc *RDBMSClient) Exec(query string, args ...interface{}) (sql.Result, error) {

	tx := rc.tx

	if tx != nil {
		return tx.Exec(query, args...)
	} else {
		return rc.db.Exec(query, args...)
	}

}

func (rc *RDBMSClient) Query(query string, args ...interface{}) (*sql.Rows, error) {
	tx := rc.tx

	if tx != nil {
		return tx.Query(query, args...)
	} else {
		return rc.db.Query(query, args...)
	}
}

func (rc *RDBMSClient) QueryRow(query string, args ...interface{}) *sql.Row {
	tx := rc.tx

	if tx != nil {
		return tx.QueryRow(query, args...)
	} else {
		return rc.db.QueryRow(query, args...)
	}
}
