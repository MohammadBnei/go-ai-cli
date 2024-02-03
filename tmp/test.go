package main

import "fmt"

type testStruct struct {
	name string
}

func two(s *testStruct) {
	s.name += " -"
	fmt.Println(s)
}

func main() {
	str := testStruct{name: "-"}
	two(&str)
	fmt.Println(str)
}
