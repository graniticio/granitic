package generate

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"path/filepath"
)

type TemplateConfigWriter func(confDir string, pg *ProjectGenerator)
type TemplateComponentWriter func(confDir string, pg *ProjectGenerator)

type ProjectGenerator struct {
	ConfWriterFunc TemplateConfigWriter
	CompWriterFunc TemplateComponentWriter
	ToolName       string
}

func (pg *ProjectGenerator) Generate() {
	a := os.Args

	if len(a) < 2 {
		pg.exitError("You must provide a name for your project")
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

	pg.mkDir(name)
	pg.mkDir(resourceDir)
	pg.mkDir(confDir)
	pg.mkDir(compDir)

	pg.CompWriterFunc(compDir, pg)
	pg.ConfWriterFunc(confDir, pg)
	pg.writeMainFile(name, projectPackage, changePackageComment)
	pg.writeGitIgnore(name)
}

func (pg *ProjectGenerator) writeMainFile(name string, projectPackage string, changePackageComment string) {

	mainFile := filepath.Join(name, "service.go")

	f := pg.OpenOutputFile(mainFile)

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

func (pg *ProjectGenerator) writeGitIgnore(name string) {

	ignoreFile := filepath.Join(name, ".gitignore")

	f := pg.OpenOutputFile(ignoreFile)

	defer f.Close()

	w := bufio.NewWriter(f)

	w.WriteString("bindings*\n")
	w.WriteString(name + "\n")
	w.Flush()

}

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
