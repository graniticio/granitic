// Copyright 2016-2022 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

/*
The grnc-project tool, used to generate skeleton project files for a new Granitic application. The generated project is module aware.

Running

	grnc-project project-name [module-name]

Will create the following files and directories:

	project-name
	project-name/.gitignore
	project-name/service.go
	project-name/resource/components/components.json
	project-name/resource/config/config.json
	project-name/go.mod

This will allow a minimal Granitic application to be built and started by running:

	cd project-name && grnc-bind && go build && ./project-name

Developers should pay attention to the import statements in the generated project-name.go file. It will contain a line similar
to:

	import "./bindings"

This is a relative import path, which will allow the project to be built and run with no knowledge of your workspace
layout, but will prevent your application being installed with 'go install' and isn't considered good Go practice.
The line should be changed to a non-relative path that reflects the layout of your Go workspace, which is most often:

	import "github.com/yourGitHubUser/yourPackage/bindings"

# Your project's module name will be the same as the project name unless you provide the module name as the second argument to this tool

The .gitignore file contains:

	bindings*
	project-name

Which prevents the output of 'grnc-bind' and 'go build' being included in your repository.
*/
package main

import (
	"bufio"
	"fmt"
	"github.com/graniticio/granitic/v3/cmd/grnc-project/generate"
	"path/filepath"
)

func main() {

	pg := newJSONProjectGenerator()

	s := generate.SettingsFromArgs(pg.ExitError)

	pg.Generate(s)

}

func newJSONProjectGenerator() *generate.ProjectGenerator {
	pg := new(generate.ProjectGenerator)
	pg.CompWriterFunc = writeComponentsFile
	pg.ConfWriterFunc = writeConfigFile
	pg.MainFileFunc = writeMainFile
	pg.ModFileFunc = writeModFile
	pg.ToolName = "grnc-project"

	return pg
}

func tab(s string, t int) string {

	for i := 0; i < t; i++ {
		s = "  " + s
	}

	return s
}

func writeConfigFile(confDir string, pg *generate.ProjectGenerator) {

	compFile := filepath.Join(confDir, "base.json")
	f := pg.OpenOutputFile(compFile)

	defer f.Close()

	w := bufio.NewWriter(f)

	w.WriteString("{\n")
	w.WriteString("}\n")

	w.Flush()

}

func writeComponentsFile(compDir string, pg *generate.ProjectGenerator) {

	compFile := filepath.Join(compDir, "common.json")
	f := pg.OpenOutputFile(compFile)

	defer f.Close()

	w := bufio.NewWriter(f)

	w.WriteString("{\n")
	w.WriteString(tab("\"packages\": [],\n", 1))
	w.WriteString(tab("\"components\": {}\n", 1))
	w.WriteString("}\n")

	w.Flush()

}

func writeMainFile(w *bufio.Writer, module string) {

	w.WriteString("package main\n\n")
	w.WriteString("import \"github.com/graniticio/granitic/v3\"\n")
	w.WriteString("import \"")
	w.WriteString(module)
	w.WriteString("/bindings\"")
	w.WriteString("\n\n")
	w.WriteString("func main() {\n")
	w.WriteString("\tgranitic.StartGranitic(bindings.Components())\n")
	w.WriteString("}\n")

}

func writeModFile(baseDir string, moduleName string, pg *generate.ProjectGenerator) {

	modFile := filepath.Join(baseDir, "go.mod")

	f := pg.OpenOutputFile(modFile)

	defer f.Close()

	w := bufio.NewWriter(f)

	fmt.Fprintf(w, "module %s\n\n", moduleName)
	fmt.Fprintf(w, "require github.com/graniticio/granitic/v3 v2\n")

	w.Flush()

}
