package httpendpoint

import (
	"net/http"
	"github.com/graniticio/granitic/iam"
)

type HttpEndPoint struct {
	MethodPatterns map[string]string
	Handler        http.Handler
}

type HttpEndpointProvider interface {
	SupportedHttpMethods() []string
	RegexPattern() string
	ServeHTTP(w *HTTPResponseWriter, req *http.Request) iam.ClientIdentity
	VersionAware() bool
	SupportsVersion(version RequiredVersion) bool
}



type RequiredVersion map[string]interface{}

type RequestedVersionExtractor interface {
	Extract(*http.Request) RequiredVersion
}