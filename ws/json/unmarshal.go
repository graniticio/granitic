package json

import (
	"encoding/json"
	"github.com/graniticio/granitic/logging"
	"github.com/graniticio/granitic/ws"
	"golang.org/x/net/context"
	"net/http"
)

type StandardJSONUnmarshaller struct {
	FrameworkLogger logging.Logger
}

func (ju *StandardJSONUnmarshaller) Unmarshall(ctx context.Context, req *http.Request, wsReq *ws.WsRequest) error {

	err := json.NewDecoder(req.Body).Decode(&wsReq.RequestBody)

	req.Body.Close()

	return err

}
