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

// Generate creates the folder structure and blank/skeleton files for a new Granitic project that will be ready to build
func (pg *ProjectGenerator) Generate() {
	a := os.Args

	if len(a) < 2 {
		pg.exitError("You must provide a name for your project")
	}

	name := a[1]
	module := name

	if len(a) > 2 {
		module = a[2]
	}

	resourceDir := filepath.Join(name, "resource")
	confDir := filepath.Join(resourceDir, "config")
	compDir := filepath.Join(resourceDir, "components")

	pg.mkDir(name)
	pg.mkDir(resourceDir)
	pg.mkDir(confDir)
	pg.mkDir(compDir)

	pg.CompWriterFunc(compDir, pg)
	pg.ConfWriterFunc(confDir, pg)
	pg.writeMainFile(name, module)
	pg.writeGitIgnore(name)
	pg.writeModFile(name, module)
}

func (pg *ProjectGenerator) writeMainFile(name string, module string) {

	mainFile := filepath.Join(name, "service.go")

	f := pg.OpenOutputFile(mainFile)

	defer f.Close()

	w := bufio.NewWriter(f)

	pg.MainFileFunc(w, module)

	w.Flush()

}

func (pg *ProjectGenerator) writeModFile(name string, module string) {

	modFile := filepath.Join(name, "go.mod")

	f := pg.OpenOutputFile(modFile)

	defer f.Close()

	w := bufio.NewWriter(f)

	fmt.Fprintf(w, "module %s\n\n", module)
	fmt.Fprintf(w, "require github.com/graniticio/granitic/v2 v2\n")

	w.Flush()

}

func (pg *ProjectGenerator) writeGitIgnore(name string) {

	ignoreFile := filepath.Join(name, ".gitignore")

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
