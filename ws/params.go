package ws

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
)

func NewWsParamsForPath(keys []string, values []string) *WsParams {

	contents := make(url.Values)
	v := len(values)
	var names []string

	for i, k := range keys {

		if i < v {
			contents[k] = []string{values[i]}
			names = append(names, k)
		}

	}

	p := new(WsParams)
	p.values = contents
	p.paramNames = names

	return p

}

func NewWsParamsForQuery(values url.Values) *WsParams {

	wp := new(WsParams)
	wp.values = values

	var names []string

	for k, _ := range values {
		names = append(names, k)
	}

	wp.paramNames = names

	return wp

}

type WsParams struct {
	values     url.Values
	paramNames []string
}

func (wp *WsParams) ParamNames() []string {
	return wp.paramNames
}

func (wp *WsParams) Exists(key string) bool {
	return wp.values[key] != nil
}

func (wp *WsParams) MultipleValues(key string) bool {

	value := wp.values[key]

	return value != nil && len(value) > 1

}

func (wp *WsParams) StringValue(key string) (string, error) {

	s := wp.values[key]

	if s == nil {
		return "", wp.noVal(key)
	}

	return s[len(s)-1], nil

}

func (wp *WsParams) BoolValue(key string) (bool, error) {

	v := wp.values[key]

	if v == nil {
		return false, wp.noVal(key)
	}

	b, err := strconv.ParseBool(v[len(v)-1])

	return b, err

}

func (wp *WsParams) FloatNValue(key string, bits int) (float64, error) {

	v := wp.values[key]

	if v == nil {
		return 0.0, wp.noVal(key)
	}

	i, err := strconv.ParseFloat(v[len(v)-1], bits)

	return i, err

}

func (wp *WsParams) IntNValue(key string, bits int) (int64, error) {

	v := wp.values[key]

	if v == nil {
		return 0, wp.noVal(key)
	}

	i, err := strconv.ParseInt(v[len(v)-1], 10, bits)

	return i, err

}

func (wp *WsParams) UIntNValue(key string, bits int) (uint64, error) {

	v := wp.values[key]

	if v == nil {
		return 0, wp.noVal(key)
	}

	i, err := strconv.ParseUint(v[len(v)-1], 10, bits)

	return i, err

}

func (wp *WsParams) noVal(key string) error {
	message := fmt.Sprintf("No value available for key %s", key)
	return errors.New(message)
}
