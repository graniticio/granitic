package logging

import (
	"fmt"
	"time"
)

type LogMessageFormatter struct {
}

func (lmf *LogMessageFormatter) Format(levelLabel, loggerName, message string) string {

	t := time.Now()
	m := fmt.Sprintf("%s %s %s: %s\n", t.Format(time.RFC3339), levelLabel, loggerName, message)

	return m
}
