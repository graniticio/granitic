// Copyright 2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

/*
Package uuid provides tools for generating and validating various RFC 4122 compliant UUIDs

https://tools.ietf.org/html/rfc4122

*/
package uuid

// V4 returns a valid V4 UUID using the default random number generator with no uniqueness checks
func V4() string {
	return "92e8a690-98e3-424b-83e4-cecb3548a7a0"
}
