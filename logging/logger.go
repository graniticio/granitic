/*
Package logging provides a component-based logging framework for user and built-in Granitic components.
*/
package logging

import (
	"fmt"
	"golang.org/x/net/context"
	"runtime"
)

type Logger interface {
	LogTracefCtx(ctx context.Context, format string, a ...interface{})
	LogDebugfCtx(ctx context.Context, format string, a ...interface{})
	LogInfofCtx(ctx context.Context, format string, a ...interface{})
	LogWarnfCtx(ctx context.Context, format string, a ...interface{})
	LogErrorfCtx(ctx context.Context, format string, a ...interface{})
	LogErrorfCtxWithTrace(ctx context.Context, format string, a ...interface{})
	LogFatalfCtx(ctx context.Context, format string, a ...interface{})
	LogAtLevelfCtx(ctx context.Context, level LogLevel, levelLabel string, format string, a ...interface{})

	LogTracef(format string, a ...interface{})
	LogDebugf(format string, a ...interface{})
	LogInfof(format string, a ...interface{})
	LogWarnf(format string, a ...interface{})
	LogErrorf(format string, a ...interface{})
	LogErrorfWithTrace(format string, a ...interface{})
	LogFatalf(format string, a ...interface{})
	LogAtLevelf(level LogLevel, levelLabel string, format string, a ...interface{})
	IsLevelEnabled(level LogLevel) bool
}

type LevelAwareLogger struct {
	globalLogThreshold LogLevel
	localLogThreshhold LogLevel
	loggerName         string
	writers            []LogWriter
	formatter          *LogMessageFormatter
}

func (lal *LevelAwareLogger) UpdateWritersAndFormatter(w []LogWriter, f *LogMessageFormatter) {
	lal.writers = w
	lal.formatter = f
}

func (lal *LevelAwareLogger) IsLevelEnabled(level LogLevel) bool {
	return level >= lal.localLogThreshhold || level >= lal.globalLogThreshold
}

func (lal *LevelAwareLogger) log(ctx context.Context, levelLabel string, level LogLevel, message string) {

	if lal.IsLevelEnabled(level) {
		m := lal.formatter.Format(ctx, levelLabel, lal.loggerName, message)

		lal.write(m)
	}

}
func (lal *LevelAwareLogger) logf(ctx context.Context, levelLabel string, level LogLevel, format string, a ...interface{}) {

	if lal.IsLevelEnabled(level) {
		message := fmt.Sprintf(format, a...)
		m := lal.formatter.Format(ctx, levelLabel, lal.loggerName, message)

		lal.write(m)
	}

}

func (lal *LevelAwareLogger) write(m string) {

	for _, w := range lal.writers {
		w.WriteMessage(m)
	}

}

func (lal *LevelAwareLogger) LogAtLevelCtx(ctx context.Context, level LogLevel, levelLabel string, message string) {
	lal.log(ctx, levelLabel, level, message)
}

func (lal *LevelAwareLogger) LogAtLevelfCtx(ctx context.Context, level LogLevel, levelLabel string, format string, a ...interface{}) {
	lal.logf(ctx, levelLabel, level, format, a...)
}

func (lal *LevelAwareLogger) LogTracefCtx(ctx context.Context, format string, a ...interface{}) {
	lal.logf(ctx, TraceLabel, Trace, format, a...)
}

func (lal *LevelAwareLogger) LogDebugfCtx(ctx context.Context, format string, a ...interface{}) {
	lal.logf(ctx, DebugLabel, Debug, format, a...)
}

func (lal *LevelAwareLogger) LogInfofCtx(ctx context.Context, format string, a ...interface{}) {
	lal.logf(ctx, InfoLabel, Info, format, a...)
}

func (lal *LevelAwareLogger) LogWarnfCtx(ctx context.Context, format string, a ...interface{}) {
	lal.logf(ctx, WarnLabel, Warn, format, a...)
}

func (lal *LevelAwareLogger) LogErrorfCtx(ctx context.Context, format string, a ...interface{}) {
	lal.logf(ctx, ErrorLabel, Error, format, a...)
}

func (lal *LevelAwareLogger) LogErrorfCtxWithTrace(ctx context.Context, format string, a ...interface{}) {
	trace := make([]byte, 2048)
	runtime.Stack(trace, false)

	format = format + "\n%s"
	a = append(a, trace)

	lal.logf(ctx, ErrorLabel, Error, format, a...)

}

func (lal *LevelAwareLogger) LogFatalfCtx(ctx context.Context, format string, a ...interface{}) {
	lal.logf(ctx, FatalLabel, Fatal, format, a...)
}

func (lal *LevelAwareLogger) LogAtLevel(level LogLevel, levelLabel string, message string) {
	lal.LogAtLevelCtx(nil, level, levelLabel, message)
}

func (lal *LevelAwareLogger) LogAtLevelf(level LogLevel, levelLabel string, format string, a ...interface{}) {
	lal.logf(nil, levelLabel, level, format, a...)
}

func (lal *LevelAwareLogger) LogTracef(format string, a ...interface{}) {
	lal.logf(nil, TraceLabel, Trace, format, a...)
}

func (lal *LevelAwareLogger) LogDebugf(format string, a ...interface{}) {
	lal.logf(nil, DebugLabel, Debug, format, a...)
}

func (lal *LevelAwareLogger) LogInfof(format string, a ...interface{}) {
	lal.logf(nil, InfoLabel, Info, format, a...)
}

func (lal *LevelAwareLogger) LogWarnf(format string, a ...interface{}) {
	lal.logf(nil, WarnLabel, Warn, format, a...)
}

func (lal *LevelAwareLogger) LogErrorf(format string, a ...interface{}) {
	lal.logf(nil, ErrorLabel, Error, format, a...)
}

func (lal *LevelAwareLogger) LogErrorfWithTrace(format string, a ...interface{}) {
	lal.LogErrorfCtxWithTrace(nil, format, a...)

}

func (lal *LevelAwareLogger) LogFatalf(format string, a ...interface{}) {
	lal.logf(nil, FatalLabel, Fatal, format, a...)
}

func (lal *LevelAwareLogger) SetGlobalThreshold(threshold LogLevel) {
	lal.globalLogThreshold = threshold
}

func (lal *LevelAwareLogger) SetLocalThreshold(threshold LogLevel) {
	lal.localLogThreshhold = threshold
}

func (lal *LevelAwareLogger) SetThreshold(threshold LogLevel) {
	lal.SetGlobalThreshold(threshold)
	lal.SetLocalThreshold(threshold)
}

func (lal *LevelAwareLogger) SetLoggerName(name string) {
	lal.loggerName = name
}

type LogRuntimeControl interface {
	SetGlobalThreshold(threshold LogLevel)
	SetLocalThreshold(threshold LogLevel)
	UpdateWritersAndFormatter([]LogWriter, *LogMessageFormatter)
}

func CreateAnonymousLogger(componentId string, threshold LogLevel) Logger {
	logger := new(LevelAwareLogger)
	logger.globalLogThreshold = threshold
	logger.localLogThreshhold = threshold
	logger.loggerName = componentId

	return logger
}
