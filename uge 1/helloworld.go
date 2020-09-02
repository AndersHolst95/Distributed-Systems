package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("Hello world!")
	var i int
	i = 3
	k := 6
	k++
	fmt.Println(k + i)
	go main()
	time.Sleep(time.Second)
}
