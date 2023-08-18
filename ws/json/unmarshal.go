// Copyright 2016-2023 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package json

import (
	"context"
	"encoding/json"
	"github.com/graniticio/granitic/v2/logging"
	"github.com/graniticio/granitic/v2/ws"
	"net/http"
)

// Unmarshaller is a component wrapper over Go's JSON decoder.
type Unmarshaller struct {
	FrameworkLogger logging.Logger
}

// Unmarshall uses Go's JSON decoder to parse a HTTP request body into a struct.
func (ju *Unmarshaller) Unmarshall(ctx context.Context, req *http.Request, wsReq *ws.Request) error {
	defer req.Body.Close()

	err := json.NewDecoder(req.Body).Decode(&wsReq.RequestBody)

	return err

}
