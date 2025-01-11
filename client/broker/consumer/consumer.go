package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/DOs0x12/TeleBot/server/v2/system"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

type uncommittedMessage struct {
	msg       kafka.Message
	timeStamp time.Time
}

type offsetWithTimeStamp struct {
	value     int64
	timeStamp time.Time
}

type KafkaConsumerData struct {
	ChatID      int64
	Value       string
	MessageUuid uuid.UUID
}

// Consumer gets data from the bot app. It implements a way to store
// the uncommitted messages and offsets to commit them later.
type KafkaConsumer struct {
	mu                  *sync.Mutex
	reader              *kafka.Reader
	uncommittedMessages map[uuid.UUID]uncommittedMessage
	offsets             map[int]offsetWithTimeStamp
}

// Create a consumer to read data from a Kafka instance.
func NewKafkaConsumer(address string, serviceName string) (*KafkaConsumer, error) {
	topicName, err := system.GetServiceToken(serviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to create a receiver: %w", err)
	}

	r := kafka.NewReader(kafka.ReaderConfig{
		GroupID:     "TeleBotClient",
		Brokers:     []string{address},
		Topic:       topicName,
		MaxBytes:    10e6,
		StartOffset: kafka.LastOffset,
	})

	rec := KafkaConsumer{
		uncommittedMessages: make(map[uuid.UUID]uncommittedMessage),
		offsets:             make(map[int]offsetWithTimeStamp),
		mu:                  &sync.Mutex{},
		reader:              r,
	}

	return &rec, nil
}

// Start receiving data from a Kafka instance. The received data is written to
// the return channel. The received messages are stored in a map.
func (r KafkaConsumer) StartGetData(ctx context.Context) <-chan KafkaConsumerData {
	dataChan := make(chan KafkaConsumerData)

	go r.consumeMessages(ctx, dataChan)

	return dataChan
}

func (r KafkaConsumer) consumeMessages(ctx context.Context, dataChan chan<- KafkaConsumerData) {
	for {
		if ctx.Err() != nil {
			break
		}

		msg, err := r.reader.FetchMessage(ctx)
		if err != nil {
			continue
		}

		msgUuid := uuid.New()
		r.mu.Lock()
		r.uncommittedMessages[msgUuid] = uncommittedMessage{msg: msg, timeStamp: time.Now()}
		r.mu.Unlock()

		var botData KafkaConsumerData
		if err = json.Unmarshal(msg.Value, &botData); err != nil {
			logrus.Error("failed to unmarshal an incoming data object", err)

			continue
		}

		botData.MessageUuid = msgUuid

		dataChan <- botData
	}

	if err := r.reader.Close(); err != nil {
		logrus.Fatal("failed to close the reader:", err)
	}
}
