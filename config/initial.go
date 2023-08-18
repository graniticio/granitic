// Copyright 2016-2023 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package config

import (
	"flag"
	"fmt"
	"github.com/graniticio/granitic/v2/instance"
	"github.com/graniticio/granitic/v2/logging"
	"os"
	"path/filepath"
	"strings"
	"time"
)

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
	InstanceID string

	// A base 64 serialised version of Granitic's built-in configuration files
	BuiltInConfig *string

	// Additional parsers to support config files in a format other than JSON
	ConfigParsers []ContentParser

	// Exit immediately after container has successfully started
	DryRun bool

	// Buffer messages and only log them when application logging has been configured
	DeferBootstrapLogging bool

	// A path to write the merged view of configuration to (then exit)
	MergedConfigPath string

	// Create a UUID and use it as the instance ID
	GenerateInstanceUUID bool
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

	v1ConfLocation := filepath.Join("resource", "config")
	defaultConfLocation := "config"

	configFilePtr := flag.String("c", defaultConfLocation, "Path to application configuration files")
	startupLogLevel := flag.String("l", "INFO", "Logging threshold for messages from components during bootstrap")
	instanceID := flag.String("i", "", "A unique identifier for this instance of the application")
	deferLogging := flag.Bool("d", false, "Defer logging messages from until application logging is configured")
	mergeFile := flag.String("m", "", "Path to a file to write merged view of config then exit")
	uuidInstanceID := flag.Bool("u", false, "Use a generated UUID as the instance ID for this application")

	flag.Parse()

	// If the default location for config is set, but doesn't exist, check to see if the Granitic v1 folder exists instead
	if *configFilePtr == defaultConfLocation {

		if !folderExists(defaultConfLocation) && folderExists(v1ConfLocation) {

			configFilePtr = &v1ConfLocation
		}
	}

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
	is.InstanceID = *instanceID
	is.DeferBootstrapLogging = *deferLogging
	is.MergedConfigPath = *mergeFile
	is.GenerateInstanceUUID = *uuidInstanceID

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

func folderExists(path string) bool {
	s, err := os.Stat(path)
	if err == nil {
		return s.IsDir()
	}
	return false
}
