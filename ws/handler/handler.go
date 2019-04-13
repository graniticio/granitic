// Copyright 2016-2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

/*
Package handler provides the types used to coordinate the processing of a web service request.

The core type in this package is WsHandler. A handler (an instance of WsHandler) must be created for every logical web service
endpoint in your application. The behaviour and configuration of handlers is described in detail at
http://granitic.io/ref/web-service-handlers but a brief description follows.

Declaring handlers

A handler is declared in your component definition file like:

	{
	  "artistHandler": {
		"type": "handler.WsHandler",
		"HTTPMethod": "GET",
		"Logic": "ref:artistLogic",
		"PathPattern": "^/artist/([\\d]+)[/]?$"
	  },

	  "artistLogic": {
		"type": "inventory.ArtistLogic"
	  }
	}

Each handler must have the following before it is considered a valid web service endpoint.

1. A regular expression that will be matched against the path component of incoming HTTP requests.

2. A single HTTP method that it will be responsible for handling. This is generally GET, POST, PUT or DELETE but any
standard or custom HTTP method can be used.

3. A 'logic' component that implements at least WsRequestProcessor (additional WsXXX interfaces can be implemented
to support advanced behaviour) OR has a method with the signature ProcessPayload(ctx context.Context, request *ws.Request, response *ws.Response, payload *YourStruct)

*/
package handler

import (
	"context"
	"errors"
	"fmt"
	"github.com/graniticio/granitic/v2/httpendpoint"
	"github.com/graniticio/granitic/v2/iam"
	"github.com/graniticio/granitic/v2/instrument"
	"github.com/graniticio/granitic/v2/ioc"
	"github.com/graniticio/granitic/v2/logging"
	"github.com/graniticio/granitic/v2/validate"
	"github.com/graniticio/granitic/v2/ws"
	"net/http"
	"reflect"
	"regexp"
)

const processPayloadFunc = "ProcessPayload"

// WsRequestProcessor specifies the minimum required of a component to be considered a 'logic' component suitable for
// use by a WsHandler.
type WsRequestProcessor interface {
	// Process performs the actual 'work' of a web service request. The reponse parameter will be modified according to
	// the output or errors that the web service caller should see.
	Process(ctx context.Context, request *ws.Request, response *ws.Response)
}

// WsPostProcessor is implemented to indicate that an object is interested in observing/modifying a web service request after processing has been completed,
// but before the HTTP response is written. Typical uses are the writing of response headers that are generic to all/most handlers or the recording of metrics.
//
// It is expected that WsPostProcessors may be shared between multiple instances of WsHandler
type WsPostProcessor interface {
	//PostProcess may modify the supplied response object if required.
	PostProcess(ctx context.Context, handlerName string, request *ws.Request, response *ws.Response)
}

// WsPreValidateManipulator is implemented to indicate that an object is interested in observing/modifying a web service request after it has been unmarshalled and parsed, but before automatic and
// application-defined validation takes place.
type WsPreValidateManipulator interface {
	// PreValidate returns true if the supplied request is in a suitable state for processing to continue.
	PreValidate(ctx context.Context, request *ws.Request, errors *ws.ServiceErrors) (proceed bool)
}

// WsRequestValidator is optionally implemented by the same object that is used as a handler's WsRequestProcessor. If implemented, the Validate method will be called
// to determine whether or not a request should proceed to processing.
type WsRequestValidator interface {
	// Validate will add one or more CategorisedServiceError objects to the supplied errors parameter if the request is not suitable for further processing.
	Validate(ctx context.Context, errors *ws.ServiceErrors, request *ws.Request)
}

// WsUnmarshallTarget  is implemented by logic components that are able to create target objects for data from a web service request to be parsed into. For example,
// a web service that supports POST requests will need an object into which the request body can be stored.
type WsUnmarshallTarget interface {
	// UnmarshallTarget returns a pointer to a struct. That struct can be used by the called to parse or map request data (body, query parameters etc) into.
	UnmarshallTarget() interface{}
}

// WsVersionAssessor allows a component to tell Granitic that the object can determine whether or not handler supports a given version of a request.
type WsVersionAssessor interface {
	// SupportsVersion returns true if the named handle is able to support the requested version.
	SupportsVersion(handlerName string, version httpendpoint.RequiredVersion) bool
}

// Templated is implemented by logic components that need to instruct the web services renderer to use a specific template to render
// a response.
type Templated interface {
	// TemplateName returns the unique name of a template to use to render response output.
	TemplateName() string

	// UseWhenError returns true if the template returned by TemplateName should be used errors are present in the response.
	UseWhenError() bool
}

// ErrorTemplate is implemented by logic components that need to instruct the web services renderer to use a specific template to render when
// errors are detected in the response.
type ErrorTemplate interface {
	//ErrorTemplateName returns the unique name of a template to use to render response output.
	ErrorTemplateName() string
}

// WsHandler co-ordinates the processing of a web service request for a particular endpoint.
// Implements ws.Provider
type WsHandler struct {

	// A component able to examine a request and see if the caller is allowed to access this endpoint.
	AccessChecker ws.AccessChecker

	// Whether or not the underlying HTTP request and response writer should be made available to request Logic.
	AllowDirectHTTPAccess bool

	// Whether or not query parameters should be automatically injected into the request body.
	AutoBindQuery bool

	// A component able to use a set of user-defined rules to validate a request.
	AutoValidator *validate.RuleValidator

	// A list of field names on the target object into which path parameters (groups in the request regex) should be bound to.
	BindPathParams []string

	// Check caller's permissions after request has been parsed (true) or before parsing (false).
	CheckAccessAfterParse bool

	// A function able to create an empty initialised struct to use as a target for request binding
	createTarget func() interface{}

	// If true, do not automatically return an error response if errors are found during the parsing and binding phases of request processing.
	DeferFrameworkErrors bool

	// If true, do not automatically return an error response if errors are found during auto validation.
	DeferAutoErrors bool

	// If true, discard the request's query parameters.
	DisableQueryParsing bool

	// If true, discard any path parameters found by match the request URI against the PathMatchPattern regex.
	DisablePathParsing bool

	// An object that provides access to application defined error messages for use during validation.
	ErrorFinder ws.ServiceErrorFinder

	// A map of fields on the request body object and the names of query parameters that should be used to populate them
	FieldQueryParam map[string]string

	// An object that provides access to built-in error messages to use when an error is found during the automated phases of request processing.
	FrameworkErrors *ws.FrameworkErrorGenerator

	// The HTTP method (GET, POST etc) that this handler supports.
	HTTPMethod string

	// A logger injected by the Granitic framework. Note this will be an application logger rather than a framework logger
	// as instances of WsHandler are considered application components.
	Log logging.Logger

	// The object representing the 'logic' behind this handler.
	Logic interface{}

	// A component injected by the Granitic framework that can map text representations of query and path parameters to Go
	// and Granitic types.
	ParamBinder *ws.ParamBinder

	// A regex that will be matched against inbound request paths to check if this handler should be used to service the request.
	PathPattern string

	// A component that might want to modify a response after it has been processed by the supplied Logic component.
	PostProcessor WsPostProcessor

	// A compponent that might want to modify a request after it has been parsed, but before it has been validated.
	PreValidateManipulator WsPreValidateManipulator

	// Stop the framework automatically adding this handler to an HTTP server.
	PreventAutoWiring bool

	// A component injected by the Granitic framework that writes the response from this handler to an HTTP response.
	ResponseWriter ws.ResponseWriter

	// Whether on not the caller needs to be authenticated (using a ws.Identifier) in order to access the logic behind this handler.
	RequireAuthentication bool

	// A component injected by the Granitic framework that can extract the body of the incoming HTTP request into a Go struct.
	Unmarshaller ws.Unmarshaller

	// A component that can examine a request to determine the calling user/service's identity.
	UserIdentifier ws.Identifier

	// A component that can check if this handler supports the version of functionality required by the caller.
	VersionAssessor   WsVersionAssessor
	bindPathParams    bool
	bindQuery         bool
	httpMethods       []string
	componentName     string
	pathRegex         *regexp.Regexp
	state             ioc.ComponentState
	validationEnabled bool
	validator         WsRequestValidator
	genericProcessor  WsRequestProcessor
}

// ProvideErrorFinder receives a component that can be used to map error codes to categorised errors.
func (wh *WsHandler) ProvideErrorFinder(finder ws.ServiceErrorFinder) {

	if wh.ErrorFinder == nil {
		wh.ErrorFinder = finder
	}
}

// ServeHTTP is the entry point called by the HTTP server once it has been determined that this handler instance
// is the correct one to handle the incoming request.
func (wh *WsHandler) ServeHTTP(ctx context.Context, w *httpendpoint.HTTPResponseWriter, req *http.Request) context.Context {

	defer func() {
		if r := recover(); r != nil {
			wh.Log.LogErrorfCtxWithTrace(ctx, "Panic recovered while trying process a request or write its response %s", r)
			wh.writePanicResponse(ctx, r, w)
		}
	}()

	if ri := instrument.InstrumentorFromContext(ctx); ri != nil {
		//This request is being instrumented, let the instrumentation have access to this handler
		ri.Amend(instrument.Handler, wh)
	}

	wsReq := new(ws.Request)
	wsReq.HTTPMethod = req.Method
	wsReq.ServingHandler = wh.ComponentName()

	if wh.AllowDirectHTTPAccess {
		da := new(ws.DirectHTTPAccess)
		da.Request = req
		da.ResponseWriter = w

		wsReq.UnderlyingHTTP = da
	}

	//Try to identify and/or authenticate the caller
	var okay bool

	if okay, ctx = wh.identifyAndAuthenticate(ctx, w, req, wsReq); !okay {

		return ctx
	}

	//Check caller has permission to use this resource
	if !wh.CheckAccessAfterParse && !wh.checkAccess(ctx, w, wsReq) {
		return ctx
	}

	//Unmarshall body, query parameters and path parameters
	wh.unmarshall(ctx, req, wsReq)
	wh.processQueryParams(ctx, req, wsReq)
	wh.processPathParams(req, wsReq)

	if wsReq.HasFrameworkErrors() && !wh.DeferFrameworkErrors {
		wh.handleFrameworkErrors(ctx, w, wsReq)
		return ctx
	}

	//Check caller has permission to use this resource
	if wh.CheckAccessAfterParse && !wh.checkAccess(ctx, w, wsReq) {
		return ctx
	}

	//Validate request
	var errors ws.ServiceErrors
	errors.ErrorFinder = wh.ErrorFinder

	wh.validateRequest(ctx, wsReq, &errors)

	if errors.HasErrors() {
		wh.writeErrorResponse(ctx, &errors, w, wsReq)

		return ctx
	}

	//Execute logic
	wh.process(ctx, wsReq, w)

	return ctx
}

func (wh *WsHandler) validateRequest(ctx context.Context, wsReq *ws.Request, errors *ws.ServiceErrors) {
	if wh.validationEnabled {
		proceed := true

		if wh.PreValidateManipulator != nil {
			proceed = wh.PreValidateManipulator.PreValidate(ctx, wsReq, errors)
		}

		if !proceed {
			return
		}

		body := wsReq.RequestBody
		ov := wh.AutoValidator

		if body == nil && ov != nil {
			wh.Log.LogWarnfCtx(ctx, "Request body is nil but an ObjectValidator is set on the handler. Automatic body validation skipped.")
		} else if ov != nil {
			sc := new(validate.SubjectContext)
			sc.Subject = body

			fe, err := ov.Validate(ctx, sc)

			if err != nil {

				wh.Log.LogErrorfCtx(ctx, "Problem encountered during automatic body validation %v", err)

				ce := wh.FrameworkErrors.HTTPError(http.StatusInternalServerError)
				errors.AddError(ce)
				return
			}

			if fe != nil && len(fe) > 0 {

				ef := wh.ErrorFinder

				for _, e := range fe {

					for _, code := range e.ErrorCodes {

						ce := ef.Find(code)
						ce.Field = e.Field
						errors.AddError(ce)

					}

				}
			}

		}

		if errors.HasErrors() && (!wh.DeferAutoErrors) {
			return
		}

		if wh.validator != nil {
			wh.validator.Validate(ctx, errors, wsReq)
		}
	}

}

func (wh *WsHandler) unmarshall(ctx context.Context, req *http.Request, wsReq *ws.Request) {

	var uf func() interface{}

	if targetSource, found := wh.Logic.(WsUnmarshallTarget); found {
		//Logic component implements WsUnmarshallTarget - use that to create target
		uf = targetSource.UnmarshallTarget
	} else if wh.createTarget != nil {
		//A function has been provided to generate targets
		uf = wh.createTarget
	} else {
		//No way of creating a target
		return
	}

	target := uf()
	wsReq.RequestBody = target

	if req.ContentLength == 0 {
		return
	}

	err := wh.Unmarshaller.Unmarshall(ctx, req, wsReq)

	if err != nil {

		wh.Log.LogDebugfCtx(ctx, "Error unmarshalling request body for %s %s %s", req.URL.Path, req.Method, err)

		m, c := wh.FrameworkErrors.MessageCode(ws.UnableToParseRequest)

		f := ws.NewUnmarshallFrameworkError(m, c)
		wsReq.AddFrameworkError(f)
	}

}

func (wh *WsHandler) processPathParams(req *http.Request, wsReq *ws.Request) {

	if wh.DisablePathParsing {
		return
	}

	re := wh.pathRegex
	params := re.FindStringSubmatch(req.URL.Path)
	wsReq.PathParams = params[1:]

	if wh.bindPathParams && len(wsReq.PathParams) > 0 {
		pp := ws.NewParamsForPath(wh.BindPathParams, wsReq.PathParams)
		wh.ParamBinder.BindPathParameters(wsReq, pp)
	}

}

func (wh *WsHandler) processQueryParams(ctx context.Context, req *http.Request, wsReq *ws.Request) {

	if wh.DisableQueryParsing {
		return
	}

	values := req.URL.Query()
	wsReq.QueryParams = ws.NewParamsForQuery(values)

	if wh.bindQuery {
		if wsReq.RequestBody == nil {
			wh.Log.LogErrorfCtx(ctx, "Query parameter binding is enabled, but no target available to bind into. Does your Logic component implement the WsUnmarshallTarget interface?")
			return
		}

		if wh.AutoBindQuery {
			wh.ParamBinder.AutoBindQueryParameters(wsReq)
		} else {
			wh.ParamBinder.BindQueryParameters(wsReq, wh.FieldQueryParam)
		}

	}

}

func (wh *WsHandler) checkAccess(ctx context.Context, w *httpendpoint.HTTPResponseWriter, wsReq *ws.Request) bool {

	ac := wh.AccessChecker

	if ac == nil {
		return true
	}

	allowed := ac.Allowed(ctx, wsReq)

	if allowed {
		return true
	}

	state := ws.NewAbnormalState(http.StatusForbidden, w)
	state.Identity = wsReq.UserIdentity
	state.WsRequest = wsReq

	wh.ResponseWriter.Write(ctx, state, ws.Abnormal)
	return false

}

func (wh *WsHandler) identifyAndAuthenticate(ctx context.Context, w *httpendpoint.HTTPResponseWriter, req *http.Request, wsReq *ws.Request) (bool, context.Context) {

	var i iam.ClientIdentity

	if wh.UserIdentifier != nil {

		i, ctx = wh.UserIdentifier.IDentify(ctx, req)
		wsReq.UserIdentity = i

		if wh.RequireAuthentication && !i.Authenticated() {

			state := ws.NewAbnormalState(http.StatusUnauthorized, w)
			state.Identity = wsReq.UserIdentity
			state.WsRequest = wsReq

			wh.ResponseWriter.Write(ctx, state, ws.Abnormal)
			return false, ctx
		}

	}

	if wsReq.UserIdentity == nil {
		wsReq.UserIdentity = iam.NewAnonymousIdentity()
	}

	return true, ctx

}

// SupportedHTTPMethods returns the HTTP method that this handler supports. Returns an array in order to
// implement Provider, but will always be a single element array.
func (wh *WsHandler) SupportedHTTPMethods() []string {
	if len(wh.httpMethods) > 0 {
		return wh.httpMethods
	}

	return []string{wh.HTTPMethod}
}

// RegexPattern returns the unparsed regex pattern that should be applicaed to the path of incoming requests to
// see if this handler should handle the request.
func (wh *WsHandler) RegexPattern() string {
	return wh.PathPattern
}

// VersionAware returns true if this handler can be considered when a user requests a specific version of functionality.
func (wh *WsHandler) VersionAware() bool {
	return wh.VersionAssessor != nil
}

// SupportsVersion returns true if this handler supports the version of functionality requested by the caller. Defers to the
// component injected into this handler's VersionAssessor field.
func (wh *WsHandler) SupportsVersion(version httpendpoint.RequiredVersion) bool {
	return wh.VersionAssessor.SupportsVersion(wh.ComponentName(), version)
}

// AutoWireable returns true if this handler should be automatically registered with any instances of httpserver.HTTPServer
// that are running in the application.
func (wh *WsHandler) AutoWireable() bool {
	return !wh.PreventAutoWiring
}

func (wh *WsHandler) handleFrameworkErrors(ctx context.Context, w *httpendpoint.HTTPResponseWriter, wsReq *ws.Request) {

	var se ws.ServiceErrors
	se.HTTPStatus = http.StatusBadRequest

	for _, fe := range wsReq.FrameworkErrors {
		se.AddNewError(ws.Client, fe.Code, fe.Message)
	}

	wh.writeErrorResponse(ctx, &se, w, wsReq)

}

func (wh *WsHandler) process(ctx context.Context, request *ws.Request, w *httpendpoint.HTTPResponseWriter) {

	defer func() {
		if r := recover(); r != nil {
			wh.Log.LogErrorfCtxWithTrace(ctx, "Panic recovered while trying process a request or write its response %s", r)
			wh.writePanicResponse(ctx, r, w)
		}
	}()

	wsRes := ws.NewResponse(wh.ErrorFinder)

	if wh.genericProcessor != nil {
		//Logic component implements WsRequestProcessor
		wh.genericProcessor.Process(ctx, request, wsRes)
	} else {
		//Call the ProcessPayload method via reflection which allows us to pass in the body of the response as a typed object
		//without knowing the type at compile time
		method := reflect.ValueOf(wh.Logic).MethodByName(processPayloadFunc)

		va := []reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(request), reflect.ValueOf(wsRes), reflect.ValueOf(request.RequestBody)}

		method.Call(va)
	}

	if wh.PostProcessor != nil {
		wh.PostProcessor.PostProcess(ctx, wh.ComponentName(), request, wsRes)
	}

	state := new(ws.ProcessState)
	state.Identity = request.UserIdentity
	state.HTTPResponseWriter = w
	state.WsResponse = wsRes
	state.WsRequest = request
	state.Status = wsRes.HTTPStatus

	// Template based response writing
	if tr, found := wh.Logic.(Templated); found {
		wsRes.Template = tr.TemplateName()
	}

	var err error

	if wsRes.HTTPStatus < 300 {
		err = wh.ResponseWriter.Write(ctx, state, ws.Normal)
	} else {
		err = wh.ResponseWriter.Write(ctx, state, ws.Abnormal)
	}

	if err != nil {
		wh.Log.LogErrorfCtx(ctx, "Problem writing response: %s", err.Error())
	}

}

func (wh *WsHandler) writeErrorResponse(ctx context.Context, errors *ws.ServiceErrors, w *httpendpoint.HTTPResponseWriter, wsReq *ws.Request) {

	l := wh.Log

	defer func() {
		if r := recover(); r != nil {
			l.LogErrorfCtxWithTrace(ctx, "Panic recovered while trying to write a response that was already in error %v", r)
		}
	}()

	state := new(ws.ProcessState)
	state.ServiceErrors = errors
	state.WsRequest = wsReq
	state.HTTPResponseWriter = w

	// Template based response writing
	if tr, found := wh.Logic.(Templated); found {

		res := new(ws.Response)
		state.WsResponse = res

		if tr.UseWhenError() {
			res.Template = tr.TemplateName()
		}

		if et, found := wh.Logic.(ErrorTemplate); found {
			res.Template = et.ErrorTemplateName()
		}

	}

	err := wh.ResponseWriter.Write(ctx, state, ws.Error)

	if err != nil {
		l.LogErrorfCtx(ctx, "Problem writing an HTTP response that was already in error", err)
	}

}

func (wh *WsHandler) writePanicResponse(ctx context.Context, r interface{}, w *httpendpoint.HTTPResponseWriter) {

	state := ws.NewAbnormalState(http.StatusInternalServerError, w)

	wh.ResponseWriter.Write(ctx, state, ws.Abnormal)

	wh.Log.LogErrorfCtx(ctx, "Panic recovered but error response served. %s", r)

}

// StartComponent is called by the IoC container. Verifies that the minimum set of fields and components and fields
// have been set (see top of this GoDoc page) and that the configuration of the handler is valid and consistent.
func (wh *WsHandler) StartComponent() error {

	if wh.state != ioc.StoppedState {
		return nil
	}

	wh.state = ioc.StartingState

	if wh.PathPattern == "" || wh.HTTPMethod == "" || wh.Logic == nil {
		return errors.New("handlers must have at least a PathPattern string, HTTPMethod string and Logic component set")
	}

	if wh.AutoValidator != nil && wh.ErrorFinder == nil {
		return errors.New("you must set ErrorFinder if you set AutoValidator. Check that the ServiceErrorManager facility is enabled")
	}

	if err := wh.checkLogicComponent(); err != nil {
		return err
	}

	validator, found := wh.Logic.(WsRequestValidator)

	wh.validationEnabled = found || wh.AutoValidator != nil

	if found {
		wh.validator = validator
	}

	wh.bindQuery = wh.AutoBindQuery || (wh.FieldQueryParam != nil && len(wh.FieldQueryParam) > 0)

	if !wh.DisablePathParsing {

		wh.bindPathParams = len(wh.BindPathParams) > 0

		r, err := regexp.Compile(wh.PathPattern)

		if err != nil {
			return err
		}

		wh.pathRegex = r
	}

	if wh.DeferAutoErrors && wh.validator == nil {
		return errors.New("if you want to defer errors generated during auto validation, your logic component must implement WsRequestValidator")
	}

	if err := wh.validateProcessPayload(); err == nil {
		//The logic attached to this handler has a ProcessPayload method. Extract a func for creating empty structs to pass to it
		wh.createTarget = wh.extractFactoryFromLogic()
	}

	wh.state = ioc.RunningState

	return nil

}

func (wh *WsHandler) checkLogicComponent() error {
	if rp, found := wh.Logic.(WsRequestProcessor); found {

		wh.genericProcessor = rp
		return nil
	}

	return wh.validateProcessPayload()
}

func (wh *WsHandler) validateProcessPayload() error {

	err := fmt.Errorf("Logic compoonent must either implement WsRequestProcessor or have method %s(ctx context.Context, request *ws.Request, response *ws.Response, payload *YourStruct)", processPayloadFunc)

	if wh.Logic == nil {
		return err
	}

	//Logic component doesn't implement WsRequestProcessor - must instead have a method called ProcessPayload
	method := reflect.ValueOf(wh.Logic).MethodByName(processPayloadFunc)

	if !method.IsValid() {
		return err
	}

	t := method.Type()

	//Quick check of parameter counts on the method signature
	if t.NumIn() != 4 || t.NumOut() != 0 {
		return err
	}

	//Check first arg is context
	if t.In(0).String() != "context.Context" {
		return err
	}

	//Check second arg is *ws.Request
	if t.In(1) != reflect.TypeOf(new(ws.Request)) {
		return err
	}

	//Check third arg is *ws.Response
	if t.In(2) != reflect.TypeOf(new(ws.Response)) {
		return err
	}

	//Check fourth arg is a pointer to a struct
	fourthArg := t.In(3)

	if fourthArg.Kind() != reflect.Ptr || fourthArg.Elem().Kind() != reflect.Struct {
		return err
	}

	return nil
}

func (wh *WsHandler) extractFactoryFromLogic() func() interface{} {

	if err := wh.validateProcessPayload(); err != nil {
		return nil
	}

	method := reflect.ValueOf(wh.Logic).MethodByName(processPayloadFunc)
	mt := method.Type()

	targetArg := mt.In(3)

	targetArg.Elem()

	t := targetArg.Elem()

	return func() interface{} {
		return reflect.New(t).Interface()
	}

}

// ComponentName implements ComponentNamer.ComponentName
func (wh *WsHandler) ComponentName() string {
	return wh.componentName
}

// SetComponentName implements ComponentNamer.SetComponentName
func (wh *WsHandler) SetComponentName(name string) {
	wh.componentName = name
}
