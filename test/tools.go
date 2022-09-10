// Copyright 2016-2022 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

/*
Package test provides tools for Granitic's unit tests.

One of Granitic's design principles is that Granitic should not introduce
dependencies on third-party libraries, so this package contains convenience methods for making Grantic's built-in unit tests
more usable and readable that would be better served by a third-party test library.

These methods are not recommended for use in user applications or tests.
*/
package test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

const (
	graniticHomeEnvVar = "GRANITIC_HOME"
	goPathEnvVar       = "GOPATH"
)

// FilePath finds the absolute path of a file that is provided relative to the testdata directory of the current package under test.
func FilePath(file string) string {

	return filepath.Join("testdata", file)
}

// ExpectString stops a test and logs an error if the string to be checked does not have the expected value.
func ExpectString(t *testing.T, check, expected string) bool {
	if expected != check {
		l := determineLine()
		t.Fatalf("%s Expected %s, actual %s", l, expected, check)
		return false
	}

	return true

}

// ExpectBool stops a test and logs an error if the bool to be checked does not have the expected value.
func ExpectBool(t *testing.T, check, expected bool) bool {
	if expected != check {
		l := determineLine()
		t.Fatalf("%s Expected %t, actual %t", l, expected, check)
		return false
	}

	return true

}

// ExpectInt stops a test and logs an error if the int to be checked does not have the expected value.
func ExpectInt(t *testing.T, check, expected int) bool {
	if expected != check {
		l := determineLine()
		t.Fatalf("%s Expected %d, actual %d", l, expected, check)
		return false
	}

	return true
}

// ExpectFloat stops a test and logs an error if the float to be checked does not have the expected value.
func ExpectFloat(t *testing.T, check, expected float64) bool {
	if expected != check {
		l := determineLine()
		t.Fatalf("%s Expected %e, actual %e", l, expected, check)
		return false
	}

	return true
}

// ExpectNil stops a test and logs an error if the value to check is not nil
func ExpectNil(t *testing.T, check interface{}) bool {
	if check == nil {
		return true
	}

	l := determineLine()

	t.Fatalf("%s Expected nil, actual %q", l, check)
	return false
}

// ExpectNotNil stops a test and logs an error if the value to check is nil
func ExpectNotNil(t *testing.T, check interface{}) bool {
	if check != nil {
		return true
	}

	l := determineLine()

	t.Fatalf("%s: Expected not nil", l)
	return false

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

// FindFacilityConfigFromWD finds the path of the folder containing the facility config based on the current working directory
func FindFacilityConfigFromWD() (string, error) {

	var err error
	var files []os.FileInfo

	path, err := os.Getwd()

	if err != nil {
		return "", err
	}

	for maxDepth := 8; maxDepth > 0; maxDepth-- {

		if files, err = ioutil.ReadDir(path); err != nil {
			return "", err
		}

		rootFilesSeen := 0

		for _, f := range files {

			n := f.Name()

			if n == "granitic.go" || n == "go.mod" || n == "LICENSE" {
				rootFilesSeen++
			}

		}

		if rootFilesSeen == 3 {
			return filepath.Join(path, "facility", "config"), nil

		}

		path = filepath.Dir(path)

	}

	return "", fmt.Errorf("Unable to locate Grantic facility config files")

}
