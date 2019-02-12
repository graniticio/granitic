// Copyright 2016-2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

/*
Package iam provides types to develop or integrate with an Identity/Access Management system.

Identity/Access Management (IAM) systems are external components that deal with authenticating (checking credentials) and
authorising (assigning and checking roles and permissions) users and access to a system. There are many third-party IAM
systems available and many developers also choose to implement their own.

As such, Granitic does not attempt to implement an IAM system, but provides types and hooks to integrate existing systems
into the web-service handling workflow.

See also

	ws.WsIdentifier
	ws.WsAccessChecker

*/
package iam

const authenticated = "Authenticated"
const anonymous = "Anonymous"
const loggableUserID = "LoggableUserID"

// Create a new ClientIdentity with the supplied log-friendly version of a user ID. The ClientIdentity will be marked
// as Authenticated and not anonymous
func NewAuthenticatedIdentity(loggableUserID string) ClientIdentity {
	i := make(ClientIdentity)
	i.SetAnonymous(false)
	i.SetAuthenticated(true)
	i.SetLoggableUserID(loggableUserID)

	return i
}

// NewAnonymousIdentity creates a new ClientIdentity for an anonymous user. The ClientIdentity will be marked as non-authenticated,
// anonymous and have a dash (-) as the loggable user ID.
func NewAnonymousIdentity() ClientIdentity {
	i := make(ClientIdentity)
	i.SetAnonymous(true)
	i.SetAuthenticated(false)
	i.SetLoggableUserID("-")

	return i
}

// ClientIdentity is a semi-structured type allowing applications to define their own representation of Identity.
type ClientIdentity map[string]interface{}

// SetAuthenticated marks this as an authenticated (true) or unauthenticated (false) Identity.
func (ci ClientIdentity) SetAuthenticated(b bool) {
	ci[authenticated] = b
}

// Authenticated indicates whether this is an authenticated (true) or unauthenticated (false) Identity.
func (ci ClientIdentity) Authenticated() bool {

	a := ci[authenticated]

	return a != nil && a.(bool)

}

// SetAnonymous called with true marks this as an anonymous Identity (e.g. no user identification was provided or trusted).
func (ci ClientIdentity) SetAnonymous(b bool) {
	ci[anonymous] = b
}

// Anonymous returns true if this Identity had no identifying information (or the provided information was not trusted)
func (ci ClientIdentity) Anonymous() bool {

	a := ci[authenticated]

	return a != nil && a.(bool)

}

// SetLoggableUserID records a string representation of the Identity that is suitable for recording in log files (e.g. a user name or real name).
func (ci ClientIdentity) SetLoggableUserID(s string) {
	ci[loggableUserID] = s
}

// LoggableUserID returns a string representation of the Identity that is suitable for recording in log files.
func (ci ClientIdentity) LoggableUserID() string {

	a := ci[loggableUserID]

	if a == nil {
		return ""
	}

	return a.(string)
}
