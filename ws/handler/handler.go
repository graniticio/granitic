package handler

import (
	"errors"
	"github.com/graniticio/granitic/httpendpoint"
	"github.com/graniticio/granitic/iam"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/logging"
	"github.com/graniticio/granitic/validate"
	"github.com/graniticio/granitic/ws"
	"golang.org/x/net/context"
	"net/http"
	"regexp"
)

// Indicates that an object is able to continue the processing of a web service request after the automated phases of
// parsing, binding, authenticating, authorising and auto-validating have been completed.
type WsRequestProcessor interface {
	Process(ctx context.Context, request *ws.WsRequest, response *ws.WsResponse)
}

// Indicates that an object is interested in observing/modifying a web service request after processing has been completed,
// but before the HTTP response is written. Typical uses are the writing of response headers that are generic to all/most handlers or the recording of metrics.
//
// It is expected that WsPostProcessors may be shared between multiple instances of WsHandler
type WsPostProcessor interface {
	PostProcess(ctx context.Context, handlerName string, request *ws.WsRequest, response *ws.WsResponse)
}

// Indicates that an object is interested in observing/modifying a web service request after it has been unmarshalled and parsed, but before automatic and
// application-defined validation takes place. If an error is encountered, or if the object decides that processing should be halted, it is expected that
// the implementing object adds one or more errors to the ws.WsResponse and returns false.
type WsPreValidateManipulator interface {
	PreValidate(ctx context.Context, request *ws.WsRequest, errors *ws.ServiceErrors) (proceed bool)
}

type WsRequestValidator interface {
	Validate(ctx context.Context, errors *ws.ServiceErrors, request *ws.WsRequest)
}

type WsUnmarshallTarget interface {
	UnmarshallTarget() interface{}
}

// Indicates that an object can determine whether or a handler supports a given version of a request.
type WsVersionAssessor interface {
	SupportsVersion(handlerName string, version httpendpoint.RequiredVersion) bool
}

//  WsHandler co-ordinates the processing of a web service request for a particular endpoint.
// Implements ws.HttpEndpointProvider
type WsHandler struct {
	AccessChecker          ws.WsAccessChecker          //
	AllowDirectHTTPAccess  bool                        // Whether or not the underlying HTTP request and response writer should be made available to request Logic.
	AutoBindQuery          bool                        // Whether or not query parameters should be automatically injected into the request body.
	AutoValidator          *validate.RuleValidator     //
	BindPathParams         []string                    // A list of fields on the request body that should be populated using elements of the request path.
	CheckAccessAfterParse  bool                        // Check caller's permissions after request has been parsed (true) or before parsing (false).
	DeferFrameworkErrors   bool                        // If true, do not automatically return an error response if errors are found during the parsing and binding phases of request processing.
	DeferAutoErrors        bool                        // If true, do not automatically return an error response if errors are found during auto validation.
	DisableQueryParsing    bool                        // If true, discard the request's query parameters.
	DisablePathParsing     bool                        // If true, discard any path parameters found by match the request URI against the PathMatchPattern regex.
	ErrorFinder            ws.ServiceErrorFinder       // An object that provides access to application defined error messages for use during validation.
	FieldQueryParam        map[string]string           // A map of fields on the request body object and the names of query parameters that should be used to populate them
	FrameworkErrors        *ws.FrameworkErrorGenerator // An object that provides access to built-in error messages to use when an error is found during the automated phases of request processing.
	HTTPMethod             string                      // The HTTP method (GET, POST etc) that this handler supports.
	Log                    logging.Logger              //
	Logic                  WsRequestProcessor          // The object representing the 'logic' behind this handler.
	ParamBinder            *ws.ParamBinder             //
	PathPattern            string                      // A regex that will be matched against inbound request paths to check if this handler should be used to service the request.
	PostProcessor          WsPostProcessor             //
	PreValidateManipulator WsPreValidateManipulator    //
	PreventAutoWiring      bool                        // Stop the framwework automatically adding this handler to an HTTP server
	ResponseWriter         ws.WsResponseWriter         //
	RequireAuthentication  bool                        // Whether on not the caller needs to be authenticated (using a ws.WsIdentifier) in order to access the logic behind this handler.
	Unmarshaller           ws.WsUnmarshaller           //
	UserIdentifier         ws.WsIdentifier             //
	VersionAssessor        WsVersionAssessor           //
	bindPathParams         bool
	bindQuery              bool
	httpMethods            []string
	componentName          string
	pathRegex              *regexp.Regexp
	state                  ioc.ComponentState
	validationEnabled      bool
	validator              WsRequestValidator
}

func (wh *WsHandler) ProvideErrorFinder(finder ws.ServiceErrorFinder) {

	if wh.ErrorFinder == nil {
		wh.ErrorFinder = finder
	}
}

//HttpEndpointProvider
func (wh *WsHandler) ServeHTTP(ctx context.Context, w *httpendpoint.HTTPResponseWriter, req *http.Request) context.Context {

	defer func() {
		if r := recover(); r != nil {
			wh.Log.LogErrorfCtxWithTrace(ctx, "Panic recovered while trying process a request or write its response %s", r)
			wh.writePanicResponse(ctx, r, w)
		}
	}()

	wsReq := new(ws.WsRequest)
	wsReq.HttpMethod = req.Method
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

func (wh *WsHandler) validateRequest(ctx context.Context, wsReq *ws.WsRequest, errors *ws.ServiceErrors) {
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

				ce := wh.FrameworkErrors.HttpError(http.StatusInternalServerError)
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

func (wh *WsHandler) unmarshall(ctx context.Context, req *http.Request, wsReq *ws.WsRequest) {

	targetSource, found := wh.Logic.(WsUnmarshallTarget)

	if found {
		target := targetSource.UnmarshallTarget()
		wsReq.RequestBody = target

		if req.ContentLength == 0 {
			return
		}

		err := wh.Unmarshaller.Unmarshall(ctx, req, wsReq)

		if err != nil {

			wh.Log.LogDebugfCtx(ctx, "Error unmarshalling request body for %s %s %s", req.URL.Path, req.Method, err)

			m, c := wh.FrameworkErrors.MessageCode(ws.UnableToParseRequest)

			f := ws.NewUnmarshallWsFrameworkError(m, c)
			wsReq.AddFrameworkError(f)
		}

	}
}

func (wh *WsHandler) processPathParams(req *http.Request, wsReq *ws.WsRequest) {

	if wh.DisablePathParsing {
		return
	}

	re := wh.pathRegex
	params := re.FindStringSubmatch(req.URL.Path)
	wsReq.PathParams = params[1:]

	if wh.bindPathParams && len(wsReq.PathParams) > 0 {
		pp := ws.NewWsParamsForPath(wh.BindPathParams, wsReq.PathParams)
		wh.ParamBinder.BindPathParameters(wsReq, pp)
	}

}

func (wh *WsHandler) processQueryParams(ctx context.Context, req *http.Request, wsReq *ws.WsRequest) {

	if wh.DisableQueryParsing {
		return
	}

	values := req.URL.Query()
	wsReq.QueryParams = ws.NewWsParamsForQuery(values)

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

func (wh *WsHandler) checkAccess(ctx context.Context, w *httpendpoint.HTTPResponseWriter, wsReq *ws.WsRequest) bool {

	ac := wh.AccessChecker

	if ac == nil {
		return true
	}

	allowed := ac.Allowed(ctx, wsReq)

	if allowed {
		return true
	} else {

		state := ws.NewAbnormalState(http.StatusForbidden, w)
		state.Identity = wsReq.UserIdentity
		state.WsRequest = wsReq

		wh.ResponseWriter.Write(ctx, state, ws.Abnormal)
		return false
	}
}

func (wh *WsHandler) identifyAndAuthenticate(ctx context.Context, w *httpendpoint.HTTPResponseWriter, req *http.Request, wsReq *ws.WsRequest) (bool, context.Context) {

	var i iam.ClientIdentity

	if wh.UserIdentifier != nil {

		i, ctx = wh.UserIdentifier.Identify(ctx, req)
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

//HttpEndpointProvider
func (wh *WsHandler) SupportedHttpMethods() []string {
	if len(wh.httpMethods) > 0 {
		return wh.httpMethods
	} else {
		return []string{wh.HTTPMethod}
	}
}

//HttpEndpointProvider
func (wh *WsHandler) RegexPattern() string {
	return wh.PathPattern
}

//HttpEndpointProvider
func (wh *WsHandler) VersionAware() bool {
	return wh.VersionAssessor != nil
}

//HttpEndpointProvider
func (wh *WsHandler) SupportsVersion(version httpendpoint.RequiredVersion) bool {
	return wh.VersionAssessor.SupportsVersion(wh.ComponentName(), version)
}

//HttpEndpointProvider
func (wh *WsHandler) AutoWireable() bool {
	return !wh.PreventAutoWiring
}

func (wh *WsHandler) handleFrameworkErrors(ctx context.Context, w *httpendpoint.HTTPResponseWriter, wsReq *ws.WsRequest) {

	var se ws.ServiceErrors
	se.HttpStatus = http.StatusBadRequest

	for _, fe := range wsReq.FrameworkErrors {
		se.AddNewError(ws.Client, fe.Code, fe.Message)
	}

	wh.writeErrorResponse(ctx, &se, w, wsReq)

}

func (wh *WsHandler) process(ctx context.Context, request *ws.WsRequest, w *httpendpoint.HTTPResponseWriter) {

	defer func() {
		if r := recover(); r != nil {
			wh.Log.LogErrorfCtxWithTrace(ctx, "Panic recovered while trying process a request or write its response %s", r)
			wh.writePanicResponse(ctx, r, w)
		}
	}()

	wsRes := ws.NewWsResponse(wh.ErrorFinder)
	wh.Logic.Process(ctx, request, wsRes)

	if wh.PostProcessor != nil {
		wh.PostProcessor.PostProcess(ctx, wh.ComponentName(), request, wsRes)
	}

	state := new(ws.WsProcessState)
	state.Identity = request.UserIdentity
	state.HTTPResponseWriter = w
	state.WsResponse = wsRes
	state.WsRequest = request
	state.Status = wsRes.HttpStatus

	// Template based response writing
	if tr, found := wh.Logic.(ws.Templated); found {
		wsRes.Template = tr.TemplateName()
	}

	var err error

	if wsRes.HttpStatus < 300 {
		err = wh.ResponseWriter.Write(ctx, state, ws.Normal)
	} else {
		err = wh.ResponseWriter.Write(ctx, state, ws.Abnormal)
	}

	if err != nil {
		wh.Log.LogErrorfCtx(ctx, "Problem writing response: %s", err.Error())
	}

}

func (wh *WsHandler) writeErrorResponse(ctx context.Context, errors *ws.ServiceErrors, w *httpendpoint.HTTPResponseWriter, wsReq *ws.WsRequest) {

	l := wh.Log

	defer func() {
		if r := recover(); r != nil {
			l.LogErrorfCtxWithTrace(ctx, "Panic recovered while trying to write a response that was already in error %v", r)
		}
	}()

	state := new(ws.WsProcessState)
	state.ServiceErrors = errors
	state.WsRequest = wsReq
	state.HTTPResponseWriter = w

	// Template based response writing
	if tr, found := wh.Logic.(ws.Templated); found {

		res := new(ws.WsResponse)
		state.WsResponse = res

		if tr.UseWhenError() {
			res.Template = tr.TemplateName()
		}

		if et, found := wh.Logic.(ws.ErrorTemplate); found {
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

func (wh *WsHandler) StartComponent() error {

	if wh.state != ioc.StoppedState {
		return nil
	}

	wh.state = ioc.StartingState

	if wh.PathPattern == "" || wh.HTTPMethod == "" || wh.Logic == nil {
		return errors.New("Handlers must have at least a PathPattern string, HttpMethod string and Logic component set.")
	}

	if wh.AutoValidator != nil && wh.ErrorFinder == nil {
		return errors.New("You must set ErrorFinder if you set AutoValidator. Is the ServiceErrorManager facility enabled?")
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
		} else {
			wh.pathRegex = r
		}

	}

	if wh.DeferAutoErrors && wh.validator == nil {
		return errors.New("If you want to defer errors generated during auto validation, your logic component must implement WsRequestValidator.")
	}

	wh.state = ioc.RunningState

	return nil

}

func (wh *WsHandler) ComponentName() string {
	return wh.componentName
}

func (wh *WsHandler) SetComponentName(name string) {
	wh.componentName = name
}
