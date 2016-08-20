package ws

import (
	"net/http"
	"github.com/graniticio/granitic/iam"
)



type WsRequest struct {
	PathParameters  map[string]string
	HttpMethod      string
	RequestBody     interface{}
	QueryParams     *WsParams
	PathParams      []string
	FrameworkErrors []*WsFrameworkError
	populatedFields map[string]bool
	UserIdentity    iam.ClientIdentity
	UnderlyingHTTP  *DirectHTTPAccess
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



type WsUnmarshaller interface {
	Unmarshall(req *http.Request, wsReq *WsRequest) error
}


type DirectHTTPAccess struct {
	ResponseWriter http.ResponseWriter
	Request *http.Request
}


