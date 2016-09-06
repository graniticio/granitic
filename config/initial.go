package config

import (
	"flag"
	"fmt"
	"github.com/graniticio/granitic/logging"
	"os"
	"strings"
	"time"
)

const (
	builtInConfigPath  = "/resource/facility-config"
	graniticHomeEnvVar = "GRANITIC_HOME"
)

func GraniticHome() string {
	return os.Getenv(graniticHomeEnvVar)
}

func checkForGraniticHome() {

	gh := GraniticHome()

	if gh == "" {
		fmt.Printf("%s environment variable is not set.\n")
		ExitError()
	}

	if strings.HasSuffix(gh, "/") || strings.HasSuffix(gh, "\\") {
		fmt.Printf("%s environment variable should not end with a / or \\.\n")
		ExitError()
	}

}

type InitialSettings struct {
	FrameworkLogLevel logging.LogLevel
	Configuration     []string
	GraniticHome      string
	StartTime         time.Time
}

func InitialSettingsFromEnvironment() *InitialSettings {

	start := time.Now()
	checkForGraniticHome()

	is := new(InitialSettings)
	is.StartTime = start
	is.GraniticHome = GraniticHome()
	is.Configuration = builtInConfigFiles()

	processCommandLineArgs(is)

	return is

}

func processCommandLineArgs(is *InitialSettings) {
	configFilePtr := flag.String("c", "resource/config", "Path to container configuration files")
	startupLogLevel := flag.String("l", "INFO", "Logging threshold for messages from components during bootstrap")
	flag.Parse()

	ll, err := logging.LogLevelFromLabel(*startupLogLevel)

	if err != nil {
		fmt.Println(err)
		ExitError()
	}

	paths := strings.Split(*configFilePtr, ",")
	userConfig, err := ExpandToFiles(paths)

	if err != nil {
		fmt.Println(err)
		ExitError()
	}

	is.Configuration = append(is.Configuration, userConfig...)
	is.FrameworkLogLevel = ll

}

func ExpandToFiles(paths []string) ([]string, error) {
	files := make([]string, 0)

	for _, path := range paths {

		expanded, err := FileListFromPath(path)

		if err != nil {
			return nil, err
		}

		files = append(files, expanded...)

	}

	return files, nil
}

func builtInConfigFiles() []string {

	d := GraniticHome() + builtInConfigPath

	files, err := FindConfigFilesInDir(d)

	if err != nil {

		fmt.Printf("Problem loading Grantic's built-in configuration from %s:\n", d)
		fmt.Println(err.Error())
		ExitError()

	}

	return files

}

func ExitError() {
	os.Exit(1)
}

func ExitNormal() {
	os.Exit(0)
}
