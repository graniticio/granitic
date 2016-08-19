package ws

import (
	"net/http"
	"github.com/graniticio/granitic/iam"
)

type WsIdentifier interface {
	Identify(req *http.Request) iam.ClientIdentity
}

type WsAccessChecker interface {
	Allowed(r *WsRequest) bool
}
