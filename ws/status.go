// Copyright 2016-2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package ws

import (
	"net/http"
	"strconv"
)

// Implemented by a component able to choose the most appropriate HTTP status code to set given the state of a Response
type HTTPStatusCodeDeterminer interface {
	// DetermineCode returns the HTTP status code that should be set on the response.
	DetermineCode(response *Response) int
}

/*
	The default HTTPStatusCodeDeterminer used by Granitic's XXXWs facilities. See the top of this page for
	rules on how this code is determined.
*/
type GraniticHTTPStatusCodeDeterminer struct {
}

// DetermineCode examines the response and returns an HTTP status code according to the rules defined at the top of this
// GoDoc page.
func (dhscd *GraniticHTTPStatusCodeDeterminer) DetermineCode(response *Response) int {
	if response.HTTPStatus != 0 {
		return response.HTTPStatus

	} else if response.Errors.HasErrors() {
		return dhscd.determineCodeFromErrors(response.Errors)

	} else {
		return http.StatusOK
	}
}

func (dhscd *GraniticHTTPStatusCodeDeterminer) determineCodeFromErrors(errors *ServiceErrors) int {

	if errors.HTTPStatus != 0 {
		return errors.HTTPStatus
	}

	sCount := 0
	cCount := 0
	lCount := 0

	for _, error := range errors.Errors {

		switch error.Category {
		case Unexpected:
			return http.StatusInternalServerError
		case HTTP:
			i, _ := strconv.Atoi(error.Code)
			return i
		case Security:
			sCount++
		case Logic:
			lCount++
		case Client:
			cCount++
		}
	}

	if sCount > 0 {
		return http.StatusUnauthorized
	}

	if cCount > 0 {
		return http.StatusBadRequest
	}

	if lCount > 0 {
		return http.StatusConflict
	}

	return http.StatusOK
}
