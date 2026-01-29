package main

import (
	"fmt"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

/*
*

	RingBuffer 是单生产者多消费者无锁模式
*/
type RingBuffer struct {
	data      []any
	_padding1 [64]byte //解决伪共享带来的cpu cache line xao
	head      uint64
	_padding2 [64]byte
	tail      uint64
	mask      uint64
	size      uint64
}

func NewRingBuffer(size uint64) *RingBuffer {
	return &RingBuffer{
		data: make([]any, size),
		size: size,
		mask: size - 1,
	}
}

func (rb *RingBuffer) Put(d any) {
	retry := 0
	for {
		head := atomic.LoadUint64(&rb.head)
		if rb.tail-head >= rb.size {
			if retry < 5 {
				runtime.Gosched()
			} else {
				time.Sleep(time.Millisecond * 100)
			}
			retry++
			continue
		}
		break
	}
	rb.data[rb.tail&rb.mask] = d
	//atomic.AddUint64(&rb.tail, 1)  下面的性能高点
	atomic.StoreUint64(&rb.tail, rb.tail+1)
}

func (rb *RingBuffer) Get() any {
	retry := 0
	for {
		tail := atomic.LoadUint64(&rb.tail)
		head := atomic.LoadUint64(&rb.head)
		if head >= tail {
			if retry < 5 {
				runtime.Gosched()
			} else {
				time.Sleep(time.Millisecond * 100)
			}
			retry++
			continue
		}
		d := rb.data[head&rb.mask]
		if atomic.CompareAndSwapUint64(&rb.head, head, head+1) {
			return d
		}
	}
}

func main() {
	rb := NewRingBuffer(1024)
	producer := func() {
		for i := 0; i < 10000; i++ {
			rb.Put(i)
		}
	}
	consumer := func(uid string) {
		for {
			fmt.Printf("%s:%d\n", uid, rb.Get())
			time.Sleep(time.Millisecond * 100)
		}
	}
	go producer()

	for i := 0; i < 100; i++ {
		go consumer(uuid.New().String())
	}
	time.Sleep(time.Second * 20)
}
