package main

import (
	"fmt"
	"net"
)

func main() {
	ServerAddr, _ := net.ResolveUDPAddr("udp", ":10001")
	ServerConn, _ := net.ListenUDP("udp", ServerAddr)
	defer ServerConn.Close()
	buf := make([]byte, 1024)
	for {
		n, addr, err := ServerConn.ReadFromUDP(buf)
		fmt.Println("Recived ", string(buf[0:n]), " from ", addr)
		if err != nil {
			fmt.Println("Error: ", err)
		}
	}
}
