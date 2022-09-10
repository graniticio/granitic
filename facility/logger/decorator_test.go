// Copyright 2016-2022 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package logger

import (
	"github.com/graniticio/granitic/v3/ioc"
	"github.com/graniticio/granitic/v3/logging"
	"github.com/graniticio/granitic/v3/test"
	"testing"
)

type AcceptLog struct {
	Log logging.Logger
}

type RejectLogStruct struct {
	Log map[string]string
}

type RejectLogPrim struct {
	Log int
}

func TestMatcher(t *testing.T) {

	d := new(applicationLogDecorator)
	d.FrameworkLogger = new(logging.ConsoleErrorLogger)

	c := new(ioc.Component)
	c.Name = "Match"
	c.Instance = new(AcceptLog)

	test.ExpectBool(t, d.OfInterest(c), true)

	c.Name = "NoMatch"
	c.Instance = new(RejectLogStruct)

	test.ExpectBool(t, d.OfInterest(c), false)

	c.Name = "NoMatchPrim"
	c.Instance = new(RejectLogPrim)

	test.ExpectBool(t, d.OfInterest(c), false)

}
