package logging

import (
	"fmt"
	"runtime"
)

type ConsoleErrorLogger struct {

}


func (l *ConsoleErrorLogger) LogTracef(format string, a ...interface{}) {
	return
}

func (l *ConsoleErrorLogger) LogDebugf(format string, a ...interface{}) {
	return
}

func (l *ConsoleErrorLogger) LogInfof(format string, a ...interface{}) {
	return
}

func (l *ConsoleErrorLogger) LogWarnf(format string, a ...interface{}) {
	return
}

func (l *ConsoleErrorLogger) LogErrorf(format string, a ...interface{}) {
	fmt.Printf(format + "\n", a...)
}

func (l *ConsoleErrorLogger) LogErrorfWithTrace(format string, a ...interface{}) {
	trace := make([]byte, 2048)
	runtime.Stack(trace, false)

	format = format + "\n%s"
	a = append(a, trace)

	l.LogErrorf(format, a...)
}

func (l *ConsoleErrorLogger) LogFatalf(format string, a ...interface{}) {
	l.LogErrorf(format, a...)
}

func (l *ConsoleErrorLogger) LogAtLevelf(level int, levelLabel string, format string, a ...interface{}) {
	if l.IsLevelEnabled(level) {
		l.LogErrorf(format, a...)
	}
}

func (l *ConsoleErrorLogger) IsLevelEnabled(level int) bool {
	return level >= Error
}
