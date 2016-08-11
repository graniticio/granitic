package ws

import (
	"net/http"
)

func NewWsResponse(errorFinder ServiceErrorFinder) *WsResponse {
	r := new(WsResponse)
	r.Errors = new(ServiceErrors)
	r.Errors.ErrorFinder = errorFinder

	r.Headers = make(map[string]string)

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
	UserIdentity    WsIdentity
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
	Headers    map[string]string
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
	WriteErrors(errors *ServiceErrors, w http.ResponseWriter) error
	WriteAbnormalStatus(status int, w http.ResponseWriter) error
}

type AbnormalStatusWriter interface {
	WriteAbnormalStatus(status int, w http.ResponseWriter) error
}

func WriteMetaData(w http.ResponseWriter, r *WsResponse, defaultHeaders map[string]string) {

	additionalHeaders := r.Headers

	for k, v := range defaultHeaders {

		if additionalHeaders == nil || additionalHeaders[k] == "" {
			w.Header().Add(k, v)
		}

	}

	if additionalHeaders != nil {
		for k, v := range additionalHeaders {
			w.Header().Add(k, v)
		}
	}

}
