package httpserver

import (
	"github.com/graniticio/granitic/ws"
	"net/http"
)

type HttpEndPoint struct {
	MethodPatterns map[string]string
	Handler        http.Handler
}

type HttpEndpointProvider interface {
	SupportedHttpMethods() []string
	RegexPattern() string
	ServeHTTP(w *ws.WsHTTPResponseWriter, req *http.Request) ws.WsIdentity
}

type ClientIdentity struct {
}
