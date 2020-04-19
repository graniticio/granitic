// Copyright 2016-2020 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package logging

import (
	"context"
	"fmt"
)

// A JsonLogFormatter is a component able to take a message to be written to a log file and format it as JSON document
type JsonLogFormatter struct {
	// A component able to extract information from a context.Context into a loggable format
	ContextFilter ContextFilter
}

// Format takes the message and prefixes it according the the rule specified in PrefixFormat or PrefixPreset
func (jlf *JsonLogFormatter) Format(ctx context.Context, levelLabel, loggerName, message string) string {
	return fmt.Sprintf("{%s %s}", loggerName, message)
}

//SetContextFilter provides the formatter with access selected data from a context
func (jlf *JsonLogFormatter) SetContextFilter(cf ContextFilter) {
	jlf.ContextFilter = cf
}
