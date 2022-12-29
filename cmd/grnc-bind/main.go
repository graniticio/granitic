// Copyright 2016-2023 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

/*
The grnc-bind tool - used to convert Granitic's component definition files and default configuration values into Go source.

Go does not support a 'type-from-name' mechanism for instantiating objects, so the container cannot create arbitrarily typed
objects at runtime. Instead, Granitic component definition files are used to generate Go source files that will be
compiled along with your application. The grnc-bind tool performs this code generation.

It also gathers together any default configuration values you have specified for your application and converts them into
Go source to be included in your application's executable.

In most cases, the grnc-bind command will be run, without arguments, in your application's root directory (the same directory
that contains your comp-def and config directories).

It will then look for component definitions and default configuration values in the following places:

  * Your Granitic installation (for built-in Granitic facilities)
  * Any dependencies from your go.mod.old file that define valid Granitic 'external facilities'
  * Your project's comp-def and config directories


The tool will merge together any .yml, .yaml and .json files containing valid component definitions/ configuration in those locations
and create a file bindings/bindings.go in your project. This file includes a single function:

	Components() *ioc.ProtoComponents

The results of that function are then included in your application's call to start Granitic. E.g.

	func main() {
		granitic.StartGranitic(bindings.Components())
	}

grnc-bind will need to be re-run whenever a component definition file is modified.

Usage of grnc-bind:

	grnc-bind [-c component-files] [-m merged-file-out] [-o generated-file] [-l log-level]

	-c string
		A comma separated list of additional component definition files or directories containing component definition files
	-m string
		The path of a file where the merged component definition file should be written to (for debugging). Execution will halt after writing.
	-o string
		Path to the Go source file that will be generated (default "bindings/bindings.go")
	-l string
		The level at which the tool will output messages: TRACE, DEBUG, INFO, ERROR, FATAL (default ERROR)
	-g string
		A path to a specific installation of Granitic that you want to use instead of automatically finding one.
*/

package main

import (
	"fmt"
	"github.com/graniticio/granitic/v3/logging"
	"os"
)

const (
	toolName = "grnc-bind"
)

func main() {

	//var s settings
	var err error

	pref := fmt.Sprintf("%s: ", toolName)
	l := logging.NewStdoutLogger(logging.Error, pref)

	// Obtain settings from command line arguments
	if _, err = settingsFromArgs(); err != nil {
		exitWithError(l, err.Error())
	}

	// Locate a valid Granitic installation
	// Find go.mod.old dependencies with valid external facilities definitions
	// Gather configuration files
	// Gather component definition files
	// Merge configuration and encrypt to disk
	// Merge component definition files
	// Generate Go source

}

func exitWithError(l logging.Logger, message string) {
	l.LogFatalf(message)
	os.Exit(1)
}
