// Copyright 2016-2019 Granitic. All rights reserved.
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
	"encoding/json"
	"github.com/graniticio/granitic/v2/cmd/grnc-bind/binder"
	"github.com/graniticio/granitic/v2/config"
	"github.com/graniticio/granitic/v2/logging"
	"io/ioutil"
)

func main() {

	b := new(binder.Binder)
	b.ToolName = "grnc-bind"
	b.Loader = new(jsonDefinitionLoader)

	s := binder.SettingsFromArgs()

	b.Bind(s)

}

// Loads JSON files from local files and remote URLs and provides a mechanism for writing the resulting merged
// file to disk
type jsonDefinitionLoader struct {
}

// LoadAndMerge reads one or more JSON from local files or HTTP URLs and merges them into a single data structure
func (jdl *jsonDefinitionLoader) LoadAndMerge(files []string) (map[string]interface{}, error) {
	jm := config.NewJSONMergerWithDirectLogging(new(logging.ConsoleErrorLogger), new(config.JSONContentParser))
	jm.MergeArrays = true

	return jm.LoadAndMergeConfig(files)
}

// WriteMerged converts the supplied data structure to JSON and writes to disk at the specified location
func (jdl *jsonDefinitionLoader) WriteMerged(data map[string]interface{}, path string) error {

	b, err := json.MarshalIndent(data, "", "\t")

	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path, b, 0644)

	if err != nil {
		return err
	}

	return nil
}
