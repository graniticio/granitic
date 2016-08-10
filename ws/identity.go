package ws

import "net/http"

type WsIdentifier interface {
	Identify(req *http.Request) WsIdentity
}

const authenticated = "Authenticated"

type WsIdentity map[string]interface{}

func (ws WsIdentity) AuthenticationStatus(b bool) {
	ws[authenticated] = b
}

func (ws WsIdentity) Authenticated() bool {

	a := ws[authenticated]

	return a != nil && a.(bool)

}
