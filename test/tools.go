package test

import (
	"os"
	"testing"
)

func TestFilePath(file string) string {
	return os.Getenv("GRANITIC_HOME") + "/test/" + file
}

func ExpectString(t *testing.T, check, expected string) bool {
	if expected != check {
		t.Errorf("Expected %s, actual %s", expected, check)
		return false
	} else {
		return true
	}
}

func ExpectBool(t *testing.T, check, expected bool) bool {
	if expected != check {
		t.Errorf("Expected %t, actual %t", expected, check)
		return false
	} else {
		return true
	}
}

func ExpectInt(t *testing.T, check, expected int) bool {
	if expected != check {
		t.Errorf("Expected %d, actual %d", expected, check)
		return false
	} else {
		return true
	}
}

func ExpectFloat(t *testing.T, check, expected float64) bool {
	if expected != check {
		t.Errorf("Expected %e, actual %e", expected, check)
		return false
	} else {
		return true
	}
}

func ExpectNil(t *testing.T, check interface{}) bool {
	if check == nil {
		return true
	} else {
		t.Errorf("Expected nil, actual %q", check)
		return false
	}
}

func ExpectNotNil(t *testing.T, check interface{}) bool {
	if check != nil {
		return true
	} else {
		t.Errorf("Expected not nil")
		return false
	}
}
