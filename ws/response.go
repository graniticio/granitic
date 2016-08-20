package ws

import (
	"net/http"
	"github.com/graniticio/granitic/httpendpoint"
)


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
	Write(res *WsResponse, w *httpendpoint.HTTPResponseWriter) error
	WriteErrors(errors *ServiceErrors, w *httpendpoint.HTTPResponseWriter) error
	WriteAbnormalStatus(status int, w *httpendpoint.HTTPResponseWriter) error
}

type AbnormalStatusWriter interface {
	WriteAbnormalStatus(status int, w *httpendpoint.HTTPResponseWriter) error
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
