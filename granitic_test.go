// Copyright 2018-2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package granitic

import (
	"github.com/graniticio/granitic/v2/cmd/grnc-bind/binder"
	"github.com/graniticio/granitic/v2/config"
	"github.com/graniticio/granitic/v2/ioc"
	"github.com/graniticio/granitic/v2/logging"
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
