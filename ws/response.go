// Copyright 2016-2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

/*
	Package ws defines types used by framework and application components involved in web service processing. For more information
	on how web services work in Granitic, see http://granitic.io/1.0/ref/web-services

	A brief explanation of the key types and concepts follows.

	Requests and responses

	Request and Response are abstractions of the HTTP request and response associated with a call to a web service
	endpoint. By default your application logic will not have access to the underlying HTTP objects (this can be overridden
	on a per-endpoint basis by setting AllowDirectHTTPAccess to true on your handler - see the package documentation for
	ws/handler for more information).

	Your application code will not directly control how data is parsed into a Request or how the data and/or errors
	in a Response are rendered to the caller. This is instead handled by the JSONWs or XMLWs facility.

	HTTP status codes are determined automatically based on the type (or lack of) errors in the Response object, but
	this behaviour can be overridden by setting an HTTP status code manually on the Response.

	Errors

	Errors can be detected or occur during all the phases of request processing (see http://granitic.io/1.0/ref/request-processing-phases).
	If errors are encountered during the parsing
	and binding phases of request processing, they are referred to as 'framework errors' as they are handled outside of
	application code. Framework errors result in one of small number of generic error messages being sent to a caller. See
	http://granitic.io/1.0/ref/errors-and-messages for information on how to override these messages or how to allow your
	application to have visibility of framework errors.

	If an error occurs during or after parsing and binding is complete, it will be recorded in the WsReponse.Errors
	field. These types of errors are called service errors. For more information on service errors, refer to the GoDoc for
	CategorisedError below or http://granitic.io/1.0/ref/errors-and-messages.

	Response writing

	The serialisation of the data in a Response to an HTTP response is handled by a component implementing ResponseWriter.
	A component of this type will be automatically created for you when you enable the JSONWs or XMLWs facility.

	Parameter binding

	Parameter binding refers to the process of automatically capturing request query parameters and injecting them into fields
	on the Request Body. It also refers to a similar process for extracting information from a request's path using regular expressions.
	See http://granitic.io/1.0/ref/parameter-binding for more details.

	IAM and versioning

	Granitic does not provide implementations of Identity Access Management or request versioning, but instead provides
	highly generic types to allow your application's implementations of these concepts to be integrated with Grantic's web
	service request processing. See the GoDoc for Identifier, AccessChecker and handler/WsVersionAssessor and the iam package for more details.

	HTTP status code determination

	Unless your application defines its own HTTPStatusCodeDeterminer, the eventual HTTP status code set on the response
	to a web service request it determined by examining the state of a Response using the following logic:

	1. If the Response.HTTPStatus field is non-zero, use that.

	2. If the Response.Errors.HTTPStatus field is non-zero, use that.

	3. If the Response.Errors structure:

	a) Contains one or more 'Unexpected' errors, use HTTP 500.

	b) Contains an 'HTTP' error, convert that error's code to a number and use that.

	c) Contains one or more 'Security' errors, use HTTP 401.

	d) Contains one or more 'Client' errors, use HTTP 400.

	e) Contains one or more 'Logic' errors, use HTTP 409.

	4. Return HTTP 200.

*/
package ws

import (
	"context"
	"github.com/graniticio/granitic/v2/httpendpoint"
	"github.com/graniticio/granitic/v2/iam"
	"net/http"
)

// An enumeration of the high-level result of processing a request. Used internally.
type Outcome uint

const (
	// A normal outcome resulting in an HTTP 200 response.
	Normal = iota

	// A outcome with anticipated and handled errors resulting in a 4xx response.
	Error

	// An unexpected or unusual outcome resulting in a 5xx response.
	Abnormal
)

// ProcessState is wrapper for current state of request processing. This type is used by
// components implementing ResponseWriter. Because a request may fail at many points during processing,
// there is no guarantee that any of the fields in this type are set, valid or complete, so this type must be used
// with caution.
type ProcessState struct {
	// The representation of the incoming request at the time processing completed or failed.
	WsRequest *Request

	// The representation of the data to be sent to the caller at the time processing completed or failed.
	WsResponse *Response

	// The HTTP output stream.
	HTTPResponseWriter *httpendpoint.HTTPResponseWriter

	// Errors detected while processing the web service request. If set, supersedes the errors present in Response field.
	ServiceErrors *ServiceErrors

	// Information about the caller or user of the web service.
	Identity iam.ClientIdentity

	// The HTTP status code to be set on the HTTP response.
	Status int
}

// NewAbnormalState creates a new ProcessState for a request that has resulted in an abnormal (HTTP 5xx) outcome).
func NewAbnormalState(status int, w *httpendpoint.HTTPResponseWriter) *ProcessState {
	state := new(ProcessState)
	state.Status = status
	state.HTTPResponseWriter = w

	return state
}

// Contains data that is relevant to the rendering of the result of a web service request to an HTTP response. This
// type is agnostic of the format (JSON, XML etc) that is to be used to render the response.
type Response struct {
	// An instruction that the HTTP status code should be set to this value (if the value is greater than 99). Generally
	// not set - the response writer will determine the correct status to use.
	HTTPStatus int

	// If the web service call resulted in data that should be written as the body of the HTTP response is stored in this field.
	// Application code must set this field explicitly.
	Body interface{}

	// All of the errors encountered while processing this request.
	Errors *ServiceErrors

	// Headers that should be set on the HTTP response.
	Headers map[string]string

	// If the type of response rendering is template based (e.g. using the XMLWs facility in template mode), this field
	// can be used to override any default templates or the template associated with the handler that created this response.
	Template string
}

// NewResponse creates a valid but empty WsReponse with Errors structure initialised.
func NewResponse(errorFinder ServiceErrorFinder) *Response {
	r := new(Response)
	r.Errors = new(ServiceErrors)
	r.Errors.ErrorFinder = errorFinder

	r.Headers = make(map[string]string)

	return r
}

// Implemented by components able write the result of a web service call to an HTTP response.
type ResponseWriter interface {
	// Write converts whatever data is present in the supplied state object to the HTTP output stream associated
	// with the current web service request.
	Write(ctx context.Context, state *ProcessState, outcome Outcome) error
}

// Implemented by components able to write a valid response even if the request resulted in an abnormal (5xx) outcome.
type AbnormalStatusWriter interface {
	// Write converts whatever data is present in the supplied state object to the HTTP output stream associated
	// with the current web service request.
	WriteAbnormalStatus(ctx context.Context, state *ProcessState) error
}

// An object that constructs response headers that are common to all web service requests. These may typically be
// caching instructions or 'processing server' records. Implementations must be extremely cautious when using
// the information in the supplied WsProcess state as some values may be nil.
type CommonResponseHeaderBuilder interface {
	BuildHeaders(ctx context.Context, state *ProcessState) map[string]string
}

// Interface for components able to convert a set of service errors into a structure suitable for serialisation.
type ErrorFormatter interface {
	// FormatErrors converts the supplied errors into a structure that a response writer will use to write the errors to
	// the current HTTP response.
	FormatErrors(errors *ServiceErrors) interface{}
}

// WriteHeaders writes the supplied map as HTTP headers.
func WriteHeaders(w http.ResponseWriter, headers map[string]string) {

	for k, v := range headers {
		w.Header().Add(k, v)
	}
}

// Implemented by components able to take the body from an Response and wrap it inside a container that will
// allow all responses to share a common structure.
type ResponseWrapper interface {
	// WrapResponse takes the supplied body and errors and wraps them in a standardised data structure.
	WrapResponse(body interface{}, errors interface{}) interface{}
}

// Merges together the headers that have been defined on the Response, the static default headers attache to this writer
// and (optionally) those constructed by the  ws.CommonResponseHeaderBuilder attached to this writer. The order of precedence,
// from lowest to highest, is static headers, constructed headers, headers in the Response.
func MergeHeaders(res *Response, ch map[string]string, dh map[string]string) map[string]string {

	merged := make(map[string]string)

	if dh != nil {
		for k, v := range dh {
			merged[k] = v
		}
	}

	if ch != nil {
		for k, v := range ch {
			merged[k] = v
		}
	}

	if res.Headers != nil {
		for k, v := range res.Headers {
			merged[k] = v
		}
	}

	return merged
}
