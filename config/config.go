package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/graniticio/granitic/logging"
	"os"
	"reflect"
	"strings"
)

const JsonPathSeparator string = "."

const (
	JsonUnknown     = -1
	JsonInt         = 0
	JsonString      = 1
	JsonArray       = 2
	JsonMap         = 3
	JsonBool        = 4
	JsonStringArray = 5
)

type ConfigAccessor struct {
	JsonData        map[string]interface{}
	FrameworkLogger logging.Logger
}

func (c *ConfigAccessor) PathExists(path string) bool {
	value := c.Value(path)

	return value != nil
}

func (c *ConfigAccessor) Value(path string) interface{} {

	splitPath := strings.Split(path, JsonPathSeparator)

	return c.configVal(splitPath, c.JsonData)

}

func (c *ConfigAccessor) ObjectVal(path string) map[string]interface{} {

	value := c.Value(path)

	if value == nil {
		return nil
	} else {
		return value.(map[string]interface{})
	}
}

func (c *ConfigAccessor) StringVal(path string) (string, error) {

	v := c.Value(path)

	if v == nil {
		return "", errors.New("No string value found at " + path)
	}

	s, found := v.(string)

	if found {
		return s, nil
	} else {
		message := fmt.Sprintf("Value at %s is %q and cannot be converted to a string", path, v)
		return "", errors.New(message)

	}

}

func (c *ConfigAccessor) IntVal(path string) (int, error) {

	v := c.Value(path)

	if v == nil {
		return 0, errors.New("No such path " + path)
	}

	f, found := v.(float64)

	if found {
		return int(f), nil
	} else {
		message := fmt.Sprintf("Value at %s is %q and cannot be converted to an int", path, v)
		return 0, errors.New(message)

	}
}

func (c *ConfigAccessor) Float64Val(path string) (float64, error) {

	v := c.Value(path)

	if v == nil {
		return 0, errors.New("No such path " + path)
	}

	f, found := v.(float64)

	if found {
		return f, nil
	} else {
		message := fmt.Sprintf("Value at %s is %q and cannot be converted to a float64", path, v)
		return 0, errors.New(message)

	}
}

func (c *ConfigAccessor) Array(path string) []interface{} {

	value := c.Value(path)

	if value == nil {
		return nil
	} else {
		return c.Value(path).([]interface{})
	}
}

func (c *ConfigAccessor) BoolVal(path string) (bool, error) {

	v := c.Value(path)

	if v == nil {
		return false, errors.New("No such path " + path)
	}

	b, found := v.(bool)

	if found {
		return b, nil
	} else {
		message := fmt.Sprintf("Value at %s is %q and cannot be converted to a bool", path, v)
		return false, errors.New(message)

	}

	return v.(bool), nil
}

func JsonType(value interface{}) int {

	switch value.(type) {
	case string:
		return JsonString
	case map[string]interface{}:
		return JsonMap
	case bool:
		return JsonBool
	case []interface{}:
		return JsonArray
	default:
		return JsonUnknown
	}
}

func (c *ConfigAccessor) configVal(path []string, jsonMap map[string]interface{}) interface{} {

	var result interface{}
	result = jsonMap[path[0]]

	if result == nil {
		return nil
	}

	if len(path) == 1 {
		return result
	} else {
		remainPath := path[1:len(path)]
		return c.configVal(remainPath, result.(map[string]interface{}))
	}
}

func (ca *ConfigAccessor) SetField(fieldName string, path string, target interface{}) error {

	if !ca.PathExists(path) {
		return errors.New("No value found at " + path)
	}

	targetReflect := reflect.ValueOf(target).Elem()
	targetField := targetReflect.FieldByName(fieldName)

	k := targetField.Type().Kind()

	switch k {
	case reflect.String:
		s, _ := ca.StringVal(path)
		targetField.SetString(s)
	case reflect.Bool:
		b, _ := ca.BoolVal(path)
		targetField.SetBool(b)
	case reflect.Int:
		i, _ := ca.IntVal(path)
		targetField.SetInt(int64(i))
	case reflect.Float64:
		f, _ := ca.Float64Val(path)
		targetField.SetFloat(f)
	case reflect.Map:
		ca.populateMapField(targetField, ca.ObjectVal(path))

	default:
		ca.FrameworkLogger.LogErrorf("Unable to use value at path %s as target field %s is not a suppported type (%s)", path, fieldName, k)
	}

	return nil
}

func (ca *ConfigAccessor) populateMapField(targetField reflect.Value, contents map[string]interface{}) {
	m := reflect.MakeMap(targetField.Type())
	targetField.Set(m)

	for k, v := range contents {

		kVal := reflect.ValueOf(k)
		vVal := reflect.ValueOf(v)

		if vVal.Kind() == reflect.Slice {
			vVal = ca.arrayVal(vVal)
		}

		m.SetMapIndex(kVal, vVal)

	}

}

//TODO support arrays other than string arrays
func (ca *ConfigAccessor) arrayVal(a reflect.Value) reflect.Value {

	v := a.Interface().([]interface{})
	l := len(v)

	if l == 0 {
		ca.FrameworkLogger.LogFatalf("Cannot use an empty array as a value in a Map.")
		os.Exit(-1)
	}

	var s reflect.Value

	switch t := v[0].(type) {
	case string:
		s = reflect.MakeSlice(reflect.TypeOf([]string{}), 0, 0)
	default:
		ca.FrameworkLogger.LogFatalf("Cannot use an array of %T as a value in a Map.", t)
		os.Exit(-1)
	}

	for _, elem := range v {

		s = reflect.Append(s, reflect.ValueOf(elem))

	}

	return s
}

func (ca *ConfigAccessor) Populate(path string, target interface{}) error {
	exists := ca.PathExists(path)

	if !exists {
		return errors.New("No such path: " + path)
	}

	object := ca.ObjectVal(path)
	data, _ := json.Marshal(object)

	json.Unmarshal(data, target)

	return nil
}
