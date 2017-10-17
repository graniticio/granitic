// Copyright 2016 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package httpendpoint

import "net/http"

// A wrapper over http.ResponseWriter that provides Granitic with better visibility on the state of response writing.
type HttpResponseWriter struct {
	rw http.ResponseWriter
	// Whether or not any data has already been sent to the underlying http.ResponseWriter.
	DataSent bool

	// The HTTP status code sent to the response or zero if no code yet sent.
	Status int

	// How many bytes have been sent to the response so far (excluding headers).
	BytesServed int
}

// Header calls through to http.ResponseWriter.Header()
func (w *HttpResponseWriter) Header() http.Header {
	return w.rw.Header()
}

// Write calls through to http.ResponseWriter.Write while keeping track of the number of bytes sent.
func (w *HttpResponseWriter) Write(b []byte) (int, error) {

	w.BytesServed += len(b)
	w.DataSent = true

	return w.rw.Write(b)
}

// WriteHeader sets the HTTP status code of the HTTP response. If this method is called more than once,
// only the first value is sent to the underlying HTTP response.
func (w *HttpResponseWriter) WriteHeader(i int) {

	if w.DataSent {
		return
	}

	w.Status = i
	w.rw.WriteHeader(i)
	w.DataSent = true
}

// Create a new WsHTTPResponseWriter wrapping the supplied http.ResponseWriter
func NewHttpResponseWriter(rw http.ResponseWriter) *HttpResponseWriter {
	w := new(HttpResponseWriter)
	w.rw = rw

	return w
}
