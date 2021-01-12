package binder

import (
	"github.com/graniticio/granitic/v2/logging"
	"github.com/graniticio/granitic/v2/test"
	"os"
	"path/filepath"
	"testing"
)

func TestValidModParsed(t *testing.T) {

	modPath := filepath.Join("validmod")
	modPath = test.FilePath(modPath)
	l := new(logging.NullLogger)

	mf, err := ParseModFile(modPath, l)

	if err != nil {
		t.Fatalf(err.Error())
	}

	if mf == nil {
		t.Fatalf("Unexpected nil parsed object")
	}

}

func TestValidModFileExistsChecking(t *testing.T) {

	modPath := filepath.Join("validmod")
	modPath = test.FilePath(modPath)

	if !CheckModFileExists(modPath) {
		t.Errorf("Expected mod file to be found at: %s", modPath)
	}

	emptyPath := filepath.Join("..", "testdata")

	if CheckModFileExists(emptyPath) {
		t.Errorf("Did not expect a mod file at: %s", modPath)
	}

}

func TestPathFinding(t *testing.T) {

	const GP = "GOPATH"
	const GMC = "GOMODCACHE"

	gpi := os.Getenv(GP)
	gmci := os.Getenv(GMC)

	l := new(logging.NullLogger)

	os.Setenv(GP, "goPathRoot")

	cachePath(l)

	os.Setenv(GP, gpi)
	os.Setenv(GMC, gmci)

}
