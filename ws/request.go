// Copyright 2016 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package ws

import (
	"github.com/graniticio/granitic/iam"
	"github.com/graniticio/granitic/types"
	"context"
	"net/http"
)

// Stores information about a web service request that has been either copied in or derived from an underlying HTTP request.
type WsRequest struct {
	// The HTTP method (GET, POST etc) of the underlying HTTP request.
	HttpMethod string

	// If the HTTP request had a body and if the handler that generated this WsRequest implements WsUnmarshallTarget,
	// then RequestBody will contain a struct representation of the request body.
	RequestBody interface{}

	// A copy of the HTTP query parameters from the underlying HTTP request with type-safe accessors.
	QueryParams *WsParams

	// Information extracted from the path portion of the HTTP request using regular expression groups with type-safe accessors.
	PathParams []string

	// Problems encountered during the parsing and binding phases of request processing.
	FrameworkErrors []*WsFrameworkError
	populatedFields types.StringSet

	// Information about the web service caller (if the handler has a WsIdentifier).
	UserIdentity iam.ClientIdentity

	//The underlying HTTP request and response  (if the handler was configured to pass
	// this information on).
	UnderlyingHTTP *DirectHTTPAccess

	//The component name of the handler that generated this WsRequest.
	ServingHandler string
}

// HasFrameworkErrors returns true if one or more framework errors have been recorded.
func (wsr *WsRequest) HasFrameworkErrors() bool {
	return len(wsr.FrameworkErrors) > 0
}

// AddFrameworkError records a framework error.
func (wsr *WsRequest) AddFrameworkError(f *WsFrameworkError) {
	wsr.FrameworkErrors = append(wsr.FrameworkErrors, f)
}

// RecordFieldAsBound is used to record the fact that a field on the RequestBody was explicitly set
// by the query/path parameter binding process.
func (wsr *WsRequest) RecordFieldAsBound(fieldName string) {
	if wsr.populatedFields == nil {
		wsr.populatedFields = new(types.OrderedStringSet)
	}

	wsr.populatedFields.Add(fieldName)
}

// WasFieldBound returns true if a field on the RequestBody was explicitly set
// by the query/path parameter binding process.
func (wsr *WsRequest) WasFieldBound(fieldName string) bool {
	return wsr.populatedFields.Contains(fieldName)
}

// BoundFields returns the name of all of the names on the RequestBody that were explicitly set
// by the query/path parameter binding process.
func (wsr *WsRequest) BoundFields() types.StringSet {

	if wsr.populatedFields == nil {
		return types.NewOrderedStringSet([]string{})
	} else {
		return wsr.populatedFields
	}

}

// Implement by components that are able to convert an HTTP request body into a struct.
type WsUnmarshaller interface {
	// Unmarshall deserialises an HTTP request body and converts it to a struct.
	Unmarshall(ctx context.Context, req *http.Request, wsReq *WsRequest) error
}

// Wraps the underlying low-level HTTP request and response writing objects.
type DirectHTTPAccess struct {
	// The HTTP response output stream.
	ResponseWriter http.ResponseWriter

	// The incoming HTTP request.
	Request *http.Request
}
