// Copyright 2016-2022 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package ws

import (
	"net/http"
	"strconv"
)

// HTTPStatusCodeDeterminer implemented by a component able to choose the most appropriate HTTP status code to set given the state of a Response
type HTTPStatusCodeDeterminer interface {
	// DetermineCode returns the HTTP status code that should be set on the response.
	DetermineCode(response *Response) int
}

// NewGraniticHTTPStatusCodeDeterminer creates a GraniticHTTPStatusCodeDeterminer using the HTTP response
// codes described in this package's package documentation
func NewGraniticHTTPStatusCodeDeterminer() *GraniticHTTPStatusCodeDeterminer {

	g := new(GraniticHTTPStatusCodeDeterminer)

	g.NoError = http.StatusOK
	g.Client = http.StatusBadRequest
	g.Security = http.StatusUnauthorized
	g.Unexpected = http.StatusInternalServerError
	g.Logic = http.StatusConflict

	return g

}

/*
GraniticHTTPStatusCodeDeterminer is the default HTTPStatusCodeDeterminer used by Granitic's XXXWs facilities. The actual
HTTP status codes returned can be customised using the fields on this struct. Use NewGraniticHTTPStatusCodeDeterminer()
to achieve the default behaviour described in this package's package documentation
*/
type GraniticHTTPStatusCodeDeterminer struct {
	NoError    int
	Client     int
	Security   int
	Unexpected int
	Logic      int
}

// DetermineCode examines the response and returns an HTTP status code according to the rules defined at the top of this
// GoDoc page.
func (dhscd *GraniticHTTPStatusCodeDeterminer) DetermineCode(response *Response) int {
	if response.HTTPStatus != 0 {
		return response.HTTPStatus

	} else if response.Errors.HasErrors() {
		return dhscd.determineCodeFromErrors(response.Errors)

	} else {
		return dhscd.NoError
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
			return dhscd.Unexpected
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
		return dhscd.Security
	}

	if cCount > 0 {
		return dhscd.Client
	}

	if lCount > 0 {
		return dhscd.Logic
	}

	return dhscd.NoError
}
