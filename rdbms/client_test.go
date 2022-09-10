package rdbms

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"github.com/graniticio/granitic/v3/logging"
	"github.com/graniticio/granitic/v3/reflecttools"
	"github.com/graniticio/granitic/v3/test"
	"github.com/graniticio/granitic/v3/types"
	"io"
	"os"
	"testing"
	"time"
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

func TestImplementsExecutor(t *testing.T) {
	var ex Client

	ex = new(ManagedClient)

	_ = ex

}

func TestPassthroughs(t *testing.T) {

	c := newRdbmsClient(db, qm, DefaultInsertWithReturnedID, logging.CreateAnonymousLogger("testLog", logging.Fatal))

	passthroughChecks(t, c)

	c = newRdbmsClient(db, qm, DefaultInsertWithReturnedID, logging.CreateAnonymousLogger("testLog", logging.Fatal))
	c.StartTransaction()

	passthroughChecks(t, c)

	c.CommitTransaction()

	c = newRdbmsClient(db, qm, DefaultInsertWithReturnedID, logging.CreateAnonymousLogger("testLog", logging.Fatal))
	c.ctx = context.Background()

	passthroughChecks(t, c)

	c.StartTransaction()

	passthroughChecks(t, c)

	c.Rollback()

}

func TestTempQueries(t *testing.T) {

	c := newRdbmsClient(db, qm, DefaultInsertWithReturnedID, logging.CreateAnonymousLogger("testLog", logging.Fatal))
	c.RegisterTempQuery("QT", "123")

	qm.reset()
	_, err := c.DeleteQIDParams("QT")

	test.ExpectNil(t, err)

	test.ExpectString(t, qm.lastQueryReturned, "")

	q, err := c.BuildQueryFromQIDParams("NQ")
	test.ExpectNil(t, err)

	test.ExpectString(t, q, "NQ")

}

func TestFindOrCreate(t *testing.T) {

	c := newRdbmsClient(db, qm, DefaultInsertWithReturnedID, logging.CreateAnonymousLogger("testLog", logging.Fatal))

	p1, p2 := testStandardParams()

	var id int64

	err := c.ExistingIDOrInsertParams("CQ", "IQ", &id, p1, p2)

	if !paramMergedCorrectly(qm.lastParams) {
		t.FailNow()
	}

	test.ExpectNil(t, err)

	test.ExpectInt(t, int(id), 1)

	drv.colNames = []string{"Int64Result"}
	drv.rowData = [][]driver.Value{{int64(8)}}

	err = c.ExistingIDOrInsertParams("CQ", "IQ", &id, p1, p2)

	test.ExpectNil(t, err)

	test.ExpectInt(t, int(id), 8)
}

func passthroughChecks(t *testing.T, c *ManagedClient) {
	drv.consumed()
	r, err := c.Query("TEST")

	test.ExpectNil(t, err)
	test.ExpectNotNil(t, r)

	test.ExpectBool(t, r.Next(), false)

	drv.colNames = []string{"Float64Result"}
	drv.rowData = [][]driver.Value{{float64(123.1)}}

	w := c.QueryRow("TEST", "A")
	test.ExpectNotNil(t, w)

	res, err := c.Exec("TEST")
	test.ExpectNil(t, err)
	test.ExpectNotNil(t, res)
}

func TestNonTxSelectMethodsWithCtx(t *testing.T) {

	c := newRdbmsClient(db, qm, DefaultInsertWithReturnedID, logging.CreateAnonymousLogger("testLog", logging.Fatal))
	c.ctx = context.Background()

	testSelectMethods(t, c)

}

func TestNonTxSelectMethods(t *testing.T) {

	c := newRdbmsClient(db, qm, DefaultInsertWithReturnedID, logging.CreateAnonymousLogger("testLog", logging.Fatal))

	testSelectMethods(t, c)

}

func TestTxSelectMethods(t *testing.T) {

	c := newRdbmsClient(db, qm, DefaultInsertWithReturnedID, logging.CreateAnonymousLogger("testLog", logging.Fatal))
	c.StartTransaction()
	testSelectMethods(t, c)
	c.CommitTransaction()

}

func TestTxSelectMethodsWithCtx(t *testing.T) {

	c := newRdbmsClient(db, qm, DefaultInsertWithReturnedID, logging.CreateAnonymousLogger("testLog", logging.Fatal))
	c.ctx = context.Background()
	c.StartTransaction()
	testSelectMethods(t, c)
	c.Rollback()

}

func TestNonTxDeleteMethods(t *testing.T) {

	c := newRdbmsClient(db, qm, DefaultInsertWithReturnedID, logging.CreateAnonymousLogger("testLog", logging.Fatal))

	testDeleteMethods(t, c)

}

func TestNonTxDeleteMethodsWithCtx(t *testing.T) {

	c := newRdbmsClient(db, qm, DefaultInsertWithReturnedID, logging.CreateAnonymousLogger("testLog", logging.Fatal))
	c.ctx = context.Background()

	testDeleteMethods(t, c)

}

func TestTxDeleteMethods(t *testing.T) {

	c := newRdbmsClient(db, qm, DefaultInsertWithReturnedID, logging.CreateAnonymousLogger("testLog", logging.Fatal))
	c.StartTransaction()
	testDeleteMethods(t, c)
	c.CommitTransaction()

}

func TestNonTxUpdateMethods(t *testing.T) {

	c := newRdbmsClient(db, qm, DefaultInsertWithReturnedID, logging.CreateAnonymousLogger("testLog", logging.Fatal))

	testUpdateMethods(t, c)

}

func TestNonTxUpdateMethodsWithCtx(t *testing.T) {

	c := newRdbmsClient(db, qm, DefaultInsertWithReturnedID, logging.CreateAnonymousLogger("testLog", logging.Fatal))
	c.ctx = context.Background()

	testUpdateMethods(t, c)

}

func TestTxUpdateMethods(t *testing.T) {

	c := newRdbmsClient(db, qm, DefaultInsertWithReturnedID, logging.CreateAnonymousLogger("testLog", logging.Fatal))
	c.StartTransaction()
	testUpdateMethods(t, c)
	c.CommitTransaction()

}

func TestTxUpdateMethodsWithCtx(t *testing.T) {

	c := newRdbmsClient(db, qm, DefaultInsertWithReturnedID, logging.CreateAnonymousLogger("testLog", logging.Fatal))
	c.ctx = context.Background()
	c.StartTransaction()
	testUpdateMethods(t, c)
	c.Rollback()

}

func TestTxInsertMethodsWithCtx(t *testing.T) {

	c := newRdbmsClient(db, qm, DefaultInsertWithReturnedID, logging.CreateAnonymousLogger("testLog", logging.Fatal))
	c.ctx = context.Background()

	c.StartTransaction()
	testInsertMethods(t, c)
	c.Rollback()

}

func TestNonTxInsertMethods(t *testing.T) {

	c := newRdbmsClient(db, qm, DefaultInsertWithReturnedID, logging.CreateAnonymousLogger("testLog", logging.Fatal))

	testInsertMethods(t, c)

}

func TestNonTxInsertMethodsWithCtx(t *testing.T) {

	c := newRdbmsClient(db, qm, DefaultInsertWithReturnedID, logging.CreateAnonymousLogger("testLog", logging.Fatal))
	c.ctx = context.Background()

	testInsertMethods(t, c)

}

func TestTxInsertMethods(t *testing.T) {

	c := newRdbmsClient(db, qm, DefaultInsertWithReturnedID, logging.CreateAnonymousLogger("testLog", logging.Fatal))
	c.StartTransaction()
	testInsertMethods(t, c)
	c.CommitTransaction()

}

func TestTxDeleteMethodsWithCtx(t *testing.T) {

	c := newRdbmsClient(db, qm, DefaultInsertWithReturnedID, logging.CreateAnonymousLogger("testLog", logging.Fatal))
	c.ctx = context.Background()

	c.StartTransaction()
	testDeleteMethods(t, c)
	c.Rollback()

}

func TestTransactionBehaviour(t *testing.T) {

	c := newRdbmsClient(db, qm, DefaultInsertWithReturnedID, logging.CreateAnonymousLogger("testLog", logging.Fatal))

	err := c.CommitTransaction()

	test.ExpectNotNil(t, err)

	err = c.StartTransaction()
	test.ExpectNil(t, err)

	err = c.CommitTransaction()
	test.ExpectNil(t, err)

	err = c.StartTransaction()
	test.ExpectNil(t, err)

	err = c.StartTransaction()
	test.ExpectNotNil(t, err)

	c = newRdbmsClient(db, qm, DefaultInsertWithReturnedID, logging.CreateAnonymousLogger("testLog", logging.Fatal))

	err = c.StartTransactionWithOptions(new(sql.TxOptions))
	test.ExpectNil(t, err)

	err = c.CommitTransaction()
	test.ExpectNil(t, err)

	err = c.StartTransactionWithOptions(new(sql.TxOptions))
	test.ExpectNil(t, err)

	err = c.StartTransaction()
	test.ExpectNotNil(t, err)

	err = c.StartTransactionWithOptions(new(sql.TxOptions))
	test.ExpectNotNil(t, err)

	c = newRdbmsClient(db, qm, DefaultInsertWithReturnedID, logging.CreateAnonymousLogger("testLog", logging.Fatal))
	c.ctx = context.Background()

	err = c.CommitTransaction()

	test.ExpectNotNil(t, err)

	err = c.StartTransaction()
	test.ExpectNil(t, err)

	err = c.CommitTransaction()
	test.ExpectNil(t, err)

	err = c.StartTransactionWithOptions(new(sql.TxOptions))
	test.ExpectNil(t, err)

	err = c.StartTransaction()
	test.ExpectNotNil(t, err)

}

func TestFragmentFinding(t *testing.T) {

	c := newRdbmsClient(db, qm, DefaultInsertWithReturnedID, logging.CreateAnonymousLogger("testLog", logging.Fatal))

	v, err := c.FindFragment("Identity")

	test.ExpectNil(t, err)
	test.ExpectString(t, v, "Identity")

}

func TestBuildQuery(t *testing.T) {

	c := newRdbmsClient(db, qm, DefaultInsertWithReturnedID, logging.CreateAnonymousLogger("testLog", logging.Fatal))

	p1, p2 := testStandardParams()

	q, err := c.BuildQueryFromQIDParams("OK", p1, p2)

	if !paramMergedCorrectly(qm.lastParams) {
		t.FailNow()
	}

	test.ExpectNil(t, err)
	test.ExpectString(t, q, "OK")

	q, err = c.BuildQueryFromQIDParams("ERROR", p1, p2)

	test.ExpectNotNil(t, err)
	test.ExpectString(t, q, "")
}

func TestIllegalResultContents(t *testing.T) {

	c := newRdbmsClient(db, qm, DefaultInsertWithReturnedID, logging.CreateAnonymousLogger("testLog", logging.Fatal))

	//SelectBindSingleQID
	drv.colNames = []string{"TimeResult"}

	drv.rowData = [][]driver.Value{{types.NewNilableString("AA")}}

	bt := new(testTarget)

	found, err := c.SelectBindSingleQID("SBSQ", bt)

	test.ExpectBool(t, found, false)
	test.ExpectNotNil(t, err)

}

func testInsertMethods(t *testing.T, c *ManagedClient) {

	p1, p2 := testStandardParams()

	var newID int64

	err := c.InsertCaptureQIDParams("ICQPs", &newID, p1, p2)

	test.ExpectNil(t, err)
	test.ExpectInt(t, int(newID), 1)

	if !paramMergedCorrectly(qm.lastParams) {
		t.FailNow()
	}

	r, err := c.InsertQIDParams("IQPs", p1, p2)

	ra, _ := r.RowsAffected()

	test.ExpectNil(t, err)
	test.ExpectInt(t, int(ra), 1)

	r, err = c.InsertQIDParams("IQPs")

	test.ExpectNil(t, err)
}

func testDeleteMethods(t *testing.T, c *ManagedClient) {

	p1, p2 := testStandardParams()

	r, err := c.DeleteQIDParam("DQP", "p1", "v1")
	test.ExpectNil(t, err)

	ra, _ := r.RowsAffected()

	test.ExpectInt(t, int(ra), 1)

	drv.forceError = true
	r, err = c.DeleteQIDParam("DQP", "p1", "v1")
	test.ExpectNotNil(t, err)
	test.ExpectNil(t, r)

	r, err = c.DeleteQIDParams("DQPs", p1, p2)
	test.ExpectNil(t, err)
	ra, _ = r.RowsAffected()

	test.ExpectInt(t, int(ra), 1)

	if !paramMergedCorrectly(qm.lastParams) {
		t.FailNow()
	}

	drv.forceError = true
	r, err = c.DeleteQIDParams("DQPs", p1, p2)
	test.ExpectNotNil(t, err)
	test.ExpectNil(t, r)

}

func testUpdateMethods(t *testing.T, c *ManagedClient) {

	p1, p2 := testStandardParams()

	r, err := c.UpdateQIDParam("DQP", "p1", "v1")
	test.ExpectNil(t, err)

	ra, _ := r.RowsAffected()

	test.ExpectInt(t, int(ra), 1)

	drv.forceError = true
	r, err = c.UpdateQIDParam("DQP", "p1", "v1")
	test.ExpectNotNil(t, err)
	test.ExpectNil(t, r)

	r, err = c.UpdateQIDParams("DQPs", p1, p2)
	test.ExpectNil(t, err)
	ra, _ = r.RowsAffected()

	test.ExpectInt(t, int(ra), 1)

	if !paramMergedCorrectly(qm.lastParams) {
		t.FailNow()
	}

	drv.forceError = true
	r, err = c.UpdateQIDParams("DQPs", p1, p2)
	test.ExpectNotNil(t, err)
	test.ExpectNil(t, r)

}

func testSelectMethods(t *testing.T, c *ManagedClient) {

	p1, p2 := testStandardParams()

	//SelectBindQID
	drv.colNames = []string{"Int64Result"}
	drv.rowData = [][]driver.Value{{int64(45)}, {int64(32)}}

	bt := new(testTarget)
	results, err := c.SelectBindQID("SBQ", bt)

	test.ExpectNil(t, err)
	test.ExpectInt(t, len(results), 2)

	test.ExpectInt(t, int(results[0].(*testTarget).Int64Result), 45)
	test.ExpectInt(t, int(results[1].(*testTarget).Int64Result), 32)

	results, err = c.SelectBindQID("SBQ", bt)

	test.ExpectNil(t, err)
	test.ExpectInt(t, len(results), 0)

	drv.forceError = true
	results, err = c.SelectBindQID("SBQ", bt)
	test.ExpectNotNil(t, err)

	//SelectBindQIDParam
	drv.colNames = []string{"Float64Result"}
	drv.rowData = [][]driver.Value{{float64(123.1)}}

	results, err = c.SelectBindQIDParam("SBQP", "p1", "v1", bt)

	test.ExpectNil(t, err)
	test.ExpectInt(t, len(results), 1)

	test.ExpectFloat(t, results[0].(*testTarget).Float64Result, float64(123.1))

	test.ExpectInt(t, len(qm.lastParams), 1)
	test.ExpectString(t, qm.lastParams["p1"].(string), "v1")

	results, err = c.SelectBindQIDParam("SBQP", "p1", "v1", bt)
	test.ExpectNil(t, err)
	test.ExpectInt(t, len(results), 0)

	drv.forceError = true
	results, err = c.SelectBindQIDParam("SBQP", "p1", "v1", bt)
	test.ExpectNotNil(t, err)

	//SelectBindQIDParams
	drv.colNames = []string{"BoolResult"}
	drv.rowData = [][]driver.Value{{true}, {false}, {true}}

	results, err = c.SelectBindQIDParams("SBQPs", bt, p1, p2)

	test.ExpectNil(t, err)
	test.ExpectInt(t, len(results), 3)

	test.ExpectBool(t, results[2].(*testTarget).BoolResult, true)

	if !paramMergedCorrectly(qm.lastParams) {
		t.FailNow()
	}

	drv.forceError = true
	results, err = c.SelectBindQIDParams("SBQPs", bt, p1, p2)
	test.ExpectNotNil(t, err)

	//SelectBindSingleQID
	drv.colNames = []string{"TimeResult"}

	drv.rowData = [][]driver.Value{{time.Now()}}

	bt = new(testTarget)

	found, err := c.SelectBindSingleQID("SBSQ", bt)

	test.ExpectNil(t, err)
	test.ExpectBool(t, found, true)
	test.ExpectBool(t, reflecttools.IsZero(bt.TimeResult), false)

	bt = new(testTarget)

	found, err = c.SelectBindSingleQID("SBSQ", bt)
	test.ExpectNil(t, err)
	test.ExpectBool(t, found, false)
	test.ExpectBool(t, reflecttools.IsZero(bt.TimeResult), true)

	drv.forceError = true
	found, err = c.SelectBindSingleQID("SBSQ", bt)
	test.ExpectNotNil(t, err)

	//SelectBindSingleQIDParam
	drv.colNames = []string{"StrResult", "Int64Result"}
	drv.rowData = [][]driver.Value{{"okay", int64(1)}, {"not", int64(2)}}

	bt = new(testTarget)

	found, err = c.SelectBindSingleQIDParam("SBSQ", "p1", "v1", bt)
	test.ExpectNotNil(t, err)
	test.ExpectInt(t, len(qm.lastParams), 1)
	test.ExpectString(t, qm.lastParams["p1"].(string), "v1")

	drv.colNames = []string{"StrResult", "Int64Result"}
	drv.rowData = [][]driver.Value{{"okay", int64(1)}}

	found, err = c.SelectBindSingleQIDParam("SBSQ", "p1", "v1", bt)

	test.ExpectNil(t, err)
	test.ExpectBool(t, found, true)

	test.ExpectInt(t, int(bt.Int64Result), 1)
	test.ExpectString(t, bt.StrResult, "okay")

	drv.forceError = true
	found, err = c.SelectBindSingleQIDParam("SBSQ", "p1", "v1", bt)
	test.ExpectNotNil(t, err)

	//SelectBindSingleQIDParams
	drv.colNames = []string{"StrResult"}
	drv.rowData = [][]driver.Value{{"okay"}}

	bt = new(testTarget)
	found, err = c.SelectBindSingleQIDParams("SBSQP", bt, p1, p2)

	test.ExpectNil(t, err)
	test.ExpectBool(t, found, true)
	test.ExpectString(t, bt.StrResult, "okay")

	if !paramMergedCorrectly(qm.lastParams) {
		t.FailNow()
	}

	found, err = c.SelectBindSingleQIDParams("SBSQP", bt, p1, p2)

	test.ExpectNil(t, err)
	test.ExpectBool(t, found, false)

	drv.forceError = true
	found, err = c.SelectBindSingleQIDParams("SBSQP", bt, p1, p2)
	test.ExpectNotNil(t, err)

	//SelectQID
	drv.colNames = []string{"StrResult"}
	drv.rowData = [][]driver.Value{{"okay"}}

	r, err := c.SelectQID("SQ")

	test.ExpectNil(t, err)
	test.ExpectBool(t, r.Next(), true)

	drv.forceError = true
	r, err = c.SelectQID("SQ")
	test.ExpectNotNil(t, err)

	//SelectQIDParam
	drv.colNames = []string{"StrResult"}
	drv.rowData = [][]driver.Value{{"okay"}, {"not"}}
	r, err = c.SelectQIDParam("SQP", "p1", "v1")
	test.ExpectNil(t, err)
	test.ExpectBool(t, r.Next(), true)
	test.ExpectBool(t, r.Next(), true)

	drv.forceError = true
	r, err = c.SelectQIDParam("SQP", "p1", "v1")
	test.ExpectNotNil(t, err)

	//SelectQIDParams
	drv.colNames = []string{"StrResult"}
	drv.rowData = [][]driver.Value{{"okay"}, {"not"}}

	r, err = c.SelectQIDParams("SQPs", p1, p2)
	test.ExpectNil(t, err)
	test.ExpectBool(t, r.Next(), true)
	test.ExpectBool(t, r.Next(), true)

	drv.forceError = true
	r, err = c.SelectQIDParams("SQPs", p1, p2)
	test.ExpectNotNil(t, err)
}

func testStandardParams() (interface{}, interface{}) {

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
	StrResult       string
	Int64Result     int64
	Float64Result   float64
	BoolResult      bool
	ByteArrayResult []byte
	TimeResult      time.Time
	Aliased         string `column:"ColumnAlias"`
	StructResult    testParam
}

type testParam struct {
	SParam  string
	NSParam *types.NilableString
	IParam  int `dbparam:"IOV"`
}

type testQueryManagerProxy struct {
	lastParams        map[string]interface{}
	lastQueryReturned string
}

func (tqm *testQueryManagerProxy) reset() {
	tqm.lastParams = nil
	tqm.lastQueryReturned = ""
}

func (tqm *testQueryManagerProxy) BuildQueryFromID(qid string, params map[string]interface{}) (string, error) {
	tqm.lastParams = params

	if qid == "ERROR" {
		tqm.lastQueryReturned = ""
		return "", errors.New("Forced error")
	}

	tqm.lastQueryReturned = qid

	return qid, nil

}

func (tqm *testQueryManagerProxy) FragmentFromID(qid string) (string, error) {

	tqm.lastQueryReturned = qid

	return qid, nil
}

type mockResult struct {
	lid int64
	ra  int64
}

func (mr mockResult) LastInsertId() (int64, error) {
	return mr.lid, nil
}

func (mr mockResult) RowsAffected() (int64, error) {
	return mr.ra, nil
}

type mockDriver struct {
	colNames   []string
	rowData    [][]driver.Value
	forceError bool
}

func (d *mockDriver) consumed() {
	d.colNames = nil
	d.rowData = nil
	d.forceError = false
}

func (d *mockDriver) Open(name string) (driver.Conn, error) {
	return newMockConn(d), nil
}

func newMockConn(d *mockDriver) *mockConn {
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

func newMockStmt(d *mockDriver) *mockStmt {
	s := new(mockStmt)
	s.d = d

	return s
}

type mockStmt struct {
	d *mockDriver
}

func (s *mockStmt) Close() error {
	return nil
}

func (s *mockStmt) NumInput() int {
	return 0
}

func (s *mockStmt) Exec(args []driver.Value) (driver.Result, error) {

	if s.d.forceError {
		drv.consumed()
		return nil, errors.New("Forced error")
	}

	mr := new(mockResult)
	mr.ra = 1
	mr.lid = 1

	drv.consumed()

	return mr, nil
}

func (s *mockStmt) Query(args []driver.Value) (driver.Rows, error) {

	if s.d.forceError {
		drv.consumed()
		return nil, errors.New("Forced error")
	}

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
	c      []string
	d      [][]driver.Value
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

	r.served++

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
