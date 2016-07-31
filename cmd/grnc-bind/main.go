package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/graniticio/granitic/config"
	"github.com/graniticio/granitic/facility/jsonmerger"
	"github.com/graniticio/granitic/logging"
	"os"
	"path"
	"strconv"
	"strings"
)

const (
	cdfLocationFlagName string = "c"
	cdfLocationDefault  string = "resource/components"
	cdfLocationHelp     string = "A comma separated list of config files or directories containing component definition files"

	ofLocationFlagName string = "o"
	ofLocationDefault  string = "bindings/bindings.go"
	ofLocationHelp     string = "Path of the Go source file that will be generated"

	llFlagName string = "l"
	llDefault  string = "ERROR"
	llHelp     string = "Minimum importance of logging to be displayed (TRACE, DEBUG, INFO, WARN, ERROR, FATAL)"

	nameField = "name"
	typeField = "type"

	deferSeparator = ":"
	refPrefix      = "ref"
	refAlias       = "r"
	confPrefix     = "conf"
	confAlias      = "c"
)

func main() {

	var cdf = flag.String(cdfLocationFlagName, cdfLocationDefault, cdfLocationHelp)
	var of = flag.String(ofLocationFlagName, ofLocationDefault, ofLocationHelp)
	var ll = flag.String(llFlagName, llDefault, llHelp)

	flag.Parse()

	command := new(CreateBindingsCommand)

	expandedFileList, err := config.ExpandToFiles(splitConfigPaths((*cdf)))

	if err != nil {
		fmt.Println("Unable to expand " + *cdf + " to a list of config files: " + err.Error())
		os.Exit(-1)
	}

	command.ComponentDefinitions = expandedFileList
	command.OutputFile = *of
	command.LogLevel = *ll

	command.Execute()

}

func splitConfigPaths(pathArgument string) []string {
	return strings.Split(pathArgument, ",")
}

type CreateBindingsCommand struct {
	logger               logging.Logger
	OutputFile           string
	ComponentDefinitions []string
	LogLevel             string
}

func (cbc *CreateBindingsCommand) Execute() int {

	cbc.configureLogging()

	jsonMerger := new(jsonmerger.JsonMerger)
	jsonMerger.Logger = cbc.logger

	mergedConfig := jsonMerger.LoadAndMergeConfig(cbc.ComponentDefinitions)

	configAccessor := config.ConfigAccessor{mergedConfig, nil}

	cbc.writeBindingsSource(cbc.OutputFile, &configAccessor)

	return 0
}

func (cbc *CreateBindingsCommand) configureLogging() {
	logLevel := logging.LogLevelFromLabel(cbc.LogLevel)

	logger := new(logging.LevelAwareLogger)

	logger.SetThreshold(logLevel)
	logger.SetLoggerName("")
	cbc.logger = logger
}

func (cbc *CreateBindingsCommand) writeBindingsSource(outPath string, configAccessor *config.ConfigAccessor) {

	cbc.logger.LogInfof("Writing binding file %s", outPath)

	os.MkdirAll(path.Dir(outPath), 0777)
	file, err := os.Create(outPath)

	if err != nil {
		cbc.logger.LogFatalf(err.(*os.PathError).Error())
		os.Exit(-1)
	}

	defer file.Close()

	writer := bufio.NewWriter(file)
	cbc.writeImports(writer, configAccessor)

	components := configAccessor.ObjectVal("components")
	componentCount := len(components)

	writer.WriteString("func Components() []*ioc.ProtoComponent {\n")

	writer.WriteString("\tpc := make([]*ioc.ProtoComponent, ")
	writer.WriteString(strconv.Itoa(componentCount))
	writer.WriteString(")\n")

	index := 0

	for name, componentJson := range components {
		component := componentJson.(map[string]interface{})

		writer.WriteString("\n\t//")
		writer.WriteString(name)
		writer.WriteString("\n")

		instanceVariableName := cbc.writeInstance(writer, configAccessor, component, name)
		componentProtoName := cbc.writeComponentWrapper(writer, configAccessor, component, name, index, instanceVariableName)

		for fieldName, fieldContents := range component {

			if !cbc.reservedFieldName(fieldName) {

				switch config.JsonType(fieldContents) {
				case config.JsonMap:
					cbc.writeMapValue(writer, instanceVariableName, fieldName, fieldContents.(map[string]interface{}))
				case config.JsonString:
					cbc.writeStringValue(writer, instanceVariableName, fieldName, fieldContents.(string), componentProtoName)
				case config.JsonBool:
					cbc.writeBoolValue(writer, instanceVariableName, fieldName, fieldContents.(bool), componentProtoName)
				case config.JsonArray:
					cbc.writeArray(writer, instanceVariableName, fieldName, fieldContents.([]interface{}), componentProtoName)
				case config.JsonUnknown:

					switch t := fieldContents.(type) {
					default:
						cbc.logger.LogErrorf("Unknown JSON type for field %s on component %s %T", fieldName, name, t)
						os.Exit(-1)
					}
				}

			}

		}

		writer.WriteString("\n\n")
		index = index + 1
	}

	writer.WriteString("\treturn pc\n")
	writer.WriteString("}\n")
	writer.Flush()
}

func (cbc *CreateBindingsCommand) writeArray(writer *bufio.Writer, instanceName string, fieldName string, fieldContents []interface{}, componentProtoName string) {

	cbc.writeAssignmentPrefix(writer, instanceName, fieldName)

	if cbc.allStrings(fieldContents) {
		cbc.writeStringArrayContents(writer, fieldContents)
	} else {
		cbc.logger.LogErrorf("Unsupported data types in array %#v", fieldContents)
		os.Exit(-1)
	}

}

func (cbc *CreateBindingsCommand) writeStringArrayContents(w *bufio.Writer, a []interface{}) {

	w.WriteString(" []string{")

	last := len(a) - 1

	for i, s := range a {

		w.WriteString(fmt.Sprintf("%q", s.(string)))

		if i != last {
			w.WriteString(", ")
		}

	}

	w.WriteString("}\n")

}

func (cbc *CreateBindingsCommand) allStrings(a []interface{}) bool {

	for _, i := range a {

		switch i.(type) {
		default:
			return false
		case string:
			continue
		}

	}

	return true
}

func (cbc *CreateBindingsCommand) writeBoolValue(writer *bufio.Writer, instanceName string, fieldName string, fieldContents bool, componentProtoName string) {

	cbc.writeAssignmentPrefix(writer, instanceName, fieldName)
	writer.WriteString(strconv.FormatBool(fieldContents))
	writer.WriteString("\n")

}

func (cbc *CreateBindingsCommand) writeStringValue(writer *bufio.Writer, instanceName string, fieldName string, fieldContents string, componentProtoName string) {

	valueElements := strings.SplitN(fieldContents, deferSeparator, 2)

	if len(valueElements) == 2 {

		prefix := valueElements[0]
		instruction := valueElements[1]

		if prefix == refPrefix || prefix == refAlias {

			writer.WriteString("\t")
			writer.WriteString(componentProtoName)
			writer.WriteString(".AddDependency(\"")
			writer.WriteString(fieldName)
			writer.WriteString("\", \"")
			writer.WriteString(instruction)
			writer.WriteString("\")\n")

			return

		} else if prefix == confPrefix || prefix == confAlias {

			writer.WriteString("\t")
			writer.WriteString(componentProtoName)
			writer.WriteString(".AddConfigPromise(\"")
			writer.WriteString(fieldName)
			writer.WriteString("\", \"")
			writer.WriteString(instruction)
			writer.WriteString("\")\n")

			return
		}
	}

	cbc.writeAssignmentPrefix(writer, instanceName, fieldName)
	writer.WriteString(fmt.Sprintf("%q", fieldContents))
	writer.WriteString("\n")
}

func (cbc *CreateBindingsCommand) writeMapValue(writer *bufio.Writer, instanceName string, fieldName string, fieldContents map[string]interface{}) {

	directField := instanceName + "." + fieldName

	writer.WriteString(directField)
	writer.WriteString(" = make(map[string]string)\n")

	for key, value := range fieldContents {
		writer.WriteString("\t")
		writer.WriteString(directField)
		writer.WriteString("[\"")
		writer.WriteString(key)
		writer.WriteString("\"] = \"")
		writer.WriteString(value.(string))
		writer.WriteString("\"\n")
	}

}

func (cbc *CreateBindingsCommand) writeAssignmentPrefix(writer *bufio.Writer, instanceName string, fieldName string) {

	writer.WriteString("\t")
	writer.WriteString(instanceName)
	writer.WriteString(".")
	writer.WriteString(fieldName)
	writer.WriteString(" = ")

}

func (cbc *CreateBindingsCommand) reservedFieldName(field string) bool {
	return field == nameField || field == typeField
}

func (cbc *CreateBindingsCommand) writeComponentWrapper(writer *bufio.Writer, configAccessor *config.ConfigAccessor, component map[string]interface{}, name string, index int, instanceName string) string {

	componentProtoName := name + "Proto"

	writer.WriteString(componentProtoName)
	writer.WriteString(" := ioc.CreateProtoComponent(")
	writer.WriteString(instanceName)
	writer.WriteString(", \"")
	writer.WriteString(name)
	writer.WriteString("\")\n\t")
	writer.WriteString("pc[")
	writer.WriteString(strconv.Itoa(index))
	writer.WriteString("] = ")
	writer.WriteString(componentProtoName)
	writer.WriteString("\n\t")
	writer.WriteString(componentProtoName)
	writer.WriteString(".Component.Name = \"")
	writer.WriteString(name)
	writer.WriteString("\"\n")

	return componentProtoName
}

func (cbc *CreateBindingsCommand) writeInstance(writer *bufio.Writer, configAccessor *config.ConfigAccessor, component map[string]interface{}, name string) string {
	instanceType := configAccessor.StringFieldVal("type", component)
	instanceName := name + "Instance"
	writer.WriteString("\t")
	writer.WriteString(instanceName)
	writer.WriteString(" := new(")
	writer.WriteString(instanceType)
	writer.WriteString(")\n\t")

	return instanceName
}

func (cbc *CreateBindingsCommand) writeImports(writer *bufio.Writer, configAccessor *config.ConfigAccessor) {
	packages := configAccessor.Array("packages")

	writer.WriteString("package bindings\n\n")
	writer.WriteString("import (\n")

	for _, packageName := range packages {
		writer.WriteString("\t\"")
		writer.WriteString(packageName.(string))
		writer.WriteString("\"\n")
	}

	writer.WriteString("\t\"github.com/graniticio/granitic/ioc\"")
	writer.WriteString("\n)\n\n")
}
