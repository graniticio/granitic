// Copyright 2016-2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package ws

import (
	"context"
	"github.com/graniticio/granitic/v2/iam"
	"github.com/graniticio/granitic/v2/types"
	"net/http"
)

// Request stores information about a web service request that has been either copied in or derived from an underlying HTTP request.
type Request struct {
	// The HTTP method (GET, POST etc) of the underlying HTTP request.
	HTTPMethod string

	// If the HTTP request had a body and if the handler that generated this Request implements WsUnmarshallTarget,
	// then RequestBody will contain a struct representation of the request body.
	RequestBody interface{}

	// A copy of the HTTP query parameters from the underlying HTTP request with type-safe accessors.
	QueryParams *types.Params

	// Information extracted from the path portion of the HTTP request using regular expression groups with type-safe accessors.
	PathParams []string

	// Problems encountered during the parsing and binding phases of request processing.
	FrameworkErrors []*FrameworkError
	populatedFields types.StringSet

	// Information about the web service caller (if the handler has a Identifier).
	UserIdentity iam.ClientIdentity

	// The underlying HTTP request and response  (if the handler was configured to pass
	// this information on).
	UnderlyingHTTP *DirectHTTPAccess

	// The component name of the handler that generated this Request.
	ServingHandler string

	// The unique ID assigned to this request and stored in the context
	ID func(ctx context.Context) string
}

// HasFrameworkErrors returns true if one or more framework errors have been recorded.
func (wsr *Request) HasFrameworkErrors() bool {
	return len(wsr.FrameworkErrors) > 0
}

// AddFrameworkError records a framework error.
func (wsr *Request) AddFrameworkError(f *FrameworkError) {
	wsr.FrameworkErrors = append(wsr.FrameworkErrors, f)
}

// RecordFieldAsBound is used to record the fact that a field on the RequestBody was explicitly set
// by the query/path parameter binding process.
func (wsr *Request) RecordFieldAsBound(fieldName string) {
	if wsr.populatedFields == nil {
		wsr.populatedFields = new(types.OrderedStringSet)
	}

	wsr.populatedFields.Add(fieldName)
}

// WasFieldBound returns true if a field on the RequestBody was explicitly set
// by the query/path parameter binding process.
func (wsr *Request) WasFieldBound(fieldName string) bool {
	return wsr.populatedFields.Contains(fieldName)
}

// BoundFields returns the name of all of the names on the RequestBody that were explicitly set
// by the query/path parameter binding process.
func (wsr *Request) BoundFields() types.StringSet {

	if wsr.populatedFields == nil {
		return types.NewOrderedStringSet([]string{})
	}

	return wsr.populatedFields
}

// Unmarshaller is implemented by components that are able to convert an HTTP request body into a struct.
type Unmarshaller interface {
	// Unmarshall deserialises an HTTP request body and converts it to a struct.
	Unmarshall(ctx context.Context, req *http.Request, wsReq *Request) error
}

// DirectHTTPAccess wraps the underlying low-level HTTP request and response writing objects.
type DirectHTTPAccess struct {
	// The HTTP response output stream.
	ResponseWriter http.ResponseWriter

	// The incoming HTTP request.
	Request *http.Request
}

type ridFuncKey string

const ridFunc ridFuncKey = "GRNCRIDFUNC"

// StoreRequestIDFunction allows the function used to extract a request from a function to be passed around in a context
func StoreRequestIDFunction(ctx context.Context, f func(context.Context) string) context.Context {

	return context.WithValue(ctx, ridFunc, f)

}

// RecoverIDFunction allows the function used to extract a request from a function to recovered from a context
func RecoverIDFunction(ctx context.Context) func(context.Context) string {

	f := ctx.Value(ridFunc)

	if f != nil {
		return f.(func(context.Context) string)
	}

	return nil
}
