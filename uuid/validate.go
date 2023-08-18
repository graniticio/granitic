// Copyright 2019-2023 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package uuid

import (
	"fmt"
	"strconv"
	"strings"
)

const uuidLen = 36
const dashOneIndex = 8
const dashTwoIndex = 13
const dashThreeIndex = 18
const dashFourIndex = 23

const v4Version = "0100"

// ValidV4 returns true if the supplied string is a valid version 4 UUID according to RFC 4122
func ValidV4(uuid string) bool {
	if !ValidFormat(uuid) {

		fmt.Println("Inv format")

		return false
	}

	csh := extractBinaryField(clockSeqAndReserved, uuid)
	thv := extractBinaryField(timeHighAndVersion, uuid)

	if thv[0:4] != v4Version {
		return false
	}

	if csh[0:2] != "10" {
		return false
	}

	return true

}

// ValidFormat returns true if the supplied string is in accordance with the ABNF defined in Section 3 of RFC 4122
func ValidFormat(uuid string) bool {

	if len(uuid) != uuidLen {
		return false
	}

	for i, c := range uuid {

		if i == dashOneIndex || i == dashTwoIndex || i == dashThreeIndex || i == dashFourIndex {
			if c != '-' {
				return false
			}
		} else if !((c >= 48 && c <= 57) || (c >= 97 && c <= 102) || (c >= 65 && c <= 70)) {
			//0-9 or a-f or A-F

			return false
		}

	}

	return true
}

type field int

const (
	timeLow field = iota
	timeMid
	timeHighAndVersion
	clockSeqAndReserved
	clockSeqLow
	node
)

/*
   UUID                   = time-low "-" time-mid "-"
                            time-high-and-version "-"
                            clock-seq-and-reserved
                            clock-seq-low "-" node
   time-low               = 4hexOctet
   time-mid               = 2hexOctet
   time-high-and-version  = 2hexOctet
   clock-seq-and-reserved = hexOctet
   clock-seq-low          = hexOctet
   node                   = 6hexOctet
   hexOctet               = hexDigit hexDigit
*/

func extractField(f field, uuid string) string {

	res := uuid

	switch f {
	case timeLow:
		res = uuid[0:8]
	case timeMid:
		res = uuid[9:13]
	case timeHighAndVersion:
		res = uuid[14:18]
	case clockSeqAndReserved:
		res = uuid[19:21]
	case clockSeqLow:
		res = uuid[21:23]
	case node:
		res = uuid[24:36]
	}

	return res
}

func extractBinaryField(f field, uuid string) string {
	var expectedLen int

	switch f {
	case timeLow:
		expectedLen = 32
	case timeMid:
		expectedLen = 16
	case timeHighAndVersion:
		expectedLen = 16
	case clockSeqAndReserved:
		expectedLen = 8
	case clockSeqLow:
		expectedLen = 8
	case node:
		expectedLen = 48
	}

	field := extractField(f, uuid)

	ui, _ := strconv.ParseUint(field, 16, 64)

	return paddedBinary(ui, expectedLen)

}

func paddedBinary(n uint64, length int) string {

	f := strconv.FormatInt(int64(n), 2)

	missing := length - len(f)

	if missing <= 0 {
		return f
	}

	var str strings.Builder

	for missing > 0 {
		str.WriteString("0")
		missing--
	}

	str.WriteString(f)

	return str.String()
}
