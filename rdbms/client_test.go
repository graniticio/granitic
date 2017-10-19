package rdbms

import (
	"database/sql"
	"testing"
	"github.com/graniticio/granitic/types"
	"fmt"
	"database/sql/driver"
	"os"
	"io"
	"github.com/graniticio/granitic/test"
	"time"
	"github.com/graniticio/granitic/reflecttools"
)


var db *sql.DB
var drv *mockDriver
var qm *testQueryManagerProxy

func TestMain(m *testing.M) {

	var err error

	drv = new(mockDriver)
	qm = new(testQueryManagerProxy)

	sql.Register("grnc-mock", drv)

	db, err = sql.Open("grnc-mock", "")

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}

	os.Exit(m.Run())
}

func TestNonTxSelectMethods(t *testing.T) {

	c := newRdbmsClient(db, qm, DefaultInsertWithReturnedId)

	testSelectMethods(t, c)

}

func TestTxSelectMethods(t *testing.T) {

	c := newRdbmsClient(db, qm, DefaultInsertWithReturnedId)
	c.StartTransaction()
	testSelectMethods(t, c)
	c.CommitTransaction()

}

func testSelectMethods(t *testing.T, c *RdbmsClient) {

	p1, p2 := testStandardParams()



	//SelectBindQId
	drv.colNames = []string{"Int64Result"}
	drv.rowData = [][]driver.Value{{int64(45)},{int64(32)}}

	bt := new(testTarget)
	results, err := c.SelectBindQId("SBQ", bt)

	test.ExpectNil(t, err)
	test.ExpectInt(t, len(results), 2)

	test.ExpectInt(t, int(results[0].(*testTarget).Int64Result), 45)
	test.ExpectInt(t, int(results[1].(*testTarget).Int64Result), 32)

	results, err = c.SelectBindQId("SBQ", bt)

	test.ExpectNil(t, err)
	test.ExpectInt(t, len(results), 0)


	//SelectBindQIdParam
	drv.colNames = []string{"Float64Result"}
	drv.rowData = [][]driver.Value{{float64(123.1)}}

	results, err = c.SelectBindQIdParam("SBQP", "p1", "v1", bt)

	test.ExpectNil(t, err)
	test.ExpectInt(t, len(results), 1)

	test.ExpectFloat(t, results[0].(*testTarget).Float64Result, float64(123.1))

	test.ExpectInt(t, len(qm.lastParams), 1)
	test.ExpectString(t, qm.lastParams["p1"].(string), "v1")

	results, err = c.SelectBindQIdParam("SBQP", "p1", "v1", bt)
	test.ExpectNil(t, err)
	test.ExpectInt(t, len(results), 0)


	//SelectBindQIdParams
	drv.colNames = []string{"BoolResult"}
	drv.rowData = [][]driver.Value{{true},{false},{true}}

	results, err = c.SelectBindQIdParams("SBQPs", bt, p1, p2)

	test.ExpectNil(t, err)
	test.ExpectInt(t, len(results), 3)

	test.ExpectBool(t, results[2].(*testTarget).BoolResult, true)

	if !paramMergedCorrectly(qm.lastParams) {
		t.FailNow()
	}


	//SelectBindSingleQId
	drv.colNames = []string{"TimeResult"}


	drv.rowData = [][]driver.Value{{time.Now()}}

	bt = new(testTarget)

	found, err := c.SelectBindSingleQId("SBSQ", bt)

	test.ExpectNil(t, err)
	test.ExpectBool(t, found, true)
	test.ExpectBool(t, reflecttools.IsZero(bt.TimeResult), false)

	bt = new(testTarget)

	found, err = c.SelectBindSingleQId("SBSQ", bt)
	test.ExpectNil(t, err)
	test.ExpectBool(t, found, false)
	test.ExpectBool(t, reflecttools.IsZero(bt.TimeResult), true)


	//SelectBindSingleQIdParam
	drv.colNames = []string{"StrResult", "Int64Result"}
	drv.rowData = [][]driver.Value{{"okay", int64(1)},{"not",int64(2)}}

	bt = new(testTarget)

	found, err = c.SelectBindSingleQIdParam("SBSQ", "p1", "v1", bt)
	test.ExpectNotNil(t, err)
	test.ExpectInt(t, len(qm.lastParams), 1)
	test.ExpectString(t, qm.lastParams["p1"].(string), "v1")

	drv.colNames = []string{"StrResult", "Int64Result"}
	drv.rowData = [][]driver.Value{{"okay", int64(1)}}

	found, err = c.SelectBindSingleQIdParam("SBSQ", "p1", "v1", bt)

	test.ExpectNil(t, err)
	test.ExpectBool(t, found, true)


	test.ExpectInt(t, int(bt.Int64Result), 1)
	test.ExpectString(t, bt.StrResult, "okay")


	//SelectBindSingleQIdParams
	drv.colNames = []string{"StrResult"}
	drv.rowData = [][]driver.Value{{"okay"}}


	bt = new(testTarget)
	found, err = c.SelectBindSingleQIdParams("SBSQP", bt, p1, p2)

	test.ExpectNil(t, err)
	test.ExpectBool(t, found, true)
	test.ExpectString(t, bt.StrResult, "okay")

	if !paramMergedCorrectly(qm.lastParams) {
		t.FailNow()
	}

	found, err = c.SelectBindSingleQIdParams("SBSQP", bt, p1, p2)

	test.ExpectNil(t, err)
	test.ExpectBool(t, found, false)


	//SelectQId
	drv.colNames = []string{"StrResult"}
	drv.rowData = [][]driver.Value{{"okay"}}

	r, err := c.SelectQId("SQ")

	test.ExpectNil(t, err)
	test.ExpectBool(t, r.Next(), true)


	//SelectQIdParam
	drv.colNames = []string{"StrResult"}
	drv.rowData = [][]driver.Value{{"okay","not"}}
	r, err = c.SelectQIdParam("SQP", "p1", "v1")
	test.ExpectNil(t, err)
	test.ExpectBool(t, r.Next(), true)
	test.ExpectBool(t, r.Next(), true)


	//SelectQIdParams
	drv.colNames = []string{"StrResult"}
	drv.rowData = [][]driver.Value{{"okay","not"}}


	r, err = c.SelectQIdParams("SQPs", p1, p2)
	test.ExpectNil(t, err)
	test.ExpectBool(t, r.Next(), true)
	test.ExpectBool(t, r.Next(), true)
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

type testTarget struct {
	StrResult string
	Int64Result int64
	Float64Result float64
	BoolResult bool
	ByteArrayResult []byte
	TimeResult time.Time
}

type testParam struct {
	SParam string
	NSParam *types.NilableString
	IParam int `dbparam:"IOV"`
}




type testQueryManagerProxy struct {
	lastParams map[string]interface{}
}

func (tqm *testQueryManagerProxy) reset() {
	tqm.lastParams = nil
}

func (tqm *testQueryManagerProxy) BuildQueryFromId(qid string, params map[string]interface{}) (string, error) {
	tqm.lastParams = params

	return qid, nil

}

func (tqm *testQueryManagerProxy) FragmentFromId(qid string) (string, error) {
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

type mockDriver struct {
	colNames []string
	rowData [][]driver.Value
}

func (d *mockDriver) consumed() {
	d.colNames = nil
	d.rowData = nil
}

func (d *mockDriver) Open(name string) (driver.Conn, error) {
	return newMockConn(d), nil
}


func newMockConn(d *mockDriver) *mockConn{
	c := new(mockConn)
	c.d = d

	return c
}

type mockConn struct {
	d *mockDriver
}

func (c *mockConn) Prepare(query string) (driver.Stmt, error) {
	return newMockStmt(c.d), nil
}

func (c *mockConn) Close() error {
	return nil
}

func (c *mockConn) Begin() (driver.Tx, error) {
	return new(mockTx), nil
}


func newMockStmt(d *mockDriver) *mockStmt{
	s := new(mockStmt)
	s.d = d

	return s
}

type mockStmt struct {
	d *mockDriver
}

func (s* mockStmt) Close() error {
	return nil
}

func (s* mockStmt) NumInput() int {
	return 0
}

func (s* mockStmt) Exec(args []driver.Value) (driver.Result, error) {
	return nil, nil
}

func (s* mockStmt) Query(args []driver.Value) (driver.Rows, error) {

	drv = s.d
	mr := newMockRows(drv.colNames, drv.rowData)

	drv.consumed()

	return mr, nil
}


func newMockRows(c []string, data [][]driver.Value) *mockRows {

	mr := new(mockRows)
	mr.d = data
	mr.c = c

	return mr
}

type mockRows struct {

	served int
	c []string
	d [][]driver.Value
}

func (r *mockRows) Columns() []string {
	return r.c
}

func (r *mockRows) Close() error {
	return nil
}

func (r *mockRows) Next(dest []driver.Value) error {

	if r.served >= len(r.d) {
		return io.EOF
	}

	for i, v := range r.d[r.served] {
		dest[i] = v
	}

	r.served += 1

	return nil
}

type mockTx struct {

}

func (t *mockTx) Commit() error {
	return nil
}

func (t *mockTx) Rollback() error {
	return nil
}