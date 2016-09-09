package xml

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"github.com/graniticio/granitic/ws"
	"golang.org/x/net/context"
	"net/http"
)

type StandardXmlUnmarshaller struct {
}

func (um *StandardXmlUnmarshaller) Unmarshall(ctx context.Context, req *http.Request, wsReq *ws.WsRequest) error {

	fmt.Println("Un")

	var b bytes.Buffer
	b.ReadFrom(req.Body)

	return xml.Unmarshal(b.Bytes(), &wsReq.RequestBody)
}
