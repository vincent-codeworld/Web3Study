package main

import (
	matcher "Web3Study/exchange/internal/matching_engine"
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	signs := make(chan os.Signal, 1)
	signal.Notify(signs, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancelFunc := context.WithCancel(context.Background())
	engine := matcher.NewMatchEngine(ctx)
	engine.Start()

	<-signs
	cancelFunc()
	go func() {
		engine.Stop()
	}()
	timeoutCtx, timeOutFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer timeOutFunc()
	<-timeoutCtx.Done()
}
