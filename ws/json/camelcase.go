// Copyright 2016-2020 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package json

import (
	"encoding/json"
	"regexp"
	"strings"
)

// CamelCase converts the supplied data into a map[string]interface{} representation then recursively renames
// any keys that start with an uppercase letter to the same name with a lowercase first letter.
func CamelCase(u interface{}) (interface{}, error) {

	re, _ := regexp.Compile("([A-Z])(.*)")
	b, err := json.Marshal(u)

	if err != nil {
		return nil, err
	}

	var j interface{}

	err = json.Unmarshal(b, &j)

	if err != nil {
		return nil, err
	}

	jmap, found := j.(map[string]interface{})

	if !found {
		return u, nil
	}

	modified := make(map[string]interface{})

	convert(jmap, modified, re)

	return modified, nil
}

func convert(jmap, modified map[string]interface{}, re *regexp.Regexp) {
	for jk, jv := range jmap {

		mk := toCamelCase(jk, re)

		vmap, found := jv.(map[string]interface{})

		if found {

			mv := make(map[string]interface{})
			convert(vmap, mv, re)
			modified[mk] = mv

		} else {
			modified[mk] = jv
		}

	}

}

func toCamelCase(key string, re *regexp.Regexp) string {

	if !re.MatchString(key) {
		return key
	}

	sm := re.FindStringSubmatch(key)

	return strings.ToLower(sm[1]) + sm[2]
}
