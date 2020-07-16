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

//ParseModFile tries to parse the mod file in the supplied directory and returns an error if parsing failed
func ParseModFile(d string, l logging.Logger) (*ModFile, error) {

	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)

	if err := os.Chdir(d); err != nil {
		return nil, fmt.Errorf("unable to change directory to %s: %s", d, err.Error())
	}

	goExec, err := exec.LookPath("go")

	if err != nil {
		return nil, fmt.Errorf("could not find go on your path. Make sure it is available in your OS PATH environment variable")
	}

	cmd := exec.Command(goExec, "mod", "edit", "--json")

	stdout, err := cmd.StdoutPipe()

	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	m := make(map[string]interface{})

	if err := json.NewDecoder(stdout).Decode(&m); err != nil {
		log.Fatal(err)
	}

	mf := new(ModFile)

	return mf, nil

}

type ModFile struct {
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
