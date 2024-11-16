package receiver

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Guise322/TeleBot/server/internal/common"
	"github.com/Guise322/TeleBot/server/internal/entities/broker"
	brCom "github.com/Guise322/TeleBot/server/internal/infrastructure/broker/common"
	"github.com/google/uuid"

	kafka "github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

type KafkaReceiver struct {
	address            string
	processingMessages map[uuid.UUID]processingMessage
	mu                 *sync.Mutex
	reader             *kafka.Reader
}

type processingMessage struct {
	msg       kafka.Message
	timeStamp time.Time
}

func NewKafkaReceiver(address string) (KafkaReceiver, error) {
	dataTopicName := "botdata"
	if err := brCom.CreateDataTopic(dataTopicName, address); err != nil {
		return KafkaReceiver{}, fmt.Errorf("an error occurs of creating the data topic: %w", err)
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		GroupID:     "regdfgd1",
		Brokers:     []string{address},
		Topic:       dataTopicName,
		Partition:   0,
		MaxBytes:    10e6,
		StartOffset: kafka.LastOffset,
	})

	cons := KafkaReceiver{
		address:            address,
		processingMessages: make(map[uuid.UUID]processingMessage),
		mu:                 &sync.Mutex{},
		reader:             reader,
	}

	return cons, nil
}

func (kr KafkaReceiver) StartReceivingData(ctx context.Context) (<-chan broker.InData, error) {
	dataChan := make(chan broker.InData)

	go kr.consumeMessages(ctx, dataChan)

	return dataChan, nil
}

func (kr KafkaReceiver) consumeMessages(ctx context.Context, dataChan chan broker.InData) {
	for {
		if ctx.Err() != nil {
			break
		}

		msg, err := kr.fetchMesWithRetries(ctx)
		if err != nil {
			logrus.Error("Can not get a message from the broker: ", err)

			continue
		}

		msgUuid := uuid.New()
		kr.mu.Lock()
		kr.processingMessages[msgUuid] = processingMessage{msg: msg, timeStamp: time.Now()}
		kr.mu.Unlock()

		commandKey := "command"
		isCommand := string(msg.Key) == commandKey
		dataChan <- broker.InData{IsCommand: isCommand, Value: string(msg.Value), MsgUuid: msgUuid}
	}

	if err := kr.reader.Close(); err != nil {
		logrus.Error("Failed to close the reader: ", err)
	}
}

func (kr KafkaReceiver) fetchMesWithRetries(ctx context.Context) (kafka.Message, error) {
	var msg kafka.Message

	act := func(ctx context.Context) error {
		var err error
		msg, err = kr.reader.FetchMessage(ctx)

		return err
	}

	err := common.ExecuteWithRetries(ctx, act)

	return msg, err
}
