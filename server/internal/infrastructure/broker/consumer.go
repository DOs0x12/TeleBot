package broker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Guise322/TeleBot/server/internal/common"
	"github.com/Guise322/TeleBot/server/internal/entities/broker"
	"github.com/google/uuid"

	kafka "github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

type KafkaConsumer struct {
	mu                  *sync.Mutex
	reader              *kafka.Reader
	uncommittedMessages map[uuid.UUID]uncommittedMessage
	offcets             map[int]offcetWithTimeStamp
}

type uncommittedMessage struct {
	msg       kafka.Message
	timeStamp time.Time
}

type offcetWithTimeStamp struct {
	value     int64
	timeStamp time.Time
}

func NewKafkaConsumer(address string) (*KafkaConsumer, error) {
	dataTopicName := "botdata"
	if err := createDataTopic(dataTopicName, address); err != nil {
		return nil, fmt.Errorf("an error occurs of creating the data topic: %w", err)
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		GroupID:     "TeleBotWorker",
		Brokers:     []string{address},
		Topic:       dataTopicName,
		MaxBytes:    10e6,
		StartOffset: kafka.LastOffset,
	})

	cons := KafkaConsumer{
		uncommittedMessages: make(map[uuid.UUID]uncommittedMessage),
		offcets:             make(map[int]offcetWithTimeStamp),
		mu:                  &sync.Mutex{},
		reader:              reader,
	}

	return &cons, nil
}

func (kr *KafkaConsumer) StartReceivingData(ctx context.Context) (<-chan broker.InData, error) {
	dataChan := make(chan broker.InData)

	go kr.consumeMessages(ctx, dataChan)

	return dataChan, nil
}

func (kr *KafkaConsumer) consumeMessages(ctx context.Context, dataChan chan broker.InData) {
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
		kr.uncommittedMessages[msgUuid] = uncommittedMessage{msg: msg, timeStamp: time.Now()}
		kr.mu.Unlock()

		commandKey := "command"
		isCommand := string(msg.Key) == commandKey
		dataChan <- broker.InData{IsCommand: isCommand, Value: string(msg.Value), MsgUuid: msgUuid}
	}

	if err := kr.reader.Close(); err != nil {
		logrus.Error("Failed to close the reader: ", err)
	}
}

func (kr *KafkaConsumer) fetchMesWithRetries(ctx context.Context) (kafka.Message, error) {
	var msg kafka.Message

	act := func(ctx context.Context) error {
		var err error
		msg, err = kr.reader.FetchMessage(ctx)

		return err
	}

	err := common.ExecuteWithRetries(ctx, act)

	return msg, err
}
