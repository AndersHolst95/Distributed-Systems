package main

import (
	"bufio"
	"fmt"
	"net"
)

func main() {
	addrs, _ := net.LookupHost("google.com")
	addr := addrs[0]
	fmt.Println(addr)
	conn, err := net.Dial("tcp", addr+":80")
	if conn != nil {
		defer conn.Close()
	}
	if err != nil {
		panic(0)
	}
	fmt.Fprint(conn, "GET /search?q=Secure+Distributed+Systems HTTP/1.1\n")
	fmt.Fprint(conn, "HOST: www.google.com")
	fmt.Fprint(conn, "\n")
	for {
		msg, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			panic(1)
		}
		fmt.Println(msg)
	}
}
