package ws

import "net/http"

type WsIdentifier interface {
	Identify(req *http.Request) WsIdentity
}

const authenticated = "Authenticated"
const anonymous = "Anonymous"
const loggableUserId = "LoggableUserId"

func NewAuthenticatedIdentity(loggableUserId string) WsIdentity {
	i := make(WsIdentity)
	i.SetAnonymous(false)
	i.SetAuthenticated(true)
	i.SetLoggableUserId(loggableUserId)

	return i
}

func NewAnonymousIdentity() WsIdentity {
	i := make(WsIdentity)
	i.SetAnonymous(true)
	i.SetAuthenticated(false)
	i.SetLoggableUserId("-")

	return i
}

type WsIdentity map[string]interface{}

func (ws WsIdentity) SetAuthenticated(b bool) {
	ws[authenticated] = b
}

func (ws WsIdentity) Authenticated() bool {

	a := ws[authenticated]

	return a != nil && a.(bool)

}

func (ws WsIdentity) SetAnonymous(b bool) {
	ws[anonymous] = b
}

func (ws WsIdentity) Anonymous() bool {

	a := ws[authenticated]

	return a != nil && a.(bool)

}

func (ws WsIdentity) SetLoggableUserId(s string) {
	ws[loggableUserId] = s
}

func (ws WsIdentity) LoggableUserId() string {

	a := ws[loggableUserId]

	if a == nil {
		return ""
	} else {
		return a.(string)
	}
}

type WsAccessChecker interface {
	Allowed(r *WsRequest) bool
}
