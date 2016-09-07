package logging

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

type LogWriter interface {
	WriteMessage(string)
	Close()
	Busy() bool
}

type ConsoleWriter struct {
}

func (cw *ConsoleWriter) WriteMessage(m string) {
	fmt.Print(m)
}

func (cw *ConsoleWriter) Close() {
}

func (cw *ConsoleWriter) Busy() bool {
	return false
}

type AsynchFileWriter struct {
	messages   chan string
	logFile    *os.File
	BufferSize int
	LogPath    string
}

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

func (afw *AsynchFileWriter) Close() {
	afw.logFile.Close()
}

func (cw *AsynchFileWriter) Busy() bool {

	return len(cw.messages) > 0
}
