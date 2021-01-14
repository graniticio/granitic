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

func TestLocalPathDetection(t *testing.T) {

	local := []string{"./test", "../test", "/test", "\\\\test", "C:\\test"}
	notLocal := []string{"github.com/some/mod/v2"}

	for _, l := range local {

		if !localPath(l) {
			t.Errorf("Expected %s to be flagged as a local path", l)
		}

	}

	for _, l := range notLocal {

		if localPath(l) {
			t.Errorf("Expected %s to be flagged as a remote path", l)
		}

	}

}

func TestPathVersionReplacement(t *testing.T) {

	reqA := requirement{
		Path:    "a",
		Version: "1",
	}

	repA1 := replacement{
		Old: modPath{
			Path:    "a",
			Version: "1",
		},
		New: modPath{
			Path:    "b",
			Version: "2",
		},
	}

	repA2 := replacement{
		Old: modPath{
			Path:    "a",
			Version: "1",
		},
		New: modPath{
			Path: "b",
		},
	}

	repA3 := replacement{
		Old: modPath{
			Path:    "a",
			Version: "1",
		},
		New: modPath{
			Path: "/b",
		},
	}

	l := replacePath(&reqA, []replacement{repA1})

	if l {
		t.Errorf("Expected detect as remote path")

	}

	if reqA.Path != "b" {
		t.Errorf("Expected path to be replaced with b")
	}

	if reqA.Version != "2" {
		t.Errorf("Expected version to be replaced with 2")
	}

	reqA.Path = "a"
	reqA.Version = "1"

	l = replacePath(&reqA, []replacement{repA2})

	if l {
		t.Errorf("Expected detect as remote path")

	}

	if reqA.Path != "b" {
		t.Errorf("Expected path to be replaced with b")
	}

	if reqA.Version != "1" {
		t.Errorf("Expected version to remain as 1")
	}

	reqA.Path = "a"
	reqA.Version = "1"

	l = replacePath(&reqA, []replacement{repA3})

	if !l {
		t.Errorf("Expected detect as local path")

	}

	if reqA.Path != "/b" {
		t.Errorf("Expected path to be replaced with /b")
	}

	if reqA.Version != "1" {
		t.Errorf("Expected version to remain as 1")
	}

}
