// Copyright 2016 Granitic. All rights reserved.
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

type GlobalLevel interface {
	GlobalLevel() LogLevel
}

type RuntimeControllableLog interface {
	SetThreshold(threshold LogLevel)
	UpdateWritersAndFormatter([]LogWriter, *LogMessageFormatter)
}

type LevelAwareLogger struct {
	global             GlobalLevel
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

	var el LogLevel

	gl := lal.global.GlobalLevel()
	ll := lal.localLogThreshhold

	if ll == All {
		el = gl
	} else {
		el = ll
	}

	return level >= el
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

func (lal *LevelAwareLogger) SetLocalThreshold(threshold LogLevel) {
	lal.localLogThreshhold = threshold
}

func (lal *LevelAwareLogger) SetThreshold(threshold LogLevel) {
	lal.SetLocalThreshold(threshold)
}

func (lal *LevelAwareLogger) SetLoggerName(name string) {
	lal.loggerName = name
}

func CreateAnonymousLogger(componentId string, threshold LogLevel) Logger {
	logger := new(LevelAwareLogger)

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

func (ls *globalLogSource) GlobalLevel() LogLevel {
	return ls.level
}
