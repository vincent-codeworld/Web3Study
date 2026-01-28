package middleware

import "github.com/IBM/sarama"

var consumer sarama.ConsumerGroup

func init() {
	client, err := sarama.NewConsumerGroup([]string{}, "", nil)
	if err != nil {
		panic(err)
	}
	consumer = client
}
