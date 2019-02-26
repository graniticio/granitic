package binder

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"flag"
	"fmt"
	"github.com/graniticio/granitic/v2/config"
	"github.com/graniticio/granitic/v2/instance"
	"github.com/graniticio/granitic/v2/logging"
	"github.com/graniticio/granitic/v2/types"
	"os"
	"path"
	"regexp"
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
	iocImport                  = "github.com/graniticio/granitic/v2/ioc"
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

	logLevelFlag    string = "l"
	logLevelDefault string = "WARN"
	logLevelHelp    string = "The level at which messages will be logged to the console (TRACE, DEBUG, WARN, INFO, ERROR, FATAL)"

	newline = "\n"

	refPrefix        = "ref:"
	refAlias         = "r:"
	refSymbol        = "+"
	refSymbolEscape  = "++"
	confPrefix       = "conf:"
	confAlias        = "c:"
	confSymbol       = "$"
	confSymbolEscape = "$$"
)

// A DefinitionLoader handles the loading of component definition files from a sequence of file paths and can write
// a merged version of those files to a location on a filesystem.
type DefinitionLoader interface {
	LoadAndMerge(files []string, log logging.Logger) (map[string]interface{}, error)
	WriteMerged(data map[string]interface{}, path string, log logging.Logger) error
}

// Settings contains output/input file locations and other variables for controlling the behaviour of this tool
type Settings struct {
	CompDefLocation *string
	BindingsFile    *string
	MergedDebugFile *string
	LogLevelLabel   *string
	LogLevel        logging.LogLevel
}

// SettingsFromArgs uses CLI parameters to populate a Settings object
func SettingsFromArgs() (Settings, error) {

	s := Settings{}

	s.CompDefLocation = flag.String(confLocationFlag, confLocationDefault, confLocationHelp)
	s.BindingsFile = flag.String(bindingsFileFlag, bindingsFileDefault, bindingsFileHelp)
	s.MergedDebugFile = flag.String(mergeLocationFlag, mergeLocationDefault, mergeLocationHelp)
	s.LogLevelLabel = flag.String(logLevelFlag, logLevelDefault, logLevelHelp)

	flag.Parse()

	if ll, err := logging.LogLevelFromLabel(*s.LogLevelLabel); err == nil {

		s.LogLevel = ll

	} else {

		return s, fmt.Errorf("Could not map %s to a valid logging level", *s.LogLevelLabel)

	}

	return s, nil

}

const defaultValuePattern = "(.*)\\((.*)\\)"

// Binder translates the components defined in component definition files into Go source code.
type Binder struct {
	Loader            DefinitionLoader
	ToolName          string
	Log               logging.Logger
	defaultValueRegex *regexp.Regexp
	errorsFound       bool
}

// Bind loads component definitions files from disk/network, merges those files into a single
// view of components and then converts the merged view into Go source code.
func (b *Binder) Bind(s Settings) {

	ca := b.loadConfig(*s.CompDefLocation)

	if *s.MergedDebugFile != "" {
		// Write the merged view of components to a file then exit
		if err := b.Loader.WriteMerged(ca.JSONData, *s.MergedDebugFile, b.Log); err != nil {
			b.exitError(err.Error())
		}

		return
	}

	b.compileRegexes()

	b.Log.LogDebugf("Writing generated bindings file to %s", *s.BindingsFile)

	f := b.openOutputFile(*s.BindingsFile)
	defer f.Close()

	w := bufio.NewWriter(f)
	b.writeBindings(w, ca)

	if b.errorsFound {
		b.exitError("Problems found. Please correct the above and re-run %s", b.ToolName)
	}

}

func (b *Binder) compileRegexes() {
	b.defaultValueRegex = regexp.MustCompile(defaultValuePattern)
}

// SerialiseBuiltinConfig takes the configuration files for Granitic's internal components (facilities) found in
// resource/facility-config and serialises them into a single string that will be embedded into your application's
// executable.
func SerialiseBuiltinConfig(log logging.Logger) string {

	log.LogDebugf("Serialising facility configuration")

	gh, err := LocateFacilityConfig(log)

	if err != nil {
		log.LogFatalf(err.Error())
		instance.ExitError()
	}

	jm := config.NewJSONMergerWithDirectLogging(log, new(config.JSONContentParser))
	jm.MergeArrays = true

	jFiles, err := config.FindJSONFilesInDir(gh)

	if err != nil {
		log.LogFatalf(err.Error())
		instance.ExitError()
	}

	if mc, err := jm.LoadAndMergeConfig(jFiles); err != nil {

		log.LogFatalf("Problem serialising Granitic's built-in config files: %s\n", err.Error())
		instance.ExitError()

	} else {

		b := bytes.Buffer{}
		e := gob.NewEncoder(&b)

		gob.Register(map[string]interface{}{})
		gob.Register([]interface{}{})

		if err := e.Encode(mc); err != nil {
			log.LogFatalf("Problem serialising Granitic's built-in config files: %s\n", err.Error())
			instance.ExitError()
		}

		ser := base64.StdEncoding.EncodeToString(b.Bytes())

		log.LogDebugf("Serialised facility configuration\n")

		return ser

	}

	return ""
}

func (b *Binder) writeBindings(w *bufio.Writer, ca *config.Accessor) {
	b.writePackage(w)
	b.writeImports(w, ca)

	c, err := ca.ObjectVal(componentsField)

	if err != nil {
		b.Log.LogFatalf("Unable to find a %s field in the merged configuration: %s", componentsField, err.Error())
		b.errorsFound = true

		return
	}

	t := b.parseTemplates(ca)

	b.writeEntryFunctionOpen(w, len(c))

	var i = 0

	b.Log.LogDebugf("Processing components:\n")

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

func (b *Binder) writeImports(w *bufio.Writer, configAccessor *config.Accessor) {

	b.Log.LogDebugf("Gathering and writing import statements")

	packages, err := configAccessor.Array(packagesField)

	if err != nil {
		b.Log.LogFatalf("Unable to find a %s field in the merged configuration: %s", packagesField, err.Error())
		b.errorsFound = true

		return
	}

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

	b.Log.LogDebugf("Import statements done\n")
}

func (b *Binder) writeEntryFunctionOpen(w *bufio.Writer, t int) {
	w.WriteString(entryFuncSignature + newline)

	a := fmt.Sprintf("%s := make([]*ioc.ProtoComponent, %d)\n\n", protoArrayVar, t)
	w.WriteString(b.tabIndent(a, 1))
}

func (b *Binder) writeComponent(w *bufio.Writer, name string, component map[string]interface{}, templates map[string]interface{}, index int) {
	baseIndent := 1

	values := make(map[string]interface{})
	refs := make(map[string]interface{})
	confPromises := make(map[string]interface{})

	log := b.Log

	log.LogDebugf("Start component %s", name)

	b.mergeValueSources(component, templates)

	if !b.validateHasTypeField(component, name) {
		log.LogDebugf("Component %s failed", name)
		return
	}

	b.writeComponentNameComment(w, name, baseIndent)
	b.writeInstanceVar(w, name, component[typeField].(string), baseIndent)
	b.writeProto(w, name, index, baseIndent)

	for field, value := range component {

		if b.isPromise(value) {

			log.LogDebugf("%s.%s has a config promise %v", name, field, value)

			confPromises[field] = value

		} else if b.isRef(value) {

			log.LogDebugf("%s.%s has a reference to component %v", name, field, value)

			refs[field] = value

		} else {

			log.LogDebugf("%s.%s has a direct value", name, field)

			values[field] = value
		}

	}

	b.writeValues(w, name, values, baseIndent)
	b.writeConfPromises(w, name, confPromises, baseIndent)
	b.writeDependencies(w, name, refs, baseIndent)

	w.WriteString(newline)
	w.WriteString(newline)

	log.LogDebugf("End component %s\n", name)

}

func (b *Binder) writeValues(w *bufio.Writer, cName string, values map[string]interface{}, tabs int) {

	if len(values) > 0 {
		w.WriteString(newline)
	}

	for k, v := range values {

		if b.reservedFieldName(k) {
			continue
		}

		if s, found := v.(string); found {
			v = b.removeEscapes(s)
		}

		init, wasMap, err := b.asGoInit(v)

		if err != nil {
			b.Log.LogErrorf("Unable to write a value to %s.%s: %s", cName, k, err.Error())
			b.errorsFound = true
			continue
		}

		s := fmt.Sprintf("%s.%s = %s\n", cName, k, init)
		w.WriteString(b.tabIndent(s, tabs))

		if wasMap {
			b.writeMapContents(w, cName, k, v.(map[string]interface{}), tabs)
		}

	}

}

func (b *Binder) removeEscapes(s string) interface{} {

	if strings.HasPrefix(s, confSymbolEscape) || strings.HasPrefix(s, refSymbolEscape) {

		return s[1:]
	}

	return s

}

func (b *Binder) writeConfPromises(w *bufio.Writer, cName string, promises map[string]interface{}, tabs int) {

	p := b.protoName(cName)

	if len(promises) > 0 {
		w.WriteString(newline)
	}

	for k, v := range promises {

		fc := b.stripRepOrConffMarker(v.(string))

		path, defaultValue := b.extractDefaultValue(fc)

		s := fmt.Sprintf("%s.%s(%s, %s)\n", p, "AddConfigPromise", b.quoteString(k), b.quoteString(path))
		w.WriteString(b.tabIndent(s, tabs))

		if defaultValue != "" {

			b.Log.LogDebugf("Found and storing a default value for %s.%s", p, k)

			s = fmt.Sprintf("%s.%s(%s, %s)\n", p, "AddDefaultValue", b.quoteString(k), b.quoteString(defaultValue))
			w.WriteString(b.tabIndent(s, tabs))
		}

	}

}

func (b *Binder) extractDefaultValue(s string) (string, string) {

	m := b.defaultValueRegex.FindStringSubmatch(s)

	if len(m) == 3 {

		return m[1], m[2]

	}

	return s, ""
}

func (b *Binder) writeDependencies(w *bufio.Writer, cName string, promises map[string]interface{}, tabs int) {

	b.Log.LogDebugf("Writing component dependencies")

	p := b.protoName(cName)

	if len(promises) > 0 {
		w.WriteString(newline)
	}

	for k, v := range promises {

		fc := b.stripRepOrConffMarker(v.(string))

		s := fmt.Sprintf("%s.%s(%s, %s)\n", p, "AddDependency", b.quoteString(k), b.quoteString(fc))
		w.WriteString(b.tabIndent(s, tabs))

	}

	b.Log.LogDebugf("Component dependencies done\n")

}

func (b *Binder) stripRepOrConffMarker(s string) string {

	if strings.HasPrefix(s, refSymbol) || strings.HasPrefix(s, confSymbol) {
		return s[1:]
	}

	return strings.SplitN(s, ":", 2)[1]
}

func (b *Binder) writeMapContents(w *bufio.Writer, iName string, fName string, contents map[string]interface{}, tabs int) {

	for k, v := range contents {

		gi, _, err := b.asGoInit(v)

		if err != nil {
			b.Log.LogErrorf("Unable to write a value to %s.%s[%s]: %s", iName, fName, b.quoteString(k), err.Error())
			b.errorsFound = true
			continue
		}

		s := fmt.Sprintf("%s.%s[%s] = %s\n", iName, fName, b.quoteString(k), gi)
		w.WriteString(b.tabIndent(s, tabs))
	}
}

func (b *Binder) asGoInit(v interface{}) (string, bool, error) {

	switch config.JSONType(v) {
	case config.JSONMap:

		s, err := b.asGoMapInit(v)
		return s, true, err
	case config.JSONArray:

		s, err := b.asGoArrayInit(v)

		return s, false, err
	default:
		return fmt.Sprintf("%#v", v), false, nil
	}
}

func (b *Binder) asGoMapInit(v interface{}) (string, error) {
	a := v.(map[string]interface{})

	at, err := b.assessMapValueType(a)

	if err != nil {
		return "", err
	}

	s := fmt.Sprintf("make(map[string]%s)", at)
	return s, nil
}

func (b *Binder) asGoArrayInit(v interface{}) (string, error) {
	a := v.([]interface{})

	at, err := b.assessArrayType(a)

	if err != nil {
		return "", err
	}

	var buf bytes.Buffer

	s := fmt.Sprintf("[]%s{", at)
	buf.WriteString(s)

	for i, m := range a {
		gi, _, _ := b.asGoInit(m)
		buf.WriteString(gi)

		if i+1 < len(a) {
			buf.WriteString(", ")
		}

	}

	s = fmt.Sprintf("}")
	buf.WriteString(s)

	return buf.String(), nil
}

func (b *Binder) assessMapValueType(a map[string]interface{}) (string, error) {

	var currentType = config.Unset
	var sampleVal interface{}

	if len(a) == 0 {
		return "", fmt.Errorf("this tool does not support empty maps as component values as the type of the map can't be determined")
	}

	for _, v := range a {

		newType := config.JSONType(v)
		sampleVal = v

		if newType == config.JSONMap {
			return "", fmt.Errorf("this tool does not support nested maps/objects as component values")
		}

		if currentType == config.Unset {
			currentType = newType
			continue
		}

		if newType != currentType {
			return "interface{}", nil
		}
	}

	if currentType == config.JSONArray {

		var at string
		var err error

		if at, err = b.assessArrayType(sampleVal.([]interface{})); err == nil {
			return fmt.Sprintf("[]%s", at), nil
		}

		return "", err
	}

	switch t := sampleVal.(type) {
	default:
		return fmt.Sprintf("%T", t), nil
	}
}

func (b *Binder) assessArrayType(a []interface{}) (string, error) {

	var currentType = config.Unset

	if len(a) == 0 {
		return "", fmt.Errorf("this tool does not support zero-length (empty) arrays as component values as the type can't be determined")
	}

	for _, v := range a {

		newType := config.JSONType(v)

		if newType == config.JSONMap || newType == config.JSONArray {
			return "", fmt.Errorf("This tool does not support multi-dimensional arrays or object arrays as component values")
		}

		if currentType == config.Unset {
			currentType = newType
			continue
		}

		if newType != currentType {
			return "interface{}", nil
		}
	}

	//All the types in the array are the same - but need to be careful with floats without decimals
	switch t := a[0].(type) {
	case int:
		for _, m := range a {
			switch at := m.(type) {
			case float64:
				//Although the first array member looks like an int, it's part of a set of floats
				return fmt.Sprintf("%T", at), nil

			}
		}

		return fmt.Sprintf("%T", t), nil

	default:
		return fmt.Sprintf("%T", t), nil
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

	return strings.HasPrefix(s, confPrefix) || strings.HasPrefix(s, confAlias) || (strings.HasPrefix(s, confSymbol) && !strings.HasPrefix(s, confSymbolEscape))
}

func (b *Binder) isRef(v interface{}) bool {
	s, found := v.(string)

	if !found {
		return false
	}

	return strings.HasPrefix(s, refPrefix) || strings.HasPrefix(s, refAlias) || (strings.HasPrefix(s, refSymbol) && !strings.HasPrefix(s, refSymbolEscape))

}

func (b *Binder) reservedFieldName(f string) bool {
	return f == templateField || f == templateFieldAlias || f == typeField || f == typeFieldAlias
}

func (b *Binder) validateHasTypeField(v map[string]interface{}, name string) bool {

	t := v[typeField]

	if t == nil {
		b.Log.LogErrorf("Component %s does not have a 'type' defined in its component defintion (or any parent templates).\n", name)
		b.errorsFound = true
		return false
	}

	_, found := t.(string)

	if !found {
		b.Log.LogErrorf("Component %s has a 'type' field defined but the value of the field is not a string.\n", name)
		b.errorsFound = true
		return false
	}

	return true

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

func (b *Binder) parseTemplates(ca *config.Accessor) map[string]interface{} {

	b.Log.LogDebugf("Processing component templates")

	flattened := make(map[string]interface{})

	if !ca.PathExists(templatesField) {
		return flattened
	}

	templates, err := ca.ObjectVal(templatesField)

	if err != nil {
		b.errorsFound = true
		b.Log.LogErrorf("Problem using the %s field in the merged component definition file: %s", templatesField, err.Error())
		return map[string]interface{}{}
	}

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

	b.Log.LogDebugf("Finished processing component templates\n")

	return flattened

}

func (b *Binder) writeSerialisedConfig(w *bufio.Writer) {

	sv := SerialiseBuiltinConfig(b.Log)

	s := fmt.Sprintf("%s := \"%s\"\n", serialisedVar, sv)

	w.WriteString(b.tabIndent(s, 1))

}

func (b *Binder) writeFrameworkModifiers(w *bufio.Writer, ca *config.Accessor) {

	tabs := 1

	s := fmt.Sprintf("%s := make(map[string]map[string]string)\n", modifierVar)
	w.WriteString(b.tabIndent(s, tabs))
	w.WriteString(newline)

	if !ca.PathExists(frameworkField) {
		return
	}

	fm, err := ca.ObjectVal(frameworkField)

	if err != nil {

		b.errorsFound = true
		b.Log.LogErrorf("Problem using the %s field in the merged component definition file: %s", frameworkField, err.Error())
	}

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

func (b *Binder) loadConfig(l string) *config.Accessor {

	log := b.Log

	log.LogDebugf("Loading component definition files from %s", l)

	s := strings.Split(l, ",")
	fl, err := config.ExpandToFilesAndURLs(s)

	if err != nil {
		m := fmt.Sprintf("Problem loading config from %s %s", l, err.Error())
		b.exitError(m)
	}

	mc, err := b.Loader.LoadAndMerge(fl, log)

	if err != nil {
		m := fmt.Sprintf("Problem merging component definition files togther: %s", err.Error())
		b.exitError(m)
	}

	ca := new(config.Accessor)
	ca.JSONData = mc
	ca.FrameworkLogger = b.Log

	if !ca.PathExists(packagesField) {
		// Add the missing packages section
		ca.JSONData[packagesField] = []interface{}{}
	}

	if !ca.PathExists(componentsField) {
		// Add the missing components section
		ca.JSONData[componentsField] = map[string]interface{}{}

	}

	return ca
}

func (b *Binder) exitError(message string, a ...interface{}) {

	b.Log.LogFatalf(message, a...)

	os.Exit(1)
}
