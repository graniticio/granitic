// Copyright 2016-2020 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package logging

import (
	"context"
	"github.com/graniticio/granitic/v2/instance"
	"github.com/graniticio/granitic/v2/test"
	"testing"
	"time"
)

func TestStdoutLoggerCanBeBuilt(t *testing.T) {

	l := NewStdoutLogger(Trace)

	l.LogTracef("MESSAGE")

	l = NewStdoutLogger(Trace, "PREFIX")

	l.LogTracef("MESSAGE")

}

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

func TestGraniticLoggerNoErrors(t *testing.T) {

	l := new(GraniticLogger)

	l.global = &globalLogSource{level: Trace}
	l.localLogThreshhold = Trace

	l.formatter = new(testMessageFomatter)

	if !l.IsLevelEnabled(Trace) {
		t.Error("Trace not enabled")
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
	//l.logAtLevelCtx(ctx, Trace, "TRACE", "")
}

func TestCreateAnonymousLoggerNoErrors(t *testing.T) {

	l := CreateAnonymousLogger("ANON", Fatal)

	l.LogInfof("TEST")
}

type testMessageFomatter struct {
}

func (tmf *testMessageFomatter) SetInstanceID(i *instance.Identifier) {

}

func (tmf *testMessageFomatter) Format(ctx context.Context, levelLabel, loggerName, message string) string {
	return message
}

func (tmf *testMessageFomatter) SetContextFilter(cf ContextFilter) {

}

func TestDeferLogging(t *testing.T) {

	l := new(GraniticLogger)

	l.global = &globalLogSource{level: Trace}
	l.localLogThreshhold = Trace

	l.formatter = new(testMessageFomatter)
	l.deferring = true

	if !l.IsLevelEnabled(Trace) {
		t.Error("Trace not enabled")
	}

	ctx := context.Background()

	dc := new(deferCount)
	l.deferLogger = dc

	l.LogInfofCtx(ctx, "INFO1")

	if dc.count != 1 {
		t.FailNow()
	}

	/*l.LogAtLevelfCtx(ctx, Info,"INFO", "INFO2")

	if dc.count != 2 {
		t.FailNow()
	}

	l.deferring = false

	l.LogInfofCtx(ctx, "INFO3")

	if dc.count != 2 {
		t.FailNow()
	}*/

}

type deferCount struct {
	count int
}

func (dc *deferCount) DeferLog(levelLabel string, level LogLevel, message string, when time.Time, logger *GraniticLogger) {
	dc.count++
}
