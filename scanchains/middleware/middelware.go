package middleware

import (
	"context"
	"errors"
	"sync"
)

var Hook *Hooker

func init() {
	Hook.hookerFuncTable = make([]HookFunc, 0)
}

type HookFunc func() error
type Hooker struct {
	hookerFuncTable []HookFunc
}

func (hook *Hooker) Register(h HookFunc) {
	hook.hookerFuncTable = append(hook.hookerFuncTable, h)
}

func (hook *Hooker) Execute(ctx context.Context) error {
	var a = struct {
		wg sync.WaitGroup
		ch chan struct{}
	}{
		ch: make(chan struct{}),
	}
	for _, h := range hook.hookerFuncTable {
		a.wg.Add(1)
		hooker := h
		go func() {
			_ = hooker()
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
