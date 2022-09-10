package grncerror

import (
	"github.com/graniticio/granitic/v3/logging"
	"github.com/graniticio/granitic/v3/types"
	"testing"
)

func TestIllegalMessageFormat(t *testing.T) {
	sem := new(ServiceErrorManager)

	sem.FrameworkLogger = new(logging.ConsoleErrorLogger)

	badCategory := []interface{}{"X", "INVALID_ARTIST", "Cannot create an artist with the information provided."}
	missingCode := []interface{}{"C", "", "Cannot create an artist with the information provided."}
	missingMessage := []interface{}{"C", "INVALID_ARTIST", ""}
	codedError := []interface{}{"C", "INVALID_ARTIST", "Cannot create an artist with the information provided."}
	dupeCode := []interface{}{"C", "INVALID_ARTIST", "Cannot create an artist with the information provided."}
	basicCodes := []interface{}{badCategory, missingCode, missingMessage, codedError, dupeCode}

	sem.LoadErrors(basicCodes)
}

func TestCheckNameSet(t *testing.T) {
	sem := createManager()

	if sem.ComponentName() == "" {
		t.FailNow()
	}
}

func TestErrorLoadingAndLookup(t *testing.T) {

	sem := createManager()

	ce := sem.Find("INVALID_ARTIST")

	if ce == nil {
		t.FailNow()
	}

	ce = sem.Find("MISSING")

	if ce != nil {
		t.FailNow()
	}

	sem.PanicOnMissing = true

	defer func() {
		if r := recover(); r != nil {
			//Expected
		} else {
			t.FailNow()
		}
	}()

	ce = sem.Find("EXPECT_PANIC")
}

func TestCodeSources(t *testing.T) {

	sem := createManager()

	es1 := ErrorSource{
		V:     true,
		CN:    "es1",
		Codes: types.NewUnorderedStringSet([]string{"INVALID_ARTIST"}),
	}

	sem.RegisterCodeUser(&es1)

	if err := sem.AllowAccess(); err != nil {
		t.FailNow()
	}

	es2 := ErrorSource{
		V:     false,
		CN:    "es2",
		Codes: types.NewUnorderedStringSet([]string{"MISSING"}),
	}

	sem = createManager()
	sem.RegisterCodeUser(&es2)

	if err := sem.AllowAccess(); err != nil {
		t.Errorf("Unexpected error %s\n", err.Error())
		t.FailNow()
	}

	es3 := ErrorSource{
		V:     true,
		CN:    "es3",
		Codes: types.NewUnorderedStringSet([]string{"MISSING"}),
	}

	sem = createManager()
	sem.RegisterCodeUser(&es3)

	if err := sem.AllowAccess(); err == nil {
		t.FailNow()
	}
}

func createManager() *ServiceErrorManager {
	sem := new(ServiceErrorManager)

	sem.FrameworkLogger = new(logging.ConsoleErrorLogger)

	codedError := []interface{}{"C", "INVALID_ARTIST", "Cannot create an artist with the information provided."}
	basicCodes := []interface{}{codedError}

	sem.LoadErrors(basicCodes)

	sem.SetComponentName("SEM")

	return sem
}

type ErrorSource struct {
	V     bool
	CN    string
	Codes types.StringSet
}

func (es *ErrorSource) ValidateMissing() bool {
	return es.V
}

func (es *ErrorSource) ErrorCodesInUse() (types.StringSet, string) {
	return es.Codes, es.CN
}
