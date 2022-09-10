// Copyright 2016-2022 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

/*
Package httpserver provides the HTTPServer facility which defines a configurable HTTP server for processing web-service requests.

The HTTPServer facility provides a server that will listen for HTTP web-service requests and map them to the web-service
endpoints defined by your application. A full description of how to configure this facility can be found at https://granitic.io/ref/http-server

This package defines two main types HTTPServer and AccessLogWriter. HTTPServer is a layer over Go's built-in http.Server adding runtime control (suspension, resumption)
and mapping of requests to instances of ws.Handler. AccessLogWriter supports Apache/Tomcat style access log formatting and writing.

Most applications will only need to enable this facility (probably changing the listen Port) and define mappings between incoming paths and application logic in their
component definition files. See handler.WsHandler for more details.
*/
package httpserver

import (
	"context"
	"errors"
	"fmt"
	"github.com/graniticio/granitic/v3/httpendpoint"
	"github.com/graniticio/granitic/v3/instrument"
	"github.com/graniticio/granitic/v3/ioc"
	"github.com/graniticio/granitic/v3/logging"
	"github.com/graniticio/granitic/v3/ws"
	"net"
	"net/http"
	"regexp"
	"sync/atomic"
	"time"
)

type registeredProvider struct {
	Provider httpendpoint.Provider
	Pattern  *regexp.Regexp
}

// HTTPServer is the server that accepts incoming HTTP requests and maps them to handlers to process them.
type HTTPServer struct {
	registeredProvidersByMethod map[string][]*registeredProvider
	unregisteredProviders       map[string]httpendpoint.Provider
	componentContainer          *ioc.ComponentContainer

	// Logger used by Granitic framework components. Automatically injected.
	FrameworkLogger logging.Logger

	// A component able to write an access log. Automatically added by this facility's builder, is access log support is enabled.
	AccessLogWriter *AccessLogWriter

	// Whether or not access logging should be enabled.
	AccessLogging bool

	// Whether or not instances of httpendpoint.Provider found in the IoC container should be automatically
	// registered with this server
	AutoFindHandlers bool

	// The TCP port on which the HTTP server should listen for requests.
	Port int

	// The IP/hostname this server should listen on, follows standard Go net package syntax. Empty string means listen on all.
	Address string

	// A component able to write valid HTTP responses in the event a user request results in an abnormal result
	// (not found, server too busy, panic in application logic). If you use the JSONWs or XMLWs facility, this is automatically injected.
	AbnormalStatusWriter ws.AbnormalStatusWriter

	// The number of HTTP requests currently being handled by the server.
	ActiveRequests int64

	// Allow request instrumentation to begin BEFORE too-busy/suspended checks. Allows instrumentation of requests that would be trivially
	// rejected, but potentially increases risk of denial-of-service if instrumentation setup causes load or consumes memory.
	AllowEarlyInstrumentation bool

	// Prevents this server from finding and using RequestInstrumentationManager components
	DisableInstrumentationAutoWire bool

	// A component able to instrument a web service request in some way
	InstrumentationManager instrument.RequestInstrumentationManager

	// How many concurrent requests the server should allow before returning 'too busy' responses to subsequent requests.
	MaxConcurrent int64

	// The HTTP status code returned with 'too busy responses'. Normally 503
	TooBusyStatus int

	// A component able to examine an incoming request and determine which version of functionality is being requested.
	VersionExtractor httpendpoint.RequestedVersionExtractor

	// A component able to use data in an HTTP request's headers to populate a context
	IDContextBuilder IdentifiedRequestContextBuilder

	state  ioc.ComponentState
	server *http.Server
}

// Container allows Granitic to inject a reference to the IOC container
func (h *HTTPServer) Container(container *ioc.ComponentContainer) {
	h.componentContainer = container
}

func (h *HTTPServer) registerProvider(endPointProvider httpendpoint.Provider) {

	for _, method := range endPointProvider.SupportedHTTPMethods() {
		var compiledRegex *regexp.Regexp
		var err error

		pattern := endPointProvider.RegexPattern()

		if compiledRegex, err = regexp.Compile(pattern); err != nil {
			h.FrameworkLogger.LogErrorf("Unable to compile regular expression from pattern %s: %s", pattern, err.Error())
		}

		h.FrameworkLogger.LogTracef("Registering %s %s", pattern, method)

		rp := registeredProvider{endPointProvider, compiledRegex}

		providersForMethod := h.registeredProvidersByMethod[method]

		if providersForMethod == nil {
			providersForMethod = make([]*registeredProvider, 1)
			providersForMethod[0] = &rp
			h.registeredProvidersByMethod[method] = providersForMethod
		} else {
			h.registeredProvidersByMethod[method] = append(providersForMethod, &rp)
		}
	}

}

// StartComponent Finds and registers any available components that implement httpendpoint.Provider (normally instances of
// handler.WsHandler) unless auto finding of handlers is disabled. The server does not actually start listening for
// requests until the IoC container calls AllowAccess.
func (h *HTTPServer) StartComponent() error {

	if h.state != ioc.StoppedState {
		return nil
	}

	h.state = ioc.StartingState
	h.registeredProvidersByMethod = make(map[string][]*registeredProvider)

	if h.AutoFindHandlers {
		for _, component := range h.componentContainer.AllComponents() {

			name := component.Name

			if provider, found := component.Instance.(httpendpoint.Provider); found && provider.AutoWireable() {
				h.FrameworkLogger.LogDebugf("Found Provider %s", name)
				h.registerProvider(provider)
			}
		}
	} else if h.unregisteredProviders != nil {

		for _, provider := range h.unregisteredProviders {

			h.registerProvider(provider)

		}

	} else {
		return errors.New("auto finding of handlers is disabled, but handlers have not been set manually")
	}

	if h.AbnormalStatusWriter == nil {

		return errors.New("no AbnormalStatusWriter set - make sure you have enabled a web services facility")
	}

	if h.InstrumentationManager == nil {
		//No RequestInstrumentationManager component injected, use a 'noop' implementation
		h.FrameworkLogger.LogDebugf("No RequestInstrumentationManager set. Using noop implementation")
		h.InstrumentationManager = new(noopRequestInstrumentationManager)
	}

	h.state = ioc.AwaitingAccessState

	return nil
}

// Suspend causes all subsequent new HTTP requests to receive a 'too busy' response until Resume is called.
func (h *HTTPServer) Suspend() error {

	if h.state != ioc.RunningState {
		return nil
	}

	h.state = ioc.SuspendedState

	return nil
}

// Resume allows subsequent requests to be processed normally (reverses the effect of calling Suspend).
func (h *HTTPServer) Resume() error {

	if h.state != ioc.SuspendedState {
		return nil
	}

	h.state = ioc.RunningState

	return nil
}

// AllowAccess starts the server listening on the configured address and port. Returns an error if the port is already in use.
func (h *HTTPServer) AllowAccess() error {

	if h.state != ioc.AwaitingAccessState {
		return nil
	}

	sm := http.NewServeMux()
	sm.Handle("/", http.HandlerFunc(h.handleAll))

	sv := new(http.Server)
	sv.Handler = sm

	listenAddress := fmt.Sprintf("%s:%d", h.Address, h.Port)

	//Check if the address is already in use
	if ln, err := net.Listen("tcp", listenAddress); err == nil {
		ln.Close()
	} else {
		return err
	}

	sv.Addr = listenAddress

	go sv.ListenAndServe()

	h.server = sv

	h.FrameworkLogger.LogInfof("Listening on %d", h.Port)

	h.state = ioc.RunningState

	return nil
}

// SetProvidersManually manually injects a set of httpendpoint.HTTPEndpointProviders when auto finding is disabled.
func (h *HTTPServer) SetProvidersManually(p map[string]httpendpoint.Provider) {
	h.unregisteredProviders = p
}

func (h *HTTPServer) writeAbnormal(ctx context.Context, status int, wrw *httpendpoint.HTTPResponseWriter, err ...error) {

	if len(err) > 0 {
		h.FrameworkLogger.LogErrorf(err[0].Error())
	}

	state := ws.NewAbnormalState(status, wrw)
	if err := h.AbnormalStatusWriter.WriteAbnormalStatus(ctx, state); err != nil {
		h.FrameworkLogger.LogErrorfCtx(ctx, err.Error())
	}

}

func (h *HTTPServer) handleAll(res http.ResponseWriter, req *http.Request) {

	var instrumentor instrument.Instrumentor
	var endInstrumentation func()
	ctx, cancelFunc := context.WithCancel(req.Context())
	defer cancelFunc()

	if h.AllowEarlyInstrumentation {
		ctx, instrumentor, endInstrumentation = h.InstrumentationManager.Begin(ctx, res, req)
		defer endInstrumentation()
	}

	wrw := httpendpoint.NewHTTPResponseWriter(res)

	if h.state != ioc.RunningState {
		// The HTTP server is suspended - reject the request
		h.writeAbnormal(ctx, h.TooBusyStatus, wrw)
		return
	}

	rCount := atomic.AddInt64(&h.ActiveRequests, 1)
	defer atomic.AddInt64(&h.ActiveRequests, -1)

	if h.MaxConcurrent > 0 && rCount > h.MaxConcurrent {
		// Too many requests already being processed
		h.writeAbnormal(ctx, h.TooBusyStatus, wrw)
		return
	}

	if instrumentor == nil {
		ctx, instrumentor, endInstrumentation = h.InstrumentationManager.Begin(ctx, res, req)
		defer endInstrumentation()
	}

	var requestID string
	received := time.Now()

	if h.IDContextBuilder != nil {

		if idCtx, err := h.IDContextBuilder.WithIdentity(ctx, req); err == nil {
			ctx = idCtx.(context.Context)
			requestID = h.IDContextBuilder.ID(idCtx)

			ctx = ws.StoreRequestIDFunction(ctx, h.IDContextBuilder.ID)

			instrumentor.Amend(instrument.RequestID, requestID)

			if h.FrameworkLogger.IsLevelEnabled(logging.Trace) {
				h.FrameworkLogger.LogTracef("Request ID: %s\n", requestID)
			}

		} else {

			//Something went wrong trying to use HTTP data to identify a context - treat as a bad request (400)
			h.writeAbnormal(ctx, http.StatusBadRequest, wrw, err)

			return

		}
	}

	matched := false

	providersByMethod := h.registeredProvidersByMethod[req.Method]

	path := req.URL.Path

	h.FrameworkLogger.LogTracef("Finding provider to handle %s %s from %d providers", path, req.Method, len(providersByMethod))

	for _, handlerPattern := range providersByMethod {

		pattern := handlerPattern.Pattern

		h.FrameworkLogger.LogTracef("Testing %s", pattern.String())

		if pattern.MatchString(path) && h.versionMatch(instrumentor, req, handlerPattern.Provider) {
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
		h.AccessLogWriter.LogRequest(ctx, req, wrw, &received, &finished)
	}

}

func (h *HTTPServer) versionMatch(ri instrument.Instrumentor, r *http.Request, p httpendpoint.Provider) bool {

	if h.VersionExtractor == nil || !p.VersionAware() {
		return true
	}

	version := h.VersionExtractor.Extract(r)

	ri.Amend(instrument.RequestVersion, version)

	return p.SupportsVersion(version)

}

// PrepareToStop sets state to Stopping. Any subsequent requests will receive a 'too busy response'
func (h *HTTPServer) PrepareToStop() {
	h.state = ioc.StoppingState

	if h.server != nil {
		h.server.Shutdown(context.Background())
	}

}

// ReadyToStop returns false is the server is currently handling any requests.
func (h *HTTPServer) ReadyToStop() (bool, error) {
	a := h.ActiveRequests
	ready := a <= 0

	if ready {
		return true, nil
	}

	return false, fmt.Errorf("HTTP server listening on %d is still serving %d request(s)", h.Port, a)

}

// Stop sets state to Stopped. Any subsequent requests will receive a 'too busy response'. Note that the HTTP
// server is still listening on its configured port and address.
func (h *HTTPServer) Stop() error {

	h.state = ioc.StoppedState

	if h.server != nil {
		h.server.Close()
	}

	return nil
}
