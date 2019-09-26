// Copyright 2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

/*
Package uuid provides tools for generating and validating various RFC 4122 compliant UUIDs

https://tools.ietf.org/html/rfc4122

*/
package uuid

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"
	"strings"
)

// V4 returns a valid V4 UUID using the default random number generator with no uniqueness checks
func V4() string {

	max128 := big.NewInt(0)

	max128.SetString("340282366920938463463374607431768211455", 10)

	r, _ := rand.Int(rand.Reader, max128)
	asBytes := r.Bytes()

	var b strings.Builder

	b.WriteString(hex.EncodeToString(asBytes[0:4]))
	b.WriteString("-")
	b.WriteString(hex.EncodeToString(asBytes[4:6]))

	//Get the hex pair that will contain the version, replace the first char with 4 but keep the second
	b.WriteString("-4")

	vByte := asBytes[6]

	b.WriteString(string(secondChar(vByte)))

	b.WriteString(hex.EncodeToString(asBytes[7:8]))
	b.WriteString("-")

	//Get the hex pair that will contain the variant and set the first two bits to 1 and 0
	varByte := asBytes[8]
	varByte |= (1 << 7)
	mask := ^(1 << 6)
	varByte &= byte(mask)
	b.WriteString(hex.EncodeToString([]byte{varByte}))

	b.WriteString(hex.EncodeToString(asBytes[9:10]))
	b.WriteString("-")
	b.WriteString(hex.EncodeToString(asBytes[10:16]))

	return b.String()
}

const hextable = "0123456789abcdef"

func secondChar(b byte) byte {
	return hextable[b&0x0f]
}
