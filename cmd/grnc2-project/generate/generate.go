package generate

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"path/filepath"
)

var allowExit = true

// TemplateConfigWriter creates a skeleton Granitic configuration file in the supplied location
type TemplateConfigWriter func(confDir string, pg *ProjectGenerator)

// TemplateComponentWriter creates a skeleton Granitic component definition file in the supplied location
type TemplateComponentWriter func(confDir string, pg *ProjectGenerator)

// ModFileWriter creates a go.mod file in the supplied location
type ModFileWriter func(baseDir string, moduleName string, pg *ProjectGenerator)

// MainFileContentWriter creates a Go source file with main function that initialises and passes control to Granitic
type MainFileContentWriter func(w *bufio.Writer, pp string)

// ProjectGenerator creates a blank Granitic project that is ready to build and start
type ProjectGenerator struct {
	ConfWriterFunc TemplateConfigWriter
	CompWriterFunc TemplateComponentWriter
	MainFileFunc   MainFileContentWriter
	ModFileFunc    ModFileWriter
	ToolName       string
}

// Settings contains the arguments for this tool
type Settings struct {
	ProjectName string
	ModuleName  string
	BaseFolder  string
}

type exitFunc func(message string, a ...interface{})

// SettingsFromArgs uses CLI parameters to populate a Settings object
func SettingsFromArgs(ef exitFunc) Settings {

	a := os.Args

	if len(a) < 2 {
		ef("You must provide a name for your project")
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

// Generate creates the folder structure and blank/skeleton files for a new Granitic project that will be ready to build
func (pg *ProjectGenerator) Generate(s Settings) {

	name := s.ProjectName
	module := s.ModuleName
	base := s.BaseFolder

	projectDir := filepath.Join(base, name)
	confDir := filepath.Join(projectDir, "config")
	compDir := filepath.Join(projectDir, "comp-def")

	pg.mkDir(projectDir)
	pg.mkDir(confDir)
	pg.mkDir(compDir)

	pg.CompWriterFunc(compDir, pg)
	pg.ConfWriterFunc(confDir, pg)
	pg.writeMainFile(base, name, module)
	pg.writeGitIgnore(base, name)
	pg.ModFileFunc(projectDir, module, pg)
}

func (pg *ProjectGenerator) writeMainFile(base, name, module string) {

	mainFile := filepath.Join(base, name, "main.go")

	f := pg.OpenOutputFile(mainFile)

	defer f.Close()

	w := bufio.NewWriter(f)

	pg.MainFileFunc(w, module)

	w.Flush()

}

func (pg *ProjectGenerator) writeGitIgnore(base, name string) {

	ignoreFile := filepath.Join(base, name, ".gitignore")

	f := pg.OpenOutputFile(ignoreFile)

	defer f.Close()

	w := bufio.NewWriter(f)

	w.WriteString("bindings*\n")
	w.WriteString(name + "\n")
	w.Flush()

}

// OpenOutputFile opens the supplied file path in create mode. Exits if there is a problem opening the file.
func (pg *ProjectGenerator) OpenOutputFile(p string) *os.File {
	os.MkdirAll(path.Dir(p), 0755)

	if f, err := os.Create(p); err != nil {
		pg.ExitError(err.Error())
	} else {
		return f
	}

	return nil
}

func (pg *ProjectGenerator) mkDir(dir string) {
	if err := os.Mkdir(dir, 0755); err != nil {
		pg.ExitError(err.Error())
	}
}

// ExitError writes
func (pg *ProjectGenerator) ExitError(message string, a ...interface{}) {

	m := fmt.Sprintf("%s: %s \n", pg.ToolName, message)

	fmt.Printf(m, a...)

	if allowExit {
		os.Exit(1)
	}
}
