package json

import (
	"encoding/json"
	"regexp"
	"strings"
)

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
