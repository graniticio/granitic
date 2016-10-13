// Copyright 2016 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package httpserver

import (
	"fmt"
	"net"
	"testing"
)

func TestPortBinding(t *testing.T) {

	addr, err := net.InterfaceAddrs()

	fmt.Println(err)

	for _, a := range addr {
		fmt.Println(a.String())
		fmt.Println(a.Network())
	}

	inf, err := net.Interfaces()

	fmt.Println(err)

	for _, i := range inf {
		fmt.Println(i.Name)
		fmt.Println(i.Addrs())
	}

	_, err = net.Listen("tcp", "[::1]:8088")

	//err = http.ListenAndServe(":8080", nil)

	fmt.Println(err)

}
