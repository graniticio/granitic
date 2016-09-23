package rdbms

import (
	"database/sql"
	"errors"
	"github.com/graniticio/granitic/dbquery"
)

func newRDBMSClient(database *sql.DB, querymanager querymanager.QueryManager, insertFunc InsertWithReturnedID) *RDBMSClient {
	rc := new(RDBMSClient)
	rc.db = database
	rc.queryManager = querymanager
	rc.lastID = insertFunc
	rc.emptyParams = make(map[string]interface{})
	rc.binder = new(RowBinder)
	return rc
}

type RDBMSClient struct {
	db           *sql.DB
	queryManager querymanager.QueryManager
	tx           *sql.Tx
	lastID       InsertWithReturnedID
	tempQueries  map[string]string
	emptyParams  map[string]interface{}
	binder       *RowBinder
}

func (rc *RDBMSClient) FindFragment(qid string) (string, error) {
	return rc.queryManager.FragmentFromID(qid)
}

func (rc *RDBMSClient) BuildQueryQIDTags(qid string, tagSource interface{}) (string, error) {
	if p, err := ParamsFromTags(tagSource); err != nil {
		return "", err
	} else {

		return rc.BuildQueryQIDParams(qid, p)
	}
}

func (rc *RDBMSClient) BuildQueryQIDParams(qid string, p map[string]interface{}) (string, error) {
	return rc.queryManager.BuildQueryFromID(qid, p)
}

func (rc *RDBMSClient) DeleteQIDTags(qid string, tagSource interface{}) (sql.Result, error) {

	if p, err := ParamsFromTags(tagSource); err != nil {
		return nil, err
	} else {

		return rc.DeleteQIDParams(qid, p)
	}

}

func (rc *RDBMSClient) DeleteQIDParams(qid string, params map[string]interface{}) (sql.Result, error) {

	return rc.execQIDParams(qid, params)

}

func (rc *RDBMSClient) DeleteQIDParam(qid string, name string, value interface{}) (sql.Result, error) {

	p := make(map[string]interface{})
	p[name] = value

	return rc.execQIDParams(qid, p)

}

func (rc *RDBMSClient) RegisterTempQuery(qid string, query string) {
	rc.tempQueries[qid] = query
}

// Finds the ID of record or if the record does not exist, inserts a new record and retrieves the newly assigned ID
func (rc *RDBMSClient) ExistingIDOrInsertTags(checkQueryId, insertQueryId string, idTarget *int64, tagSource ...interface{}) error {

	if p, err := ParamsFromTags(tagSource...); err != nil {
		return err
	} else {
		return rc.ExistingIDOrInsertParams(checkQueryId, insertQueryId, idTarget, p)
	}

}

func (rc *RDBMSClient) ExistingIDOrInsertParams(checkQueryId, insertQueryId string, idTarget *int64, p map[string]interface{}) error {

	if found, err := rc.SelectBindSingleQIDParams(checkQueryId, p, idTarget); err != nil {
		return err
	} else if found {
		return nil
	} else {

		if err = rc.InsertCaptureQIDParams(insertQueryId, p, idTarget); err != nil {
			return err
		}

	}

	return nil
}

func (rc *RDBMSClient) InsertQIDTags(qid string, tagSource interface{}) (sql.Result, error) {

	if p, err := ParamsFromTags(tagSource); err != nil {
		return nil, err
	} else {

		return rc.InsertQIDParams(qid, p)
	}

}

func (rc *RDBMSClient) InsertQIDParams(qid string, params map[string]interface{}) (sql.Result, error) {

	return rc.execQIDParams(qid, params)

}

func (rc *RDBMSClient) InsertCaptureQIDTags(qid string, tagSource interface{}, target *int64) error {
	if p, err := ParamsFromTags(tagSource); err != nil {
		return err
	} else {
		return rc.InsertCaptureQIDParams(qid, p, target)
	}
}

func (rc *RDBMSClient) InsertCaptureQIDParams(qid string, params map[string]interface{}, target *int64) error {

	if query, err := rc.buildQuery(qid, params); err != nil {
		return err
	} else {

		return rc.lastID(query, rc, target)
	}

}

func (rc *RDBMSClient) SelectBindSingleQID(qid string, target interface{}) (bool, error) {
	return rc.SelectBindSingleQIDParams(qid, rc.emptyParams, target)
}

func (rc *RDBMSClient) SelectBindSingleQIDTags(qid string, tagSource interface{}, target interface{}) (bool, error) {
	if p, err := ParamsFromTags(tagSource); err != nil {
		return false, err
	} else {
		return rc.SelectBindSingleQIDParams(qid, p, target)
	}
}

func (rc *RDBMSClient) SelectBindSingleQIDParam(qid string, name string, value interface{}, target interface{}) (bool, error) {
	p := make(map[string]interface{})
	p[name] = value

	return rc.SelectBindSingleQIDParams(qid, p, target)
}

func (rc *RDBMSClient) SelectBindSingleQIDParams(qid string, params map[string]interface{}, target interface{}) (bool, error) {

	var r *sql.Rows
	var err error

	if r, err = rc.SelectQIDParams(qid, params); err != nil {
		return false, err
	}

	defer r.Close()

	return rc.binder.BindRow(r, target)

}

func (rc *RDBMSClient) SelectBindQID(qid string, template interface{}) ([]interface{}, error) {
	return rc.SelectBindQIDParams(qid, rc.emptyParams, template)
}

func (rc *RDBMSClient) SelectBindQIDParam(qid string, name string, value interface{}, template interface{}) ([]interface{}, error) {
	p := make(map[string]interface{})
	p[name] = value

	return rc.SelectBindQIDParams(qid, p, template)
}

func (rc *RDBMSClient) SelectBindQIDParams(qid string, params map[string]interface{}, template interface{}) ([]interface{}, error) {

	if r, err := rc.SelectQIDParams(qid, params); err != nil {
		return nil, err
	} else {

		defer r.Close()

		return rc.binder.BindRows(r, template)

	}

}

func (rc *RDBMSClient) SelectBindQIDTags(qid string, tagSource interface{}, template interface{}) ([]interface{}, error) {
	if p, err := ParamsFromTags(tagSource); err != nil {
		return nil, err
	} else {
		return rc.SelectBindQIDParams(qid, p, template)
	}
}

func (rc *RDBMSClient) SelectQID(qid string) (*sql.Rows, error) {
	return rc.SelectQIDParams(qid, rc.emptyParams)
}

func (rc *RDBMSClient) SelectQIDParam(qid string, name string, value interface{}) (*sql.Rows, error) {
	p := make(map[string]interface{})
	p[name] = value

	return rc.SelectQIDParams(qid, p)
}

func (rc *RDBMSClient) SelectQIDParams(qid string, params map[string]interface{}) (*sql.Rows, error) {
	query, err := rc.buildQuery(qid, params)

	if err != nil {
		return nil, err
	}

	return rc.Query(query)

}

func (rc *RDBMSClient) UpdateQIDTags(qid string, tagSource interface{}) (sql.Result, error) {

	if p, err := ParamsFromTags(tagSource); err != nil {
		return nil, err
	} else {

		return rc.UpdateQIDParams(qid, p)
	}

}

func (rc *RDBMSClient) UpdateQIDParams(qid string, params map[string]interface{}) (sql.Result, error) {

	return rc.execQIDParams(qid, params)

}

func (rc *RDBMSClient) UpdateQIDParam(qid string, name string, value interface{}) (sql.Result, error) {

	p := make(map[string]interface{})
	p[name] = value

	return rc.execQIDParams(qid, p)

}

func (rc *RDBMSClient) execQIDParams(qid string, params map[string]interface{}) (sql.Result, error) {

	if query, err := rc.buildQuery(qid, params); err != nil {
		return nil, err
	} else {
		return rc.Exec(query)
	}

}

func (rc *RDBMSClient) buildQuery(qid string, p map[string]interface{}) (string, error) {

	tq := rc.tempQueries[qid]

	if tq != "" {
		return tq, nil
	} else {
		return rc.queryManager.BuildQueryFromID(qid, p)
	}

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
