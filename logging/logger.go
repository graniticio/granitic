/*
Package logging provides a component-based logging framework for user and built-in Granitic components.
*/
package logging

import (
	"fmt"
	"runtime"
	"time"
)

type Logger interface {
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
}

func (lal *LevelAwareLogger) IsLevelEnabled(level LogLevel) bool {
	return level >= lal.localLogThreshhold || level >= lal.globalLogThreshold
}

func (lal *LevelAwareLogger) log(prefix string, level LogLevel, message string) {

	if lal.IsLevelEnabled(level) {
		t := time.Now()
		fmt.Printf("%s %s %s: %s\n", t.Format(time.RFC3339), prefix, lal.loggerName, message)
	}

}
func (lal *LevelAwareLogger) logf(levelLabel string, level LogLevel, format string, a ...interface{}) {

	if lal.IsLevelEnabled(level) {
		t := time.Now()
		message := fmt.Sprintf(format, a...)
		fmt.Printf("%s %s %s: %s\n", t.Format(time.RFC3339), levelLabel, lal.loggerName, message)
	}

}

func (lal *LevelAwareLogger) LogAtLevel(level LogLevel, levelLabel string, message string) {
	lal.log(levelLabel, level, message)
}

func (lal *LevelAwareLogger) LogAtLevelf(level LogLevel, levelLabel string, format string, a ...interface{}) {
	lal.logf(levelLabel, level, format, a...)
}

func (lal *LevelAwareLogger) LogTracef(format string, a ...interface{}) {
	lal.logf(TraceLabel, Trace, format, a...)
}

func (lal *LevelAwareLogger) LogDebugf(format string, a ...interface{}) {
	lal.logf(DebugLabel, Debug, format, a...)
}

func (lal *LevelAwareLogger) LogInfof(format string, a ...interface{}) {
	lal.logf(InfoLabel, Info, format, a...)
}

func (lal *LevelAwareLogger) LogWarnf(format string, a ...interface{}) {
	lal.logf(WarnLabel, Warn, format, a...)
}

func (lal *LevelAwareLogger) LogErrorf(format string, a ...interface{}) {
	lal.logf(ErrorLabel, Error, format, a...)
}

func (lal *LevelAwareLogger) LogErrorfWithTrace(format string, a ...interface{}) {
	trace := make([]byte, 2048)
	runtime.Stack(trace, false)

	format = format + "\n%s"
	a = append(a, trace)

	lal.logf(ErrorLabel, Error, format, a...)

}

func (lal *LevelAwareLogger) LogFatalf(format string, a ...interface{}) {
	lal.logf(FatalLabel, Fatal, format, a...)
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

type LogThresholdControl interface {
	SetGlobalThreshold(threshold LogLevel)
	SetLocalThreshold(threshold LogLevel)
}

func CreateAnonymousLogger(componentId string, threshold LogLevel) Logger {
	logger := new(LevelAwareLogger)
	logger.globalLogThreshold = threshold
	logger.localLogThreshhold = threshold
	logger.loggerName = componentId

	return logger
}
