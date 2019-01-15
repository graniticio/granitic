// Copyright 2016-2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package json

import (
	"context"
	"encoding/json"
	"github.com/graniticio/granitic/logging"
	"github.com/graniticio/granitic/ws"
	"net/http"
)

// Component wrapper over Go's JSON decoder.
type StandardJSONUnmarshaller struct {
	FrameworkLogger logging.Logger
}

// Unmarshall uses Go's JSON decoder to parse a HTTP request body into a struct.
func (ju *StandardJSONUnmarshaller) Unmarshall(ctx context.Context, req *http.Request, wsReq *ws.WsRequest) error {
	defer req.Body.Close()

	err := json.NewDecoder(req.Body).Decode(&wsReq.RequestBody)

	return err

}
