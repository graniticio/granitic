// Copyright 2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

/*
Package uuid provides tools for generating and validating various RFC 4122 compliant UUIDs

https://tools.ietf.org/html/rfc4122

*/
package uuid

import (
	"crypto/rand"
	"math/big"
	"strings"
)

var default128Gen random128

// V4 returns a valid V4 UUID using the default random number generator with no uniqueness checks
func V4() string {

	if default128Gen == nil {
		default128Gen = newCrypto128().Generate
	}

	return V4Custom(default128Gen, StandardEncoder)

}

// V4Custom returns a valid V4 UUID with a custom random number generator and encoder
func V4Custom(gen random128, e encoder) string {

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

type random128 func() []byte

type encoder func([]byte) string

// StandardEncoder encodes the supplied 128 bit number as a standard dash seperated UUID
func StandardEncoder(asBytes []byte) string {
	var sb strings.Builder

	for i, b := range asBytes {

		if (i == 4) || (i == 6) || (i == 8) || (i == 10) {
			sb.WriteRune('-')
		}

		sb.WriteByte(hextable[b>>4])
		sb.WriteByte(hextable[b&0x0f])
	}

	return sb.String()
}

func newCrypto128() *crypto128 {

	c := new(crypto128)
	c.max = big.NewInt(0)

	c.max.SetString("340282366920938463463374607431768211455", 10)

	return c
}

type crypto128 struct {
	max *big.Int
}

func (c *crypto128) Generate() []byte {
	r, _ := rand.Int(rand.Reader, c.max)
	asBytes := r.Bytes()

	missingBytes := 16 - len(asBytes)

	if missingBytes > 0 {
		pad := make([]byte, missingBytes)
		asBytes = append(pad, asBytes...)
	}

	return asBytes
}
