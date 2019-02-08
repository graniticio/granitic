// Copyright 2016-2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package xml

import (
	"bytes"
	"context"
	"encoding/xml"
	"github.com/graniticio/granitic/v2/ws"
	"net/http"
)

// Component wrapper over Go's xml.Unmarshal method
type Unmarshaller struct {
}

// Unmarshall decodes XML into a Go struct using Go's builtin xml.Unmarshal method.
func (um *Unmarshaller) Unmarshall(ctx context.Context, req *http.Request, wsReq *ws.Request) error {
	defer req.Body.Close()

	var b bytes.Buffer
	b.ReadFrom(req.Body)

	err := xml.Unmarshal(b.Bytes(), &wsReq.RequestBody)

	return err
}
