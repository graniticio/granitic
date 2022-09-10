// Copyright 2016-2022 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package logging

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// LogWriter is implemented by components able to write a log message (to a file, console etc)
type LogWriter interface {
	// WriteMessage request that the supplied message be written. Depending on implementation, may be asynchronous.
	WriteMessage(string)

	// Close any resources (file handles etc) that this LogWriter might be holding.
	Close()

	// Returns true if the LogWriter is currently in the process of writing a log line.
	Busy() bool
}

// ConsoleWriter is an implementation of LogWriter that sends messages to the console/stdout using the fmt.Print method
type ConsoleWriter struct {
}

// WriteMessage passes message to fmt.Print
func (cw *ConsoleWriter) WriteMessage(m string) {
	fmt.Print(m)
}

// Close does nothing
func (cw *ConsoleWriter) Close() {
}

// Busy always returns false
func (cw *ConsoleWriter) Busy() bool {
	return false
}

// FixedPrefixConsoleWriter is an implementation of LogWriter that sends messages to the console/stdout with a static prefix in front of every line
type FixedPrefixConsoleWriter struct {
	Prefix string
}

// WriteMessage passes message to fmt.Print
func (cw *FixedPrefixConsoleWriter) WriteMessage(m string) {
	fmt.Printf("%s%s", cw.Prefix, m)
}

// Close does nothing
func (cw *FixedPrefixConsoleWriter) Close() {
}

// Busy always returns false
func (cw *FixedPrefixConsoleWriter) Busy() bool {
	return false
}

// AsynchFileWriter is an implementation of LogWriter that appends a message to a file. Messages will be written
// asynchronously as long as the number of messages queued for writing does not exceed the value of BufferSize
type AsynchFileWriter struct {
	messages chan string
	logFile  *os.File

	//The number of messages that can be queued for writing before calls to WriteMessage block.
	BufferSize int

	//The file (absolute path or relative to application's working directory) that log messages should be appended to.
	LogPath string
}

// WriteMessage queues a message for writing and returns immediately, as long as the number of queued messages does not
// exceed BufferSize.
func (afw *AsynchFileWriter) WriteMessage(m string) {
	afw.messages <- m
}

func (afw *AsynchFileWriter) watchLineBuffer() {
	for {
		line := <-afw.messages

		f := afw.logFile

		if f != nil {
			f.WriteString(line)
		}

	}
}

// Init creates a channel to act as a buffer for queued messages
func (afw *AsynchFileWriter) Init() error {

	afw.messages = make(chan string, afw.BufferSize)

	err := afw.openFile()

	go afw.watchLineBuffer()

	return err

}

func (afw *AsynchFileWriter) openFile() error {
	logPath := afw.LogPath

	if len(strings.TrimSpace(logPath)) == 0 {
		return errors.New("File logging is enabled, but no path to a log file specified")
	}

	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)

	if err != nil {
		return err
	}

	afw.logFile = f
	return nil
}

// Close closes the log file
func (afw *AsynchFileWriter) Close() {
	afw.logFile.Close()
}

// Busy returns true if one or more messages are queued for writing.
func (afw *AsynchFileWriter) Busy() bool {

	return len(afw.messages) > 0
}
