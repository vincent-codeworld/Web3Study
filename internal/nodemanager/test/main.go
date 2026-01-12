package main

import "fmt"

func main() {
	s := make([]string, 10)
	s = append(s, "Hello World")
	for _, ss := range s {
		fmt.Println(ss)
	}
}
