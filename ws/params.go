// Copyright 2016-2022 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package ws

import (
	"github.com/graniticio/granitic/v3/types"
	"net/url"
	"strconv"
)

// NewParamsForPath creates a Params used to store the elements of a request
// path extracted using regular expression groups.
func NewParamsForPath(targets []string, values []string) *types.Params {

	contents := make(url.Values)
	v := len(values)
	var names []string

	for i, k := range targets {

		if i < v {
			contents[strconv.Itoa(i)] = []string{values[i]}
			names = append(names, k)
		}

	}

	p := types.NewParams(contents, names)

	return p

}

// NewParamsForQuery creates a Params storing the HTTP query parameters from a request.
func NewParamsForQuery(values url.Values) *types.Params {

	var names []string

	for k := range values {
		names = append(names, k)
	}

	return types.NewParams(values, names)

}
