package ws

import (
	"github.com/graniticio/granitic/iam"
	"golang.org/x/net/context"
	"net/http"
)

type WsIdentifier interface {
	Identify(ctx context.Context, req *http.Request) (iam.ClientIdentity, context.Context)
}

type WsAccessChecker interface {
	Allowed(ctx context.Context, r *WsRequest) bool
}
