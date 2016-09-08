package ws

import (
	"github.com/graniticio/granitic/iam"
	"github.com/graniticio/granitic/types"
	"golang.org/x/net/context"
	"net/http"
)

type WsRequest struct {
	HttpMethod      string
	RequestBody     interface{}
	QueryParams     *WsParams
	PathParams      []string
	FrameworkErrors []*WsFrameworkError
	populatedFields types.StringSet
	UserIdentity    iam.ClientIdentity
	UnderlyingHTTP  *DirectHTTPAccess
	ServingHandler  string
}

func (wsr *WsRequest) HasFrameworkErrors() bool {
	return len(wsr.FrameworkErrors) > 0
}

func (wsr *WsRequest) AddFrameworkError(f *WsFrameworkError) {
	wsr.FrameworkErrors = append(wsr.FrameworkErrors, f)
}

func (wsr *WsRequest) RecordFieldAsBound(fieldName string) {
	if wsr.populatedFields == nil {
		wsr.populatedFields = new(types.OrderedStringSet)
	}

	wsr.populatedFields.Add(fieldName)
}

func (wsr *WsRequest) WasFieldBound(fieldName string) bool {
	return wsr.populatedFields.Contains(fieldName)
}

func (wsr *WsRequest) BoundFields() types.StringSet {

	if wsr.populatedFields == nil {
		return types.NewOrderedStringSet([]string{})
	} else {
		return wsr.populatedFields
	}

}

type WsUnmarshaller interface {
	Unmarshall(ctx context.Context, req *http.Request, wsReq *WsRequest) error
}

type DirectHTTPAccess struct {
	ResponseWriter http.ResponseWriter
	Request        *http.Request
}
