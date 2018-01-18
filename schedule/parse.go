// Copyright 2018 Granitic. All rights reserved.
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

const day_single = "DAY"
const day_single_space = day_single + " "
const minute_single = "MINUTE"
const minute_single_space = minute_single + " "
const hour_single = "HOUR"
const hour_single_space = hour_single + " "
const second_single = "SECOND"
const second_single_space = hour_single + " "
const double_space = "  "
const single_space = " "
const at = "AT"

const dayDuration = time.Hour * 24

func parseEvery(every string) (*interval, error) {
	return parseEveryFromGivenNow(every, time.Now())
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

	for strings.Contains(norm, double_space) {
		norm = strings.Replace(norm, double_space, single_space, -1)
	}

	tokens := strings.Split(norm, " ")

	if len(tokens) < 2 {
		return i, errors.New(m)
	}

	//Make such the first token is a postive integer
	frequencyValue, okay := validValue(tokens[0])

	if !okay {
		em := fmt.Sprintf("%s: %s cannot be interpreted as a positive integer greater than 1", m, tokens[0])
		return i, errors.New(em)
	}

	frequencyUnit, okay := validUnit(tokens[1])

	if !okay {
		em := fmt.Sprintf("%s: %s cannot be interpreted as a valid unit (seconds, minutes, hours, days) ", m, tokens[1])
		return i, errors.New(em)
	}

	i.Frequency = time.Duration(frequencyValue) * frequencyUnit

	if len(tokens) == 2 {

		i.Mode = OFFSET_FROM_START
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

	_, err := extractTimeElements(components[1:], i.Frequency)

	if err != nil {
		return err
	}

	switch {
	case i.Frequency == time.Minute:
	case i.Frequency == time.Hour:
	case i.Frequency == dayDuration:
	default:
		return errors.New("'at' can only be used if the frequency is a single minute, hour or day")
	}

	return nil
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
	} else if freq == time.Minute {

		seconds := s[2]

		if hi, err := strconv.Atoi(seconds); err == nil && hi >= 0 && hi <= 59 {
			te.second = hi
		} else {
			m := fmt.Sprintf("%s is not a valid second (must be 00-59)", seconds)
			return te, errors.New(m)
		}

	}

	return timeElements{}, nil

}

func singleUnit(s string) bool {

	result := (s == day_single || strings.HasPrefix(s, day_single_space) ||
		s == hour_single || strings.HasPrefix(s, hour_single_space) ||
		s == minute_single || strings.HasPrefix(s, minute_single_space) ||
		s == second_single || strings.HasPrefix(s, second_single_space))

	return result

}

func validValue(s string) (int64, bool) {
	if v, err := strconv.ParseInt(s, 10, 64); err != nil || v < 1 {
		return 0, false
	} else {
		return v, true
	}
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
	OFFSET_FROM_START intervalMode = iota
	ACTUAL_START_TIME
)

type intervalToken struct {
}

type timeElements struct {
	hour, minute, second int
}
