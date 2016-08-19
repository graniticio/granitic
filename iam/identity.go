package iam

const authenticated = "Authenticated"
const anonymous = "Anonymous"
const loggableUserId = "LoggableUserId"

func NewAuthenticatedIdentity(loggableUserId string) ClientIdentity {
	i := make(ClientIdentity)
	i.SetAnonymous(false)
	i.SetAuthenticated(true)
	i.SetLoggableUserId(loggableUserId)

	return i
}

func NewAnonymousIdentity() ClientIdentity {
	i := make(ClientIdentity)
	i.SetAnonymous(true)
	i.SetAuthenticated(false)
	i.SetLoggableUserId("-")

	return i
}

type ClientIdentity map[string]interface{}

func (ci ClientIdentity) SetAuthenticated(b bool) {
	ci[authenticated] = b
}

func (ci ClientIdentity) Authenticated() bool {

	a := ci[authenticated]

	return a != nil && a.(bool)

}

func (ci ClientIdentity) SetAnonymous(b bool) {
	ci[anonymous] = b
}

func (ci ClientIdentity) Anonymous() bool {

	a := ci[authenticated]

	return a != nil && a.(bool)

}

func (ci ClientIdentity) SetLoggableUserId(s string) {
	ci[loggableUserId] = s
}

func (ci ClientIdentity) LoggableUserId() string {

	a := ci[loggableUserId]

	if a == nil {
		return ""
	} else {
		return a.(string)
	}
}
