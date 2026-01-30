package middleware

import (
	"github.com/segmentio/kafka-go"
)

var Consumer *kafka.Reader

func init() {
	Consumer = kafka.NewReader(kafka.ReaderConfig{
		// 手动提交
		CommitInterval: 0,
	})
}
