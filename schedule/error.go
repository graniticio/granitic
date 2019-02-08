// Copyright 2018-2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package schedule

import "fmt"

func NewAllowRetryErrorf(template string, args ...interface{}) error {
	return &AllowRetryError{
		message: fmt.Sprintf(template, args...),
	}
}

type AllowRetryError struct {
	message string
}

func (e *AllowRetryError) Error() string {
	return e.message
}
