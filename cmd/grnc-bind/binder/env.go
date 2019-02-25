package binder

import (
	"fmt"
	"github.com/graniticio/granitic/v2/config"
	"github.com/graniticio/granitic/v2/logging"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	graniticHomeEnvVar = "GRANITIC_HOME"
	goPathEnvVar       = "GOPATH"
)

// LocateFacilityConfig determines where on your filesystem you have checked out Granitic. This is used when code needs
// access to the configuration for Granitic's built-in facility components which is stored under resource/facility-config
func LocateFacilityConfig(log logging.Logger) (string, error) {

	var graniticPath string
	var err error

	notFound := fmt.Errorf("unable to find a Granitic installation - checked go.mod file, %s envrionment variable and standard checkout location under %s", graniticHomeEnvVar, goPathEnvVar)

	log.LogDebugf("Looking for an installation of Granitic")

	if modPath := installationFromModule(log); modPath != "" {

		// If the project has a go.mod file, try and work out where Granitic
		log.LogDebugf("Using location from go.mod file")

		graniticPath = modPath
	} else if envPath := os.Getenv(graniticHomeEnvVar); envPath != "" {
		// See if the GRANITIC_HOME environment variable is set and points to a valid install
		log.LogDebugf("Using location from %s environment variable", graniticHomeEnvVar)

		graniticPath = envPath

	} else if goPath := os.Getenv(goPathEnvVar); goPath != "" {
		// See if Granitic is checked out in a standard location under GOPATH
		possiblePath := filepath.Join(goPath, "src", "github.com", "graniticio", "granitic")

		if _, readErr := ioutil.ReadDir(possiblePath); readErr == nil {

			log.LogDebugf("Using standard checkout location under %s environment variable", goPathEnvVar)

			graniticPath = possiblePath
		} else {
			err = notFound
		}
	} else {
		err = notFound
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
	resourcePath := filepath.Join(path, "resource", "facility-config")

	if _, err := config.FindJSONFilesInDir(resourcePath); err != nil {
		return path, false
	}

	return resourcePath, true
}

func installationFromModule(log logging.Logger) string {

	return ""

}
