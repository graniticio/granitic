package httpserver

import (
	"fmt"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/logging"
	"net/http"
	"regexp"
	"time"
)

type RegisteredProvider struct {
	Provider HttpEndpointProvider
	Pattern  *regexp.Regexp
}

type HttpServer struct {
	registeredProvidersByMethod map[string][]*RegisteredProvider
	componentContainer          *ioc.ComponentContainer
	FrameworkLogger             logging.Logger
	AccessLogWriter             *AccessLogWriter
	AccessLogging               bool
	Port                        int
	ContentType                 string
	Encoding                    string
}

func (hs *HttpServer) Container(container *ioc.ComponentContainer) {
	hs.componentContainer = container
}

func (hs *HttpServer) registerProvider(endPointProvider HttpEndpointProvider) {

	for _, method := range endPointProvider.SupportedHttpMethods() {

		pattern := endPointProvider.RegexPattern()
		compiledRegex, regexError := regexp.Compile(pattern)

		if regexError != nil {
			hs.FrameworkLogger.LogErrorf("Unable to compile regular expression from pattern %s: %s", pattern, regexError.Error())
		}

		hs.FrameworkLogger.LogTracef("Registering %s %s", pattern, method)

		rp := RegisteredProvider{endPointProvider, compiledRegex}

		providersForMethod := hs.registeredProvidersByMethod[method]

		if providersForMethod == nil {
			providersForMethod = make([]*RegisteredProvider, 1)
			providersForMethod[0] = &rp
			hs.registeredProvidersByMethod[method] = providersForMethod
		} else {
			hs.registeredProvidersByMethod[method] = append(providersForMethod, &rp)
		}
	}

}

func (hs *HttpServer) StartComponent() error {

	hs.registeredProvidersByMethod = make(map[string][]*RegisteredProvider)

	for name, component := range hs.componentContainer.AllComponents() {
		provider, found := component.Instance.(HttpEndpointProvider)

		if found {
			hs.FrameworkLogger.LogDebugf("Found HttpEndpointProvider %s", name)

			hs.registerProvider(provider)

		}
	}

	return nil
}

func (hs *HttpServer) AllowAccess() error {
	http.Handle("/", http.HandlerFunc(hs.handleAll))

	listenAddress := fmt.Sprintf(":%d", hs.Port)

	go http.ListenAndServe(listenAddress, nil)

	hs.FrameworkLogger.LogInfof("HTTP server started listening on %d", hs.Port)

	return nil
}

func (h *HttpServer) handleAll(responseWriter http.ResponseWriter, request *http.Request) {

	received := time.Now()
	matched := false

	contentType := fmt.Sprintf("%s; charset=%s", h.ContentType, h.Encoding)
	responseWriter.Header().Set("Content-Type", contentType)

	providersByMethod := h.registeredProvidersByMethod[request.Method]

	path := request.URL.Path

	h.FrameworkLogger.LogTracef("Finding provider to handle %s %s from %d providers", path, request.Method, len(providersByMethod))

	wrw := new(wrappedResponseWriter)
	wrw.rw = responseWriter

	for _, handlerPattern := range providersByMethod {

		pattern := handlerPattern.Pattern

		h.FrameworkLogger.LogTracef("Testing %s", pattern.String())

		if pattern.MatchString(path) {
			h.FrameworkLogger.LogTracef("Matches %s", pattern.String())
			matched = true
			handlerPattern.Provider.ServeHTTP(wrw, request)
		}
	}

	if !matched {
		h.handleNotFound(request, wrw)
	}

	if h.AccessLogging {
		finished := time.Now()
		h.AccessLogWriter.LogRequest(request, wrw, &received, &finished, nil)
	}

}

func (h *HttpServer) handleNotFound(req *http.Request, res *wrappedResponseWriter) {

	http.NotFound(res, req)

}

type wrappedResponseWriter struct {
	rw          http.ResponseWriter
	Status      int
	BytesServed int
}

func (wrw *wrappedResponseWriter) Header() http.Header {
	return wrw.rw.Header()
}

func (wrw *wrappedResponseWriter) Write(b []byte) (int, error) {

	wrw.BytesServed += len(b)

	return wrw.rw.Write(b)
}

func (wrw *wrappedResponseWriter) WriteHeader(i int) {
	wrw.Status = i
	wrw.rw.WriteHeader(i)
}
