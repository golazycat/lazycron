package main

import "fmt"

type A struct {
	a int
	b int
}

type B struct {
	A
	s string
}

func main() {

	b := B{
		A: A{
			a: 1,
			b: 2,
		},
		s: "hello",
	}

	a := b.A

	fmt.Printf("%+v\n", a)
}
