package main

import (
	"fmt"
	"strconv"
)

var sendernames = [5]string{"Alice", "Bob", "Carsten", "Dennis", "Elisa"}
var recievernames = [5]string{"Frederik", "Gary", "Hailey", "Isabel", "Jesper"}

func send(c chan string, myname string) {
	for i := 0; i < 1000; i++ {
		c <- myname + "#" + strconv.Itoa(i)
	}
}

func receive(c chan string, myname string) {
	i := 0
	for {
		msg := <-c
		fmt.Println(myname + "#" + strconv.Itoa(i) + " " + msg)
		i++
	}
}

func main() {
	c := make(chan string)
	for i := 0; i < 5; i++ {
		go send(c, sendernames[i])
		go receive(c, recievernames[i])
	}

	receive(c, "Kacey")
}
