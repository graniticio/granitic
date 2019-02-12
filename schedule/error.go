// Copyright 2018-2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package schedule

import "fmt"

// NewAllowRetryErrorf creates a an AllowRetryError with the supplied message
func NewAllowRetryErrorf(template string, args ...interface{}) error {
	return &AllowRetryError{
		message: fmt.Sprintf(template, args...),
	}
}

// AllowRetryError indicates a non-fatal problem that means retrying the task is permissible
type AllowRetryError struct {
	message string
}

// Error returns the error message
func (e *AllowRetryError) Error() string {
	return e.message
}
