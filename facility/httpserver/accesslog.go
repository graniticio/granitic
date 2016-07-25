package httpserver

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const DefaultLogBufferLength = 10
const percent = "%"
const hyphen = "-"
const unsupported = "???"
const presetCommonName = "common"
const presetCommonFormat = "%h %l %u %t \"%r\" %s %b"
const presetCombinedName = "combined"
const presetCombinedFormat = "%h %l %u %t \"%r\" %s %b \"%{Referer}i\" \"%{User-agent}i\""

const presetframeworkName = "framework"
const presetFrameworkFormat = "%h XFF[%{X-Forwarded-For}i] %l %u [%{02/Jan/2006:15:04:05 Z0700}t] \"%m %U%q\" %s %bB %{us}Tμs"

const formatRegex = "\\%[a-zA-Z]|\\%\\%|\\%{[^}]*}[a-zA-Z]"
const varModifiedRegex = "\\%{([^}]*)}([a-zA-Z])"
const commonLogDateFormat = "[02/Jan/2006:15:04:05 -0700]"

type LogFormatPlaceHolder int

const (
	Unsupported = iota
	RemoteHost
	ClientId
	UserId
	ReceivedTime
	RequestLine
	StatusCode
	BytesReturned
	BytesReturnedClf
	RequestHeader
	PercentSymbol
	Method
	Path
	Query
	ProcessTimeMicro
	ProcessTime
)

type LogLineElementType int

const (
	Text = iota
	Placeholder
	PlaceholderWithVar
)

type LogLineElement struct {
	elementType     LogLineElementType
	placeholderType LogFormatPlaceHolder
	content         string
	variable        string
}

func newTextLogLineElement(text string) *LogLineElement {

	e := new(LogLineElement)
	e.elementType = Text
	e.content = text

	return e
}

func newPlaceholderLineElement(phType LogFormatPlaceHolder) *LogLineElement {

	e := new(LogLineElement)
	e.elementType = Placeholder
	e.placeholderType = phType

	return e
}

func newPlaceholderWithVarLineElement(phType LogFormatPlaceHolder, variable string) *LogLineElement {

	e := new(LogLineElement)
	e.elementType = PlaceholderWithVar
	e.placeholderType = phType
	e.variable = variable

	return e
}

type AccessLogWriter struct {
	logFile       *os.File
	LogPath       string
	LogLineFormat string
	LogLinePreset string
	UtcTimes      bool
	elements      []*LogLineElement
	lines         chan string
}

func (alw *AccessLogWriter) LogRequest(req *http.Request, res *wrappedResponseWriter, rec *time.Time, fin *time.Time, id *IdentityMap) {

	alw.lines <- alw.buildLine(req, res, rec, fin, id)

}

func (alw *AccessLogWriter) buildLine(req *http.Request, res *wrappedResponseWriter, rec *time.Time, fin *time.Time, id *IdentityMap) string {
	var b bytes.Buffer

	if alw.UtcTimes {
		utcRec := rec.UTC()
		utcFin := fin.UTC()

		rec = &utcRec
		fin = &utcFin
	}

	for _, e := range alw.elements {

		switch e.elementType {
		case Text:
			b.WriteString(e.content)
		case Placeholder:
			b.WriteString(alw.findValue(e, req, res, rec, fin, id))
		case PlaceholderWithVar:
			b.WriteString(alw.findValueWithVar(e, req, res, rec, fin, id))
		}
	}

	b.WriteString("\n")

	return b.String()

}

func (alw *AccessLogWriter) StartComponent() error {

	alw.lines = make(chan string, DefaultLogBufferLength)

	err := alw.configureLogFormat()

	if err != nil {
		return err
	}

	err = alw.openFile()

	go alw.watchLineBuffer()

	return err

}

func (alw *AccessLogWriter) watchLineBuffer() {
	for {
		line := <-alw.lines

		f := alw.logFile

		if f != nil {
			f.WriteString(line)
		}

	}
}

func (alw *AccessLogWriter) openFile() error {
	logPath := alw.LogPath

	if len(strings.TrimSpace(logPath)) == 0 {
		return errors.New("HTTP server access log is enabled, but no path to a log file specified")
	}

	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)

	if err != nil {
		return err
	}

	alw.logFile = f
	return nil
}

func (alw *AccessLogWriter) configureLogFormat() error {

	f := alw.LogLineFormat
	pre := alw.LogLinePreset

	if f == "" && pre == "" {
		return errors.New("You must specify either a format for access log lines or the name of a preset format (neither has been provided).")
	}

	if f != "" && pre != "" {
		return errors.New("You must specify either a format for access log lines OR the name of a preset format (BOTH have been provided).")
	}

	if pre != "" {

		if pre == presetCommonName {
			return alw.parseFormat(presetCommonFormat)

		} else if pre == presetCombinedName {
			return alw.parseFormat(presetCombinedFormat)

		} else if pre == presetframeworkName {
			return alw.parseFormat(presetFrameworkFormat)

		} else {
			message := fmt.Sprintf("%s is not a supported preset for access log lines", pre)
			return errors.New(message)
		}

	}

	return alw.parseFormat(f)
}

func (alw *AccessLogWriter) parseFormat(format string) error {

	lineRe := regexp.MustCompile(formatRegex)
	varRe := regexp.MustCompile(varModifiedRegex)

	placeholders := lineRe.FindAllString(format, -1)
	textFragments := lineRe.Split(format, -1)
	firstMatch := lineRe.FindStringIndex(format)
	startsWithPh := (firstMatch[0] == 0) && textFragments[0] != ""

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
				err = alw.addPlaceholder(ph, varRe)
				alw.addTextElement(text)
			} else {
				alw.addTextElement(text)
				err = alw.addPlaceholder(ph, varRe)
			}

		} else if phAvail {
			ph := placeholders[i]
			err = alw.addPlaceholder(ph, varRe)

		} else if tAvail {
			text := textFragments[i]
			alw.addTextElement(text)
		}

		if err != nil {
			return err
		}

	}

	return nil
}

func (alw *AccessLogWriter) addTextElement(text string) {

	if text != "" {
		e := newTextLogLineElement(text)
		alw.elements = append(alw.elements, e)
	}
}

func (alw *AccessLogWriter) addPlaceholder(ph string, re *regexp.Regexp) error {

	if len(ph) == 2 {

		formatTypeCode := ph[1:2]

		lfph := alw.mapPlaceholder(formatTypeCode)

		if lfph == Unsupported {
			message := fmt.Sprintf("%s is not a supported field for formatting access log lines", ph)
			return errors.New(message)
		} else {
			e := newPlaceholderLineElement(lfph)
			alw.elements = append(alw.elements, e)
		}

	} else {

		v := re.FindStringSubmatch(ph)
		arg := v[1]
		formatTypeCode := v[2]

		lfph := alw.mapPlaceholder(formatTypeCode)

		if lfph == Unsupported {
			message := fmt.Sprintf("%s is not a supported field for formatting access log lines", ph)
			return errors.New(message)
		} else {
			e := newPlaceholderWithVarLineElement(lfph, arg)
			alw.elements = append(alw.elements, e)
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

func (alw *AccessLogWriter) mapPlaceholder(ph string) LogFormatPlaceHolder {

	switch ph {
	default:
		return Unsupported
	case "%":
		return PercentSymbol
	case "b":
		return BytesReturnedClf
	case "B":
		return BytesReturned
	case "D":
		return ProcessTimeMicro
	case "h":
		return RemoteHost
	case "i":
		return RequestHeader
	case "l":
		return ClientId
	case "m":
		return Method
	case "q":
		return Query
	case "r":
		return RequestLine
	case "s":
		return StatusCode
	case "t":
		return ReceivedTime
	case "T":
		return ProcessTime
	case "u":
		return UserId
	case "U":
		return Path
	}

}

func (alw *AccessLogWriter) findValueWithVar(element *LogLineElement, req *http.Request, res *wrappedResponseWriter, received *time.Time, finished *time.Time, id *IdentityMap) string {
	switch element.placeholderType {
	case RequestHeader:
		return alw.requestHeader(element.variable, req)

	case ReceivedTime:
		return received.Format(element.variable)

	case ProcessTime:

		switch element.variable {
		case "s":
			return alw.processTime(received, finished, time.Second)
		case "us":
			return alw.processTime(received, finished, time.Microsecond)
		case "ms":
			return alw.processTime(received, finished, time.Millisecond)
		default:
			return "??"

		}

	default:
		return unsupported

	}
}

func (alw *AccessLogWriter) findValue(element *LogLineElement, req *http.Request, res *wrappedResponseWriter, received *time.Time, finished *time.Time, id *IdentityMap) string {

	switch element.placeholderType {

	case PercentSymbol:
		return percent

	case BytesReturnedClf:
		if res.BytesServed == 0 {
			return hyphen
		} else {
			return (strconv.Itoa(res.BytesServed))
		}

	case BytesReturned:
		return (strconv.Itoa(res.BytesServed))

	case RemoteHost:
		return req.RemoteAddr

	case ClientId:
		return hyphen

	case UserId:
		return alw.userId(id)

	case Method:
		return req.Method

	case Path:
		return req.URL.Path

	case Query:
		return alw.query(req)

	case RequestLine:
		return alw.requestLine(req)

	case ReceivedTime:
		return received.Format(commonLogDateFormat)

	case StatusCode:
		return strconv.Itoa(res.Status)

	case ProcessTimeMicro:
		return alw.processTime(received, finished, time.Microsecond)

	case ProcessTime:
		return alw.processTime(received, finished, time.Second)

	default:
		return unsupported

	}

}

func (alw *AccessLogWriter) processTime(rec *time.Time, fin *time.Time, unit time.Duration) string {
	spent := fin.Sub(*rec)

	return strconv.FormatInt(int64(spent/unit), 10)
}

func (alw *AccessLogWriter) query(req *http.Request) string {

	q := req.URL.RawQuery

	if q == "" {
		return q
	} else {

		return "?" + q
	}

}

func (alw *AccessLogWriter) requestHeader(name string, req *http.Request) string {

	value := req.Header.Get(name)

	if value == "" {
		return hyphen

	} else {
		return value
	}

}

func (alw *AccessLogWriter) requestLine(req *http.Request) string {

	return fmt.Sprintf("%s %s %s", req.Method, req.RequestURI, req.Proto)

}

func (alw *AccessLogWriter) userId(id *IdentityMap) string {

	if id == nil || id.PublicUserId() == "" {
		return hyphen

	} else {
		return id.PublicUserId()
	}

}

func (alw *AccessLogWriter) PrepareToStop() {

}

func (alw *AccessLogWriter) ReadyToStop() (bool, error) {
	return true, nil
}

func (alw *AccessLogWriter) Stop() error {

	if alw.logFile != nil {
		return alw.logFile.Close()
	}

	return nil
}
