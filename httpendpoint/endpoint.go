package httpendpoint

import (
	"golang.org/x/net/context"
	"net/http"
)

type HttpEndPoint struct {
	MethodPatterns map[string]string
	Handler        http.Handler
}

type HttpEndpointProvider interface {
	SupportedHttpMethods() []string
	RegexPattern() string
	ServeHTTP(ctx context.Context, w *HTTPResponseWriter, req *http.Request) context.Context
	VersionAware() bool
	SupportsVersion(version RequiredVersion) bool
}

type RequiredVersion map[string]interface{}

type RequestedVersionExtractor interface {
	Extract(*http.Request) RequiredVersion
}
