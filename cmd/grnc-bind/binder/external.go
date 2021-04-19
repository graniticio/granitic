// Copyright 2021 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package binder

import (
	"fmt"
	"strings"
)

// ExternalFacilities holds information about the code and config defined in Go module
// dependencies that should be compiled into this application.
type ExternalFacilities struct {
	Info []*ExternalFacility
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
