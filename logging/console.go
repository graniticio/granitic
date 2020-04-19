// Copyright 2016-2020 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package logging

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
)

// ConsoleErrorLogger is an implementation of logging.Logger that writes errors and fatal messages to the console/command line using Go's fmt.Printf function.
// Messages at all other levels are ignored. This implementation is used by Granitic's command line tools and is not
// recommended for use in user applications but can by useful for unit tests.
type ConsoleErrorLogger struct {
	w io.Writer
}

// LogTracefCtx is ignored - messages sent to this method are discarded.
func (l *ConsoleErrorLogger) LogTracefCtx(ctx context.Context, s string, a ...interface{}) {
	return
}

// LogDebugfCtx is ignored - messages sent to this method are discarded.
func (l *ConsoleErrorLogger) LogDebugfCtx(ctx context.Context, s string, a ...interface{}) {
	return
}

// LogInfofCtx is ignored - messages sent to this method are discarded.
func (l *ConsoleErrorLogger) LogInfofCtx(ctx context.Context, s string, a ...interface{}) {
	return
}

// LogWarnfCtx is ignored - messages sent to this method are discarded.
func (l *ConsoleErrorLogger) LogWarnfCtx(ctx context.Context, s string, a ...interface{}) {
	return
}

// LogErrorfCtx uses fmt.printf to write the supplied message to the console.
func (l *ConsoleErrorLogger) LogErrorfCtx(ctx context.Context, format string, a ...interface{}) {

	if l.w == nil {
		l.w = os.Stdout
	}

	fmt.Fprintf(l.w, format+"\n", a...)
}

// LogErrorfCtxWithTrace uses fmt.printf to write the supplied message to the console and appends a stack trace.
func (l *ConsoleErrorLogger) LogErrorfCtxWithTrace(ctx context.Context, format string, a ...interface{}) {
	l.LogErrorfWithTrace(format, a...)
}

// LogFatalfCtx uses fmt.printf to write the supplied message to the console.
func (l *ConsoleErrorLogger) LogFatalfCtx(ctx context.Context, format string, a ...interface{}) {
	l.LogErrorf(format, a...)
}

// LogAtLevelfCtx uses fmt.printf to write the supplied message to the console.
func (l *ConsoleErrorLogger) LogAtLevelfCtx(ctx context.Context, level LogLevel, levelLabel string, format string, a ...interface{}) {
	if l.IsLevelEnabled(level) {
		l.LogErrorf(format, a...)
	}
}

// LogTracef is ignored - messages sent to this method are discarded.
func (l *ConsoleErrorLogger) LogTracef(format string, a ...interface{}) {
	return
}

// LogDebugf is ignored - messages sent to this method are discarded.
func (l *ConsoleErrorLogger) LogDebugf(format string, a ...interface{}) {
	return
}

// LogInfof is ignored - messages sent to this method are discarded.
func (l *ConsoleErrorLogger) LogInfof(format string, a ...interface{}) {
	return
}

// LogWarnf is ignored - messages sent to this method are discarded.
func (l *ConsoleErrorLogger) LogWarnf(format string, a ...interface{}) {
	return
}

// LogErrorf uses fmt.Fprintf to write the supplied message to the console (or other writer)
func (l *ConsoleErrorLogger) LogErrorf(format string, a ...interface{}) {

	if l.w == nil {
		l.w = os.Stdout
	}

	fmt.Fprintf(l.w, format+"\n", a...)
}

// LogErrorfWithTrace uses fmt.printf to write the supplied message to the console and appends
// a stack trace.
func (l *ConsoleErrorLogger) LogErrorfWithTrace(format string, a ...interface{}) {
	trace := make([]byte, 2048)
	runtime.Stack(trace, false)

	format = format + "\n%s"
	a = append(a, trace)

	l.LogErrorf(format, a...)
}

// LogFatalf uses fmt.printf to write the supplied message to the console.
func (l *ConsoleErrorLogger) LogFatalf(format string, a ...interface{}) {
	l.LogErrorf(format, a...)
}

// LogAtLevelf uses fmt.printf to write the supplied message to the console.
func (l *ConsoleErrorLogger) LogAtLevelf(level LogLevel, levelLabel string, format string, a ...interface{}) {
	if l.IsLevelEnabled(level) {
		l.LogErrorf(format, a...)
	}
}

// IsLevelEnabled returns true if the supplied level is >= Error
func (l *ConsoleErrorLogger) IsLevelEnabled(level LogLevel) bool {
	return level >= Error
}
