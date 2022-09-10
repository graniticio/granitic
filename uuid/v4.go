// Copyright 2019-2022 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

/*
Package uuid provides tools for generating and validating various RFC 4122 compliant UUIDs

https://tools.ietf.org/html/rfc4122

Default V4 UUID generation is benchmarked at ~500ns per generation (2.6 GHz Core i7 (I7-8850H))
and does not benefit from any buffering/pre-generation.
*/
package uuid

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/base64"
	"io"
)

var default128Gen Generate16Byte

// V4 returns a valid V4 UUID using the default random number generator with no uniqueness checks
func V4() string {

	return V4Custom(GenerateCryptoRand, StandardEncoder)

}

// V4Custom returns a valid V4 UUID with a custom random number generator and EncodeFrom16Byte
func V4Custom(gen Generate16Byte, e EncodeFrom16Byte) string {

	asBytes := gen()

	//Get the hex pair that will contain the variant and set the first two bits to 1 and 0
	varByte := asBytes[8]
	varByte |= (1 << 7)
	mask := ^(1 << 6)
	varByte &= byte(mask)

	asBytes[8] = varByte

	//Get the hex pair that will contain the version, replace the first four bits with 0100
	vByte := asBytes[6]
	mask = ^(1 << 7)
	vByte &= byte(mask)
	vByte |= (1 << 6)
	mask = ^(1 << 5)
	vByte &= byte(mask)
	mask = ^(1 << 4)
	vByte &= byte(mask)

	asBytes[6] = vByte

	return StandardEncoder(asBytes)
}

const hextable = "0123456789abcdef"

// Bytes16 is a 128-bit number represented as 16 bytes
type Bytes16 [16]byte

// Generate16Byte is a type of function able to create unsigned 128-bit numbers represented as a sequence of 16 bytes.
type Generate16Byte func() Bytes16

// EncodeFrom16Byte takes a 16 byte/128-bit representation of a UUID and encodes it as a string
type EncodeFrom16Byte func(Bytes16) string

// StandardEncoder encodes the supplied 128 bit number as a standard dash separated UUID
func StandardEncoder(asBytes Bytes16) string {
	buf := make([]byte, 36)

	writeIndex := 0

	for i, b := range asBytes {

		if (i == 4) || (i == 6) || (i == 8) || (i == 10) {
			buf[writeIndex] = ('-')
			writeIndex++
		}

		buf[writeIndex] = hextable[b>>4]
		buf[writeIndex+1] = hextable[b&0x0f]

		writeIndex += 2
	}

	return string(buf)
}

// Base32Encoder encodes the supplied 128 bit number as a base32 string
func Base32Encoder(asBytes Bytes16) string {
	return base32.StdEncoding.EncodeToString(asBytes[:])
}

// Base64Encoder encodes the supplied 128 bit number as a base32 string
func Base64Encoder(asBytes Bytes16) string {
	return base64.StdEncoding.EncodeToString(asBytes[:])
}

// GenerateCryptoRand creates a random 128 bit number using crypto/rand.Reader as a source
func GenerateCryptoRand() Bytes16 {

	asBytes := Bytes16{}

	io.ReadFull(rand.Reader, asBytes[:])

	return asBytes
}
