// Copyright 2016 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package ws

import (
	"github.com/graniticio/granitic/iam"
	"golang.org/x/net/context"
	"net/http"
)

// WsIdentifier is implemented by components that are able to identify a caller based on a raw HTTP request (normally from
// headers and cookies). Implementations of this interface may return a new Context that supersedes the supplied Context.
type WsIdentifier interface {
	// Identify returns information about the caller derived request and a Context that might be different from the supplied Context.
	Identify(ctx context.Context, req *http.Request) (iam.ClientIdentity, context.Context)
}

// WsIdentifier is implemented by components that are able to determine if a caller is allowed to have a request processed.
type WsAccessChecker interface {
	// Allowed returns true if the caller is allowed to have this request processed, false otherwise.
	Allowed(ctx context.Context, r *WsRequest) bool
}
