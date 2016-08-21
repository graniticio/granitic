package json

import (
	"encoding/json"
	"github.com/graniticio/granitic/logging"
	"github.com/graniticio/granitic/ws"
	"net/http"
)

type StandardJSONUnmarshaller struct {
	FrameworkLogger logging.Logger
}

func (ju *StandardJSONUnmarshaller) Unmarshall(req *http.Request, wsReq *ws.WsRequest) error {

	err := json.NewDecoder(req.Body).Decode(&wsReq.RequestBody)

	return err

}
