package topic

import (
	"context"
	"net"
	"regexp"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/topics"
)

func CreateTopicIfNotExist(ctx context.Context, topicName string, addr net.Addr) error {
	cl := &kafka.Client{Addr: addr}
	exists, err := topicExists(ctx, topicName, cl)
	if err != nil {
		return err
	}

	if !exists {
		req := &kafka.CreateTopicsRequest{
			Addr: cl.Addr,
			Topics: []kafka.TopicConfig{
				{Topic: topicName, NumPartitions: 1, ReplicationFactor: 1},
			},
		}

		_, err := cl.CreateTopics(ctx, req)
		if err != nil {
			return err
		}
	}

	return nil
}

func topicExists(ctx context.Context, topicName string, client *kafka.Client) (bool, error) {
	topicReg := regexp.MustCompile(topicName)
	topics, err := topics.ListRe(ctx, client, topicReg)
	if err != nil {
		return false, err
	}

	return len(topics) > 0, nil
}
