package config

import (
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

type ConfigValue interface{}

type ConfigAccessor struct {
	JsonData        map[string]interface{}
	FrameworkLogger logging.Logger
}

func (c *ConfigAccessor) PathExists(path string) bool {
	value := c.Value(path)

	return value != nil
}

func (c *ConfigAccessor) Value(path string) ConfigValue {

	splitPath := strings.Split(path, JsonPathSeparator)

	return c.configValue(splitPath, c.JsonData)

}

func (c *ConfigAccessor) ObjectVal(path string) map[string]interface{} {

	value := c.Value(path)

	if value == nil {
		return nil
	} else {
		return value.(map[string]interface{})
	}
}

func (c *ConfigAccessor) StringVal(path string) string {
	return c.Value(path).(string)
}

func (c *ConfigAccessor) StringFieldVal(field string, object map[string]interface{}) string {
	return object[field].(string)
}

func (c *ConfigAccessor) IntValue(path string) int {
	return int(c.Value(path).(float64))
}

func (c *ConfigAccessor) Float64Value(path string) float64 {
	return c.Value(path).(float64)
}

func (c *ConfigAccessor) Array(path string) []interface{} {

	value := c.Value(path)

	if value == nil {
		return nil
	} else {
		return c.Value(path).([]interface{})
	}
}

func (c *ConfigAccessor) BoolValue(path string) bool {
	return c.Value(path).(bool)
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

func (c *ConfigAccessor) configValue(path []string, jsonMap map[string]interface{}) interface{} {

	var result interface{}
	result = jsonMap[path[0]]

	if result == nil {
		return nil
	}

	if len(path) == 1 {
		return result
	} else {
		remainPath := path[1:len(path)]
		return c.configValue(remainPath, result.(map[string]interface{}))
	}
}

func (ca *ConfigAccessor) SetField(fieldName string, path string, target interface{}) {

	targetReflect := reflect.ValueOf(target).Elem()
	targetField := targetReflect.FieldByName(fieldName)

	k := targetField.Type().Kind()

	switch k {
	case reflect.String:
		targetField.SetString(ca.StringVal(path))
	case reflect.Bool:
		targetField.SetBool(ca.BoolValue(path))
	case reflect.Int:
		targetField.SetInt(int64(ca.IntValue(path)))
	case reflect.Map:
		ca.populateMapField(targetField, ca.ObjectVal(path))

	default:
		ca.FrameworkLogger.LogErrorf("Unable to use value at path %s as target field %s is not a suppported type (%s)", path, fieldName, k)
	}

}

func (ca *ConfigAccessor) populateMapField(targetField reflect.Value, contents map[string]interface{}) {
	m := reflect.MakeMap(targetField.Type())
	targetField.Set(m)

	for k, v := range contents {

		kVal := reflect.ValueOf(k)
		vVal := reflect.ValueOf(v)

		if vVal.Kind() == reflect.Slice {
			vVal = ca.arrayValue(vVal)
		}

		m.SetMapIndex(kVal, vVal)

	}

}

//TODO support arrays other than string arrays
func (ca *ConfigAccessor) arrayValue(a reflect.Value) reflect.Value {

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

func (ca *ConfigAccessor) Populate(path string, target interface{}) {
	exists := ca.PathExists(path)

	if exists {
		targetReflect := reflect.ValueOf(target).Elem()
		targetType := targetReflect.Type()
		numFields := targetType.NumField()

		for i := 0; i < numFields; i++ {

			fieldName := targetType.Field(i).Name

			expectedConfigPath := path + JsonPathSeparator + fieldName

			if ca.PathExists(expectedConfigPath) {
				ca.SetField(fieldName, expectedConfigPath, target)
			}

		}

	} else {
		ca.FrameworkLogger.LogErrorf("Trying to populate an object from a JSON object, but the base path %s does not exist", path)
	}

}
