package postgres

import "Web3Study/middleware"

func init() {
	middleware.Hook.Register(close)
}

func close() error {
	return nil
}
