// Copyright 2016 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package logging

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"
)

// The default prefix format for log lines
const PresetFormatFramework = "%{02/Jan/2006:15:04:05 Z0700}t %P [%c] "
const FrameworkPresetPrefix = "framework"
const formatRegex = "\\%[a-zA-Z]|\\%\\%|\\%{[^}]*}[a-zA-Z]"
const varModifiedRegex = "\\%{([^}]*)}([a-zA-Z])"
const percent = "%"

const unsupported = "???"

type prefixFormatPlaceHolder int

const (
	unsupportedPH = iota
	percentSymbolPH
	logTimePH
	logLevelFullPH
	logLevelInitialPH
	logLevelFullPaddedPH
	componentNamePH
	componentNameTruncPH
	ctxValuePH
)

type logPrefixElementType int

const (
	textElement = iota
	placeholderElement
	placeholderWithVarElement
)

type prefixElement struct {
	elementType     logPrefixElementType
	placeholderType prefixFormatPlaceHolder
	content         string
	variable        string
}

func newTextLogLineElement(text string) *prefixElement {

	e := new(prefixElement)
	e.elementType = textElement
	e.content = text

	return e
}

func newPlaceholderLineElement(phType prefixFormatPlaceHolder) *prefixElement {

	e := new(prefixElement)
	e.elementType = placeholderElement
	e.placeholderType = phType

	return e
}

func newPlaceholderWithVarLineElement(phType prefixFormatPlaceHolder, variable string) *prefixElement {

	e := new(prefixElement)
	e.elementType = placeholderWithVarElement
	e.placeholderType = phType
	e.variable = variable

	return e
}

// NewFrameworkLogMessageFormatter creates a new LogMessageFormatter using the default 'framework' pattern for log line
// prefixes and UTC timestamps.
func NewFrameworkLogMessageFormatter() *LogMessageFormatter {
	lmf := new(LogMessageFormatter)
	lmf.UtcTimes = true
	lmf.PrefixPreset = FrameworkPresetPrefix

	lmf.Init()

	return lmf
}

// A component able to take a message to be written to a log file and prefix it with a formatted template
// which can include log times, data from a Context etc.
type LogMessageFormatter struct {
	elements []*prefixElement

	// The pattern to be used as a template when generating prefixes. Mutally exclusive with PrefixPreset
	PrefixFormat string

	// The name of a pre-defined prefix template (e.g. 'framework'). Mutally exclusive with PrefixFormat
	PrefixPreset string

	// Convert timestamps in prefixes to UTC
	UtcTimes bool

	// The symbol to use in place of an unset variable in a log line prefix.
	Unset string
}

// Format takes the message and prefixes it according the the rule specified in PrefixFormat or PrefixPreset
func (lmf *LogMessageFormatter) Format(ctx context.Context, levelLabel, loggerName, message string) string {
	var b bytes.Buffer
	var t time.Time

	if ctx == nil {
		ctx = context.Background()
	}

	if lmf.UtcTimes {
		t = time.Now().UTC()
	} else {
		t = time.Now()
	}

	for _, e := range lmf.elements {

		switch e.elementType {
		case textElement:
			b.WriteString(e.content)
		case placeholderElement:
			b.WriteString(lmf.findValue(e, levelLabel, loggerName, &t))
		case placeholderWithVarElement:
			b.WriteString(lmf.findValueWithVar(ctx, e, levelLabel, loggerName, &t))
		}
	}

	b.WriteString(message)
	b.WriteString("\n")

	return b.String()
}

func (alw *LogMessageFormatter) findValueWithVar(ctx context.Context, element *prefixElement, levelLabel, loggerName string, loggedAt *time.Time) string {
	switch element.placeholderType {
	case logTimePH:
		return loggedAt.Format(element.variable)
	case componentNameTruncPH:
		return truncOrPad(loggerName, element.variable)
	case ctxValuePH:
		return alw.ctxValue(ctx, element.variable)
	default:
		return unsupported

	}
}

func (alw *LogMessageFormatter) ctxValue(ctx context.Context, key string) string {

	var v interface{}

	if v = ctx.Value(key); v != nil {

		return fmt.Sprintf("%v", v)

	} else {
		return alw.Unset
	}

}

func (alw *LogMessageFormatter) findValue(element *prefixElement, levelLabel, loggerName string, loggedAt *time.Time) string {

	switch element.placeholderType {

	case percentSymbolPH:
		return percent

	case componentNamePH:
		return loggerName

	case logLevelFullPH:
		return levelLabel

	case logLevelInitialPH:
		return string(levelLabel[0])
	case logLevelFullPaddedPH:
		return padRightTo(levelLabel, 5)
	default:
		return unsupported

	}

}

// Checks that a valid format has been provided for the log message prefixes.
func (lmf *LogMessageFormatter) Init() error {

	f := lmf.PrefixFormat
	pre := lmf.PrefixPreset

	if f == "" && pre == "" {
		return errors.New("You must specify either a format for the prefix to log messages or the name of a preset format (neither has been provided).")
	}

	if f != "" && pre != "" {
		return errors.New("You must specify either a format for the prefix to log messages OR the name of a preset format (BOTH have been provided).")
	}

	if f != "" {
		return lmf.parseFormat(f)
	} else {

		if pre == FrameworkPresetPrefix {
			return lmf.parseFormat(PresetFormatFramework)

		} else {
			message := fmt.Sprintf("%s is not a supported preset for log prefixes", pre)
			return errors.New(message)
		}

	}

}

func (lmf *LogMessageFormatter) parseFormat(format string) error {

	lineRe := regexp.MustCompile(formatRegex)
	varRe := regexp.MustCompile(varModifiedRegex)

	placeholders := lineRe.FindAllString(format, -1)
	textFragments := lineRe.Split(format, -1)
	firstMatch := lineRe.FindStringIndex(format)

	var startsWithPh bool

	if len(firstMatch) > 0 {
		startsWithPh = (firstMatch[0] == 0) && textFragments[0] != ""
	}

	phCount := len(placeholders)
	tCount := len(textFragments)

	maxCount := intMax(phCount, tCount)

	for i := 0; i < maxCount; i++ {

		phAvail := i < phCount
		tAvail := i < tCount
		var err error

		if phAvail && tAvail {

			ph := placeholders[i]
			text := textFragments[i]

			if startsWithPh {
				err = lmf.addPlaceholder(ph, varRe)
				lmf.addTextElement(text)
			} else {
				lmf.addTextElement(text)
				err = lmf.addPlaceholder(ph, varRe)
			}

		} else if phAvail {
			ph := placeholders[i]
			err = lmf.addPlaceholder(ph, varRe)

		} else if tAvail {
			text := textFragments[i]
			lmf.addTextElement(text)
		}

		if err != nil {
			return err
		}

	}

	return nil
}

func (alw *LogMessageFormatter) addTextElement(text string) {

	if text != "" {
		e := newTextLogLineElement(text)
		alw.elements = append(alw.elements, e)
	}
}

func (lmf *LogMessageFormatter) addPlaceholder(ph string, re *regexp.Regexp) error {

	if len(ph) == 2 {

		formatTypeCode := ph[1:2]

		lfph := lmf.mapPlaceholder(formatTypeCode)

		if lfph == unsupportedPH {
			message := fmt.Sprintf("%s is not a supported field for formatting the prefix to log lines", ph)
			return errors.New(message)
		} else {
			e := newPlaceholderLineElement(lfph)
			lmf.elements = append(lmf.elements, e)
		}

	} else {

		v := re.FindStringSubmatch(ph)
		arg := v[1]
		formatTypeCode := v[2]

		lfph := lmf.mapPlaceholder(formatTypeCode)

		if lfph == unsupportedPH {
			message := fmt.Sprintf("%s is not a supported field for formatting the prefix to log lines", ph)
			return errors.New(message)
		} else {
			e := newPlaceholderWithVarLineElement(lfph, arg)
			lmf.elements = append(lmf.elements, e)
		}

	}

	return nil
}

func intMax(x, y int) int {
	if x > y {
		return x
	} else {
		return y
	}
}

func truncOrPad(s string, sp string) string {

	p, err := strconv.Atoi(sp)

	if err != nil {
		return s
	}

	if len(s) > p {
		return s[0:p]
	} else {
		return padRightTo(s, p)
	}

}

func padRightTo(s string, p int) string {

	l := len(s)

	if l >= p {
		return s
	}

	var b bytes.Buffer

	b.WriteString(s)

	for i := l; i < p; i++ {
		b.WriteString(" ")
	}

	return b.String()
}

func (alw *LogMessageFormatter) mapPlaceholder(ph string) prefixFormatPlaceHolder {

	switch ph {
	default:
		return unsupportedPH
	case "%":
		return percentSymbolPH
	case "t":
		return logTimePH
	case "L":
		return logLevelFullPH
	case "l":
		return logLevelInitialPH
	case "P":
		return logLevelFullPaddedPH
	case "c":
		return componentNamePH
	case "C":
		return componentNameTruncPH
	case "X":
		return ctxValuePH
	}

}
