package main

import (
	mw "Web3Study/scanchains/middleware"
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	sign := make(chan os.Signal, 1)
	signal.Notify(sign, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)

	<-sign
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	_ = mw.Hook.Execute(ctx)
}
