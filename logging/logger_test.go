// Copyright 2016-2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package logging

import (
	"context"
	"github.com/graniticio/granitic/v2/test"
	"testing"
)

func TestThresholdDetection(t *testing.T) {

	g := new(globalLogSource)

	lal := new(GraniticLogger)
	lal.global = g

	g.level = All
	lal.localLogThreshhold = All

	test.ExpectBool(t, lal.IsLevelEnabled(Debug), true)

	g.level = Error
	test.ExpectBool(t, lal.IsLevelEnabled(Debug), false)

	lal.localLogThreshhold = Debug

	test.ExpectBool(t, lal.IsLevelEnabled(Debug), true)

	g.level = Trace
	test.ExpectBool(t, lal.IsLevelEnabled(Trace), false)

	lal.localLogThreshhold = All
	test.ExpectBool(t, lal.IsLevelEnabled(Trace), true)

}

func TestNilLogging(t *testing.T) {

	var l Logger
	l = new(NullLogger)

	if l.IsLevelEnabled(Trace) {
		t.FailNow()
	}

	ctx := context.Background()

	l.LogDebugf("")
	l.LogAtLevelf(Trace, "TRACE", "")
	l.LogAtLevelfCtx(ctx, Trace, "TRACE", "")
	l.LogDebugfCtx(ctx, "")
	l.LogErrorf("")
	l.LogErrorfCtx(ctx, "")
	l.LogErrorfCtxWithTrace(ctx, "")
	l.LogErrorfWithTrace("")
	l.LogErrorfWithTrace("")
	l.LogFatalf("")
	l.LogFatalfCtx(ctx, "")
	l.LogInfof("")
	l.LogInfofCtx(ctx, "")
	l.LogTracef("")
	l.LogTracefCtx(ctx, "")
	l.LogWarnf("")
	l.LogWarnfCtx(ctx, "")

}
