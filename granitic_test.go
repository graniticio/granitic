// Copyright 2018 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package granitic

import (
	"github.com/graniticio/granitic/cmd/grnc-bind/binder"
	"github.com/graniticio/granitic/config"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/logging"
	"testing"
)

func TestStartWithSettings(t *testing.T) {
	frameworkModifiers := make(map[string]map[string]string)
	protoComponents := make([]*ioc.ProtoComponent, 0)

	bic := binder.SerialiseBuiltinConfig()

	pc := ioc.NewProtoComponents(protoComponents, frameworkModifiers, &bic)

	is := new(config.InitialSettings)
	is.FrameworkLogLevel = logging.Fatal
	is.DryRun = true
	is.Configuration = []string{"resource/test/config/simple.json"}
	StartGraniticWithSettings(pc, is)

}
