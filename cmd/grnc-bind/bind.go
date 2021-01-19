// Copyright 2016-2021 Granitic. All rights reserved.
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

The results of that function are then included in your application's call to start Granitic. E.g.

	func main() {
		granitic.StartGranitic(bindings.Components())
	}

grnc-bind will need to be re-run whenever a component definition file is modified.

Usage of grnc-bind:

	grnc-bind [-c component-files] [-m merged-file-out] [-o generated-file] [-l log-level]

	-c string
		A comma separated list of component definition files or directories containing component definition files (default "resource/components")
	-m string
		The path of a file where the merged component definition file should be written to. Execution will halt after writing.
	-o string
		Path to the Go source file that will be generated (default "bindings/bindings.go")
	-l string
		The level at which the tool will output messages: TRACE, DEBUG, INFO, ERROR, FATAL (default ERROR)
    -x
		Enable the discovery of external facilities in Go modules referenced in this application's go.mod file
    -i string
		A comma separated list of module paths (e.g. github.com/someaccount/somemod/someversion) that should be ignored when looking for external facilities


*/
package main

import (
	"fmt"
	"github.com/graniticio/granitic/v2/cmd/grnc-bind/binder"
	"github.com/graniticio/granitic/v2/config"
	"github.com/graniticio/granitic/v2/logging"
	"os"
)

func main() {

	b := new(binder.Binder)
	b.ToolName = "grnc-bind"
	b.Loader = binder.NewJsonDefinitionLoader()
	b.SupportedExtensions = new(config.JSONContentParser).Extensions()

	s, err := binder.SettingsFromArgs()

	if err != nil {
		fmt.Printf("%s: %s\n", b.ToolName, err.Error())
		os.Exit(1)
	}

	pref := fmt.Sprintf("%s: ", b.ToolName)
	b.Log = logging.NewStdoutLogger(s.LogLevel, pref)

	b.Bind(s)

}
