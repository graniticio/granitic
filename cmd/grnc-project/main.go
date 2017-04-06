// Copyright 2016 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

/*
	The grnc-project tool, used to generate skeleton project files for a new Granitic application.

	Running

		grnc-project project-name [package]

	Will create the following files and directories:

		project-name
		project-name/.gitignore
		project-name/service.go
		project-name/resource/components/components.json
		project-name/resource/config/config.json

	This will allow a minimal Granitic application to be built and started by running:

		cd project-name && grnc-bind && go build && ./project-name

	Developers should pay attention to the import statements in the generated project-name.go file. It will contain a line similar
	to:

		import "./bindings"

	This is a relative import path, which will allow the project to be built and run with no knowledge of your workspace
	layout, but will prevent your application being installed with 'go install' and isn't considered good Go practice.
	The line should be changed to a non-relative path that reflects the layout of your Go workspace, which is most often:

		import "github.com/yourGitHubUser/yourPackage/bindings"

	You can specify your project's package as the second argument to the grnc-project tool

	The .gitignore file contains:

		bindings*
		project-name

	Which prevents the output of 'grnc-bind' and 'go build' being included in your repository.
*/
package main

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"path/filepath"
)

func main() {

	a := os.Args

	if len(a) < 2 {
		exitError("You must provide a name for your project")
	}

	projectPackage := "."
	changePackageComment := "  //Change to a non-relative path if you want to use 'go install'"

	if len(a) > 2 {
		projectPackage = a[2]
		changePackageComment = ""
	}

	name := a[1]
	resourceDir := filepath.Join(name, "resource")
	confDir := filepath.Join(resourceDir, "config")
	compDir := filepath.Join(resourceDir, "components")

	mkDir(name)
	mkDir(resourceDir)
	mkDir(confDir)
	mkDir(compDir)

	writeComponentsFile(compDir)
	writeConfigFile(confDir)
	writeMainFile(name, projectPackage, changePackageComment)
	writeGitIgnore(name)
}

func writeMainFile(name string, projectPackage string, changePackageComment string) {

	mainFile := filepath.Join(name, "service.go")

	f := openOutputFile(mainFile)

	defer f.Close()

	w := bufio.NewWriter(f)

	w.WriteString("package main\n\n")
	w.WriteString("import \"github.com/graniticio/granitic\"\n")
	w.WriteString("import \"")
	w.WriteString(projectPackage)
	w.WriteString("/bindings\"")
	w.WriteString(changePackageComment)
	w.WriteString("\n\n")
	w.WriteString("func main() {\n")
	w.WriteString("\tgranitic.StartGranitic(bindings.Components())\n")
	w.WriteString("}\n")
	w.Flush()

}

func writeGitIgnore(name string) {

	ignoreFile := filepath.Join(name, ".gitignore")

	f := openOutputFile(ignoreFile)

	defer f.Close()

	w := bufio.NewWriter(f)

	w.WriteString("bindings*\n")
	w.WriteString(name + "\n")
	w.Flush()

}

func writeConfigFile(confDir string) {

	compFile := filepath.Join(confDir, "config.json")
	f := openOutputFile(compFile)

	defer f.Close()

	w := bufio.NewWriter(f)

	w.WriteString("{\n")
	w.WriteString("}\n")

	w.Flush()

}

func writeComponentsFile(compDir string) {

	compFile := filepath.Join(compDir, "components.json")
	f := openOutputFile(compFile)

	defer f.Close()

	w := bufio.NewWriter(f)

	w.WriteString("{\n")
	w.WriteString(tab("\"packages\": [],\n", 1))
	w.WriteString(tab("\"components\": {}\n", 1))
	w.WriteString("}\n")

	w.Flush()

}

func openOutputFile(p string) *os.File {
	os.MkdirAll(path.Dir(p), 0755)

	if f, err := os.Create(p); err != nil {
		exitError(err.Error())
	} else {
		return f
	}

	return nil
}

func mkDir(dir string) {
	if err := os.Mkdir(dir, 0755); err != nil {
		exitError(err.Error())
	}
}

func exitError(message string, a ...interface{}) {

	m := "grnc-project: " + message + "\n"

	fmt.Printf(m, a...)
	os.Exit(1)
}

func tab(s string, t int) string {

	for i := 0; i < t; i++ {
		s = "  " + s
	}

	return s
}
