package binder

import (
	"bufio"
	"fmt"
	"github.com/graniticio/granitic/v2/config"
	"github.com/graniticio/granitic/v2/logging"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	graniticHomeEnvVar = "GRANITIC_HOME"
	goPathEnvVar       = "GOPATH"
	reqRegex           = ".*github.com/graniticio/granitic/(v[\\d]*)[\\s]*(\\S*)"
	replaceRegex       = ".*github.com/graniticio/granitic/v[\\d]*[\\s]*=>[\\s]*(\\S*)"
)

// LocateFacilityConfig determines where on your filesystem you have checked out Granitic. This is used when code needs
// access to the configuration for Granitic's built-in facility components which is stored under resource/facility-config
func LocateFacilityConfig(log logging.Logger) (string, error) {

	var graniticPath string
	var err error

	notFound := fmt.Errorf("unable to find a Granitic installation - checked go.mod file, %s environment variable and standard checkout location under %s", graniticHomeEnvVar, goPathEnvVar)

	log.LogDebugf("Looking for an installation of Granitic")

	if modPath, okay := installationFromModule(log); okay {

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
	resourcePath := filepath.Join(path, "facility", "config")

	if _, err := config.FindJSONFilesInDir(resourcePath); err != nil {
		return path, false
	}

	return resourcePath, true
}

func installationFromModule(log logging.Logger) (string, bool) {

	requiredVersion := ""
	replacePath := ""
	majorVersion := ""

	var f *os.File
	var err error

	if f, err = os.Open("go.mod"); err != nil { // os.Open defaults to read only

		fmt.Println("Failed to open")
		return "", false
	}

	defer f.Close()

	reqRe := regexp.MustCompile(reqRegex)
	repRe := regexp.MustCompile(replaceRegex)

	s := bufio.NewScanner(f)

	for s.Scan() {
		line := s.Text()

		reqMatches := reqRe.FindStringSubmatch(line)

		if len(reqMatches) >= 3 {
			majorVersion = reqMatches[1]
			requiredVersion = reqMatches[2]
		}

		repMatches := repRe.FindStringSubmatch(line)

		if len(repMatches) >= 2 {
			replacePath = repMatches[1]
			break
		}

	}

	if replacePath != "" {
		log.LogDebugf("Found a replace path for Granitic in go.mod: %s", replacePath)
		return replacePath, true
	}

	requiredVersion = strings.TrimSpace(requiredVersion)
	majorVersion = strings.TrimSpace(majorVersion)

	if requiredVersion != "" {

		fullVersion := fmt.Sprintf("%s@%s", majorVersion, requiredVersion)

		log.LogDebugf("Found a required version for Granitic in go.mod: %s", fullVersion)

		if goPath := os.Getenv(goPathEnvVar); goPath != "" {

			modPath := filepath.Join(goPath, "pkg", "mod", "github.com", "graniticio", "granitic", fullVersion)

			log.LogDebugf("Looking for downloaded Granitic module at %s", modPath)

			if _, err := os.Stat(modPath); !os.IsNotExist(err) {
				return modPath, true
			}

			log.LogWarnf("Expected to find a downloaded version of Granitic at %s - make sure you have run 'go mod download'", modPath)

		} else {

			log.LogWarnf("GOPATH not set - unable to try and find downloaded modules")

		}

	}

	return "", false

}
