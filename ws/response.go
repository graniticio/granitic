package ws

import (
	"net/http"
	"github.com/graniticio/granitic/httpendpoint"
	"github.com/graniticio/granitic/iam"
)

type WsOutcome uint

const (
	Normal = iota
	Error
	Abnormal
)

type WsProcessState struct {
	WsRequest *WsRequest
	WsResponse *WsResponse
	HTTPResponseWriter *httpendpoint.HTTPResponseWriter
	ServiceErrors *ServiceErrors
	Identity iam.ClientIdentity
	Status int
}

func NewAbnormalState(status int, w *httpendpoint.HTTPResponseWriter) *WsProcessState {
	state := new(WsProcessState)
	state.Status = status
	state.HTTPResponseWriter = w

	return state
}

type WsResponse struct {
	HttpStatus int
	Body       interface{}
	Errors     *ServiceErrors
	Headers    map[string]string
}

func NewWsResponse(errorFinder ServiceErrorFinder) *WsResponse {
	r := new(WsResponse)
	r.Errors = new(ServiceErrors)
	r.Errors.ErrorFinder = errorFinder

	r.Headers = make(map[string]string)

	return r
}

type WsResponseWriter interface {
	Write(state *WsProcessState, outcome WsOutcome) error
}

type AbnormalStatusWriter interface {
	WriteAbnormalStatus(state *WsProcessState) error
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
