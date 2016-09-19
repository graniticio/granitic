package main

import (
	"bufio"
	"fmt"
	"os"
	"path"
)

func main() {

	a := os.Args

	if len(a) < 2 {
		exitError("You must provide a name for your project")
	}

	name := a[1]
	resourceDir := name + "/resource"
	confDir := resourceDir + "/config"
	compDir := resourceDir + "/components"

	mkDir(name)
	mkDir(resourceDir)
	mkDir(confDir)
	mkDir(compDir)

	writeComponentsFile(compDir)
	writeConfigFile(confDir)
	writeMainFile(name)
	writeGitIgnore(name)
}

func writeMainFile(name string) {

	mainFile := name + "/" + name + ".go"

	f := openOutputFile(mainFile)

	defer f.Close()

	w := bufio.NewWriter(f)

	w.WriteString("package main\n\n")
	w.WriteString("import \"github.com/graniticio/granitic\"\n")
	w.WriteString("import \"./bindings\"  //Change to a non-relative path if you want to use 'go install'\n\n")
	w.WriteString("func main() {\n")
	w.WriteString("\tgranitic.StartGranitic(bindings.Components())\n")
	w.WriteString("}\n")
	w.Flush()

}

func writeGitIgnore(name string) {

	ignoreFile := name + "/.gitignore"

	f := openOutputFile(ignoreFile)

	defer f.Close()

	w := bufio.NewWriter(f)

	w.WriteString("bindings*\n")
	w.WriteString(name + "\n")
	w.Flush()

}

func writeConfigFile(confDir string) {

	compFile := confDir + "/config.json"
	f := openOutputFile(compFile)

	defer f.Close()

	w := bufio.NewWriter(f)

	w.WriteString("{\n")
	w.WriteString("}\n")

	w.Flush()

}

func writeComponentsFile(compDir string) {

	compFile := compDir + "/components.json"
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
