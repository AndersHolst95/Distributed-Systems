package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()
	myEnd := conn.LocalAddr().String()
	otherEnd := conn.RemoteAddr().String()
	for {
		msg, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Println("Ending session with " + otherEnd)
			return
		} else {
			fmt.Print("From "+otherEnd+" to "+myEnd+": ", string(msg))
			titlemsg := strings.Title(msg)
			conn.Write([]byte(titlemsg))
		}
	}
}

func main() {
	ln, _ := net.Listen("tcp", ":18081")
	defer ln.Close()
	for {
		fmt.Println("Listening for connection...")
		conn, _ := ln.Accept()
		fmt.Println("Got a connection...")
		go handleConnection(conn)
	}
}
