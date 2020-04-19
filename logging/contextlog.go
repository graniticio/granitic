// Copyright 2016-2020 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package logging

import "context"

// FilteredContextData holds information that has been extracted from a context.Context in a format suitable for logging
type FilteredContextData map[string]string

// ContextFilter takes a context and extracts some or all of the data in it a returns in a form suitable for
// inclusion in application and access logs.
type ContextFilter interface {
	Extract(ctx context.Context) FilteredContextData
}
