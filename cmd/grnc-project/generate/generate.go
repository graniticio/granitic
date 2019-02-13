package generate

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"path/filepath"
)

// TemplateConfigWriter creates a skeleton Granitic configuration file in the supplied location
type TemplateConfigWriter func(confDir string, pg *ProjectGenerator)

// TemplateComponentWriter creates a skeleton Granitic component definition file in the supplied location
type TemplateComponentWriter func(confDir string, pg *ProjectGenerator)

// MainFileContentWriter creates a Go source file with main function that initialises and passes control to Granitic
type MainFileContentWriter func(w *bufio.Writer, pp string)

// ProjectGenerator creates a blank Granitic project that is ready to build and start
type ProjectGenerator struct {
	ConfWriterFunc TemplateConfigWriter
	CompWriterFunc TemplateComponentWriter
	MainFileFunc   MainFileContentWriter
	ToolName       string
}

// Settings contains the arguments for this tool
type Settings struct {
	ProjectName string
	ModuleName  string
	BaseFolder  string
}

// SettingsFromArgs uses CLI parameters to populate a Settings object
func (pg *ProjectGenerator) SettingsFromArgs() Settings {

	a := os.Args

	if len(a) < 2 {
		pg.exitError("You must provide a name for your project")
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
	resourceDir := filepath.Join(base, name, "resource")
	confDir := filepath.Join(resourceDir, "config")
	compDir := filepath.Join(resourceDir, "components")

	pg.mkDir(projectDir)
	pg.mkDir(resourceDir)
	pg.mkDir(confDir)
	pg.mkDir(compDir)

	pg.CompWriterFunc(compDir, pg)
	pg.ConfWriterFunc(confDir, pg)
	pg.writeMainFile(base, name, module)
	pg.writeGitIgnore(base, name)
	pg.writeModFile(base, name, module)
}

func (pg *ProjectGenerator) writeMainFile(base, name, module string) {

	mainFile := filepath.Join(base, name, "service.go")

	f := pg.OpenOutputFile(mainFile)

	defer f.Close()

	w := bufio.NewWriter(f)

	pg.MainFileFunc(w, module)

	w.Flush()

}

func (pg *ProjectGenerator) writeModFile(base, name, module string) {

	modFile := filepath.Join(base, name, "go.mod")

	f := pg.OpenOutputFile(modFile)

	defer f.Close()

	w := bufio.NewWriter(f)

	fmt.Fprintf(w, "module %s\n\n", module)
	fmt.Fprintf(w, "require github.com/graniticio/granitic/v2 v2\n")

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
		pg.exitError(err.Error())
	} else {
		return f
	}

	return nil
}

func (pg *ProjectGenerator) mkDir(dir string) {
	if err := os.Mkdir(dir, 0755); err != nil {
		pg.exitError(err.Error())
	}
}

func (pg *ProjectGenerator) exitError(message string, a ...interface{}) {

	m := fmt.Sprintf("%s: %s \n", pg.ToolName, message)

	fmt.Printf(m, a...)
	os.Exit(1)
}
