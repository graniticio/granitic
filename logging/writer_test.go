package logging

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConsoleWriter(t *testing.T) {

	cw := new(ConsoleWriter)

	//Exepcted to do nothing so subsequent methods should run without problems
	cw.Close()

	if cw.Busy() {
		t.Errorf("ConsoleWriter should never return busy")
	}

	cw.WriteMessage("CONSOLE LOG TEST\n")

	var lw LogWriter

	lw = cw

	lw.Close()

}

func TestFixedPrefixConsoleWriter(t *testing.T) {

	cw := new(FixedPrefixConsoleWriter)
	cw.Prefix = "PREFIX"

	var lw LogWriter

	lw = cw

	//Exepcted to do nothing so subsequent methods should run without problems
	lw.Close()

	if lw.Busy() {
		t.Errorf("FixedPrefixConsoleWriter should never return busy")
	}

	lw.WriteMessage("CONSOLE LOG TEST WITH PREFIX\n")

}

func TestAsynchFileLoggingMissingFile(t *testing.T) {

	alw := new(AsynchFileWriter)

	alw.BufferSize = 1

	if err := alw.Init(); err == nil {
		t.Errorf("Expected an error when no log file specified")
	}

}

func TestAsynchFileLoggingIllegalFile(t *testing.T) {

	alw := new(AsynchFileWriter)
	alw.LogPath = "///:/"

	alw.BufferSize = 1

	if err := alw.Init(); err == nil {
		t.Errorf("Expected an error illegal log file  name specified")
	}

}

func TestAsynchFileLogging(t *testing.T) {

	fn := filepath.Join(os.TempDir(), "grnc-test-app.log")

	os.Remove(fn)

	f, err := os.Create(fn)
	defer os.Remove(fn)

	if err != nil {
		t.Fatalf("Unable to create tmp file for logging test %s", err.Error())
	} else {
		f.Close()
	}

	alw := new(AsynchFileWriter)

	alw.BufferSize = 1

	alw.LogPath = fn

	if err := alw.Init(); err != nil {
		t.Fatalf("Unable to init log file writer %s", err.Error())
	}

	m := "BUFFERED MESSAGE"

	alw.WriteMessage(m)

	for alw.Busy() {
	}

	b, err := os.ReadFile(fn)

	if err != nil {
		t.Fatalf("Unable to read tmp file for logging test %s", err.Error())
	}

	s := string(b)

	if !strings.HasPrefix(s, m) {
		t.Errorf("Could not find logged message in file")
	}

	alw.Close()

}
