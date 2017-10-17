package rdbms

import (
	"database/sql"
	"context"
	"testing"
	"github.com/graniticio/granitic/types"
	"fmt"
)

func TestNonTxSelectMethods(t *testing.T) {

	/*qm := new(testQueryManagerProxy)
	db := new(testDbProxy)
	p1, p2 := testStandardParams()

	c := newRdbmsClient(db, qm, DefaultInsertWithReturnedId)

	bt := new(testTargetSingle)

	_, err := c.SelectBindSingleQIDParams("SBSQP", bt, p1, p2)

	fmt.Println(err)
	fmt.Println(db.lastMethod)

	if !paramMergedCorrectly(qm.lastParams) {
		t.FailNow()
	}*/
}

func testStandardParams() (interface{}, interface{}){

	tp := new(testParam)

	tp.IParam = 44
	tp.NSParam = types.NewNilableString("NS")
	tp.SParam = "S"


	pm := make(map[string]interface{})

	pm["NSParam"] = "NS1"
	pm["BParam"] = false

	return tp, pm
}

func paramMergedCorrectly(p map[string]interface{}) bool {

	if len(p) != 4 {
		fmt.Printf("Expected 4 params got %d\n", len(p))
		return false
	}

	return true
}

type testTargetSingle struct {

}

type testParam struct {
	SParam string
	NSParam *types.NilableString
	IParam int `dbparam:"IOV"`
}


type testDbProxy struct {
	lastMethod string
	lastQuery string
}

func (tdp *testDbProxy) used() bool {
	return tdp.lastMethod != ""
}

func (tdp *testDbProxy) reset() {
	tdp.lastMethod = ""
	tdp.lastQuery = ""
}

func (tdp *testDbProxy) Exec(query string, args ...interface{}) (sql.Result, error){

	tdp.lastMethod = "Exec"
	tdp.lastQuery = query

	return mockResult{}, nil
}

func (tdp *testDbProxy) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {

	tdp.lastMethod = "ExecContext"
	tdp.lastQuery = query

	return mockResult{}, nil
}

func (tdp *testDbProxy) Query(query string, args ...interface{}) (*sql.Rows, error) {
	tdp.lastMethod = "Query"
	tdp.lastQuery = query

	r := new(sql.Rows)



	return r, nil
}

func (tdp *testDbProxy) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	tdp.lastMethod = "QueryContext"
	tdp.lastQuery = query

	return new(sql.Rows), nil
}

func (tdp *testDbProxy) QueryRow(query string, args ...interface{}) *sql.Row {
	tdp.lastMethod = "QueryRow"
	tdp.lastQuery = query

	return new(sql.Row)
}

func (tdp *testDbProxy) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	tdp.lastMethod = "QueryRowContext"
	tdp.lastQuery = query

	return new(sql.Row)
}

func (tdp *testDbProxy) Begin() (*sql.Tx, error) {
	return nil, nil
}

func (tdp *testDbProxy) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return nil, nil
}

type testQueryManagerProxy struct {
	lastParams map[string]interface{}
}

func (tqm *testQueryManagerProxy) reset() {
	tqm.lastParams = nil
}

func (tqm *testQueryManagerProxy) BuildQueryFromID(qid string, params map[string]interface{}) (string, error) {
	tqm.lastParams = params

	return qid, nil

}

func (tqm *testQueryManagerProxy) FragmentFromID(qid string) (string, error) {
	return qid, nil
}

type mockResult struct {
	lid int64
	ra int64
}

func (mr mockResult) LastInsertId() (int64, error) {
	return mr.lid, nil
}

func  (mr mockResult) RowsAffected() (int64, error) {
	return mr.ra, nil
}