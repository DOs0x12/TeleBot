package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Guise322/TeleBot/server/common"
	"github.com/Guise322/TeleBot/server/internal/entities/service"

	kafka "github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

type KafkaReceiver struct {
	address string
}

func NewKafkaReceiver(address string) KafkaReceiver {
	return KafkaReceiver{address: address}
}

func (kr KafkaReceiver) StartReceivingData(ctx context.Context) (<-chan service.InData, error) {
	dataTopicName := "botdata"
	if err := createDataTopic(dataTopicName, kr.address); err != nil {
		return nil, fmt.Errorf("an error occurs of creating the data topic: %w", err)
	}

	dataChan := make(chan service.InData)

	r := kafka.NewReader(kafka.ReaderConfig{
		GroupID:     "regdfgd1",
		Brokers:     []string{kr.address},
		Topic:       dataTopicName,
		Partition:   0,
		MaxBytes:    10e6,
		StartOffset: kafka.LastOffset,
	})

	go consumeMessages(ctx, dataChan, r)

	return dataChan, nil
}

func consumeMessages(ctx context.Context, dataChan chan service.InData, r *kafka.Reader) {
	for {
		if ctx.Err() != nil {
			break
		}

		waitTime := 500 * time.Millisecond

		msg, err := r.FetchMessage(ctx)
		if err != nil {
			logrus.Errorf("Can not get a message from the broker: %v", err)

			common.WaitWithContext(ctx, waitTime)

			continue
		}

		if !commitMesWithRetries(ctx, msg, r) {
			continue
		}

		commandKey := "command"
		isCommand := string(msg.Key) == commandKey
		dataChan <- service.InData{IsCommand: isCommand, Value: string(msg.Value)}

		common.WaitWithContext(ctx, waitTime)
	}

	if err := r.Close(); err != nil {
		logrus.Fatal("Failed to close the reader:", err)
	}
}

func commitMesWithRetries(ctx context.Context, msg kafka.Message, r *kafka.Reader) bool {
	retryNum := 10
	waitTime := 5 * time.Second

	for i := 0; i < retryNum; i++ {
		err := r.CommitMessages(ctx, msg)
		if err == nil {
			return true
		}

		logrus.Errorf("Can not commit a message: %v", err)

		common.WaitWithContext(ctx, waitTime)
	}

	return false
}
