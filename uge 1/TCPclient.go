package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

var conn net.Conn

func main() {
	conn, _ := net.Dial("tcp", "127.0.0.1:18081")
	defer conn.Close()
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("> ")
		text, err := reader.ReadString('\n')
		if text == "quit\n" {
			return
		}
		fmt.Fprint(conn, text)
		msg, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			return
		}
		fmt.Print("From server: " + msg)
	}
}
