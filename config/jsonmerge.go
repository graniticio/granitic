package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/graniticio/granitic/instance"
	"github.com/graniticio/granitic/logging"
	"io/ioutil"
	"net/http"
)

const jsonMergerComponentName string = instance.FrameworkPrefix + "JsonMerger"

type JSONObject map[string]interface{}

func NewJSONMerger(flm *logging.ComponentLoggerManager) *JSONMerger {
	jm := new(JSONMerger)

	jm.Logger = flm.CreateLogger(jsonMergerComponentName)

	return jm
}

type JSONMerger struct {
	Logger logging.Logger
}

func (jm *JSONMerger) LoadAndMergeConfig(files []string) (map[string]interface{}, error) {

	var mergedConfig map[string]interface{}
	var jsonData []byte
	var err error

	for index, fileName := range files {

		if isURL(fileName) {
			jm.Logger.LogTracef("Acessing URL %s", fileName)

			jsonData, err = jm.loadFromURL(fileName)

		} else {
			jm.Logger.LogTracef("Reading file %s", fileName)

			jsonData, err = ioutil.ReadFile(fileName)
		}

		if err != nil {
			m := fmt.Sprintf("Problem reading data from file/URL %s: %s", fileName, err)
			return nil, errors.New(m)
		}

		var loadedConfig interface{}
		err = json.Unmarshal(jsonData, &loadedConfig)

		if err != nil {
			m := fmt.Sprintf("Problem parsing data from a file or URL (%s) as JSON : %s", fileName, err)
			return nil, errors.New(m)
		}

		additionalConfig := loadedConfig.(map[string]interface{})

		if index == 0 {
			mergedConfig = additionalConfig
		} else {
			mergedConfig = jm.merge(mergedConfig, additionalConfig)
		}

	}

	return mergedConfig, nil
}

func (jm *JSONMerger) loadFromURL(url string) ([]byte, error) {

	r, err := http.Get(url)

	if err != nil {
		return nil, err
	}

	if r.StatusCode >= 400 {
		m := fmt.Sprintf("HTTP %d", r.StatusCode)
		return nil, errors.New(m)
	}

	var b bytes.Buffer

	b.ReadFrom(r.Body)
	r.Body.Close()

	return b.Bytes(), nil
}

func (jm *JSONMerger) merge(base, additional map[string]interface{}) map[string]interface{} {

	for key, value := range additional {

		if existingEntry, ok := base[key]; ok {

			existingEntryType := JsonType(existingEntry)
			newEntryType := JsonType(value)

			if existingEntryType == JsonMap && newEntryType == JsonMap {
				jm.merge(existingEntry.(map[string]interface{}), value.(map[string]interface{}))
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
