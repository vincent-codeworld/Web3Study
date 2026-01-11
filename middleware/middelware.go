package middleware

import (
	"context"
	"errors"
	"sync"
)

var Hook *Hooker

func init() {
	Hook.hookerTable = make([]Closer, 0)
}

type Hooker struct {
	hookerTable []Closer
}

func (hook *Hooker) Register(c Closer) {
	hook.hookerTable = append(hook.hookerTable, c)
}

func (hook *Hooker) Execute(ctx context.Context) error {
	var a = struct {
		wg sync.WaitGroup
		ch chan struct{}
	}{
		ch: make(chan struct{}),
	}
	for _, h := range hook.hookerTable {
		a.wg.Add(1)
		hooker := h
		go func() {
			_ = hooker.close()
			a.wg.Done()
		}()
	}
	go func() {
		a.wg.Wait()
		a.ch <- struct{}{}
	}()
	select {
	case <-ctx.Done():
		return errors.New("hookerTable excecution time out")
	case <-a.ch:
		return nil
	}
}

type Closer interface {
	close() error
}
