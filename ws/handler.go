package ws

import (
	"fmt"
	"github.com/graniticio/granitic/logging"
	"net/http"
	"regexp"
)

//Implements HttpEndpointProvider
type WsHandler struct {
	Unmarshaller         WsUnmarshaller
	HttpMethod           string
	HttpMethods          []string
	PathMatchPattern     string
	Logic                WsRequestProcessor
	ResponseWriter       WsResponseWriter
	ErrorResponseWriter  WsAbnormalResponseWriter
	Log                  logging.Logger
	StatusDeterminer     HttpStatusCodeDeterminer
	ErrorFinder          ServiceErrorFinder
	FrameworkErrors		 *FrameworkErrorGenerator
	RevealPanicDetails   bool
	DisableQueryParsing  bool
	DisablePathParsing   bool
	DeferFrameworkErrors bool
	BindQueryParams      map[string]string
	BindPathParams       []string
	ParamBinder          *ParamBinder
	AutoBindQuery        bool
	validate             bool
	validator            WsRequestValidator
	bindQuery            bool
	bindPathParams       bool
	pathRegex            *regexp.Regexp
}

func (wh *WsHandler) ProvideErrorFinder(finder ServiceErrorFinder) {
	wh.ErrorFinder = finder
}

//HttpEndpointProvider
func (wh *WsHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	defer func() {
		if r := recover(); r != nil {
			wh.writePanicResponse(r, w)
		}
	}()

	wsReq := new(WsRequest)
	wsReq.HttpMethod = req.Method

	err := wh.unmarshall(req, wsReq)

	if err != nil {
		wh.handleUnmarshallError(err, w, wsReq)
		return
	}

	wh.processQueryParams(req, wsReq)
	wh.processPathParams(req, wsReq)

	var errors ServiceErrors
	errors.ErrorFinder = wh.ErrorFinder

	if wh.validate {
		wh.validator.Validate(&errors, wsReq)
	}

	if errors.HasErrors() {
		wh.writeErrorResponse(&errors, w)
	} else {

		wh.process(wsReq, w)

	}

}

func (wh *WsHandler) unmarshall(req *http.Request, wsReq *WsRequest) error {

	targetSource, found := wh.Logic.(WsUnmarshallTarget)

	if found {
		target := targetSource.UnmarshallTarget()
		wsReq.RequestBody = target

		if req.ContentLength == 0 {
			return nil
		} else {
			return wh.Unmarshaller.Unmarshall(req, wsReq)
		}
	}

	return nil
}

func (wh *WsHandler) processPathParams(req *http.Request, wsReq *WsRequest) {

	if wh.DisablePathParsing {
		return
	}

	re := wh.pathRegex
	params := re.FindStringSubmatch(req.URL.Path)
	wsReq.PathParams = params[1:]

	if wh.bindPathParams {
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
		}

	}

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

func (wh *WsHandler) handleUnmarshallError(err error, w http.ResponseWriter, wsReq *WsRequest) {
	wh.Log.LogWarnf("Error unmarshalling request body %s", err)

	if wh.DeferFrameworkErrors {
		//Add a framework error for a validator to pick up later
		f := NewUnmarshallWsFrameworkError(err.Error())
		wsReq.AddFrameworkError(f)

	} else {

		var se ServiceErrors
		se.HttpStatus = http.StatusBadRequest

		e := wh.FrameworkErrors.Error(UnableToParseRequest, Client)
		se.AddError(e)

		wh.writeErrorResponse(&se, w)
	}

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

	errors := wsRes.Errors

	if errors.HasErrors() {
		wh.writeErrorResponse(errors, w)

	} else {
		status := wh.StatusDeterminer.DetermineCode(wsRes)
		w.WriteHeader(status)
		wh.ResponseWriter.Write(wsRes, w)
	}

}

func (wh *WsHandler) writeErrorResponse(errors *ServiceErrors, w http.ResponseWriter) {

	l := wh.Log

	defer func() {
		if r := recover(); r != nil {
			l.LogErrorfWithTrace("Panic recovered while trying to write a response that was already in error %s", r)
		}
	}()

	status := wh.StatusDeterminer.DetermineCodeFromErrors(errors)

	err := wh.ErrorResponseWriter.WriteWithErrors(status, errors, w)

	if err != nil {
		l.LogErrorf("Problem writing an HTTP response that was already in error", err)
	}

}

func (wh *WsHandler) writePanicResponse(r interface{}, w http.ResponseWriter) {

	var se ServiceErrors
	se.HttpStatus = http.StatusInternalServerError

	var message string

	if wh.RevealPanicDetails {
		message = fmt.Sprintf("Unhandled error %s", r)

	} else {
		message = "A unexpected error occured while processing this request."
	}

	wh.Log.LogErrorf("Panic recovered but error response served. %s", r)

	se.AddNewError(Unexpected, "UNXP", message)

	wh.writeErrorResponse(&se, w)
}

func (wh *WsHandler) StartComponent() error {

	validator, found := wh.Logic.(WsRequestValidator)

	wh.validate = found

	if found {
		wh.validator = validator
	}

	wh.bindQuery = wh.AutoBindQuery || (wh.BindQueryParams != nil && len(wh.BindQueryParams) > 0)

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
