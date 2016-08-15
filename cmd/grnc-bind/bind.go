package main

import (
	"flag"
	"github.com/graniticio/granitic/config"
	"strings"
	"fmt"
	"os"
	"github.com/graniticio/granitic/facility/jsonmerger"
	"github.com/graniticio/granitic/logging"
	"encoding/json"
	"io/ioutil"
	"path"
	"bufio"
)

const (

	packagesField = "packages"
	componentsField = "components"
	templatesField = "templates"
	templateField = "template"
	bindingsPackage = "bindings"
	iocImport = "github.com/graniticio/granitic/ioc"
	entryFuncSignature = "func Components() []*ioc.ProtoComponent {"
	protoArrayVar = "pc"
	confLocationFlag string = "c"
	confLocationDefault string = "resource/components"
	confLocationHelp string = "A comma separated list of component definition files or directories containing component definition files"

	bindingsFileFlag string = "o"
	bindingsFileDefault string = "bindings/bindings.go"
	bindingsFileHelp string = "Path to the Go source file that will be generated"

	mergeLocationFlag string = "m"
	mergeLocationDefault string = ""
	mergeLocationHelp string = "The path of a file where the merged component defintion file should be written to. Execution will halt after writing."

	newline = "\n"
	nameField = "name"
	typeField = "type"

	deferSeparator = ":"
	refPrefix      = "ref"
	refAlias       = "r"
	confPrefix     = "conf"
	confAlias      = "c"
)


func main() {

	var confLocation = flag.String(confLocationFlag, confLocationDefault, confLocationHelp)
	var bindingsfile = flag.String(bindingsFileFlag, bindingsFileDefault, bindingsFileHelp)
	var mergedComponentsFile = flag.String(mergeLocationFlag, mergeLocationDefault, mergeLocationHelp)

	flag.Parse()

	ca := loadConfig(*confLocation)

	if (*mergedComponentsFile != "") {
		writeMergedAndExit(ca, *mergedComponentsFile)
	}

	f := openOutputFile(*bindingsfile)
	defer f.Close()

	w := bufio.NewWriter(f)
	writeBindings(w, ca)
}

func writeBindings(w *bufio.Writer, ca *config.ConfigAccessor) {
	writePackage(w)
	writeImports(w, ca)

	c := ca.ObjectVal(componentsField)
	t := parseTemplates(ca)

	writeEntryFunctionOpen(w, len(c))

	for name, v := range c {

		writeComponent(w, name, v.((map[string]interface{})), t)
	}

	writeEntryFunctionClose(w)
	w.Flush()
}


func loadConfig(l string) *config.ConfigAccessor{

	s := strings.Split(l, ",")
	fl, err := config.ExpandToFiles(s)

	if err != nil {
		m := fmt.Sprintf("Problem loading config from %s %s", l, err.Error())
		fatal(m)
	}

	jm := new(jsonmerger.JsonMerger)
	jm.Logger = new(logging.ConsoleErrorLogger)

	mc := jm.LoadAndMergeConfig(fl)

	ca := new(config.ConfigAccessor)
	ca.JsonData = mc
	ca.FrameworkLogger = new(logging.ConsoleErrorLogger)

	if !ca.PathExists(packagesField) || !ca.PathExists(componentsField){
		m := fmt.Sprintf("The merged component definition file must contain a %s and a %s section.\n", packagesField, componentsField)
		fatal(m)

	}

	return ca
}

func writePackage(w *bufio.Writer) {

	l := fmt.Sprintf("package %s\n\n", bindingsPackage)
	w.WriteString(l)
}


func writeImports(w *bufio.Writer, configAccessor *config.ConfigAccessor) {
	packages := configAccessor.Array(packagesField)

	w.WriteString("import (\n")

	iocImp := tabIndent(quoteString(iocImport), 1)
	w.WriteString(iocImp + newline)

	for _, packageName := range packages {

		i := quoteString(packageName.(string))
		i = tabIndent(i, 1)
		w.WriteString(i + newline)
	}

	w.WriteString(")\n\n")
}

func writeEntryFunctionOpen(w *bufio.Writer, i int) {
	w.WriteString(entryFuncSignature + newline)

	a := fmt.Sprintf("%s := make([]*ioc.ProtoComponent, %d)\n\n", protoArrayVar, i)
	w.WriteString(tabIndent(a, 1))
}

func writeComponent(w *bufio.Writer, name string, values map[string]interface{}, templates map[string]interface{}) {
	baseIdent := 1

	mergeValueSources(values, templates)
	validateHasType(values, name)

	writeComponentNameComment(w, name, baseIdent)
	writeInstanceVar(w, name, values[typeField].(string), baseIdent)




	w.WriteString(newline)

}

func writeComponentNameComment(w *bufio.Writer, n string, i int) {
	s := fmt.Sprintf("//%s\n", n)
	w.WriteString(tabIndent(s, i))
}

func writeInstanceVar(w *bufio.Writer, n string, t string, i int) {
	s := fmt.Sprintf("%s := new(%s)\n",n,t)
	w.WriteString(tabIndent(s, i))
}



func writeEntryFunctionClose(w *bufio.Writer) {
	a := fmt.Sprintf("}\n")
	w.WriteString(a)
}

func validateHasType(v map[string]interface{}, name string) {

	t := v[typeField]

	if t == nil {
		m := fmt.Sprintf("Component %s does not have a 'type' defined in its component defintion (or any parent templates).\n", name)
		fatal(m)
	}

	_, found := t.(string)

	if !found {
		m := fmt.Sprintf("Component %s has a 'type' field defined but the value of the field is not a string.\n", name)
		fatal(m)
	}

}

func mergeValueSources(c map[string]interface{}, t map[string]interface{}){

	if c[templateField] != nil {
		flatten(c, t, c[templateField].(string))
	}
}

func quoteString(s string) string{
	return fmt.Sprintf("\"%s\"", s)
}

func tabIndent(s string, t int) string{

	for i := 0; i < t; i++ {
		s = "\t" + s
	}

	return s
}


func writeMergedAndExit(ca *config.ConfigAccessor, f string) {

	b, err := json.MarshalIndent(ca.JsonData, "", "\t")

	if err != nil {
		fatal(err.Error())
	}

	err = ioutil.WriteFile(f, b, 0644)

	if err != nil {
		fatal(err.Error())
	}

	os.Exit(0)
}



func openOutputFile(p string) *os.File {
	os.MkdirAll(path.Dir(p), 0777)
	f, err := os.Create(p)

	if err != nil {
		m := fmt.Sprintf(err.Error() + "\n")
		fatal(m)
	}

	return f
}

func parseTemplates(ca *config.ConfigAccessor) map[string]interface{} {

	flattened := make(map[string]interface{})

	if !ca.PathExists(templatesField) {
		return flattened
	}

	templates := ca.ObjectVal(templatesField)

	for n, t := range templates {
		checkForTemplateLoop(t.(map[string]interface{}), templates, []string{n})

		ft := make(map[string]interface{})
		flatten(ft, templates, n)

		flattened[n] = ft

	}


	return flattened

}

func flatten(target map[string]interface{}, templates map[string]interface{}, tname string) {

	if templates[tname] == nil{
		fmt.Printf("No template %s\n", tname)
		return
	}

	parent := templates[tname].(map[string]interface{})

	for k, v := range parent {

		if target[k] == nil && k != templateField{
			target[k] = v
		}

	}

	if parent[templateField] != nil {
		flatten(target, templates, parent[templateField].(string))
	}


}

func checkForTemplateLoop(template map[string]interface{}, templates map[string]interface{}, chain []string) {

	if template[templateField] == nil {
		return
	}

	p := template[templateField].(string)

	if contains(chain, p) {
		message := fmt.Sprintf("Invalid template inheritance %v\n", append(chain, p))
		fatal(message)
	}

	if templates[p] ==  nil{
		message := fmt.Sprintf("No template exists with name %s\n", p)
		fatal(message)
	}

	checkForTemplateLoop(templates[p].(map[string]interface{}), templates, append(chain, p))


}

func contains(a []string, c string) bool{
	for _, s := range a {
		if s == c {
			return true
		}
	}

	return false
}


func fatal(m string) {
	fmt.Printf(m)
	os.Exit(-1)
}

