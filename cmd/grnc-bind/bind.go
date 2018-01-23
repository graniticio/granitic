// Copyright 2016-2018 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

/*
	The grnc-bind tool - used to convert Granitic's JSON component definition files into Go source.

	Go does not support a 'type-from-name' mechanism for instantiating objects, so the container cannot create arbitrarily typed
	objects at runtime. Instead, Granitic component definition files are used to generate Go source files that will be
	compiled along with your application. The grnc-bind tool performs this code generation.

	In most cases, the grnc-bind command will be run, without arguments, in your application's root directory (the same folder
	that contains your resources directory. The tool will merge together any .json files found in resources/components and
	create a file bindings/bindings.go. This file includes a single function:

		Components() *ioc.ProtoComponents

	The results of that function are then included in your application's call to start Granticic. E.g.

		func main() {
			granitic.StartGranitic(bindings.Components())
		}

	grnc-bind will need to be re-run whenever a component definition file is modified.

	Usage of grnc-bind:

		grnc-bind [-c component-files] [-m merged-file-out] [-o generated-file]

		-c string
			A comma separated list of component definition files or directories containing component definition files (default "resource/components")
		-m string
			The path of a file where the merged component defintion file should be written to. Execution will halt after writing.
		-o string
			Path to the Go source file that will be generated (default "bindings/bindings.go")

*/
package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/graniticio/granitic/config"
	"github.com/graniticio/granitic/instance"
	"github.com/graniticio/granitic/logging"
	"github.com/graniticio/granitic/types"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

const (
	packagesField      = "packages"
	componentsField    = "components"
	frameworkField     = "frameworkModifiers"
	templatesField     = "templates"
	templateField      = "compTemplate"
	templateFieldAlias = "ct"
	typeField          = "type"
	typeFieldAlias     = "t"

	protoSuffix = "Proto"
	modsSuffix  = "Mods"

	bindingsPackage            = "bindings"
	iocImport                  = "github.com/graniticio/granitic/ioc"
	entryFuncSignature         = "func Components() *ioc.ProtoComponents {"
	protoArrayVar              = "protoComponents"
	modifierVar                = "frameworkModifiers"
	serialisedVar              = "ser"
	confLocationFlag    string = "c"
	confLocationDefault string = "resource/components"
	confLocationHelp    string = "A comma separated list of component definition files or directories containing component definition files"

	bindingsFileFlag    string = "o"
	bindingsFileDefault string = "bindings/bindings.go"
	bindingsFileHelp    string = "Path to the Go source file that will be generated containing your JSON component details"

	mergeLocationFlag    string = "m"
	mergeLocationDefault string = ""
	mergeLocationHelp    string = "The path of a file where the merged component definition file should be written to. Execution will halt after writing."

	newline = "\n"

	refPrefix  = "ref:"
	refAlias   = "r:"
	confPrefix = "conf:"
	confAlias  = "c:"
)

func main() {

	var confLocation = flag.String(confLocationFlag, confLocationDefault, confLocationHelp)
	var bindingsFile = flag.String(bindingsFileFlag, bindingsFileDefault, bindingsFileHelp)
	var mergedComponentsFile = flag.String(mergeLocationFlag, mergeLocationDefault, mergeLocationHelp)

	flag.Parse()

	ca := loadConfig(*confLocation)

	if *mergedComponentsFile != "" {
		writeMergedAndExit(ca, *mergedComponentsFile)
	}

	f := openOutputFile(*bindingsFile)
	defer f.Close()

	w := bufio.NewWriter(f)
	writeBindings(w, ca)

}

func serialiseBuiltinConfig() string {
	gh := config.GraniticHome()

	ghr := path.Join(gh, "resource", "facility-config")

	if fcf, err := config.FindConfigFilesInDir(ghr); err != nil {
		fmt.Printf("%s does not seem to contain a valid Granitic installation. Check your %s and/or %s environment variables\n", gh, "GRANITIC_HOME", "GOPATH")
		instance.ExitError()
	} else {

		jm := new(config.JsonMerger)
		jm.MergeArrays = true
		jm.Logger = new(logging.ConsoleErrorLogger)

		if mc, err := jm.LoadAndMergeConfig(fcf); err != nil {

			fmt.Printf("Problem serialising Granitic's built-in config files: %s\n", err.Error())
			instance.ExitError()

		} else {

			b := bytes.Buffer{}
			e := gob.NewEncoder(&b)

			gob.Register(map[string]interface{}{})
			gob.Register([]interface{}{})

			if err := e.Encode(mc); err != nil {
				fmt.Printf("Problem serialising Granitic's built-in config files: %s\n", err.Error())
				instance.ExitError()
			}

			ser := base64.StdEncoding.EncodeToString(b.Bytes())
			return ser

		}

	}

	return ""
}

func writeBindings(w *bufio.Writer, ca *config.ConfigAccessor) {
	writePackage(w)
	writeImports(w, ca)

	c, err := ca.ObjectVal(componentsField)
	checkErr(err)

	t := parseTemplates(ca)

	writeEntryFunctionOpen(w, len(c))

	var i = 0

	for name, v := range c {

		writeComponent(w, name, v.((map[string]interface{})), t, i)
		i++
	}

	writeSerialisedConfig(w)
	writeFrameworkModifiers(w, ca)

	writeEntryFunctionClose(w)
	w.Flush()
}

func writePackage(w *bufio.Writer) {

	l := fmt.Sprintf("package %s\n\n", bindingsPackage)
	w.WriteString(l)
}

func writeImports(w *bufio.Writer, configAccessor *config.ConfigAccessor) {
	packages, err := configAccessor.Array(packagesField)
	checkErr(err)

	seen := types.NewEmptyOrderedStringSet()

	w.WriteString("import (\n")

	iocImp := tabIndent(quoteString(iocImport), 1)
	w.WriteString(iocImp + newline)

	for _, packageName := range packages {

		p := packageName.(string)

		if seen.Contains(p) {
			continue
		} else {
			seen.Add(p)
		}

		i := quoteString(p)
		i = tabIndent(i, 1)
		w.WriteString(i + newline)
	}

	w.WriteString(")\n\n")
}

func writeEntryFunctionOpen(w *bufio.Writer, t int) {
	w.WriteString(entryFuncSignature + newline)

	a := fmt.Sprintf("%s := make([]*ioc.ProtoComponent, %d)\n\n", protoArrayVar, t)
	w.WriteString(tabIndent(a, 1))
}

func writeComponent(w *bufio.Writer, name string, component map[string]interface{}, templates map[string]interface{}, index int) {
	baseIdent := 1

	values := make(map[string]interface{})
	refs := make(map[string]interface{})
	confPromises := make(map[string]interface{})

	mergeValueSources(component, templates)
	validateHasTypeField(component, name)

	writeComponentNameComment(w, name, baseIdent)
	writeInstanceVar(w, name, component[typeField].(string), baseIdent)
	writeProto(w, name, index, baseIdent)

	for field, value := range component {

		if isPromise(value) {
			confPromises[field] = value

		} else if isRef(value) {
			refs[field] = value

		} else {
			values[field] = value
		}

	}

	writeValues(w, name, values, baseIdent)
	writeDeferred(w, name, confPromises, baseIdent, "AddConfigPromise")
	writeDeferred(w, name, refs, baseIdent, "AddDependency")

	w.WriteString(newline)
	w.WriteString(newline)

}

func writeValues(w *bufio.Writer, cName string, values map[string]interface{}, tabs int) {

	if len(values) > 0 {
		w.WriteString(newline)
	}

	for k, v := range values {

		if reservedFieldName(k) {
			continue
		}

		init, wasMap := asGoInit(v)

		s := fmt.Sprintf("%s.%s = %s\n", cName, k, init)
		w.WriteString(tabIndent(s, tabs))

		if wasMap {
			writeMapContents(w, cName, k, v.(map[string]interface{}), tabs)
		}

	}

}

func writeDeferred(w *bufio.Writer, cName string, promises map[string]interface{}, tabs int, funcName string) {

	p := protoName(cName)

	if len(promises) > 0 {
		w.WriteString(newline)
	}

	for k, v := range promises {

		fc := strings.SplitN(v.(string), ":", 2)[1]

		s := fmt.Sprintf("%s.%s(%s, %s)\n", p, funcName, quoteString(k), quoteString(fc))
		w.WriteString(tabIndent(s, tabs))

	}

}

func writeMapContents(w *bufio.Writer, iName string, fName string, contents map[string]interface{}, tabs int) {

	for k, v := range contents {

		gi, _ := asGoInit(v)

		s := fmt.Sprintf("%s.%s[%s] = %s\n", iName, fName, quoteString(k), gi)
		w.WriteString(tabIndent(s, tabs))
	}
}

func asGoInit(v interface{}) (string, bool) {

	switch config.JsonType(v) {
	case config.JsonMap:
		return asGoMapInit(v), true
	case config.JsonArray:
		return asGoArrayInit(v), false
	default:
		return fmt.Sprintf("%#v", v), false
	}
}

func asGoMapInit(v interface{}) string {
	a := v.(map[string]interface{})

	at := assessMapValueType(a)

	s := fmt.Sprintf("make(map[string]%s)", at)
	return s
}

func asGoArrayInit(v interface{}) string {
	a := v.([]interface{})

	at := assessArrayType(a)

	var b bytes.Buffer

	s := fmt.Sprintf("[]%s{", at)
	b.WriteString(s)

	for i, m := range a {
		gi, _ := asGoInit(m)
		b.WriteString(gi)

		if i+1 < len(a) {
			b.WriteString(", ")
		}

	}

	s = fmt.Sprintf("}")
	b.WriteString(s)

	return b.String()
}

func assessMapValueType(a map[string]interface{}) string {

	var currentType = config.Unset
	var sampleVal interface{}

	if len(a) == 0 {
		exitError("This tool does not support empty maps as component values as the type of the map can't be determined.")
	}

	for _, v := range a {

		newType := config.JsonType(v)
		sampleVal = v

		if newType == config.JsonMap {
			exitError("This tool does not support nested maps/objects as component values.\n")
		}

		if currentType == config.Unset {
			currentType = newType
			continue
		}

		if newType != currentType {
			return "interface{}"
		}
	}

	if currentType == config.JsonArray {
		return "[]" + assessArrayType(sampleVal.([]interface{}))
	}

	switch t := sampleVal.(type) {
	default:
		return fmt.Sprintf("%T", t)
	}
}

func assessArrayType(a []interface{}) string {

	var currentType = config.Unset

	if len(a) == 0 {
		exitError("This tool does not support zero-length (empty) arrays as component values as the type can't be determined.")
	}

	for _, v := range a {

		newType := config.JsonType(v)

		if newType == config.JsonMap || newType == config.JsonArray {
			exitError("This tool does not support multi-dimensional arrays or object arrays as component values\n")
		}

		if currentType == config.Unset {
			currentType = newType
			continue
		}

		if newType != currentType {
			return "interface{}"
		}
	}

	switch t := a[0].(type) {
	default:
		return fmt.Sprintf("%T", t)
	}
}

func writeComponentNameComment(w *bufio.Writer, n string, i int) {
	s := fmt.Sprintf("//%s\n", n)
	w.WriteString(tabIndent(s, i))
}

func writeInstanceVar(w *bufio.Writer, n string, ct string, tabs int) {
	s := fmt.Sprintf("%s := new(%s)\n", n, ct)
	w.WriteString(tabIndent(s, tabs))
}

func writeProto(w *bufio.Writer, n string, index int, tabs int) {

	p := protoName(n)

	s := fmt.Sprintf("%s := ioc.CreateProtoComponent(%s, %s)\n", p, n, quoteString(n))
	w.WriteString(tabIndent(s, tabs))
	s = fmt.Sprintf("%s[%d] = %s\n", protoArrayVar, index, p)
	w.WriteString(tabIndent(s, tabs))
}

func writeEntryFunctionClose(w *bufio.Writer) {
	a := fmt.Sprintf("\treturn ioc.NewProtoComponents(%s, %s, &%s)\n}\n", protoArrayVar, modifierVar, serialisedVar)
	w.WriteString(a)
}

func protoName(n string) string {
	return n + protoSuffix
}

func isPromise(v interface{}) bool {

	s, found := v.(string)

	if !found {
		return false
	}

	return strings.HasPrefix(s, confPrefix) || strings.HasPrefix(s, confAlias)
}

func isRef(v interface{}) bool {
	s, found := v.(string)

	if !found {
		return false
	}

	return strings.HasPrefix(s, refPrefix) || strings.HasPrefix(s, refAlias)

}

func reservedFieldName(f string) bool {
	return f == templateField || f == templateFieldAlias || f == typeField || f == typeFieldAlias
}

func validateHasTypeField(v map[string]interface{}, name string) {

	t := v[typeField]

	if t == nil {
		m := fmt.Sprintf("Component %s does not have a 'type' defined in its component defintion (or any parent templates).\n", name)
		exitError(m)
	}

	_, found := t.(string)

	if !found {
		m := fmt.Sprintf("Component %s has a 'type' field defined but the value of the field is not a string.\n", name)
		exitError(m)
	}

}

func mergeValueSources(c map[string]interface{}, t map[string]interface{}) {

	replaceAliases(c)

	if c[templateField] != nil {
		flatten(c, t, c[templateField].(string))
	}
}

func quoteString(s string) string {
	return fmt.Sprintf("\"%s\"", s)
}

func tabIndent(s string, t int) string {

	for i := 0; i < t; i++ {
		s = "\t" + s
	}

	return s
}

func writeMergedAndExit(ca *config.ConfigAccessor, f string) {

	b, err := json.MarshalIndent(ca.JsonData, "", "\t")

	if err != nil {
		exitError(err.Error())
	}

	err = ioutil.WriteFile(f, b, 0644)

	if err != nil {
		exitError(err.Error())
	}

	os.Exit(0)
}

func openOutputFile(p string) *os.File {
	os.MkdirAll(path.Dir(p), 0777)
	f, err := os.Create(p)

	if err != nil {
		m := fmt.Sprintf(err.Error() + "\n")
		exitError(m)
	}

	return f
}

func parseTemplates(ca *config.ConfigAccessor) map[string]interface{} {

	flattened := make(map[string]interface{})

	if !ca.PathExists(templatesField) {
		return flattened
	}

	templates, err := ca.ObjectVal(templatesField)
	checkErr(err)

	for _, template := range templates {
		replaceAliases(template.(map[string]interface{}))
	}

	for n, template := range templates {

		t := template.(map[string]interface{})

		checkForTemplateLoop(t, templates, []string{n})

		ft := make(map[string]interface{})
		flatten(ft, templates, n)

		flattened[n] = ft

	}

	return flattened

}

func writeSerialisedConfig(w *bufio.Writer) {

	sv := serialiseBuiltinConfig()

	s := fmt.Sprintf("%s := \"%s\"\n", serialisedVar, sv)

	w.WriteString(tabIndent(s, 1))

}

func writeFrameworkModifiers(w *bufio.Writer, ca *config.ConfigAccessor) {

	tabs := 1

	s := fmt.Sprintf("%s := make(map[string]map[string]string)\n", modifierVar)
	w.WriteString(tabIndent(s, tabs))
	w.WriteString(newline)

	if !ca.PathExists(frameworkField) {
		return
	}

	fm, err := ca.ObjectVal(frameworkField)
	checkErr(err)

	for fc, mods := range fm {

		n := fc + modsSuffix

		s := fmt.Sprintf("%s := make(map[string]string)\n", n)
		w.WriteString(tabIndent(s, tabs))

		s = fmt.Sprintf("%s[%s] = %s\n", modifierVar, quoteString(fc), n)
		w.WriteString(tabIndent(s, tabs))

		for f, d := range mods.(map[string]interface{}) {

			s := fmt.Sprintf("%s[%s] = %s\n", n, quoteString(f), quoteString(d.(string)))
			w.WriteString(tabIndent(s, tabs))

		}

		w.WriteString(newline)

	}

}

func replaceAliases(vs map[string]interface{}) {
	tma := vs[templateFieldAlias]

	if tma != nil {
		delete(vs, templateFieldAlias)
		vs[templateField] = tma
	}

	tya := vs[typeFieldAlias]

	if tya != nil {

		delete(vs, typeFieldAlias)
		vs[typeField] = tya

	}
}

func flatten(target map[string]interface{}, templates map[string]interface{}, tname string) {

	if templates[tname] == nil {
		fmt.Printf("No template %s\n", tname)
		return
	}

	parent := templates[tname].(map[string]interface{})

	for k, v := range parent {

		if target[k] == nil && k != templateField {
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
		exitError(message)
	}

	if templates[p] == nil {
		message := fmt.Sprintf("No template exists with name %s\n", p)
		exitError(message)
	}

	checkForTemplateLoop(templates[p].(map[string]interface{}), templates, append(chain, p))

}

func contains(a []string, c string) bool {
	for _, s := range a {
		if s == c {
			return true
		}
	}

	return false
}

func loadConfig(l string) *config.ConfigAccessor {

	s := strings.Split(l, ",")
	fl, err := config.ExpandToFilesAndURLs(s)

	if err != nil {
		m := fmt.Sprintf("Problem loading config from %s %s", l, err.Error())
		exitError(m)
	}

	jm := new(config.JsonMerger)
	jm.MergeArrays = true
	jm.Logger = new(logging.ConsoleErrorLogger)

	mc, err := jm.LoadAndMergeConfig(fl)

	if err != nil {
		m := fmt.Sprintf("Problem merging JSON files togther: %s", err.Error())
		exitError(m)
	}

	ca := new(config.ConfigAccessor)
	ca.JsonData = mc
	ca.FrameworkLogger = new(logging.ConsoleErrorLogger)

	if !ca.PathExists(packagesField) || !ca.PathExists(componentsField) {
		m := fmt.Sprintf("The merged component definition file must contain a %s and a %s section.\n", packagesField, componentsField)
		exitError(m)

	}

	return ca
}

func exitError(message string, a ...interface{}) {

	m := "grnc-ctl: " + message + "\n"

	fmt.Printf(m, a...)
	os.Exit(1)
}

func checkErr(e error) {
	if e != nil {
		exitError(e.Error())
	}
}
