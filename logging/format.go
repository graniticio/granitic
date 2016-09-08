package logging

import (
	"bytes"
	"errors"
	"fmt"
	"golang.org/x/net/context"
	"regexp"
	"strconv"
	"time"
)

const presetFrameworkFormat = "%{02/Jan/2006:15:04:05 Z0700}t %P %c "
const FrameworkPresetPrefix = "framework"
const formatRegex = "\\%[a-zA-Z]|\\%\\%|\\%{[^}]*}[a-zA-Z]"
const varModifiedRegex = "\\%{([^}]*)}([a-zA-Z])"
const percent = "%"
const hyphen = "-"

const unsupported = "???"

type prefixFormatPlaceHolder int

const (
	Unsupported = iota
	PercentSymbol
	LogTime
	LogLevelFull
	LogLevelInitial
	LogLevelFullPadded
	ComponentName
	ComponentNameTrunc
)

type logPrefixElementType int

const (
	Text = iota
	Placeholder
	PlaceholderWithVar
)

type prefixElement struct {
	elementType     logPrefixElementType
	placeholderType prefixFormatPlaceHolder
	content         string
	variable        string
}

func newTextLogLineElement(text string) *prefixElement {

	e := new(prefixElement)
	e.elementType = Text
	e.content = text

	return e
}

func newPlaceholderLineElement(phType prefixFormatPlaceHolder) *prefixElement {

	e := new(prefixElement)
	e.elementType = Placeholder
	e.placeholderType = phType

	return e
}

func newPlaceholderWithVarLineElement(phType prefixFormatPlaceHolder, variable string) *prefixElement {

	e := new(prefixElement)
	e.elementType = PlaceholderWithVar
	e.placeholderType = phType
	e.variable = variable

	return e
}

func NewFrameworkLogMessageFormatter() *LogMessageFormatter {
	lmf := new(LogMessageFormatter)
	lmf.UtcTimes = true
	lmf.PrefixPreset = FrameworkPresetPrefix

	lmf.Init()

	return lmf
}

type LogMessageFormatter struct {
	elements     []*prefixElement
	PrefixFormat string
	PrefixPreset string
	UtcTimes     bool
}

func (lmf *LogMessageFormatter) Format(ctx context.Context, levelLabel, loggerName, message string) string {
	var b bytes.Buffer
	var t time.Time

	if lmf.UtcTimes {
		t = time.Now().UTC()
	} else {
		t = time.Now()
	}

	for _, e := range lmf.elements {

		switch e.elementType {
		case Text:
			b.WriteString(e.content)
		case Placeholder:
			b.WriteString(lmf.findValue(e, levelLabel, loggerName, &t))
		case PlaceholderWithVar:
			b.WriteString(lmf.findValueWithVar(e, levelLabel, loggerName, &t))
		}
	}

	b.WriteString(message)
	b.WriteString("\n")

	return b.String()
}

func (alw *LogMessageFormatter) findValueWithVar(element *prefixElement, levelLabel, loggerName string, loggedAt *time.Time) string {
	switch element.placeholderType {
	case LogTime:
		return loggedAt.Format(element.variable)
	case ComponentNameTrunc:
		return truncOrPad(loggerName, element.variable)
	default:
		return unsupported

	}
}

func (alw *LogMessageFormatter) findValue(element *prefixElement, levelLabel, loggerName string, loggedAt *time.Time) string {

	switch element.placeholderType {

	case PercentSymbol:
		return percent

	case ComponentName:
		return loggerName

	case LogLevelFull:
		return levelLabel

	case LogLevelInitial:
		return string(levelLabel[0])
	case LogLevelFullPadded:
		return padRightTo(levelLabel, 5)
	default:
		return unsupported

	}

}

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
			return lmf.parseFormat(presetFrameworkFormat)

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

		if lfph == Unsupported {
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

		if lfph == Unsupported {
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
		return Unsupported
	case "%":
		return PercentSymbol
	case "t":
		return LogTime
	case "L":
		return LogLevelFull
	case "l":
		return LogLevelInitial
	case "P":
		return LogLevelFullPadded
	case "c":
		return ComponentName
	case "C":
		return ComponentNameTrunc
	}

}
