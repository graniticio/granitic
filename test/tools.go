// Copyright 2016-2019 Granitic. All rights reserved.
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
	"github.com/graniticio/granitic/v2/instance"
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

// FilePath finds the absolute path of a file that is provided relative to the resource/test directory.
func FilePath(file string) string {

	//Check if GRANITIC_HOME explicitly set
	graniticPath := os.Getenv(graniticHomeEnvVar)

	if graniticPath == "" {

		if gopath := os.Getenv(goPathEnvVar); gopath == "" {

			fmt.Printf("Neither %s or %s environment variable is not set. Cannot find Granitic\n", graniticHomeEnvVar, goPathEnvVar)
			instance.ExitError()

		} else {

			graniticPath = filepath.Join(gopath, "src", "github.com", "graniticio", "granitic")

			if _, err := ioutil.ReadDir(graniticPath); err != nil {

				fmt.Printf("%s environment variable is not set and cannot find Granitic in the default install path of %s (your GOPATH variable is set to %s)\n", graniticHomeEnvVar, graniticPath, gopath)
				instance.ExitError()
			}

		}

	}
	return filepath.Join(graniticPath, "resource", "test", file)
}

// ExpectString stops a test and logs an error if the string to be checked does not have the expected value.
func ExpectString(t *testing.T, check, expected string) bool {
	if expected != check {
		l := determineLine()
		t.Fatalf("%s Expected %s, actual %s", l, expected, check)
		return false
	} else {
		return true
	}
}

// ExpectBool stops a test and logs an error if the bool to be checked does not have the expected value.
func ExpectBool(t *testing.T, check, expected bool) bool {
	if expected != check {
		l := determineLine()
		t.Fatalf("%s Expected %t, actual %t", l, expected, check)
		return false
	} else {
		return true
	}
}

// ExpectInt stops a test and logs an error if the int to be checked does not have the expected value.
func ExpectInt(t *testing.T, check, expected int) bool {
	if expected != check {
		l := determineLine()
		t.Fatalf("%s Expected %d, actual %d", l, expected, check)
		return false
	} else {
		return true
	}
}

// ExpectFloat stops a test and logs an error if the float to be checked does not have the expected value.
func ExpectFloat(t *testing.T, check, expected float64) bool {
	if expected != check {
		l := determineLine()
		t.Fatalf("%s Expected %e, actual %e", l, expected, check)
		return false
	} else {
		return true
	}
}

// ExpectNil stops a test and logs an error if the value to check is not nil
func ExpectNil(t *testing.T, check interface{}) bool {
	if check == nil {
		return true
	} else {
		l := determineLine()

		t.Fatalf("%s Expected nil, actual %q", l, check)
		return false
	}
}

// ExpectNil stops a test and logs an error if the value to check is nil
func ExpectNotNil(t *testing.T, check interface{}) bool {
	if check != nil {
		return true
	} else {

		l := determineLine()

		t.Fatalf("%s: Expected not nil", l)
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
