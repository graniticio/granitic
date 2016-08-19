package ws

import "net/http"

// A wrapper over http.ResponseWriter that provides Granitic with better visibility on the state of response writing.
type WsHTTPResponseWriter struct {
	rw          http.ResponseWriter
	DataSent	bool // Whether or not any data has already been sent to the underlying http.ResponseWriter.
	Status      int // The HTTP status code sent to the response or zero if no code yet sent.
	BytesServed int // How many bytes have been sent to the response so far, excluding headers.
}

func (w *WsHTTPResponseWriter) Header() http.Header {
	return w.rw.Header()
}

func (w *WsHTTPResponseWriter) Write(b []byte) (int, error) {

	w.BytesServed += len(b)
	w.DataSent = true

	return w.rw.Write(b)
}

// Set the HTTP status code of the HTTP response. If this method is called more that once,
// only the first value is sent to the underlying HTTP response.
func (w *WsHTTPResponseWriter) WriteHeader(i int) {

	if w.DataSent {
		return
	}

	w.Status = i
	w.rw.WriteHeader(i)
	w.DataSent = true
}

// Create a new WsHTTPResponseWriter wrapping the supplied http.ResponseWriter
func NewWsHTTPResponseWriter(rw http.ResponseWriter) *WsHTTPResponseWriter{
	w := new(WsHTTPResponseWriter)
	w.rw = rw

	return w
}


