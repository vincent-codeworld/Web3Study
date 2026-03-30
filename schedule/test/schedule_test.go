package test

import (
	"fmt"
	"testing"
)

func TestA(t *testing.T) {
	cpu := 16
	i := 1 << (uint(cpu) % 8)
	fmt.Println(i)
}
