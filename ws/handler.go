package ws

import (
	"github.com/graniticio/granitic/logging"
	"net/http"
	"regexp"
)

//Implements HttpEndpointProvider
type WsHandler struct {
	Unmarshaller          WsUnmarshaller
	HttpMethod            string
	HttpMethods           []string
	PathMatchPattern      string
	Logic                 WsRequestProcessor
	ResponseWriter        WsResponseWriter
	Log                   logging.Logger
	ErrorFinder           ServiceErrorFinder
	FrameworkErrors       *FrameworkErrorGenerator
	DisableQueryParsing   bool
	DisablePathParsing    bool
	DeferFrameworkErrors  bool
	RequireAuthentication bool
	FieldQueryParam       map[string]string
	BindPathParams        []string
	ParamBinder           *ParamBinder
	UserIdentifier        WsIdentifier
	AccessChecker         WsAccessChecker
	CheckAccessAfterParse bool
	AutoBindQuery         bool
	validate              bool
	validator             WsRequestValidator
	bindQuery             bool
	bindPathParams        bool
	pathRegex             *regexp.Regexp
	componentName         string
}

func (wh *WsHandler) ProvideErrorFinder(finder ServiceErrorFinder) {
	wh.ErrorFinder = finder
}

//HttpEndpointProvider
func (wh *WsHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	defer func() {
		if r := recover(); r != nil {
			wh.Log.LogErrorfWithTrace("Panic recovered while trying process a request or write its response %s", r)
			wh.writePanicResponse(r, w)
		}
	}()

	wsReq := new(WsRequest)
	wsReq.HttpMethod = req.Method

	//Try to identify and/or authenticate the caller
	if !wh.identifyAndAuthenticate(w, req, wsReq) {
		return
	}

	//Check caller has permission to use this resource
	if !wh.CheckAccessAfterParse && !wh.checkAccess(w, wsReq) {
		return
	}

	//Unmarshall body, query parameters and path parameters
	wh.unmarshall(req, wsReq)
	wh.processQueryParams(req, wsReq)
	wh.processPathParams(req, wsReq)

	if wsReq.HasFrameworkErrors() && !wh.DeferFrameworkErrors {
		wh.handleFrameworkErrors(w, wsReq)
		return
	}

	//Check caller has permission to use this resource
	if wh.CheckAccessAfterParse && !wh.checkAccess(w, wsReq) {
		return
	}

	//Validate request
	var errors ServiceErrors
	errors.ErrorFinder = wh.ErrorFinder

	if wh.validate {
		wh.validator.Validate(&errors, wsReq)
	}

	if errors.HasErrors() {
		wh.writeErrorResponse(&errors, w)

		return
	}

	//Execute logic
	wh.process(wsReq, w)

}

func (wh *WsHandler) unmarshall(req *http.Request, wsReq *WsRequest) {

	targetSource, found := wh.Logic.(WsUnmarshallTarget)

	if found {
		target := targetSource.UnmarshallTarget()
		wsReq.RequestBody = target

		if req.ContentLength == 0 {
			return
		}

		err := wh.Unmarshaller.Unmarshall(req, wsReq)

		if err != nil {

			wh.Log.LogDebugf("Error unmarshalling request body for %s %s %s", req.URL.Path, req.Method, err)

			m, c := wh.FrameworkErrors.MessageCode(UnableToParseRequest)

			f := NewUnmarshallWsFrameworkError(m, c)
			wsReq.AddFrameworkError(f)
		}

	}
}

func (wh *WsHandler) processPathParams(req *http.Request, wsReq *WsRequest) {

	if wh.DisablePathParsing {
		return
	}

	re := wh.pathRegex
	params := re.FindStringSubmatch(req.URL.Path)
	wsReq.PathParams = params[1:]

	if wh.bindPathParams && len(wsReq.PathParams) > 0 {
		pp := NewWsParamsForPath(wh.BindPathParams, wsReq.PathParams)
		wh.ParamBinder.AutoBindPathParameters(wsReq, pp)
	}

}

func (wh *WsHandler) processQueryParams(req *http.Request, wsReq *WsRequest) {

	if wh.DisableQueryParsing {
		return
	}

	values := req.URL.Query()
	wsReq.QueryParams = NewWsParamsForQuery(values)

	if wh.bindQuery {
		if wsReq.RequestBody == nil {
			wh.Log.LogErrorf("Query parameter binding is enabled, but no target available to bind into. Does your Logic component implement the WsUnmarshallTarget interface?")
			return
		}

		if wh.AutoBindQuery {
			wh.ParamBinder.AutoBindQueryParameters(wsReq)
		} else {
			wh.ParamBinder.BindQueryParameters(wsReq, wh.FieldQueryParam)
		}

	}

}

func (wh *WsHandler) checkAccess(w http.ResponseWriter, wsReq *WsRequest) bool {

	ac := wh.AccessChecker

	if ac == nil {
		return true
	}

	allowed := ac.Allowed(wsReq)

	if allowed {
		return true
	} else {
		wh.ResponseWriter.WriteAbnormalStatus(http.StatusForbidden, w)
		return false
	}
}

func (wh *WsHandler) identifyAndAuthenticate(w http.ResponseWriter, req *http.Request, wsReq *WsRequest) bool {

	if wh.UserIdentifier != nil {
		i := wh.UserIdentifier.Identify(req)
		wsReq.UserIdentity = i

		if wh.RequireAuthentication && !i.Authenticated() {
			wh.ResponseWriter.WriteAbnormalStatus(http.StatusUnauthorized, w)
			return false
		}

	}

	return true

}

//HttpEndpointProvider
func (wh *WsHandler) SupportedHttpMethods() []string {
	if len(wh.HttpMethods) > 0 {
		return wh.HttpMethods
	} else {
		return []string{wh.HttpMethod}
	}
}

//HttpEndpointProvider
func (wh *WsHandler) RegexPattern() string {
	return wh.PathMatchPattern
}

func (wh *WsHandler) handleFrameworkErrors(w http.ResponseWriter, wsReq *WsRequest) {

	var se ServiceErrors
	se.HttpStatus = http.StatusBadRequest

	for _, fe := range wsReq.FrameworkErrors {
		se.AddNewError(Client, fe.Code, fe.Message)
	}

	wh.writeErrorResponse(&se, w)

}

func (wh *WsHandler) process(jsonReq *WsRequest, w http.ResponseWriter) {

	defer func() {
		if r := recover(); r != nil {
			wh.Log.LogErrorfWithTrace("Panic recovered while trying process a request or write its response %s", r)
			wh.writePanicResponse(r, w)
		}
	}()

	wsRes := NewWsResponse(wh.ErrorFinder)
	wh.Logic.Process(jsonReq, wsRes)

	wh.ResponseWriter.Write(wsRes, w)

}

func (wh *WsHandler) writeErrorResponse(errors *ServiceErrors, w http.ResponseWriter) {

	l := wh.Log

	defer func() {
		if r := recover(); r != nil {
			l.LogErrorfWithTrace("Panic recovered while trying to write a response that was already in error %s", r)
		}
	}()

	err := wh.ResponseWriter.WriteErrors(errors, w)

	if err != nil {
		l.LogErrorf("Problem writing an HTTP response that was already in error", err)
	}

}

func (wh *WsHandler) writePanicResponse(r interface{}, w http.ResponseWriter) {

	wh.ResponseWriter.WriteAbnormalStatus(http.StatusInternalServerError, w)

	wh.Log.LogErrorf("Panic recovered but error response served. %s", r)

}

func (wh *WsHandler) StartComponent() error {

	validator, found := wh.Logic.(WsRequestValidator)

	wh.validate = found

	if found {
		wh.validator = validator
	}

	wh.bindQuery = wh.AutoBindQuery || (wh.FieldQueryParam != nil && len(wh.FieldQueryParam) > 0)

	if !wh.DisablePathParsing {

		wh.bindPathParams = len(wh.BindPathParams) > 0

		r, err := regexp.Compile(wh.PathMatchPattern)

		if err != nil {
			return err
		} else {
			wh.pathRegex = r
		}

	}

	return nil

}

func (wh *WsHandler) ComponentName() string {
	return wh.componentName
}

func (wh *WsHandler) SetComponentName(name string) {
	wh.componentName = name
}
