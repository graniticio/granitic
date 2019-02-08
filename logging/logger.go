// Copyright 2016-2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

/*
	Package logging provides a component-based logging framework for user and built-in Granitic components.

	Logging in Granitic is covered in detail at http://granitic.io/1.0/ref/logging a brief description of the key concepts and
	types follows.

	Component logging

	Every struct that is registered as a component in Granitic's IoC container has the option of having a Logger injected into it.
	Components are classified as framework components (built-in components that are created when you enable a facility in
	your application) and application components - named components defined in your application's component definition files.

	Obtaining a Logger

	Any component in your application will have a Logger injected into it if the underlying struct for that component declares a field:

		Log logging.Logger

	The Logger is injected during the 'decoration' phase of container startup (see package documentation for the ioc pacakge). This
	means the Logger is safe to use in any method in your component.

	Log levels

	A log level is a label indicating the relative significance of a message to be logged. In order of significance they are:

		TRACE, DEBUG, INFO, WARN, ERROR, FATAL

	These levels have no set meanings, but a rough set of meanings follows:

	TRACE

	Very detailed low level messages showing almost line-by-line commentary of a procedure.

	DEBUG

	Information that might be significant when trying to diagnose a fault (connection URLs, resource utilisation etc).

	INFO

	Status information that might be of interest to outside observers of an application (ready messages, declaring which ports
	HTTP servers are listening to, shutdown notifications).

	WARN

	An undesirable but managed situation where application or request integrity has not suffered (approaching a resource limit,
	having to retry a connection to an external system).

	ERROR

	A problem that has not affected application integrity, but has caused a user request to terminate abnormally (problem inserting into
	a database, downstream system unavailable after retries).

	FATAL

	A serious problem that has affected the integrity of the application such that it should be restarted or has crashed.

	Log methods

	The Logger interface (see below) provides methods to log a message at a particular level. E.g.

		Log.LogDebugf("A %s message", "DEBUG")


	Global and component thresholds

	A log level threshold is the minimum significance of a message that will be actually written to a log file or console. Granitic maintains
	a separate global log level for application components and framework components. This separation means that you could, for example,
	set your application's global log level to DEBUG without having's Granitic's built-in components filling your log files with clutter.

	The log levels can be adjusted in the configuration of your application by setting the following configuration in your
	application's configuration file:

		{
		  "FrameworkLogger":{
		    "GlobalLogLevel": "INFO"
		  },
		  "ApplicationLogger":{
		    "GlobalLogLevel": "INFO"
		  }
		}

	The above example is the default setting for Granitic applications, meaning that only messages with a log level of INFO or
	higher will actually be written to the console and/or log file.

	The global log level can be overridden for individual components by setting configuration similar to:

		{
		  "FrameworkLogger":{
		    "GlobalLogLevel": "FATAL",
		    "ComponentLogLevels": {
			  "grncHttpServer": "TRACE"
		    }
		  },
		  "ApplicationLogger":{
		    "GlobalLogLevel": "INFO",
		    "ComponentLogLevels": {
			  "noisyComponent": "ERROR"
		    }
		  }
		}

	In this example, all framework components are effectively silenced (apart from fatal errors), but the HTTP server is allowed to output TRACE
	messages. Application components are allowed to log at INFO and above, but one component is too chatty so is only
	allowed to log at ERROR or above.

	Log output

	Output of log messages to file and console is controlled by the LogWriting configuration element. The default settings
	look something like:

		{
		  "LogWriting": {
			"EnableConsoleLogging": true,
			"EnableFileLogging": false,
			"File": {
			  "LogPath": "./granitic.log",
			  "BufferSize": 50
			},
			"Format": {
			  "UtcTimes":     true,
			  "Unset": "-"
			}
		  }
		}

	For more information on these settings, refer to http://granitic.io/1.0/ref/logging#output

	Runtime control

	Global log levels and component log levels can be changed at runtime, if your application has the RuntimeCtl facility
	enabled. See http://granitic.io/1.0/ref/runtime-ctl for more information

	Log message prefixes

	Every message written to a log file or console can be given a customisable prefix containing meta-data like time
	of logging or information from a Context. See http://granitic.io/1.0/ref/logging#prefix


*/
package logging

import (
	"context"
	"fmt"
	"runtime"
)

/*
	The interface used by application code to submit a message to potentially be logged to a file or console. Methods
	are of the form

		Log[Level]f
		Log[Level]fCtx

	The Ctx versions of methods accept a Context. The Context is made available to the LogWriter implementations that
	write to files so that information in the Context can be potentially included in log line prefixes.

	The f suffix indicates that the methods accept the same templates and variadic arguments as fmt.Printf
*/
type Logger interface {
	//LogTracefCtx log a message at TRACE level with a Context
	LogTracefCtx(ctx context.Context, format string, a ...interface{})

	//LogDebugfCtx log a message at DEBUG level with a Context
	LogDebugfCtx(ctx context.Context, format string, a ...interface{})

	//LogInfofCtx log a message at INFO level with a Context
	LogInfofCtx(ctx context.Context, format string, a ...interface{})

	//LogWarnfCtx log a message at WARN level with a Context
	LogWarnfCtx(ctx context.Context, format string, a ...interface{})

	//LogErrorfCtx log a message at ERROR level with a Context
	LogErrorfCtx(ctx context.Context, format string, a ...interface{})

	//LogErrorfCtxWithTrace log a message at ERROR level with a Context. Message output will be followed by a partial stack trace.
	LogErrorfCtxWithTrace(ctx context.Context, format string, a ...interface{})

	//LogFatalfCtx log a message at FATAL level with a Context
	LogFatalfCtx(ctx context.Context, format string, a ...interface{})

	//LogAtLevelfCtx log at the specified level with a Context
	LogAtLevelfCtx(ctx context.Context, level LogLevel, levelLabel string, format string, a ...interface{})

	//LogTracef log a message at TRACE level
	LogTracef(format string, a ...interface{})

	//LogDebugf log a message at DEBUG level
	LogDebugf(format string, a ...interface{})

	//LogInfof log a message at INFO level
	LogInfof(format string, a ...interface{})

	//LogWarnf log a message at WARN level
	LogWarnf(format string, a ...interface{})

	//LogErrorf log a message at ERROR level
	LogErrorf(format string, a ...interface{})

	//LogErrorfWithTrace log a message at ERROR level. Message output will be followed by a partial stack trace.
	LogErrorfWithTrace(format string, a ...interface{})

	//LogFatalf log a message at FATAL level
	LogFatalf(format string, a ...interface{})

	//LogAtLevelfCtx log at the specified level
	LogAtLevelf(level LogLevel, levelLabel string, format string, a ...interface{})

	//IsLevelEnabled returns true if a message at the supplied level would acutally be logged. Useful to check
	//if the construction of a message would be expensive or slow.
	IsLevelEnabled(level LogLevel) bool
}

// Implemented by Loggers able to state what the current global log level is
type GlobalLevel interface {
	//GlobalLevel returns this component's current global log level
	GlobalLevel() LogLevel
}

// Implemented by loggers that can be modified at runtime
type RuntimeControllableLog interface {
	//SetThreshold sets the logger's log level to the specified level
	SetThreshold(threshold LogLevel)

	//UpdateWritersAndFormatter causes the Logger to discard its current LogWriters and LogMessageFormatter in favour of the ones supplied.
	UpdateWritersAndFormatter([]LogWriter, *LogMessageFormatter)
}

// Standard implementation of Logger which respects both a global log level and a specific level for this Logger.
type GraniticLogger struct {
	global             GlobalLevel
	localLogThreshhold LogLevel
	loggerName         string
	writers            []LogWriter
	formatter          *LogMessageFormatter
}

// See RuntimeControllableLog.UpdateWritersAndFormatter
func (grl *GraniticLogger) UpdateWritersAndFormatter(w []LogWriter, f *LogMessageFormatter) {
	grl.writers = w
	grl.formatter = f
}

// See Logger.IsLevelEnabled
func (grl *GraniticLogger) IsLevelEnabled(level LogLevel) bool {

	var el LogLevel

	gl := grl.global.GlobalLevel()
	ll := grl.localLogThreshhold

	if ll == All {
		el = gl
	} else {
		el = ll
	}

	return level >= el
}

func (grl *GraniticLogger) log(ctx context.Context, levelLabel string, level LogLevel, message string) {

	if grl.IsLevelEnabled(level) {
		m := grl.formatter.Format(ctx, levelLabel, grl.loggerName, message)

		grl.write(m)
	}

}
func (grl *GraniticLogger) logf(ctx context.Context, levelLabel string, level LogLevel, format string, a ...interface{}) {

	if grl.IsLevelEnabled(level) {
		message := fmt.Sprintf(format, a...)
		m := grl.formatter.Format(ctx, levelLabel, grl.loggerName, message)

		grl.write(m)
	}

}

func (grl *GraniticLogger) write(m string) {

	for _, w := range grl.writers {
		w.WriteMessage(m)
	}

}

func (grl *GraniticLogger) logAtLevelCtx(ctx context.Context, level LogLevel, levelLabel string, message string) {
	grl.log(ctx, levelLabel, level, message)
}

// See Logger.LogAtLevelfCtx
func (grl *GraniticLogger) LogAtLevelfCtx(ctx context.Context, level LogLevel, levelLabel string, format string, a ...interface{}) {
	grl.logf(ctx, levelLabel, level, format, a...)
}

// See Logger.LogTracefCtx
func (grl *GraniticLogger) LogTracefCtx(ctx context.Context, format string, a ...interface{}) {
	grl.logf(ctx, TraceLabel, Trace, format, a...)
}

// See Logger.LogDebugfCtx
func (grl *GraniticLogger) LogDebugfCtx(ctx context.Context, format string, a ...interface{}) {
	grl.logf(ctx, DebugLabel, Debug, format, a...)
}

// See Logger.LogInfofCtx
func (grl *GraniticLogger) LogInfofCtx(ctx context.Context, format string, a ...interface{}) {
	grl.logf(ctx, InfoLabel, Info, format, a...)
}

// See Logger.LogWarnfCtx
func (grl *GraniticLogger) LogWarnfCtx(ctx context.Context, format string, a ...interface{}) {
	grl.logf(ctx, WarnLabel, Warn, format, a...)
}

// See Logger.LogErrorfCtx
func (grl *GraniticLogger) LogErrorfCtx(ctx context.Context, format string, a ...interface{}) {
	grl.logf(ctx, ErrorLabel, Error, format, a...)
}

// See Logger.LogErrorfCtxWithTrace
func (grl *GraniticLogger) LogErrorfCtxWithTrace(ctx context.Context, format string, a ...interface{}) {
	trace := make([]byte, 2048)
	runtime.Stack(trace, false)

	format = format + "\n%s"
	a = append(a, trace)

	grl.logf(ctx, ErrorLabel, Error, format, a...)

}

// See Logger.LogFatalfCtx
func (grl *GraniticLogger) LogFatalfCtx(ctx context.Context, format string, a ...interface{}) {
	grl.logf(ctx, FatalLabel, Fatal, format, a...)
}

// See Logger.LogAtLevelf
func (grl *GraniticLogger) LogAtLevelf(level LogLevel, levelLabel string, format string, a ...interface{}) {
	grl.logf(nil, levelLabel, level, format, a...)
}

// See Logger.LogTracef
func (grl *GraniticLogger) LogTracef(format string, a ...interface{}) {
	grl.logf(nil, TraceLabel, Trace, format, a...)
}

// See Logger.LogDebugf
func (grl *GraniticLogger) LogDebugf(format string, a ...interface{}) {
	grl.logf(nil, DebugLabel, Debug, format, a...)
}

// See Logger.LogInfof
func (grl *GraniticLogger) LogInfof(format string, a ...interface{}) {
	grl.logf(nil, InfoLabel, Info, format, a...)
}

// See Logger.LogWarnf
func (grl *GraniticLogger) LogWarnf(format string, a ...interface{}) {
	grl.logf(nil, WarnLabel, Warn, format, a...)
}

// See Logger.LogErrorf
func (grl *GraniticLogger) LogErrorf(format string, a ...interface{}) {
	grl.logf(nil, ErrorLabel, Error, format, a...)
}

// See Logger.LogErrorfWithTrace
func (grl *GraniticLogger) LogErrorfWithTrace(format string, a ...interface{}) {
	grl.LogErrorfCtxWithTrace(nil, format, a...)

}

// See Logger.LogFatalf
func (grl *GraniticLogger) LogFatalf(format string, a ...interface{}) {
	grl.logf(nil, FatalLabel, Fatal, format, a...)
}

// Sets the log threshold for this Logger
func (grl *GraniticLogger) SetLocalThreshold(threshold LogLevel) {
	grl.localLogThreshhold = threshold
}

// CreateAnonymousLogger creates a new Logger without attaching it to a LogManager. Useful for tests.
func CreateAnonymousLogger(componentId string, threshold LogLevel) Logger {
	logger := new(GraniticLogger)

	gls := new(globalLogSource)
	gls.level = threshold

	logger.global = gls
	logger.localLogThreshhold = threshold
	logger.loggerName = componentId

	return logger
}

type globalLogSource struct {
	level LogLevel
}

// See GlobalLevel.GlobalLevel
func (ls *globalLogSource) GlobalLevel() LogLevel {
	return ls.level
}
