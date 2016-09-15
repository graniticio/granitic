package xml

import (
	"bytes"
	"encoding/xml"
	"github.com/graniticio/granitic/ws"
	"golang.org/x/net/context"
	"net/http"
)

type StandardXmlUnmarshaller struct {
}

func (um *StandardXmlUnmarshaller) Unmarshall(ctx context.Context, req *http.Request, wsReq *ws.WsRequest) error {

	var b bytes.Buffer
	b.ReadFrom(req.Body)

	err := xml.Unmarshal(b.Bytes(), &wsReq.RequestBody)

	req.Body.Close()

	return err
}
