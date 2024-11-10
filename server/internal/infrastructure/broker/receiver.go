package broker

import (
	"context"
	"fmt"

	"github.com/Guise322/TeleBot/server/internal/common"
	"github.com/Guise322/TeleBot/server/internal/entities/broker"

	kafka "github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

type KafkaReceiver struct {
	address string
}

func NewKafkaReceiver(address string) KafkaReceiver {
	return KafkaReceiver{address: address}
}

func (kr KafkaReceiver) StartReceivingData(ctx context.Context) (<-chan broker.InData, error) {
	dataTopicName := "botdata"
	if err := createDataTopic(dataTopicName, kr.address); err != nil {
		return nil, fmt.Errorf("an error occurs of creating the data topic: %w", err)
	}

	dataChan := make(chan broker.InData)

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

func consumeMessages(ctx context.Context, dataChan chan broker.InData, r *kafka.Reader) {
	for {
		if ctx.Err() != nil {
			break
		}

		msg, err := fetchMesWithRetries(ctx, r)
		if err != nil {
			logrus.Error("Can not get a message from the broker: ", err)

			continue
		}

		err = commitMesWithRetries(ctx, msg, r)
		if err != nil {
			logrus.Error("Can not commit a message in the broker: ", err)

			continue
		}

		commandKey := "command"
		isCommand := string(msg.Key) == commandKey
		dataChan <- broker.InData{IsCommand: isCommand, Value: string(msg.Value)}
	}

	if err := r.Close(); err != nil {
		logrus.Error("Failed to close the reader: ", err)
	}
}

func commitMesWithRetries(ctx context.Context, msg kafka.Message, r *kafka.Reader) error {
	act := func(ctx context.Context) error {
		return r.CommitMessages(ctx, msg)
	}

	return common.ExecuteWithRetries(ctx, act)
}

func fetchMesWithRetries(ctx context.Context, r *kafka.Reader) (kafka.Message, error) {
	var msg kafka.Message

	act := func(ctx context.Context) error {
		var err error
		msg, err = r.FetchMessage(ctx)

		return err
	}

	err := common.ExecuteWithRetries(ctx, act)

	return msg, err
}
