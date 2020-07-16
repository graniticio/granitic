package binder

import (
	"github.com/graniticio/granitic/v2/logging"
	"github.com/graniticio/granitic/v2/test"
	"os"
	"testing"
)

func TestInvalidDirectoryAndWorkingDir(t *testing.T) {

	wd, _ := os.Getwd()

	illegalDir := "A!@$55l;''"

	m, err := ParseModFile(illegalDir, logging.NewStdoutLogger(logging.Fatal, ""))

	if m != nil {
		t.FailNow()
	}

	if err == nil {
		t.Errorf("Expected an error on unparseable directory")
	}

	if nwd, _ := os.Getwd(); nwd != wd {
		t.Errorf("Expected working dir to be %s but it is %s", wd, nwd)
	}

}

func TestNoModFound(t *testing.T) {

	wd, _ := os.Getwd()

	if CheckModFileExists(wd) {
		t.FailNow()
	}

	d := test.TestDataJoin("validmod")

	if !CheckModFileExists(d) {
		t.FailNow()
	}

}

func TestValidModOkay(t *testing.T) {

	d := test.TestDataJoin("validmod")

	m, err := ParseModFile(d, logging.NewStdoutLogger(logging.Fatal, ""))

	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}

	if m == nil {
		t.Errorf("Expected valid ModFile, got nil")
	}

}
