// Copyright 2016-2020 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package logging

import "context"

// NullLogger is a logger that ignores all supplied messages and always returns false for 'is level enabled' tests.
// Used when the application logging facility has been disabled
type NullLogger struct {
}

// LogTracefCtx does nothing
func (n NullLogger) LogTracefCtx(ctx context.Context, format string, a ...interface{}) {

}

// LogDebugfCtx does nothing
func (n NullLogger) LogDebugfCtx(ctx context.Context, format string, a ...interface{}) {
}

// LogInfofCtx does nothing
func (n NullLogger) LogInfofCtx(ctx context.Context, format string, a ...interface{}) {
}

// LogWarnfCtx does nothing
func (n NullLogger) LogWarnfCtx(ctx context.Context, format string, a ...interface{}) {
}

// LogErrorfCtx does nothing
func (n NullLogger) LogErrorfCtx(ctx context.Context, format string, a ...interface{}) {
}

// LogErrorfCtxWithTrace does nothing
func (n NullLogger) LogErrorfCtxWithTrace(ctx context.Context, format string, a ...interface{}) {
}

// LogFatalfCtx does nothing
func (n NullLogger) LogFatalfCtx(ctx context.Context, format string, a ...interface{}) {
}

// LogAtLevelfCtx does nothing
func (n NullLogger) LogAtLevelfCtx(ctx context.Context, level LogLevel, levelLabel string, format string, a ...interface{}) {
}

// LogTracef does nothing
func (n NullLogger) LogTracef(format string, a ...interface{}) {
}

// LogDebugf does nothing
func (n NullLogger) LogDebugf(format string, a ...interface{}) {
}

// LogInfof does nothing
func (n NullLogger) LogInfof(format string, a ...interface{}) {
}

// LogWarnf does nothing
func (n NullLogger) LogWarnf(format string, a ...interface{}) {
}

// LogErrorf does nothing
func (n NullLogger) LogErrorf(format string, a ...interface{}) {
}

// LogErrorfWithTrace does nothing
func (n NullLogger) LogErrorfWithTrace(format string, a ...interface{}) {
}

// LogFatalf does nothing
func (n NullLogger) LogFatalf(format string, a ...interface{}) {
}

// LogAtLevelf does nothing
func (n NullLogger) LogAtLevelf(level LogLevel, levelLabel string, format string, a ...interface{}) {
}

// IsLevelEnabled always returns false
func (n NullLogger) IsLevelEnabled(level LogLevel) bool {
	return false
}
