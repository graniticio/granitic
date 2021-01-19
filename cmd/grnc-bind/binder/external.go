// Copyright 2021 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package binder

// ExternalFacilities holds information about the code and config defined in Go module
// dependencies that should be compiled into this application.
type ExternalFacilities struct {
	Info []*ExternalFacility
}

type ExternalFacility struct {
	Name          string
	ModulePath    string
	ModuleVersion string
	Manifest      string
	Components    string
	Config        string
}

// Manifest is the structure into which an external facility's manifest file is parsed into
type Manifest struct {
	Namespace          string
	ExternalFacilities map[string]*definition
}

type definition struct {
	Namespace string
}
