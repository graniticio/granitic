package main

import (
	"github.com/graniticio/granitic/v3/cmd/grnc-project/generate"
	"os"
	"path/filepath"
	"testing"
)

func TestProjectCreation(t *testing.T) {

	s := generate.Settings{
		BaseFolder:  os.TempDir(),
		ModuleName:  "my-mod",
		ProjectName: "my-proj",
	}

	pf := filepath.Join(s.BaseFolder, s.ProjectName)

	os.RemoveAll(pf)

	pg := newJSONProjectGenerator()

	pg.Generate(s)

}
