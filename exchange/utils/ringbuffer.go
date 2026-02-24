package utils

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
type RingBuffer[T any] struct {
	head uint64
	_    [56]byte
	tail uint64
	_    [56]byte
	data []T
	mask uint64
	size uint64
}

func NewRingBuffer[T any](size uint64) *RingBuffer[T] {
	return &RingBuffer[T]{
		data: make([]T, size),
		size: size,
		mask: size - 1,
	}
}

func (rb *RingBuffer[T]) Put(d T) {
	for {
		head := atomic.LoadUint64(&rb.head)
		if rb.tail-head >= rb.size {
			runtime.Gosched()
			continue
		}
		break
	}
	rb.data[rb.tail&rb.mask] = d
	//atomic.AddUint64(&rb.tail, 1)  下面的性能高点
	atomic.StoreUint64(&rb.tail, rb.tail+1)
}

func (rb *RingBuffer[T]) Get() T {
	for {
		tail := atomic.LoadUint64(&rb.tail)
		head := atomic.LoadUint64(&rb.head)
		if head >= tail {
			runtime.Gosched()
			continue
		}
		d := rb.data[head&rb.mask]
		if atomic.CompareAndSwapUint64(&rb.head, head, head+1) {
			return d
		}
	}
}

func main() {
	rb := NewRingBuffer[int](1024)
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
