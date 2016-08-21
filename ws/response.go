package ws

import (
	"github.com/graniticio/granitic/httpendpoint"
	"github.com/graniticio/granitic/iam"
	"net/http"
)

type WsOutcome uint

const (
	Normal = iota
	Error
	Abnormal
)

type WsProcessState struct {
	WsRequest          *WsRequest
	WsResponse         *WsResponse
	HTTPResponseWriter *httpendpoint.HTTPResponseWriter
	ServiceErrors      *ServiceErrors
	Identity           iam.ClientIdentity
	Status             int
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

// An object that constructs response headers that are common to all web service requests. These may typically be
// caching instructions or 'processing server' records. Implementations must be extremely cautious when using
// the information in the supplied WsProcess state as some values may be nil.
type WsCommonResponseHeaderBuilder interface {
	BuildHeaders(state *WsProcessState) map[string]string
}

// Interface for components able to convert a set of service errors into a structure suitable for serialisation.
type ErrorFormatter interface {
	FormatErrors(errors *ServiceErrors) interface{}
}

func WriteHeaders(w http.ResponseWriter, headers map[string]string) {

	for k, v := range headers {
		w.Header().Add(k, v)
	}
}

type ResponseWrapper interface {
	WrapResponse(body interface{}, errors interface{}) interface{}
}
