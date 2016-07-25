package httpserver

import (
	"net/http"
)

type HttpEndPoint struct {
	MethodPatterns map[string]string
	Handler        http.Handler
}

type HttpEndpointProvider interface {
	SupportedHttpMethods() []string
	RegexPattern() string
	ServeHTTP(w http.ResponseWriter, req *http.Request)
}

type ClientIdentity struct {
}
