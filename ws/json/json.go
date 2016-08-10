package json

import (
	"encoding/json"
	"github.com/graniticio/granitic/logging"
	"github.com/graniticio/granitic/ws"
	"net/http"
)

type DefaultJsonUnmarshaller struct {
	FrameworkLogger logging.Logger
}

func (jdu *DefaultJsonUnmarshaller) Unmarshall(req *http.Request, wsReq *ws.WsRequest) error {

	err := json.NewDecoder(req.Body).Decode(&wsReq.RequestBody)

	return err

}
