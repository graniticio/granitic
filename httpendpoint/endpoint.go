package httpendpoint

import (
	"net/http"
	"github.com/graniticio/granitic/ws"
	"github.com/graniticio/granitic/iam"
)

type HttpEndPoint struct {
	MethodPatterns map[string]string
	Handler        http.Handler
}

type HttpEndpointProvider interface {
	SupportedHttpMethods() []string
	RegexPattern() string
	ServeHTTP(w *ws.WsHTTPResponseWriter, req *http.Request) iam.ClientIdentity
	VersionAware() bool
	SupportsVersion(version RequiredVersion)
}



type RequiredVersion map[string]interface{}
