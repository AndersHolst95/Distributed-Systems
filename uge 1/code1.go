package main

import (
	"fmt"
)

func multi(x int, y int) (string, int) {
	return "the answer is", x + y
}

// Named is a sturct
type Named struct {
	name string // Member of a class
}

// PrintName is something
func (nm *Named) PrintName(n int) {
	if n < 0 {
		panic(-1)
	}
	for i := 0; i < n; i++ {
		fmt.Print(nm.name + "\n")
	}
}

func main() {
	var i int
	i = 21
	j := 21

	decr := func() int {
		j = j - 7
		return j
	}

	str, m := multi(i, j)
	defer fmt.Println(str, m)

	fmt.Println(decr())
	fmt.Println(decr())

	nm1 := Named{name: "Jesper"}
	nm2 := &Named{}

	nm1.PrintName(2)
	nm2.PrintName(2)
	nm2.name = "Claudio"
	nm2.PrintName(2)
	var nm3 *Named = nm2
	nm2.name = "Ivan"
	nm3.PrintName(2)
	nm3.PrintName(-1)
	fmt.Println("Will we make it here?")
}
