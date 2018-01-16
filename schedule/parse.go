// Copyright 2018 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package schedule

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"errors"
)

func parseEvery(every string) (*interval, error) {

	m := fmt.Sprintf("Cannot parse recurring schedule [%s]", every)
	i := new(interval)

	alphaNumRe := regexp.MustCompile("[^a-zA-Z0-9 ]+")

	norm := alphaNumRe.ReplaceAllString(every, "")

	norm = strings.TrimSpace(strings.ToUpper(norm))

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
		em := fmt.Sprintf("%s: %s cannot be interpreted as a valid unit (seconds, hours, days) ", m, tokens[1])
		return i, errors.New(em)
	}

	i.Frequency = time.Duration(frequencyValue) * frequencyUnit

	i.Mode = OFFSET_FROM_START
	i.CalculatedAt = time.Now()

	return i, nil

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
	case "SECONDS":
		return time.Second, true
	case "HOUR":
	case "HOURS":
		return time.Hour, true
	case "DAY":
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
