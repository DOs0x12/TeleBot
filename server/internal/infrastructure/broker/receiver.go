package service

import (
	"context"
	"fmt"
	"time"

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

			time.Sleep(waitTime)

			continue
		}

		commandKey := "command"
		isCommand := string(msg.Key) == commandKey

		dataChan <- service.InData{IsCommand: isCommand, Value: string(msg.Value)}
		err = r.CommitMessages(ctx, msg)
		if err != nil {
			logrus.Errorf("Can not commit a message: %v", err)

			time.Sleep(waitTime)

			continue
		}

		time.Sleep(waitTime)
	}

	if err := r.Close(); err != nil {
		logrus.Fatal("failed to close reader:", err)
	}
}
