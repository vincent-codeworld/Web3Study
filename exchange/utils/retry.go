package utils

import "fmt"

func Retry(t int, f func() error) error {
	if t <= 0 {
		return fmt.Errorf("invalid retry count: %d", t)
	}
	var err error
	for i := 0; i < t; i++ {
		if err = f(); err == nil {
			break
		}
	}
	return err
}
