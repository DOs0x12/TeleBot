package producer

import (
	"context"
	"fmt"

	"github.com/DOs0x12/TeleBot/server/internal/common/retry"
	"github.com/DOs0x12/TeleBot/server/internal/entities/broker"

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

func (kt KafkaProducer) TransmitData(ctx context.Context, data broker.DataTo) error {
	act := func(ctx context.Context) error {
		return kt.sendMessage(ctx, data)
	}

	return retry.ExecuteWithRetries(ctx, act)
}

func (kt KafkaProducer) sendMessage(ctx context.Context, data broker.DataTo) error {
	if lastCommand == "" && data.CommName == "" {
		logrus.Warn("Got an empty message")

		return nil
	}

	if data.CommName == "" {
		data.CommName = lastCommand
	}

	msg := kafka.Message{Topic: data.Token, Value: []byte(data.Value)}
	err := kt.w.WriteMessages(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to write messages: %w", err)
	}

	if data.CommName != "" {
		lastCommand = data.Token
	}

	return nil
}

func (kt KafkaProducer) Close() {
	if err := kt.w.Close(); err != nil {
		logrus.Error("Failed to close writer:", err)
	}
}
