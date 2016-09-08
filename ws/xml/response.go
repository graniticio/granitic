package xml

import (
	"github.com/graniticio/granitic/logging"
	"github.com/graniticio/granitic/ws"
	"golang.org/x/net/context"
	"net/http"
)

type StandardXMLResponseWriter struct {
	FrameworkLogger  logging.Logger
	StatusDeterminer ws.HttpStatusCodeDeterminer
	FrameworkErrors  *ws.FrameworkErrorGenerator
	DefaultHeaders   map[string]string
}

func (rw *StandardXMLResponseWriter) Write(ctx context.Context, state *ws.WsProcessState, outcome ws.WsOutcome) error {
	return nil
}

func (rw *StandardXMLResponseWriter) WriteAbnormalStatus(ctx context.Context, state *ws.WsProcessState) error {
	return nil
}

type StandardXmlUnmarshaller struct {
}

func (um *StandardXmlUnmarshaller) Unmarshall(ctx context.Context, req *http.Request, wsReq *ws.WsRequest) error {

	return nil
}
