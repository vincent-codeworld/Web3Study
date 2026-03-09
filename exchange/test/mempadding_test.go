package test

import (
	"fmt"
	"sync/atomic"
	"testing"
)

func TestMemPadding(t *testing.T) {
	type BadStruct struct {
		A uint32 // 4 字节
		//B uint64 // 8 字节
		B atomic.Uint64
	}
	instance := BadStruct{}
	a := atomic.AddUint64(&instance.B, 1)
	fmt.Println(a)
}
