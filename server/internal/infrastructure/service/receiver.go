package service

import (
	"context"

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

func (kr KafkaReceiver) StartReceivingData(ctx context.Context) <-chan service.InData {
	dataTopicName := "botdata"
	createDataTopic(dataTopicName, kr.address)
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

	return dataChan
}

func consumeMessages(ctx context.Context, dataChan chan service.InData, r *kafka.Reader) {
	for {
		if ctx.Err() != nil {
			break
		}

		msg, err := r.FetchMessage(ctx)
		if err != nil {
			continue
		}

		commandKey := "command"
		isCommand := string(msg.Key) == commandKey

		dataChan <- service.InData{IsCommand: isCommand, Value: string(msg.Value)}
		r.CommitMessages(ctx, msg)
	}

	if err := r.Close(); err != nil {
		logrus.Fatal("failed to close reader:", err)
	}
}
