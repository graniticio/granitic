// Copyright 2016 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package config

import (
	"flag"
	"fmt"
	"github.com/graniticio/granitic/instance"
	"github.com/graniticio/granitic/logging"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	graniticHomeEnvVar = "GRANITIC_HOME"
)

// Returns the value of the OS environment variable GRANITIC_HOME. This is expected to be the path
// (without a trailing /) where Grantic has been installed (e.g. /home/youruser/go/src/github.com/graniticio/granitic).
// This is required so that configuration files for built-in facilities can be loaded from the resource sub-directory
// of the Granitic installation.
func GraniticHome() string {
	return os.Getenv(graniticHomeEnvVar)
}

func checkForGraniticHome() {

	gh := GraniticHome()

	if gh == "" {
		fmt.Printf("%s environment variable is not set.\n")
		instance.ExitError()
	}

	if strings.HasSuffix(gh, "/") || strings.HasSuffix(gh, "\\") {
		fmt.Printf("%s environment variable should not end with a / or \\.\n")
		instance.ExitError()
	}

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

	// The path (without trailing slash) where Granitic has been installed.
	GraniticHome string

	// The time at which the application was started (to allow accurate timing of the IoC container start process).
	StartTime time.Time

	// An (optional) unique identifier for this instance of a Granitic application.
	InstanceId string
}

// InitialSettingsFromEnvironment builds an InitialSettings and populates it with defaults or the values of command line
// arguments if available.
func InitialSettingsFromEnvironment() *InitialSettings {

	start := time.Now()
	checkForGraniticHome()

	is := new(InitialSettings)
	is.StartTime = start
	is.GraniticHome = GraniticHome()
	is.Configuration = BuiltInConfigFiles()

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

// BuiltInConfigFiles provides a list of file paths, each of which points to one of the default configuration files
// for Granitic's built-in facilities. This is useful when programmatically building an InitialSettings object.
func BuiltInConfigFiles() []string {

	builtInConfigPath := filepath.Join(GraniticHome(), "resource", "facility-config")

	files, err := FindConfigFilesInDir(builtInConfigPath)

	if err != nil {

		fmt.Printf("Problem loading Grantic's built-in configuration from %s:\n", builtInConfigPath)
		fmt.Println(err.Error())
		instance.ExitError()

	}

	return files

}
