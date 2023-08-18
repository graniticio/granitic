// Copyright 2018-2023 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package schedule

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const daySingle = "DAY"
const daySingleSpace = daySingle + " "
const minuteSingle = "MINUTE"
const minuteSingleSpace = minuteSingle + " "
const hourSingle = "HOUR"
const hourSingleSpace = hourSingle + " "
const secondSingle = "SECOND"
const secondSingleSpace = hourSingle + " "
const doubleSpace = "  "
const singleSpace = " "
const at = "AT"

const dayDuration = time.Hour * 24

func parseEvery(every string) (*interval, error) {
	return parseEveryFromGivenNow(every, time.Now())
}

func parseNaturalToDuration(interval string) (time.Duration, error) {

	tokens := strings.Split(strings.ToUpper(interval), " ")

	if len(tokens) < 2 {
		m := fmt.Sprintf("%s cannot be parsed as a numeric value and a unit of time", interval)
		return 0, errors.New(m)
	}

	var v int64
	var u time.Duration
	var err error

	if v, u, err = parseValueUnit(tokens[0], tokens[1]); err != nil {
		return 0, err
	}

	return time.Duration(v) * u, nil

}

func parseValueUnit(value, unit string) (int64, time.Duration, error) {
	//Make such the first token is a positive integer
	frequencyValue, okay := validValue(value)

	if !okay {
		em := fmt.Sprintf("%s cannot be interpreted as a positive integer greater than 1", value)
		return 0, 0, errors.New(em)
	}

	frequencyUnit, okay := validUnit(unit)

	if !okay {
		em := fmt.Sprintf("%s cannot be interpreted as a valid unit (seconds, minutes, hours, days) ", unit)
		return 0, 0, errors.New(em)
	}

	return frequencyValue, frequencyUnit, nil
}

func parseEveryFromGivenNow(every string, now time.Time) (*interval, error) {

	m := fmt.Sprintf("Cannot parse recurring schedule [%s] ", every)
	i := new(interval)

	alphaNumRe := regexp.MustCompile("[^a-zA-Z0-9 ]+")

	norm := alphaNumRe.ReplaceAllString(every, "")

	norm = strings.TrimSpace(strings.ToUpper(norm))

	firstChar := string(norm[0])

	if _, err := strconv.Atoi(firstChar); err != nil {

		if singleUnit(norm) {
			norm = "1 " + norm
		} else {
			return i, errors.New(m)
		}

	}

	for strings.Contains(norm, doubleSpace) {
		norm = strings.Replace(norm, doubleSpace, singleSpace, -1)
	}

	tokens := strings.Split(norm, " ")

	if len(tokens) < 2 {
		return i, errors.New(m)
	}

	frequencyValue, frequencyUnit, err := parseValueUnit(tokens[0], tokens[1])

	if err != nil {
		return nil, errors.New(m + err.Error())
	}

	i.Frequency = time.Duration(frequencyValue) * frequencyUnit

	if len(tokens) == 2 {

		i.Mode = OffsetFromStart
		i.CalculatedAt = now

		return i, nil
	}

	if len(tokens) < 4 {
		return i, errors.New(m)
	}

	modifier := tokens[2]

	if modifier == at {
		err := configureRunAtModifier(tokens[3], i, now)

		if err != nil {
			err = errors.New(m + err.Error())
		}

		return i, err
	}

	return i, errors.New(m)
}

func configureRunAtModifier(offset string, i *interval, now time.Time) error {

	var re = regexp.MustCompile(`(?m)^(.{2}):?(.{2}):?((?:.{2})?)$`)

	if !re.MatchString(offset) {
		return errors.New("time following 'at' should be of the form HHMM, HHMMSS, HH:MM or HH:MM:SS")
	}

	components := re.FindStringSubmatch(offset)

	te, err := extractTimeElements(components[1:], i.Frequency)

	if err != nil {
		return err
	}

	i.ActualStart = calculateFirstRun(te, i.Frequency)
	i.Mode = ActualStartTime

	return nil
}

func calculateFirstRun(te timeElements, freq time.Duration) time.Time {

	now := time.Now()

	date := now.Format("2006-01-02")

	if te.hour == -1 {
		te.hour = now.Hour()
	}

	if te.minute == -1 {
		te.minute = now.Minute()
	}

	proposedTime := fmt.Sprintf("%s %02d:%02d:%02d", date, te.hour, te.minute, te.second)

	runTime, _ := time.Parse("2006-01-02 15:04:05", proposedTime)

	if runTime.Before(time.Now()) {
		runTime = runTime.Add(freq)
	}

	return runTime

}

func extractTimeElements(s []string, freq time.Duration) (timeElements, error) {

	te := timeElements{}

	hours := s[0]

	if freq < dayDuration {
		//Ignore hours
		te.hour = -1
	} else if hi, err := strconv.Atoi(hours); err == nil && hi >= 0 && hi < 24 {
		te.hour = hi
	} else {
		m := fmt.Sprintf("%s is not a valid hour (must be 00-23)", hours)
		return te, errors.New(m)
	}

	minutes := s[1]

	if freq < time.Hour {
		//Ignore minutes
		te.minute = -1
	} else if hi, err := strconv.Atoi(minutes); err == nil && hi >= 0 && hi <= 59 {
		te.minute = hi
	} else {
		m := fmt.Sprintf("%s is not a valid minute (must be 00-59)", minutes)
		return te, errors.New(m)
	}

	ec := len(s)

	m := fmt.Sprintf("you must provide the second past the minute that the task should at (e.g. HH:MM:30)")

	if (ec < 3 && freq == time.Minute) || (freq == time.Minute && s[2] == "") {

		return te, errors.New(m)
	} else if s[2] != "" {

		seconds := s[2]

		if hi, err := strconv.Atoi(seconds); err == nil && hi >= 0 && hi <= 59 {
			te.second = hi
		} else {
			m := fmt.Sprintf("%s is not a valid second (must be 00-59)", seconds)
			return te, errors.New(m)
		}

	}

	return te, nil

}

func singleUnit(s string) bool {

	result := (s == daySingle || strings.HasPrefix(s, daySingleSpace) ||
		s == hourSingle || strings.HasPrefix(s, hourSingleSpace) ||
		s == minuteSingle || strings.HasPrefix(s, minuteSingleSpace) ||
		s == secondSingle || strings.HasPrefix(s, secondSingleSpace))

	return result

}

func validValue(s string) (int64, bool) {

	var v int64
	var err error

	if v, err = strconv.ParseInt(s, 10, 64); err != nil || v < 1 {
		return 0, false
	}

	return v, true

}

func validUnit(s string) (time.Duration, bool) {

	switch s {
	case "SECOND":
		fallthrough
	case "SECONDS":
		return time.Second, true
	case "MINUTE":
		fallthrough
	case "MINUTES":
		return time.Minute, true
	case "HOUR":
		fallthrough
	case "HOURS":
		return time.Hour, true
	case "DAY":
		fallthrough
	case "DAYS":
		return time.Hour * time.Duration(24), true

	}

	return 0, false
}

type interval struct {
	Mode            intervalMode
	OffsetFromStart time.Duration
	ActualStart     time.Time
	Frequency       time.Duration
	CalculatedAt    time.Time
}

type intervalMode int

const (
	//OffsetFromStart indicates the interval between invocations is relative to the time the application started
	OffsetFromStart intervalMode = iota

	//ActualStartTime indicates that the invocation will run at a specified time
	ActualStartTime
)

type intervalToken struct {
}

type timeElements struct {
	hour, minute, second int
}
