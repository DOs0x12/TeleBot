package producer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/DOs0x12/TeleBot/server/v2/internal/entities/broker"
	"github.com/DOs0x12/TeleBot/server/v2/retry"

	kafka "github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

type KafkaProducer struct {
	w *kafka.Writer
}

type ProducerDataDto struct {
	CommName string
	ChatID   int64
	Value    []byte
	IsFile   bool
}

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
	rCnt := 5
	rDur := 1 * time.Second
	return retry.ExecuteWithRetries(ctx, act, rCnt, rDur)
}

var (
	lastCommand,
	lastToken string
)

func (kt KafkaProducer) sendMessage(ctx context.Context, data broker.DataTo) error {
	if lastCommand == "" && data.CommName == "" {
		logrus.Warn("Got an empty message")

		return nil
	}

	if data.CommName == "" {
		data.CommName = lastCommand
		data.Token = lastToken
	}

	dataDto := ProducerDataDto{
		CommName: data.CommName,
		ChatID:   data.ChatID,
		Value:    data.Value,
		IsFile:   data.IsFile,
	}

	rawData, err := json.Marshal(dataDto)
	if err != nil {
		return fmt.Errorf("failed to marshal data to send: %w", err)
	}

	msg := kafka.Message{Topic: data.Token, Value: rawData}
	err = kt.w.WriteMessages(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to write messages: %w", err)
	}

	lastCommand = data.CommName
	lastToken = data.Token

	return nil
}

func (kt KafkaProducer) Close() error {
	if err := kt.w.Close(); err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	return nil
}
