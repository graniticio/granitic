package binder

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"flag"
	"fmt"
	"github.com/graniticio/granitic/config"
	"github.com/graniticio/granitic/instance"
	"github.com/graniticio/granitic/logging"
	"github.com/graniticio/granitic/types"
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
	bindingsFileHelp    string = "Path to the Go source file that will be generated to instatiate your components"

	mergeLocationFlag    string = "m"
	mergeLocationDefault string = ""
	mergeLocationHelp    string = "The path of a file where the merged component definition file should be written to. Execution will halt after writing."

	newline = "\n"

	refPrefix         = "ref:"
	refAlias          = "r:"
	confPrefix        = "conf:"
	confAlias         = "c:"
	emptyStructPrefix = "empty-struct:"
	emptyStructAlias  = "es:"
)

type DefinitionLoader interface {
	LoadAndMerge(files []string) (map[string]interface{}, error)
	WriteMerged(data map[string]interface{}, path string) error
}

type Binder struct {
	Loader   DefinitionLoader
	ToolName string
}

func (b *Binder) Bind() {
	var confLocation = flag.String(confLocationFlag, confLocationDefault, confLocationHelp)
	var bindingsFile = flag.String(bindingsFileFlag, bindingsFileDefault, bindingsFileHelp)
	var mergedComponentsFile = flag.String(mergeLocationFlag, mergeLocationDefault, mergeLocationHelp)

	flag.Parse()

	ca := b.loadConfig(*confLocation)

	if *mergedComponentsFile != "" {
		if err := b.Loader.WriteMerged(ca.JsonData, *mergedComponentsFile); err != nil {
			b.exitError(err.Error())
		}
	}

	f := b.openOutputFile(*bindingsFile)
	defer f.Close()

	w := bufio.NewWriter(f)
	b.writeBindings(w, ca)
}

func SerialiseBuiltinConfig() string {
	gh := config.GraniticHome()

	ghr := path.Join(gh, "resource", "facility-config")
	
	fcf, err := config.FindConfigFilesInDir(ghr);
	if err != nil {
		fmt.Printf("%s does not seem to contain a valid Granitic installation. Check your %s and/or %s environment variables\n", gh, "GRANITIC_HOME", "GOPATH")
		instance.ExitError()
	}
	
	jm := config.NewJsonMergerWithDirectLogging(new(logging.ConsoleErrorLogger), new(config.JsonContentParser))
	jm.MergeArrays = true
	mc, err := jm.LoadAndMergeConfig(fcf)
	if err != nil {
		fmt.Printf("Problem serialising Granitic's built-in config files: %s\n", err.Error())
		instance.ExitError()
	} 

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

func (b *Binder) writeBindings(w *bufio.Writer, ca *config.ConfigAccessor) {
	b.writePackage(w)
	b.writeImports(w, ca)

	c, err := ca.ObjectVal(componentsField)
	b.checkErr(err)

	t := b.parseTemplates(ca)

	b.writeEntryFunctionOpen(w, len(c))

	var i = 0

	for name, v := range c {

		b.writeComponent(w, name, v.((map[string]interface{})), t, i)
		i++
	}

	b.writeSerialisedConfig(w)
	b.writeFrameworkModifiers(w, ca)

	b.writeEntryFunctionClose(w)
	w.Flush()
}

func (b *Binder) writePackage(w *bufio.Writer) {

	l := fmt.Sprintf("package %s\n\n", bindingsPackage)
	w.WriteString(l)
}

func (b *Binder) writeImports(w *bufio.Writer, configAccessor *config.ConfigAccessor) {
	packages, err := configAccessor.Array(packagesField)
	b.checkErr(err)

	seen := types.NewEmptyOrderedStringSet()

	w.WriteString("import (\n")

	iocImp := b.tabIndent(b.quoteString(iocImport), 1)
	w.WriteString(iocImp + newline)

	for _, packageName := range packages {

		p := packageName.(string)

		if seen.Contains(p) {
			continue
		} else {
			seen.Add(p)
		}

		i := b.quoteString(p)
		i = b.tabIndent(i, 1)
		w.WriteString(i + newline)
	}

	w.WriteString(")\n\n")
}

func (b *Binder) writeEntryFunctionOpen(w *bufio.Writer, t int) {
	w.WriteString(entryFuncSignature + newline)

	a := fmt.Sprintf("%s := make([]*ioc.ProtoComponent, %d)\n\n", protoArrayVar, t)
	w.WriteString(b.tabIndent(a, 1))
}

func (b *Binder) writeComponent(w *bufio.Writer, name string, component map[string]interface{}, templates map[string]interface{}, index int) {
	baseIdent := 1

	values := make(map[string]interface{})
	refs := make(map[string]interface{})
	emptyStructs := make(map[string]interface{})
	confPromises := make(map[string]interface{})

	b.mergeValueSources(component, templates)
	b.validateHasTypeField(component, name)

	b.writeComponentNameComment(w, name, baseIdent)
	b.writeInstanceVar(w, name, component[typeField].(string), baseIdent)
	b.writeProto(w, name, index, baseIdent)

	for field, value := range component {

		if b.isPromise(value) {
			confPromises[field] = value

		} else if b.isRef(value) {
			refs[field] = value

		} else if b.isEmptyStruct(value) {
			emptyStructs[field] = value
		} else {
			values[field] = value
		}

	}

	b.writeValues(w, name, values, baseIdent)
	b.writeDeferred(w, name, confPromises, baseIdent, "AddConfigPromise")
	b.writeDeferred(w, name, refs, baseIdent, "AddDependency")
	b.writeEmptyStructFunctions(w, name, emptyStructs, baseIdent)

	w.WriteString(newline)
	w.WriteString(newline)

}

func (b *Binder) writeValues(w *bufio.Writer, cName string, values map[string]interface{}, tabs int) {

	if len(values) > 0 {
		w.WriteString(newline)
	}

	for k, v := range values {

		if b.reservedFieldName(k) {
			continue
		}

		init, wasMap := b.asGoInit(v)

		s := fmt.Sprintf("%s.%s = %s\n", cName, k, init)
		w.WriteString(b.tabIndent(s, tabs))

		if wasMap {
			b.writeMapContents(w, cName, k, v.(map[string]interface{}), tabs)
		}

	}

}

func (b *Binder) writeDeferred(w *bufio.Writer, cName string, promises map[string]interface{}, tabs int, funcName string) {

	p := b.protoName(cName)

	if len(promises) > 0 {
		w.WriteString(newline)
	}

	for k, v := range promises {

		fc := strings.SplitN(v.(string), ":", 2)[1]

		s := fmt.Sprintf("%s.%s(%s, %s)\n", p, funcName, b.quoteString(k), b.quoteString(fc))
		w.WriteString(b.tabIndent(s, tabs))

	}

}

func (b *Binder) writeEmptyStructFunctions(w *bufio.Writer, cName string, emptyStructs map[string]interface{}, tabs int) {

	if len(emptyStructs) > 0 {
		w.WriteString(newline)
	}

	for k, v := range emptyStructs {

		reqType := strings.SplitN(v.(string), ":", 2)[1]

		s := fmt.Sprintf("%s.%s = func() interface{} {\n", cName, k)
		w.WriteString(b.tabIndent(s, tabs))

		s = fmt.Sprintf("\treturn new(%s)\n", reqType)
		w.WriteString(b.tabIndent(s, tabs))

		w.WriteString(b.tabIndent("}\n\n", tabs))

	}

}

func (b *Binder) writeMapContents(w *bufio.Writer, iName string, fName string, contents map[string]interface{}, tabs int) {

	for k, v := range contents {

		gi, _ := b.asGoInit(v)

		s := fmt.Sprintf("%s.%s[%s] = %s\n", iName, fName, b.quoteString(k), gi)
		w.WriteString(b.tabIndent(s, tabs))
	}
}

func (b *Binder) asGoInit(v interface{}) (string, bool) {

	switch config.JsonType(v) {
	case config.JsonMap:
		return b.asGoMapInit(v), true
	case config.JsonArray:
		return b.asGoArrayInit(v), false
	default:
		return fmt.Sprintf("%#v", v), false
	}
}

func (b *Binder) asGoMapInit(v interface{}) string {
	a := v.(map[string]interface{})

	at := b.assessMapValueType(a)

	s := fmt.Sprintf("make(map[string]%s)", at)
	return s
}

func (b *Binder) asGoArrayInit(v interface{}) string {
	a := v.([]interface{})

	at := b.assessArrayType(a)

	var buf bytes.Buffer

	s := fmt.Sprintf("[]%s{", at)
	buf.WriteString(s)

	for i, m := range a {
		gi, _ := b.asGoInit(m)
		buf.WriteString(gi)

		if i+1 < len(a) {
			buf.WriteString(", ")
		}

	}

	s = fmt.Sprintf("}")
	buf.WriteString(s)

	return buf.String()
}

func (b *Binder) assessMapValueType(a map[string]interface{}) string {

	var currentType = config.Unset
	var sampleVal interface{}

	if len(a) == 0 {
		b.exitError("This tool does not support empty maps as component values as the type of the map can't be determined.")
	}

	for _, v := range a {

		newType := config.JsonType(v)
		sampleVal = v

		if newType == config.JsonMap {
			b.exitError("This tool does not support nested maps/objects as component values.\n")
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
		return "[]" + b.assessArrayType(sampleVal.([]interface{}))
	}

	switch t := sampleVal.(type) {
	default:
		return fmt.Sprintf("%T", t)
	}
}

func (b *Binder) assessArrayType(a []interface{}) string {

	var currentType = config.Unset

	if len(a) == 0 {
		b.exitError("This tool does not support zero-length (empty) arrays as component values as the type can't be determined.")
	}

	for _, v := range a {

		newType := config.JsonType(v)

		if newType == config.JsonMap || newType == config.JsonArray {
			b.exitError("This tool does not support multi-dimensional arrays or object arrays as component values\n")
		}

		if currentType == config.Unset {
			currentType = newType
			continue
		}

		if newType != currentType {
			return "interface{}"
		}
	}

	//All the types in the array are the same - but need to be careful with floats without decimals
	switch t := a[0].(type) {
	case int:
		for _, m := range a {
			switch at := m.(type) {
			case float64:
				//Although the first array member looks like an int, it's part of a set of floats
				return fmt.Sprintf("%T", at)

			}
		}

		return fmt.Sprintf("%T", t)

	default:
		return fmt.Sprintf("%T", t)
	}

}

func (b *Binder) writeComponentNameComment(w *bufio.Writer, n string, i int) {
	s := fmt.Sprintf("//%s\n", n)
	w.WriteString(b.tabIndent(s, i))
}

func (b *Binder) writeInstanceVar(w *bufio.Writer, n string, ct string, tabs int) {
	s := fmt.Sprintf("%s := new(%s)\n", n, ct)
	w.WriteString(b.tabIndent(s, tabs))
}

func (b *Binder) writeProto(w *bufio.Writer, n string, index int, tabs int) {

	p := b.protoName(n)

	s := fmt.Sprintf("%s := ioc.CreateProtoComponent(%s, %s)\n", p, n, b.quoteString(n))
	w.WriteString(b.tabIndent(s, tabs))
	s = fmt.Sprintf("%s[%d] = %s\n", protoArrayVar, index, p)
	w.WriteString(b.tabIndent(s, tabs))
}

func (b *Binder) writeEntryFunctionClose(w *bufio.Writer) {
	a := fmt.Sprintf("\treturn ioc.NewProtoComponents(%s, %s, &%s)\n}\n", protoArrayVar, modifierVar, serialisedVar)
	w.WriteString(a)
}

func (b *Binder) protoName(n string) string {
	return n + protoSuffix
}

func (b *Binder) isPromise(v interface{}) bool {

	s, found := v.(string)

	if !found {
		return false
	}

	return strings.HasPrefix(s, confPrefix) || strings.HasPrefix(s, confAlias)
}

func (b *Binder) isRef(v interface{}) bool {
	s, found := v.(string)

	if !found {
		return false
	}

	return strings.HasPrefix(s, refPrefix) || strings.HasPrefix(s, refAlias)

}

func (b *Binder) isEmptyStruct(v interface{}) bool {

	s, found := v.(string)

	if !found {
		return false
	}

	return strings.HasPrefix(s, emptyStructPrefix) || strings.HasPrefix(s, emptyStructAlias)
}

func (b *Binder) reservedFieldName(f string) bool {
	return f == templateField || f == templateFieldAlias || f == typeField || f == typeFieldAlias
}

func (b *Binder) validateHasTypeField(v map[string]interface{}, name string) {

	t := v[typeField]

	if t == nil {
		m := fmt.Sprintf("Component %s does not have a 'type' defined in its component defintion (or any parent templates).\n", name)
		b.exitError(m)
	}

	_, found := t.(string)

	if !found {
		m := fmt.Sprintf("Component %s has a 'type' field defined but the value of the field is not a string.\n", name)
		b.exitError(m)
	}

}

func (b *Binder) mergeValueSources(c map[string]interface{}, t map[string]interface{}) {

	b.replaceAliases(c)

	if c[templateField] != nil {
		b.flatten(c, t, c[templateField].(string))
	}
}

func (b *Binder) quoteString(s string) string {
	return fmt.Sprintf("\"%s\"", s)
}

func (b *Binder) tabIndent(s string, t int) string {

	for i := 0; i < t; i++ {
		s = "\t" + s
	}

	return s
}

func (b *Binder) openOutputFile(p string) *os.File {
	os.MkdirAll(path.Dir(p), 0777)
	f, err := os.Create(p)

	if err != nil {
		m := fmt.Sprintf(err.Error() + "\n")
		b.exitError(m)
	}

	return f
}

func (b *Binder) parseTemplates(ca *config.ConfigAccessor) map[string]interface{} {

	flattened := make(map[string]interface{})

	if !ca.PathExists(templatesField) {
		return flattened
	}

	templates, err := ca.ObjectVal(templatesField)
	b.checkErr(err)

	for _, template := range templates {
		b.replaceAliases(template.(map[string]interface{}))
	}

	for n, template := range templates {

		t := template.(map[string]interface{})

		b.checkForTemplateLoop(t, templates, []string{n})

		ft := make(map[string]interface{})
		b.flatten(ft, templates, n)

		flattened[n] = ft

	}

	return flattened

}

func (b *Binder) writeSerialisedConfig(w *bufio.Writer) {

	sv := SerialiseBuiltinConfig()

	s := fmt.Sprintf("%s := \"%s\"\n", serialisedVar, sv)

	w.WriteString(b.tabIndent(s, 1))

}

func (b *Binder) writeFrameworkModifiers(w *bufio.Writer, ca *config.ConfigAccessor) {

	tabs := 1

	s := fmt.Sprintf("%s := make(map[string]map[string]string)\n", modifierVar)
	w.WriteString(b.tabIndent(s, tabs))
	w.WriteString(newline)

	if !ca.PathExists(frameworkField) {
		return
	}

	fm, err := ca.ObjectVal(frameworkField)
	b.checkErr(err)

	for fc, mods := range fm {

		n := fc + modsSuffix

		s := fmt.Sprintf("%s := make(map[string]string)\n", n)
		w.WriteString(b.tabIndent(s, tabs))

		s = fmt.Sprintf("%s[%s] = %s\n", modifierVar, b.quoteString(fc), n)
		w.WriteString(b.tabIndent(s, tabs))

		for f, d := range mods.(map[string]interface{}) {

			s := fmt.Sprintf("%s[%s] = %s\n", n, b.quoteString(f), b.quoteString(d.(string)))
			w.WriteString(b.tabIndent(s, tabs))

		}

		w.WriteString(newline)

	}

}

func (b *Binder) replaceAliases(vs map[string]interface{}) {
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

func (b *Binder) flatten(target map[string]interface{}, templates map[string]interface{}, tname string) {

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
		b.flatten(target, templates, parent[templateField].(string))
	}

}

func (b *Binder) checkForTemplateLoop(template map[string]interface{}, templates map[string]interface{}, chain []string) {

	if template[templateField] == nil {
		return
	}

	p := template[templateField].(string)

	if b.contains(chain, p) {
		message := fmt.Sprintf("Invalid template inheritance %v\n", append(chain, p))
		b.exitError(message)
	}

	if templates[p] == nil {
		message := fmt.Sprintf("No template exists with name %s\n", p)
		b.exitError(message)
	}

	b.checkForTemplateLoop(templates[p].(map[string]interface{}), templates, append(chain, p))

}

func (b *Binder) contains(a []string, c string) bool {
	for _, s := range a {
		if s == c {
			return true
		}
	}

	return false
}

func (b *Binder) loadConfig(l string) *config.ConfigAccessor {

	s := strings.Split(l, ",")
	fl, err := config.ExpandToFilesAndURLs(s)

	if err != nil {
		m := fmt.Sprintf("Problem loading config from %s %s", l, err.Error())
		b.exitError(m)
	}

	mc, err := b.Loader.LoadAndMerge(fl)

	if err != nil {
		m := fmt.Sprintf("Problem merging component definition files togther: %s", err.Error())
		b.exitError(m)
	}

	ca := new(config.ConfigAccessor)
	ca.JsonData = mc
	ca.FrameworkLogger = new(logging.ConsoleErrorLogger)

	if !ca.PathExists(packagesField) || !ca.PathExists(componentsField) {
		m := fmt.Sprintf("The merged component definition file must contain a %s and a %s section.\n", packagesField, componentsField)
		b.exitError(m)

	}

	return ca
}

func (b *Binder) exitError(message string, a ...interface{}) {

	m := fmt.Sprintf("%s: %s\n", b.ToolName, message)

	fmt.Printf(m, a...)
	os.Exit(1)
}

func (b *Binder) checkErr(e error) {
	if e != nil {
		b.exitError(e.Error())
	}
}
