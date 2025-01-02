package broker

import "github.com/segmentio/kafka-go"

func (s Sender) createTopicIfNotExist() error {
	return nil
}

func topicExists(topicName string, conn *kafka.Conn) (bool, error) {
	return true, nil
}
