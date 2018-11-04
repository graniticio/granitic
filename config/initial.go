// Copyright 2016-2018 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package config

import (
	"flag"
	"fmt"
	"github.com/graniticio/granitic/instance"
	"github.com/graniticio/granitic/logging"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	graniticHomeEnvVar = "GRANITIC_HOME"
	goPathEnvVar       = "GOPATH"
)

// Returns the value of the OS environment variable GRANITIC_HOME. This is expected to be the path
// (without a trailing /) where Grantic has been installed (e.g. /home/youruser/go/src/github.com/graniticio/granitic).
// This is required so that configuration files for built-in facilities can be loaded from the resource sub-directory
// of the Granitic installation.
func GraniticHome() string {

	//Check if GRANITIC_HOME explicitly set
	graniticPath := os.Getenv(graniticHomeEnvVar)

	if graniticPath == "" {

		if gopath := os.Getenv(goPathEnvVar); gopath == "" {

			fmt.Printf("Neither %s or %s environment variable is not set. Cannot find Granitic\n", graniticHomeEnvVar, goPathEnvVar)
			instance.ExitError()

		} else {

			graniticPath = filepath.Join(gopath, "src", "github.com", "graniticio", "granitic")

			if _, err := ioutil.ReadDir(graniticPath); err != nil {

				fmt.Printf("%s environment variable is not set and cannot find Granitic in the default install path of %s (your GOPATH variable is set to %s)\n", graniticHomeEnvVar, graniticPath, gopath)
				instance.ExitError()
			}

		}

	}

	resourcePath := filepath.Join(graniticPath, "resource", "facility-config")

	if _, err := FindConfigFilesInDir(resourcePath); err != nil {
		fmt.Printf("%s does not seem to contain a valid Granitic installation. Check your %s and/or %s environment variables\n", graniticPath, graniticHomeEnvVar, goPathEnvVar)
		instance.ExitError()
	}

	return graniticPath

}

// InitialSettings contains settings that are needed by Granitic before configuration files can be loaded and parsed. See package
// granitic for more information on how this is used when starting a Granitic application with command line arguments
// or supplying an instance of this struct programmatically.
type InitialSettings struct {

	// The level at which messages from Granitic components used before configuration files can be loaded (the initiator,
	// logger, IoC container, JSON merger etc.) will be logged. As soon as configuration has been loaded, this value will
	// be discarded in favour of LogLevels defined in that configuration.
	FrameworkLogLevel logging.LogLevel

	// Files, directories and URLs from which JSON configuration should be loaded and merged.
	Configuration []string

	// The time at which the application was started (to allow accurate timing of the IoC container start process).
	StartTime time.Time

	// An (optional) unique identifier for this instance of a Granitic application.
	InstanceId string

	// A base 64 serialised version of Granitic's built-in configuration files
	BuiltInConfig *string

	// Additional parsers to support config files in a format other than JSON
	ConfigParsers []ContentParser
}

// InitialSettingsFromEnvironment builds an InitialSettings and populates it with defaults or the values of command line
// arguments if available.
func InitialSettingsFromEnvironment() *InitialSettings {

	start := time.Now()

	is := new(InitialSettings)
	is.StartTime = start
	is.Configuration = make([]string, 0)

	processCommandLineArgs(is)

	return is

}

func processCommandLineArgs(is *InitialSettings) {
	configFilePtr := flag.String("c", "resource/config", "Path to container configuration files")
	startupLogLevel := flag.String("l", "INFO", "Logging threshold for messages from components during bootstrap")
	instanceId := flag.String("i", "", "A unique identifier for this instance of the application")
	flag.Parse()

	ll, err := logging.LogLevelFromLabel(*startupLogLevel)

	if err != nil {
		fmt.Println(err)
		instance.ExitError()
	}

	paths := strings.Split(*configFilePtr, ",")
	userConfig, err := ExpandToFilesAndURLs(paths)

	if err != nil {
		fmt.Println(err)
		instance.ExitError()
	}

	is.Configuration = append(is.Configuration, userConfig...)
	is.FrameworkLogLevel = ll
	is.InstanceId = *instanceId

}

// ExpandToFilesAndURLs takes a slice that may be a mixture of URLs, file paths
// and directory paths and returns a list of URLs and file paths, recursively expanding any non-empty
// directories into a list of files. Returns an error if there is a problem traversing directories of if any of the
// supplied file paths does not exist.
func ExpandToFilesAndURLs(paths []string) ([]string, error) {
	files := make([]string, 0)

	for _, path := range paths {

		if isURL(path) {
			files = append(files, path)
			continue
		}

		expanded, err := FileListFromPath(path)

		if err != nil {
			return nil, err
		}

		files = append(files, expanded...)

	}

	return files, nil
}

func isURL(u string) bool {
	return strings.HasPrefix(u, "http:") || strings.HasPrefix(u, "https:")
}
