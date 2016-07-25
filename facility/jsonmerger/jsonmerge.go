package jsonmerger

import (
	"encoding/json"
	"github.com/graniticio/granitic/config"
	"github.com/graniticio/granitic/logging"
	"io/ioutil"
)

type JsonObject map[string]interface{}

type JsonMerger struct {
	Logger logging.Logger
}

func (jm *JsonMerger) LoadAndMergeConfig(files []string) map[string]interface{} {

	var mergedConfig map[string]interface{}

	for index, fileName := range files {

		jm.Logger.LogTracef("Reading %s", fileName)

		jsonData, err := ioutil.ReadFile(fileName)
		jm.check(err)

		var loadedConfig interface{}
		err = json.Unmarshal(jsonData, &loadedConfig)
		jm.check(err)

		additionalConfig := loadedConfig.(map[string]interface{})

		if index == 0 {
			mergedConfig = additionalConfig
		} else {
			mergedConfig = jm.merge(mergedConfig, additionalConfig)
		}

	}

	return mergedConfig
}

func (jm *JsonMerger) merge(base, additional map[string]interface{}) map[string]interface{} {

	for key, value := range additional {

		if existingEntry, ok := base[key]; ok {

			existingEntryType := config.JsonType(existingEntry)
			newEntryType := config.JsonType(value)

			if existingEntryType == config.JsonMap && newEntryType == config.JsonMap {
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

func (jm *JsonMerger) check(e error) {
	if e != nil {
		panic(e)
	}
}
