package logging

import (
	"sort"
	"testing"
)

func TestInitialLogLevels(t *testing.T) {

	i := make(map[string]interface{})

	i["X"] = TraceLabel

	clm := CreateComponentLoggerManager(Error, i, []LogWriter{}, nil, false)

	def := clm.CreateLogger("A")

	if def.IsLevelEnabled(Warn) {
		t.Fail()
	}

	if !def.IsLevelEnabled(Error) {
		t.Fail()
	}

	cust := clm.CreateLoggerAtLevel("A", Warn)

	if !cust.IsLevelEnabled(Warn) {
		t.Fail()
	}

	pre := clm.CreateLogger("X")

	if !pre.IsLevelEnabled(Trace) {
		t.Fail()
	}

}

func TestInitialLogLevelsSetAfterCreation(t *testing.T) {

	i := make(map[string]interface{})

	i["X"] = TraceLabel
	i["A"] = WarnLabel

	clm := CreateComponentLoggerManager(Error, make(map[string]interface{}), []LogWriter{}, nil, false)

	def := clm.CreateLogger("A")

	if def.IsLevelEnabled(Warn) {
		t.Fail()
	}

	if !def.IsLevelEnabled(Error) {
		t.Fail()
	}

	cust := clm.CreateLoggerAtLevel("A", Warn)

	if !cust.IsLevelEnabled(Warn) {
		t.Fail()
	}

	clm.SetInitialLogLevels(i)

	pre := clm.CreateLogger("X")

	if !pre.IsLevelEnabled(Trace) {
		t.Fail()
	}

}

func TestListCurrentLevels(t *testing.T) {

	i := make(map[string]interface{})

	i["X"] = TraceLabel
	i["A"] = WarnLabel

	clm := CreateComponentLoggerManager(Error, i, []LogWriter{}, nil, false)

	clm.CreateLogger("A")
	clm.CreateLogger("A")
	clm.CreateLogger("X")

	cl := clm.CurrentLevels()

	if len(cl) != 2 {
		t.Fail()
	}

	sort.Sort(ByName{ComponentLevels: cl})

	l := cl[0]

	if l.Name == "A" && l.Level != Warn {
		t.Fail()
	}

	l = cl[1]

	if l.Name == "X" && l.Level != Trace {
		t.Fail()
	}

}

func TestRecoverExistingLogger(t *testing.T) {

	i := make(map[string]interface{})

	clm := CreateComponentLoggerManager(Error, i, []LogWriter{}, nil, false)

	def := clm.CreateLogger("A")

	rep := clm.LoggerByName("A")

	if rep == nil || rep != def {
		t.Fail()
	}

}

func TestChangeThreshold(t *testing.T) {
	i := make(map[string]interface{})

	clm := CreateComponentLoggerManager(Error, i, []LogWriter{}, nil, false)

	def := clm.CreateLogger("A")

	if def.IsLevelEnabled(Warn) {
		t.Fail()
	}

	clm.SetGlobalThreshold(Warn)

	if !def.IsLevelEnabled(Warn) {
		t.Fail()
	}
}

func TestDisabling(t *testing.T) {

	i := make(map[string]interface{})

	clm := CreateComponentLoggerManager(Error, i, []LogWriter{}, nil, false)

	clm.CreateLogger("A")

	clm.Disable()

	clm.CreateLogger("A")
	clm.CreateLoggerAtLevel("B", Error)

	if !clm.IsDisabled() {
		t.Fail()
	}

}

func TestLifecycle(t *testing.T) {

	w := new(dummyWriter)
	w.b = true

	i := make(map[string]interface{})

	clm := CreateComponentLoggerManager(Error, i, []LogWriter{w}, nil, false)

	if b, err := clm.ReadyToStop(); b || err != nil {
		t.Errorf("Expected no error and not ready got %v %v", err, b)
	}

	w.b = false

	if b, err := clm.ReadyToStop(); !b || err != nil {
		t.Errorf("Expected no error and  ready got %v %v", err, b)
	}

	clm.PrepareToStop()

	if err := clm.Stop(); w.c == false || err != nil {
		t.Errorf("Expected no error and writer ready to be closed %v %v", err, w.c)
	}

}

type dummyWriter struct {
	b bool
	c bool
}

// WriteMessage request that the supplied message be written. Depending on implementation, may be asynchronous.
func (dw *dummyWriter) WriteMessage(string) {}

// Close any resources (file handles etc) that this LogWriter might be holding.
func (dw *dummyWriter) Close() {
	dw.c = true
}

// Returns true if the LogWriter is currently in the process of writing a log line.
func (dw *dummyWriter) Busy() bool {

	return dw.b
}
