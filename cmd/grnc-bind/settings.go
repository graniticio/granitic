// Copyright 2016-2022 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/graniticio/granitic/v3/logging"
	"log"
	"os"
	"os/exec"
	"strings"
)

const (
	compLocationFlag    string = "c"
	compLocationDefault string = ""
	compLocationHelp    string = "A comma separated list of component definition files or directories containing component definition files"

	bindingsFileFlag    string = "o"
	bindingsFileDefault string = "bindings/bindings.go"
	bindingsFileHelp    string = "Path to the Go source file that will be generated to instatiate your components"

	mergeLocationFlag    string = "m"
	mergeLocationDefault string = ""
	mergeLocationHelp    string = "The path of a file where the merged component definition file should be written to. Execution will halt after writing."

	logLevelFlag    string = "l"
	logLevelDefault string = "WARN"
	logLevelHelp    string = "The level at which messages will be logged to the console (TRACE, DEBUG, WARN, INFO, ERROR, FATAL)"
)

// settings contains output/input file locations and other variables for controlling the behaviour of this tool
type settings struct {
	CompDefLocation *string
	BindingsFile    *string
	MergedDebugFile *string
	LogLevelLabel   *string
	LogLevel        logging.LogLevel
}

// settingsFromArgs uses CLI parameters to populate a settings object
func settingsFromArgs() (settings, error) {

	s := settings{}

	s.CompDefLocation = flag.String(compLocationFlag, compLocationDefault, compLocationHelp)
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

/*
func locateGraniticInstallation(log logging.Logger, s settings) (string, error) {

	var graniticPath string
	var err error

	notFound := fmt.Errorf("unable to find a Granitic installation - checked go.mod.old file, %s environment variable and standard checkout location under %s", graniticHomeEnvVar, goPathEnvVar)

	log.LogDebugf("Looking for an installation of Granitic")

	if modPath, okay := installationFromModule(log); okay {

		// If the project has a go.mod.old file, try and work out where Granitic
		log.LogDebugf("Using location from go.mod.old file")

		graniticPath = modPath
	} else if envPath := os.Getenv(graniticHomeEnvVar); envPath != "" {
		// See if the GRANITIC_HOME environment variable is set and points to a valid install
		log.LogDebugf("Using location from %s environment variable", graniticHomeEnvVar)

		graniticPath = envPath

	} else if goPath := os.Getenv(goPathEnvVar); goPath != "" {
		// See if Granitic is checked out in a standard location under GOPATH
		possiblePath := filepath.Join(goPath, "src", "github.com", "graniticio", "granitic")

		if _, readErr := os.ReadDir(possiblePath); readErr == nil {

			log.LogDebugf("Using standard checkout location under %s environment variable", goPathEnvVar)

			graniticPath = possiblePath
		} else {
			err = notFound
		}
	} else {
		err = notFound
	}

	//For developers of Granitic or users running tests of their Granitic installation, check if we're inside the Granitic folders
	if err != nil {

		if wd, fileErr := os.Getwd(); fileErr == nil {

			fmt.Println(wd)

			if pos := strings.LastIndex(wd, pathSuffix); pos > 0 {

				graniticPath = wd[:pos+len(pathSuffix)]

				err = nil

			}
		}
	}

	if graniticPath != "" {
		if confPath, okay := validateInstallation(graniticPath); okay {
			graniticPath = confPath
		} else {
			err = fmt.Errorf("%s does not seem to contain a valid Granitic installation. Check your %s and/or %s environment variables", graniticPath, graniticHomeEnvVar, goPathEnvVar)
		}
	}

	log.LogDebugf("Using Granitic facility configuration files at %s", graniticPath)

	return graniticPath, err

}

func validateInstallation(path string) (string, bool) {
	resourcePath := filepath.Join(path, "facility", "config")

	if _, err := config.FindJSONFilesInDir(resourcePath); err != nil {
		return path, false
	}

	return resourcePath, true
}
*/

func findGoModCache() (string, error) {

	goExec, err := exec.LookPath("go")

	if err != nil {
		return "", fmt.Errorf("could not find go on your path. Make sure it is available in your OS PATH environment variable")
	}

	cmd := exec.Command(goExec, "env", "GOMODCACHE")

	stdout, err := cmd.StdoutPipe()

	if err != nil {
		return "", err
	}

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(stdout)

	path := buf.String()
	path = strings.TrimSpace(path)

	_, err = os.ReadDir(path)

	if err != nil {
		return "", fmt.Errorf("Unable to open GOMODCACHE directory: %s", err.Error())
	}

	return path, nil

}
