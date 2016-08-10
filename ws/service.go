package ws

import (
	"net/http"
)

func NewWsResponse(errorFinder ServiceErrorFinder) *WsResponse {
	r := new(WsResponse)
	r.Errors = new(ServiceErrors)
	r.Errors.ErrorFinder = errorFinder

	return r
}

type WsRequest struct {
	PathParameters  map[string]string
	HttpMethod      string
	RequestBody     interface{}
	QueryParams     *WsParams
	PathParams      []string
	FrameworkErrors []*WsFrameworkError
	populatedFields map[string]bool
	UserIdentity    *WsIdentity
}

func (wsr *WsRequest) HasFrameworkErrors() bool {
	return len(wsr.FrameworkErrors) > 0
}

func (wsr *WsRequest) AddFrameworkError(f *WsFrameworkError) {
	wsr.FrameworkErrors = append(wsr.FrameworkErrors, f)
}

func (wsr *WsRequest) RecordFieldAsPopulated(fieldName string) {
	if wsr.populatedFields == nil {
		wsr.populatedFields = make(map[string]bool)
	}

	wsr.populatedFields[fieldName] = true
}

func (wsr *WsRequest) WasFieldPopulated(fieldName string) bool {
	return wsr.populatedFields[fieldName] != false
}

type WsResponse struct {
	HttpStatus int
	Body       interface{}
	Errors     *ServiceErrors
}

type WsRequestProcessor interface {
	Process(request *WsRequest, response *WsResponse)
}

type WsRequestValidator interface {
	Validate(errors *ServiceErrors, request *WsRequest)
}

type WsUnmarshallTarget interface {
	UnmarshallTarget() interface{}
}

type WsUnmarshaller interface {
	Unmarshall(req *http.Request, wsReq *WsRequest) error
}

type WsResponseWriter interface {
	Write(res *WsResponse, w http.ResponseWriter) error
}

type WsAbnormalResponseWriter interface {
	WriteWithErrors(status int, errors *ServiceErrors, w http.ResponseWriter) error
}
