// Copyright 2016-2022 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/graniticio/granitic/v3/instance"
	"github.com/graniticio/granitic/v3/logging"
	"net/http"
	"os"
	"strings"
)

const jsonMergerComponentName string = instance.FrameworkPrefix + "JSONMerger"

// A ContentParser can take a []byte of some structured file type (e.g. YAML, JSON() and convert into a map[string]interface{} representation
type ContentParser interface {
	ParseInto(data []byte, target interface{}) error
	Extensions() []string
	ContentTypes() []string
}

// EmptyFileError is a marker error to indicate that a config file was logically empty
type EmptyFileError struct {
	Message string
}

// Error returns the message associated with this error
func (efe EmptyFileError) Error() string {
	return efe.Message
}

// JSONContentParser supports the loading and parsing of JSON configuration and component definition files
type JSONContentParser struct {
}

// ParseInto takes a byte array that is assumed to be serialised JSON and attempts to parse that into the supplied target object
func (jcp *JSONContentParser) ParseInto(data []byte, target interface{}) error {
	return json.Unmarshal(data, &target)
}

// Extensions returns the list of filename extensions (lowercase, without leading dot) that will be considered to be JSON files.
func (jcp *JSONContentParser) Extensions() []string {
	return []string{"json"}
}

// ContentTypes returns the MIME media types/HTTP content-types that will be considered to represent JSON
func (jcp *JSONContentParser) ContentTypes() []string {
	return []string{"application/json", "application/x-javascript", "text/javascript", "text/x-javascript", "text/x-json"}
}

// NewJSONMergerWithManagedLogging creates a JSONMerger with a Logger managed by Granitic
func NewJSONMergerWithManagedLogging(flm *logging.ComponentLoggerManager, cp ContentParser) *JSONMerger {

	l := flm.CreateLogger(jsonMergerComponentName)

	return NewJSONMergerWithDirectLogging(l, cp)

}

// NewJSONMergerWithDirectLogging creates a JSONMerger that uses the supplied logger
func NewJSONMergerWithDirectLogging(l logging.Logger, cp ContentParser) *JSONMerger {

	jm := new(JSONMerger)
	jm.Logger = l
	jm.DefaultParser = cp

	jm.parserByContent = make(map[string]ContentParser)
	jm.parserByFile = make(map[string]ContentParser)

	jm.RegisterContentParser(cp)

	return jm
}

// A JSONMerger can merge a sequence of JSON configuration files (from a filesystem or HTTP URL) into a single
// view of configuration that will be used to configure Grantic's facilities and the user's IoC components. See the top
// of this page for a brief explanation of how merging works.
type JSONMerger struct {
	// Logger used by Granitic framework components. Automatically injected.
	Logger logging.Logger

	// True if arrays should be joined when merging; false if the entire conetnts of the array should be overwritten.
	MergeArrays bool

	DefaultParser ContentParser

	parserByFile    map[string]ContentParser
	parserByContent map[string]ContentParser
}

// LoadAndMergeConfig takes a list of file paths or URIs to JSON files and merges them into a single in-memory object representation.
// See the top of this page for a brief explanation of how merging works. Returns an error if a remote URI returned a 4xx or 5xx response code,
// a file or folder could not be accessed or if two files could not be merged dued to JSON parsing errors.
func (jm *JSONMerger) LoadAndMergeConfig(files []string) (map[string]interface{}, error) {
	mergedConfig := make(map[string]interface{})

	return jm.LoadAndMergeConfigWithBase(mergedConfig, files)
}

// RegisterContentParser adds the supplied ContentParser to the set of parsers that might be eligible to parse a file/stream
func (jm *JSONMerger) RegisterContentParser(cp ContentParser) {

	for _, ct := range cp.ContentTypes() {

		jm.parserByContent[strings.ToLower(ct)] = cp

	}

	for _, ext := range cp.Extensions() {

		jm.parserByFile[strings.ToLower(ext)] = cp

	}

}

// LoadAndMergeConfigWithBase takes a representation of a JSON file that has already been parsed and merges the contents
// of the supplied list of JSON files into that in-memory representation and returns the file merged version.
func (jm *JSONMerger) LoadAndMergeConfigWithBase(config map[string]interface{}, files []string) (map[string]interface{}, error) {

	var jsonData []byte
	var err error

	for _, fileName := range files {

		var cp ContentParser

		if isURL(fileName) {
			//Read config from a remote URL
			jm.Logger.LogTracef("Acessing URL %s", fileName)

			jsonData, cp, err = jm.loadFromURL(fileName)

		} else {
			//Read config from a filesystem file
			jm.Logger.LogTracef("Reading file %s", fileName)

			ext := jm.extractExtension(fileName)

			if jm.parserByFile[ext] != nil {
				jm.Logger.LogTracef("Found ContentParser for extension %s", ext)
				cp = jm.parserByFile[ext]
			} else {
				jm.Logger.LogTracef("Skipping file with unsupported extension %s", ext)
				continue
			}

			jsonData, err = os.ReadFile(fileName)
		}

		if err != nil {
			return nil, fmt.Errorf("Problem reading data from file/URL %s: %s", fileName, err)
		}

		var loadedConfig interface{}

		err = cp.ParseInto(jsonData, &loadedConfig)

		if err != nil {

			if _, found := err.(EmptyFileError); found {
				jm.Logger.LogWarnf("Config file/URL %s is empty", fileName)
			} else {
				return nil, fmt.Errorf("Problem parsing data from a file or URL (%s) as JSON : %s", fileName, err)
			}
		}

		additionalConfig := loadedConfig.(map[string]interface{})

		config = jm.merge(config, additionalConfig)

	}

	return config, nil
}

func (jm *JSONMerger) extractExtension(path string) string {

	c := strings.Split(path, ".")

	if len(c) == 1 {
		return ""
	}

	return strings.ToLower(c[len(c)-1])
}

func (jm *JSONMerger) loadFromURL(url string) ([]byte, ContentParser, error) {

	r, err := http.Get(url)

	if err != nil {
		return nil, nil, err
	}

	cp := jm.DefaultParser

	if ct := r.Header.Get("content-type"); ct != "" {
		ct = strings.Split(ct, ";")[0]
		ct = strings.TrimSpace(ct)
		ct = strings.ToLower(ct)

		if jm.parserByContent[ct] != nil {
			jm.Logger.LogDebugf("Found content parser for %s", ct)
			cp = jm.parserByContent[ct]
		}

	}

	if r.StatusCode >= 400 {
		m := fmt.Sprintf("HTTP %d", r.StatusCode)
		return nil, nil, errors.New(m)
	}

	var b bytes.Buffer

	b.ReadFrom(r.Body)
	r.Body.Close()

	return b.Bytes(), cp, nil
}

func (jm *JSONMerger) merge(base, additional map[string]interface{}) map[string]interface{} {

	for key, value := range additional {

		if existingEntry, ok := base[key]; ok {

			existingEntryType := JSONType(existingEntry)
			newEntryType := JSONType(value)

			if existingEntryType == JSONMap && newEntryType == JSONMap {
				jm.merge(existingEntry.(map[string]interface{}), value.(map[string]interface{}))
			} else if jm.MergeArrays && existingEntryType == JSONArray && newEntryType == JSONArray {
				base[key] = jm.mergeArrays(existingEntry.([]interface{}), value.([]interface{}))
			} else {
				base[key] = value
			}
		} else {
			jm.Logger.LogTracef("Adding %s", key)

			base[key] = value
		}

	}

	return base
}

func (jm *JSONMerger) mergeArrays(a []interface{}, b []interface{}) []interface{} {
	return append(a, b...)
}
