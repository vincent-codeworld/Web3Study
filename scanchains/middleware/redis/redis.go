package redis

import (
	"Web3Study/scanchains/middleware"
)

func init() {
	middleware.Hook.Register(close)
}

func close() error {
	return nil
}
