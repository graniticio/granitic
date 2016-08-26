package test

import (
	"os"
	"runtime"
	"strings"
	"testing"
)

func TestFilePath(file string) string {
	return os.Getenv("GRANITIC_HOME") + "/test/" + file
}

func ExpectString(t *testing.T, check, expected string) bool {
	if expected != check {
		l := determineLine()
		t.Errorf("%s Expected %s, actual %s", l, expected, check)
		return false
	} else {
		return true
	}
}

func ExpectBool(t *testing.T, check, expected bool) bool {
	if expected != check {
		l := determineLine()
		t.Errorf("%s Expected %t, actual %t", l, expected, check)
		return false
	} else {
		return true
	}
}

func ExpectInt(t *testing.T, check, expected int) bool {
	if expected != check {
		l := determineLine()
		t.Errorf("%s Expected %d, actual %d", l, expected, check)
		return false
	} else {
		return true
	}
}

func ExpectFloat(t *testing.T, check, expected float64) bool {
	if expected != check {
		l := determineLine()
		t.Errorf("%s Expected %e, actual %e", l, expected, check)
		return false
	} else {
		return true
	}
}

func ExpectNil(t *testing.T, check interface{}) bool {
	if check == nil {
		return true
	} else {
		l := determineLine()

		t.Errorf("%s Expected nil, actual %q", l, check)
		return false
	}
}

func ExpectNotNil(t *testing.T, check interface{}) bool {
	if check != nil {
		return true
	} else {

		l := determineLine()

		t.Errorf("%s: Expected not nil", l)
		return false
	}
}

func determineLine() string {
	trace := make([]byte, 2048)
	runtime.Stack(trace, false)

	splitTrace := strings.SplitN(string(trace), "\n", -1)

	for _, l := range splitTrace {

		if strings.Contains(l, "granitic/test") {
			continue
		}

		if strings.HasPrefix(l, "\t") {
			trimmed := strings.TrimSpace(l)
			p := strings.SplitN(trimmed, " +", -1)[0]

			f := strings.SplitN(p, string(os.PathSeparator), -1)

			return f[len(f)-1]

		}
	}

	return ""
}
