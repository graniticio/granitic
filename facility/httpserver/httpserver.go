// Copyright 2016-2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

/*
	Package httpserver provides the HTTPServer facility which defines a configurable HTTP server for processing web-service requests.

	The HTTPServer facility provides a server that will listen for HTTP web-service requests and map them to the web-service
	endpoints defined by your application. A full description of how to configure this facility can be found at http://granitic.io/1.0/ref/http-server

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
	"github.com/graniticio/granitic/v2/httpendpoint"
	"github.com/graniticio/granitic/v2/instrument"
	"github.com/graniticio/granitic/v2/ioc"
	"github.com/graniticio/granitic/v2/logging"
	"github.com/graniticio/granitic/v2/ws"
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

	// The name of a component in the IoC container that should be used as an AbnormalStatusWriter if one is not being auto-injected.
	AbnormalStatusWriterName string

	// The number of HTTP requests currently being handled by the server.
	ActiveRequests int64

	// Allow request instrumentation to begin BEFORE too-busy/suspended checks. Allows instrumentation of requests that would be trivially
	// rejected, but potentially increases risk of denial-of-service if instrumentation setup causes load or consumes memory.
	AllowEarlyInstrumentation bool

	// How many concurrent requests the server should allow before returning 'too busy' responses to subsequent requests.
	MaxConcurrent int64

	// The HTTP status code returned with 'too busy responses'. Normally 503
	TooBusyStatus int

	// A component able to examine an incoming request and determine which version of functionality is being requested.
	VersionExtractor httpendpoint.RequestedVersionExtractor

	// A component able to use data in an HTTP request's headers to populate a context
	IdContextBuilder IdentifiedRequestContextBuilder

	// ID of a component that implements instrument.RequestInstrumentationManager
	RequestInstrumentationManagerName string

	state          ioc.ComponentState
	server         *http.Server
	reqInstManager instrument.RequestInstrumentationManager
}

// Implements ioc.ContainerAccessor
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
		return errors.New("Auto finding of handlers is disabled, but handlers have not been set manually.")
	}

	if h.AbnormalStatusWriter == nil {
		return errors.New("No AbnormalStatusWriter set.")
	}

	if rid := h.RequestInstrumentationManagerName; rid == "" {
		//No RequestInstrumentationManager component specified, use a 'noop' implementation
		h.reqInstManager = new(noopRequestInstrumentationManager)
	} else {

		if c := h.componentContainer.ComponentByName(rid); c == nil {
			return fmt.Errorf("No component named %s exists - was specified in the RequestInstrumentationManagerName field", rid)
		} else {

			if rim, found := c.Instance.(instrument.RequestInstrumentationManager); found {
				h.reqInstManager = rim
			} else {
				return fmt.Errorf("Component %s exists, but does not implement instrument.RequestInstrumentationManager. Was specified in the RequestInstrumentationManagerName field", rid)
			}

		}

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

// Resume allows subsequent requests to be processed normally (reserves the effect of calling Suspend).
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
		ctx, instrumentor, endInstrumentation = h.reqInstManager.Begin(ctx, res, req)
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
		ctx, instrumentor, endInstrumentation = h.reqInstManager.Begin(ctx, res, req)
		defer endInstrumentation()
	}

	var requestId string
	received := time.Now()

	if h.IdContextBuilder != nil {

		if idCtx, err := h.IdContextBuilder.WithIdentity(ctx, req); err != nil {

			//Something went wrong trying to use HTTP data to identify a context - treat as a bad request (400)
			h.writeAbnormal(ctx, http.StatusBadRequest, wrw, err)

			return
		} else {

			ctx = idCtx.(context.Context)
			requestId = h.IdContextBuilder.Id(idCtx)

			instrumentor.Amend(instrument.REQUEST_ID, requestId)

			if h.FrameworkLogger.IsLevelEnabled(logging.Trace) {
				h.FrameworkLogger.LogTracef("Request ID: %s\n", requestId)
			}

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
		h.AccessLogWriter.LogRequest(req, wrw, &received, &finished, ctx)
	}

}

func (h *HTTPServer) versionMatch(ri instrument.Instrumentor, r *http.Request, p httpendpoint.Provider) bool {

	if h.VersionExtractor == nil || !p.VersionAware() {
		return true
	}

	version := h.VersionExtractor.Extract(r)

	ri.Amend(instrument.REQUEST_VERSION, version)

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
	} else {

		message := fmt.Sprintf("HTTP server listening on %d is still serving %d request(s)", h.Port, a)

		return false, errors.New(message)
	}
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
