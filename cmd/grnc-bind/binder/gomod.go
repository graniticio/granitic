package binder

import (
	"encoding/json"
	"fmt"
	"github.com/graniticio/granitic/v2/logging"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

// FindExternalFacilities parses the first level of modules imported by this application's go.mod file and
// tries to find properly defined Granitic external facilities
func FindExternalFacilities(l logging.Logger) (*ExternalFacilities, error) {

	cwd, _ := os.Getwd()

	if m, err := ParseModFile(cwd, l); err != nil {
		return nil, err
	} else {

		return modulesToFacilities(m, l)
	}
}

// ExternalFacilities holds information about the code and config defined in Go module
// dependencies that should be compiled into this application.
type ExternalFacilities struct {
}

func modulesToFacilities(mf *modFile, l logging.Logger) (*ExternalFacilities, error) {

	cp, err := cachePath(l)

	if err != nil {
		return nil, err
	}

	l.LogDebugf("Expecting downloaded packages to be in %s\n", cp)

	return nil, nil
}

func cachePath(l logging.Logger) (string, error) {
	gmc := os.Getenv("GOMODCACHE")

	if gmc != "" {
		l.LogDebugf("GOMODCACHE environment variable set. Using as location for downloaded modules")
		return filepath.Join(gmc, "cache", "download"), nil
	}

	gmc = os.Getenv("GOPATH")

	if gmc != "" {
		l.LogDebugf("GOPATH environment variable set. Using $GOPATH/pkg/mod for downloaded modules")

		return filepath.Join(gmc, "pkg", "mod", "cache", "download"), nil
	}

	l.LogWarnf("Neither GOPATH nor GOMODCACHE environment variable set. Assuming user home directory contains go artifacts")

	hd, err := os.UserHomeDir()

	if err != nil {
		return "", err
	}

	return filepath.Join(hd, "go", "pkg", "mod", "cache", "download"), nil
}

//ParseModFile tries to parse the mod file in the supplied directory and returns an error if parsing failed
func ParseModFile(d string, l logging.Logger) (*modFile, error) {

	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)

	if err := os.Chdir(d); err != nil {
		return nil, fmt.Errorf("unable to change directory to %s: %s", d, err.Error())
	}

	goExec, err := exec.LookPath("go")

	if err != nil {
		return nil, fmt.Errorf("could not find the 'go' executable on your path. Make sure it is available in your OS PATH environment variable")
	}

	cmd := exec.Command(goExec, "mod", "edit", "--json")

	stdout, err := cmd.StdoutPipe()

	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	m := modFile{}

	if err := json.NewDecoder(stdout).Decode(&m); err != nil {
		return nil, err
	}

	return &m, nil

}

// Module represents a mod file on disk
type Module struct {
	Name         string
	Version      string
	ExpectedPath string
}

// CheckModFileExists makes sure that a go.mod file exists in the supplied directory
func CheckModFileExists(d string) bool {

	f := filepath.Join(d, "go.mod")

	info, err := os.Stat(f)

	if os.IsNotExist(err) || info.IsDir() {
		return false
	}

	return true
}

type modFile struct {
	Require []requirement
}

type requirement struct {
	Path    string
	Version string
}
