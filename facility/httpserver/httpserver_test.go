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
