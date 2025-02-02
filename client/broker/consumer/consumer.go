package consumer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/DOs0x12/TeleBot/server/v2/broker_data"
	"github.com/DOs0x12/TeleBot/server/v2/system"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

type KafkaConsumerDataDto struct {
	CommName string
	ChatID   int64
	Value    string
}

// KafkaConsumerData is the data to work with a consumer.
type KafkaConsumerData struct {
	CommName    string
	ChatID      int64
	Value       string
	MessageUuid uuid.UUID
}

// KafkaConsumer gets data from the bot app. It implements a way to store
// the uncommitted messages and offsets to commit them later.
type KafkaConsumer struct {
	offsetService             broker_data.OffsetService
	uncommittedMessageService broker_data.UncommitedMessageService
	reader                    *kafka.Reader
}

// NewKafkaConsumer creates a consumer to read data from a Kafka instance.
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
		offsetService:             broker_data.NewOffsetService(),
		uncommittedMessageService: broker_data.NewUncommitedMessageService(),
		reader:                    r,
	}

	return &rec, nil
}

// StartGetData starts receiving data from a Kafka instance. The received data is written to
// the return channel. The received messages are stored in a map.
func (r KafkaConsumer) StartGetData(ctx context.Context) <-chan KafkaConsumerData {
	dataChan := make(chan KafkaConsumerData)

	go r.consumeMessages(ctx, dataChan)
	r.uncommittedMessageService.StartCleanupUncommittedMessages(ctx)
	r.offsetService.StartCleanupOffsets(ctx)

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

		msgUuid := r.uncommittedMessageService.AddMsgToUncommitted(msg)

		var botDataDto KafkaConsumerDataDto
		if err = json.Unmarshal(msg.Value, &botDataDto); err != nil {
			logrus.Error("failed to unmarshal an incoming data object", err)

			continue
		}

		botData := KafkaConsumerData{
			CommName:    castFromTgCommand(botDataDto.CommName),
			ChatID:      botDataDto.ChatID,
			Value:       botDataDto.Value,
			MessageUuid: msgUuid,
		}

		dataChan <- botData
	}

	if err := r.reader.Close(); err != nil {
		logrus.Fatal("failed to close the reader:", err)
	}
}

func castFromTgCommand(tgCommand string) string {
	if tgCommand == "" {
		return tgCommand
	}

	tgCommChar := byte('/')

	firtsCommChar := tgCommand[0]

	if firtsCommChar != tgCommChar {
		return tgCommand
	}

	return tgCommand[1:]
}
