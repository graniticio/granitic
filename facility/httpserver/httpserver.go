/*
Package httpserver provides a configurable HTTP server for processing web-service requests.
*/
package httpserver

import (
	"errors"
	"fmt"
	"github.com/graniticio/granitic/httpendpoint"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/logging"
	"github.com/graniticio/granitic/ws"
	"golang.org/x/net/context"
	"net"
	"net/http"
	"regexp"
	"sync/atomic"
	"time"
)

type RegisteredProvider struct {
	Provider httpendpoint.HttpEndpointProvider
	Pattern  *regexp.Regexp
}

type HTTPServer struct {
	registeredProvidersByMethod map[string][]*RegisteredProvider
	componentContainer          *ioc.ComponentContainer
	FrameworkLogger             logging.Logger
	AccessLogWriter             *AccessLogWriter
	AccessLogging               bool
	Port                        int
	Address                     string
	AbnormalStatusWriter        ws.AbnormalStatusWriter
	AbnormalStatusWriterName    string
	ActiveRequests              int64
	MaxConcurrent               int64
	TooBusyStatus               int
	VersionExtractor            httpendpoint.RequestedVersionExtractor
	available                   bool
}

func (h *HTTPServer) Container(container *ioc.ComponentContainer) {
	h.componentContainer = container
}

func (h *HTTPServer) registerProvider(endPointProvider httpendpoint.HttpEndpointProvider) {

	for _, method := range endPointProvider.SupportedHttpMethods() {
		var compiledRegex *regexp.Regexp
		var err error

		pattern := endPointProvider.RegexPattern()

		if compiledRegex, err = regexp.Compile(pattern); err != nil {
			h.FrameworkLogger.LogErrorf("Unable to compile regular expression from pattern %s: %s", pattern, err.Error())
		}

		h.FrameworkLogger.LogTracef("Registering %s %s", pattern, method)

		rp := RegisteredProvider{endPointProvider, compiledRegex}

		providersForMethod := h.registeredProvidersByMethod[method]

		if providersForMethod == nil {
			providersForMethod = make([]*RegisteredProvider, 1)
			providersForMethod[0] = &rp
			h.registeredProvidersByMethod[method] = providersForMethod
		} else {
			h.registeredProvidersByMethod[method] = append(providersForMethod, &rp)
		}
	}

}

func (h *HTTPServer) StartComponent() error {

	h.registeredProvidersByMethod = make(map[string][]*RegisteredProvider)

	for name, component := range h.componentContainer.AllComponents() {

		if provider, found := component.Instance.(httpendpoint.HttpEndpointProvider); found {
			h.FrameworkLogger.LogDebugf("Found HttpEndpointProvider %s", name)
			h.registerProvider(provider)
		}
	}

	if h.AbnormalStatusWriter == nil {
		return errors.New("No AbnormalStatusWriter set.")
	}

	return nil
}

func (h *HTTPServer) AllowAccess() error {

	http.Handle("/", http.HandlerFunc(h.handleAll))

	listenAddress := fmt.Sprintf("%s:%d", h.Address, h.Port)

	//Check if the address is already in use
	if ln, err := net.Listen("tcp", listenAddress); err == nil {
		ln.Close()
	} else {
		return err
	}

	go http.ListenAndServe(listenAddress, nil)

	h.FrameworkLogger.LogInfof("HTTP server started listening on %d", h.Port)

	h.available = true

	return nil
}

func (h *HTTPServer) handleAll(res http.ResponseWriter, req *http.Request) {

	wrw := httpendpoint.NewHTTPResponseWriter(res)
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	if !h.available {
		state := ws.NewAbnormalState(h.TooBusyStatus, wrw)
		if err := h.AbnormalStatusWriter.WriteAbnormalStatus(ctx, state); err != nil {
			h.FrameworkLogger.LogErrorfCtx(ctx, err.Error())
		}
		return
	}

	rCount := atomic.AddInt64(&h.ActiveRequests, 1)
	defer atomic.AddInt64(&h.ActiveRequests, -1)

	if h.MaxConcurrent > 0 && rCount > h.MaxConcurrent {
		state := ws.NewAbnormalState(h.TooBusyStatus, wrw)
		if err := h.AbnormalStatusWriter.WriteAbnormalStatus(ctx, state); err != nil {
			h.FrameworkLogger.LogErrorfCtx(ctx, err.Error())
		}
		return
	}

	received := time.Now()
	matched := false

	providersByMethod := h.registeredProvidersByMethod[req.Method]

	path := req.URL.Path

	h.FrameworkLogger.LogTracef("Finding provider to handle %s %s from %d providers", path, req.Method, len(providersByMethod))

	for _, handlerPattern := range providersByMethod {

		pattern := handlerPattern.Pattern

		h.FrameworkLogger.LogTracef("Testing %s", pattern.String())

		if pattern.MatchString(path) && h.versionMatch(req, handlerPattern.Provider) {
			h.FrameworkLogger.LogTracef("Matches %s", pattern.String())
			matched = true
			ctx = handlerPattern.Provider.ServeHTTP(ctx, wrw, req)
		}
	}

	if !matched {
		state := ws.NewAbnormalState(http.StatusNotFound, wrw)

		if err := h.AbnormalStatusWriter.WriteAbnormalStatus(ctx, state); err != nil {
			h.FrameworkLogger.LogErrorfCtx(ctx, err.Error())
		}
	}

	if h.AccessLogging {
		finished := time.Now()
		h.AccessLogWriter.LogRequest(req, wrw, &received, &finished, ctx)
	}

}

func (h *HTTPServer) versionMatch(r *http.Request, p httpendpoint.HttpEndpointProvider) bool {

	if h.VersionExtractor == nil || !p.VersionAware() {
		return true
	}

	version := h.VersionExtractor.Extract(r)

	return p.SupportsVersion(version)

}

func (h *HTTPServer) PrepareToStop() {
	h.available = false
}

func (h *HTTPServer) ReadyToStop() (bool, error) {
	a := h.ActiveRequests
	ready := a <= 0

	if ready {
		return true, nil
	} else {

		message := fmt.Sprintf("HTTP server listening on %d is still serving %d request(s)", h.Port, a)

		return false, errors.New(message)
	}
}

func (h *HTTPServer) Stop() error {
	return nil
}
