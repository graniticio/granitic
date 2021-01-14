// Copyright 2021 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package binder

// ExternalFacilities holds information about the code and config defined in Go module
// dependencies that should be compiled into this application.
type ExternalFacilities struct {
	Info []*ExternalFacility
}

type ExternalFacility struct {
	ModulePath    string
	ModuleVersion string
	Manifest      string
	Components    string
	Config        string
}
