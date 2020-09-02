package main

import (
	"fmt"
	"net"
	"strconv"
)

func main() {
	addrs, _ := net.LookupHost("google.com")
	for indx, addr := range addrs {
		fmt.Println("Address number " + strconv.Itoa(indx) + ": " + addr)
	}
}
