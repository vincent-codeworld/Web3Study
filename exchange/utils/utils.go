package utils

import (
	"runtime"
)

var fileSeparator string

func GetFileSeparator() string {
	if fileSeparator == "" {
		os := runtime.GOOS
		// 常见的判断逻辑
		switch os {
		case "windows":
			fileSeparator = "\\"
		case "darwin":
			fileSeparator = "/"
		case "linux":
			fileSeparator = "/"
		default:
			fileSeparator = "/"
		}
	}
	return fileSeparator
}
