package broker

import (
	"context"
	"fmt"
	"strings"

	"github.com/Guise322/TeleBot/server/internal/common"
	"github.com/Guise322/TeleBot/server/internal/entities/broker"

	kafka "github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

type KafkaProducer struct {
	w *kafka.Writer
}

var lastCommand string

func NewKafkaProducer(address string) KafkaProducer {
	w := &kafka.Writer{
		Addr:     kafka.TCP(address),
		Balancer: &kafka.LeastBytes{},
	}

	return KafkaProducer{w: w}
}

func (kt KafkaProducer) TransmitData(ctx context.Context, data broker.OutData) error {
	act := func(ctx context.Context) error {
		return kt.sendMessage(ctx, data)
	}

	return common.ExecuteWithRetries(ctx, act)
}

func (kt KafkaProducer) sendMessage(ctx context.Context, data broker.OutData) error {
	if lastCommand == "" && data.CommName == "" {
		logrus.Warn("Get an empty message")

		return nil
	}

	if data.CommName == "" {
		data.CommName = lastCommand
	}

	topicName := strings.Trim(data.CommName, "/")

	msg := kafka.Message{Topic: topicName, Value: []byte(data.Value)}
	err := kt.w.WriteMessages(ctx, msg)
	if err != nil {
		if err == kafka.UnknownTopicOrPartition {
			logrus.WithField("topiName", topicName).Warn("An unknown topic, create the one")
			createDataTopic(topicName, kt.w.Addr.String())
			err = kt.w.WriteMessages(ctx, msg)
		}
	}

	if err == nil {
		if data.CommName != "" {
			lastCommand = data.CommName
		}

		return nil
	}

	return fmt.Errorf("failed to write messages: %w", err)
}

func (kt KafkaProducer) Close() {
	if err := kt.w.Close(); err != nil {
		logrus.Error("Failed to close writer:", err)
	}
}