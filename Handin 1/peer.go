package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

var conns []net.Conn
var msgSent map[string]bool

func peer() {
	msgSent = make(map[string]bool)

	fmt.Println("Please enter the address of a peer")
	reader := bufio.NewReader(os.Stdin)
	addr, errip := reader.ReadString('\n')
	if errip != nil {
		fmt.Println("GØR NOGET")
	}

	//addr = addr[:len(addr)-2] // Remove the \n delimiter
	addr = strings.TrimRight(addr, "\r\n")

	fmt.Println("I am trying to connect to " + addr)
	// Try to connect
	conn, err := net.Dial("tcp", addr)
	if err == nil {
		conns = append(conns, conn)
		go msgReceiver(conn)
		fmt.Println("Connected to address")
	} else {
		fmt.Println("No peer at address")
	}

	// Create server
	ln, _ := net.Listen("tcp", ":0")
	fmt.Println("Server waiting for connection at " + ln.Addr().String())
	go openConnection(ln)

	fmt.Println("Ready for input. Send your message!")
	for {
		msg, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		broadcastMsg(msg)
	}
}

func msgReceiver(conn net.Conn) {
	for {
		msg, _ := bufio.NewReader(conn).ReadString('\n')
		broadcastMsg(msg)
	}
}

func broadcastMsg(msg string) {
	if !msgSent[msg] { // Message has NOT been sent before
		msgSent[msg] = true
		fmt.Print(msg)
		for _, conn := range conns {
			conn.Write([]byte(msg))
		}
	}
}

func openConnection(ln net.Listener) {
	fmt.Println("Waiting for connection...")
	conn, _ := ln.Accept()
	conns = append(conns, conn)
	go msgReceiver(conn)
	openConnection(ln)
}

func main() {
	msgSent = make(map[string]bool)
	peer()
}
