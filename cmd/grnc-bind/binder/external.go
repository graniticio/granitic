// Copyright 2021 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package binder

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

// ExternalFacilities holds information about the code and config defined in Go module
// dependencies that should be compiled into this application.
type ExternalFacilities struct {
	Info         []*ExternalFacility
	TempConfFile string
	TempCompFile string
}

// Found returns true if at least one valid external facility has been found
func (ex *ExternalFacilities) Found() bool {

	return len(ex.Info) > 0

}

type ExternalFacility struct {
	Namespace     string
	ModuleName    string
	ModulePath    string
	ModuleVersion string
	Manifest      *Manifest
	Components    string
	Config        string
}

// Manifest is the structure into which an external facility's manifest file is parsed into
type Manifest struct {
	Namespace  string
	Facilities map[string]*definition
}

type definition struct {
	Disabled bool
	Depends  []string
	Builder  string
}

// WriteTempFacilityFiles creates a temporary component and config file containing the builders and default states for
// external facilities
func WriteTempFacilityFiles(facilities *ExternalFacilities) (err error) {

	var compFile, confFile *os.File

	if compFile, err = ioutil.TempFile("", "comp.*.json"); err != nil {
		return fmt.Errorf("unable to create a temporary component file: %s", err.Error())
	}

	if confFile, err = ioutil.TempFile("", "conf.*.json"); err != nil {
		return fmt.Errorf("unable to create a temporary config file: %s", err.Error())
	}

	defer compFile.Close()
	defer confFile.Close()

	confMap := make(map[string]interface{})
	facilityStatus := make(map[string]interface{})

	confMap["ExternalFacilities"] = facilityStatus

	for _, ef := range facilities.Info {

		nsm := facilityStatus[ef.Namespace]

		if nsm == nil {
			facilityStatus[ef.Namespace] = make(map[string]bool)

		}

		tsm := facilityStatus[ef.Namespace].(map[string]bool)

		for fn, fm := range ef.Manifest.Facilities {

			tsm[fn] = !fm.Disabled

		}

	}

	fw := bufio.NewWriter(confFile)
	cfWriter := json.NewEncoder(fw)

	if err := cfWriter.Encode(confMap); err != nil {
		return fmt.Errorf("unable to write temporary config file to %s: %s", confFile.Name(), err.Error())
	} else {
		fw.Flush()
	}

	fmt.Println(confFile.Name())

	facilities.TempCompFile = compFile.Name()
	facilities.TempConfFile = confFile.Name()

	return nil
}

type confMap map[string]interface{}

func (cm confMap) addExternal(ns, name string, state bool) {

	ex := cm["External"]

	if ex == nil {
		ex = make(map[string]interface{})
		cm["External"] = ex
	}

	ans := ex.(map[string]interface{})[ns]

	if ans == nil {
		ans = make(map[string]interface{})
		ex.(map[string]interface{})[ns] = ans
	}

}

func validateManifest(m *Manifest) error {

	ns := m.Namespace

	if len(ns) == 0 {
		return fmt.Errorf("manifests must specify a Namespace")
	}

	if len(ns) != len(strings.TrimSpace(ns)) {

		return fmt.Errorf("the Namespace field in a manifest must not have leading or trailing whitespace (was [%s])", ns)
	}

	if !validGoName(ns) {
		return fmt.Errorf("the Namespace should only contain letters and numbers and must start with a letter (was [%s])", ns)
	}

	return validateDefinitions(m)

}

func validateDefinitions(m *Manifest) error {

	if len(m.Facilities) == 0 {
		return nil
	}

	for name, def := range m.Facilities {

		if !validGoName(name) {
			return fmt.Errorf("the external facility named %s is not a valid facility name (should follow same rules as a Go identifier", name)
		}

		if len(def.Builder) > 0 {

			if len(strings.TrimSpace(def.Builder)) == 0 {
				return fmt.Errorf("the Builder field for a facility definition is empty once trimmed")
			}
		}

		for _, dep := range def.Depends {

			if err := validDependency(dep); err != nil {
				return err
			}

		}

	}

	return nil
}

func validDependency(d string) error {
	if len(strings.TrimSpace(d)) == 0 {
		return fmt.Errorf("a dependency in a facility defition is emtpy once trimmed")
	}

	np := strings.Split(d, ".")

	if len(np) > 2 {
		return fmt.Errorf("%s is not a valid dependency. Should be [Namespace.Dependency] or [Dependency]", d)
	}

	for _, dp := range np {

		if !validGoName(dp) {
			return fmt.Errorf("%s is not a valid dependency. %s is not a valid Go identifier", d, dp)
		}

	}

	return nil

}
