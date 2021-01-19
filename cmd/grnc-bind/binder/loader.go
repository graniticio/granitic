// Copyright 2016-2021 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package binder

import (
	"encoding/json"
	"fmt"
	"github.com/graniticio/granitic/v2/config"
	"github.com/graniticio/granitic/v2/logging"
	"io/ioutil"
	"os"
)

// NewJsonDefinitionLoader returns an instance of jsonDefinitionLoader
func NewJsonDefinitionLoader() DefinitionLoader {
	return new(jsonDefinitionLoader)
}

// Loads JSON files from local files and remote URLs and provides a mechanism for writing the resulting merged
// file to disk
type jsonDefinitionLoader struct {
}

// LoadAndMerge reads one or more JSON from local files or HTTP URLs and merges them into a single data structure
func (jdl *jsonDefinitionLoader) LoadAndMerge(files []string, log logging.Logger) (map[string]interface{}, error) {
	jm := config.NewJSONMergerWithDirectLogging(log, new(config.JSONContentParser))
	jm.MergeArrays = true

	return jm.LoadAndMergeConfig(files)
}

// WriteMerged converts the supplied data structure to JSON and writes to disk at the specified location
func (jdl *jsonDefinitionLoader) WriteMerged(data map[string]interface{}, path string, log logging.Logger) error {

	b, err := json.MarshalIndent(data, "", "\t")

	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path, b, 0644)

	if err != nil {
		return err
	}

	return nil
}

func (jdl *jsonDefinitionLoader) FacilityManifest(path string) (*Manifest, error) {

	mf, err := os.Open(path)

	if err != nil {
		return nil, fmt.Errorf("unable to open manifest file at %s: %s", path, err.Error())
	}

	defer mf.Close()

	b, err := ioutil.ReadAll(mf)

	if err != nil {
		return nil, fmt.Errorf("unable to read manifest file at %s: %s", path, err.Error())
	}

	m := new(Manifest)

	err = json.Unmarshal(b, m)

	if err != nil {
		return nil, fmt.Errorf("unable to parse manifest file at %s: %s", path, err.Error())
	}

	return m, nil
}
