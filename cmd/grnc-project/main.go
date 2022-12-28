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
	project-name/resource/components/components.yaml
	project-name/resource/config/config.yaml
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
	_ "embed"
	"fmt"
	"github.com/graniticio/granitic/v3"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
)

const usage = "grnc-project project-name [module-name]"
const perms os.FileMode = 0755
const compDir = "comp-def"
const configDir = "config"

//go:embed templates/gomod.template
var goModTemplate string

//go:embed templates/maingo.template
var mainGoTemplate string

//go:embed templates/gitignore.template
var ignoreTemplate string

//go:embed templates/config.template
var configTemplate string

//go:embed templates/components.template
var compTemplate string

func main() {

	s := settingsFromArgs()

	createDirStructure(&s)

	m := make(map[string]string)

	m["Module"] = s.ModuleName
	m["FullVersion"] = granitic.Version
	m["MajorVersion"] = grncMajorVersion()

	createModuleFile(s, m)
	createMainFile(s, m)
	createIgnoreFile(s, m)
	createComponentFile(s, m)
	createConfigFile(s, m)
}

func createModuleFile(s Settings, m map[string]string) {

	modFile := filepath.Join(s.projectDir, "go.mod")

	fileFromTemplate(modFile, goModTemplate, m)

}

func createMainFile(s Settings, m map[string]string) {

	modFile := filepath.Join(s.projectDir, "main.go")

	fileFromTemplate(modFile, mainGoTemplate, m)

}

func createIgnoreFile(s Settings, m map[string]string) {

	modFile := filepath.Join(s.projectDir, ".gitignore")

	fileFromTemplate(modFile, ignoreTemplate, m)

}

func createConfigFile(s Settings, m map[string]string) {

	modFile := filepath.Join(s.projectDir, configDir, "base.yaml")

	fileFromTemplate(modFile, configTemplate, m)

}

func createComponentFile(s Settings, m map[string]string) {

	modFile := filepath.Join(s.projectDir, compDir, "base.yaml")

	fileFromTemplate(modFile, compTemplate, m)

}

func fileFromTemplate(path string, templateContent string, data any) {
	t, err := template.New("TMP").Parse(templateContent)

	if err != nil {
		exitWithError(err.Error())
	}

	f := openOutputFile(path)

	defer f.Close()

	w := bufio.NewWriter(f)

	t.Execute(w, data)

	w.Flush()
}

func createDirStructure(s *Settings) {

	s.projectDir = s.ProjectName

	s.confDir = filepath.Join(s.projectDir, "config")
	s.compDir = filepath.Join(s.projectDir, "comp-def")

	mkDir(s.projectDir)
	mkDir(s.confDir)
	mkDir(s.compDir)
}

// Settings contains the arguments for this tool
type Settings struct {
	ProjectName string
	ModuleName  string
	projectDir  string
	compDir     string
	confDir     string
}

// settingsFromArgs uses CLI parameters to populate a Settings object
func settingsFromArgs() Settings {

	a := os.Args

	if len(a) < 2 {

		message := fmt.Sprintf("Usage:\n\n\t%s\n", usage)

		exitWithError(message)
	}

	project := a[1]
	module := project

	if len(a) > 2 {
		module = a[2]
	}

	return Settings{
		ModuleName:  module,
		ProjectName: project,
	}

}

func exitWithError(message string) {
	fmt.Fprintln(os.Stdout, message)
	os.Exit(1)
}

func mkDir(dir string) {
	if err := os.Mkdir(dir, perms); err != nil {
		exitWithError(err.Error())
	}
}

func openOutputFile(p string) *os.File {
	os.MkdirAll(path.Dir(p), perms)

	if f, err := os.Create(p); err != nil {
		exitWithError(err.Error())
	} else {
		return f
	}

	return nil
}

func grncMajorVersion() string {
	return strings.Split(granitic.Version, ".")[0]
}
