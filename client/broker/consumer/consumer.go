package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/DOs0x12/TeleBot/server/v3/broker_data"
	"github.com/DOs0x12/TeleBot/server/v3/system"
	"github.com/DOs0x12/TeleBot/server/v3/tmp_storage"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

type KafkaConsumerDataDto struct {
	CommName string
	ChatID   int64
	Value    []byte
	IsFile   bool
}

// KafkaConsumerData is the data to work with a consumer.
type KafkaConsumerData struct {
	CommName    string
	ChatID      int64
	Value       []byte
	MessageUuid uuid.UUID
	IsFile      bool
}

// KafkaConsumer gets data from the bot app. It implements a way to store
// the uncommitted messages and offsets to commit them later.
type KafkaConsumer struct {
	offsetService   broker_data.OffsetService
	uncomMsgStorage tmp_storage.TmpStorage
	reader          *kafka.Reader
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
		offsetService:   broker_data.NewOffsetService(),
		uncomMsgStorage: tmp_storage.NewTmpStorage(),
		reader:          r,
	}

	return &rec, nil
}

// StartGetData starts receiving data from a Kafka instance. The received data is written to
// the return channel. The received messages are stored in a map.
func (r KafkaConsumer) StartGetData(ctx context.Context) <-chan KafkaConsumerData {
	dataChan := make(chan KafkaConsumerData)

	go r.consumeMessages(ctx, dataChan)
	msgLifetime := 48 * time.Hour
	clPeriod := 1 * time.Hour
	r.uncomMsgStorage.StartCleanupOldObjs(ctx, msgLifetime, clPeriod)
	r.offsetService.StartCleanupOffsets(ctx)

	return dataChan
}

func (r KafkaConsumer) consumeMessages(ctx context.Context, dataChan chan<- KafkaConsumerData) {
	defer func() {
		close(dataChan)
		if err := r.reader.Close(); err != nil {
			logrus.Fatal("failed to close the reader:", err)
		}
	}()

	for {
		if ctx.Err() != nil {
			break
		}

		msg, err := r.reader.FetchMessage(ctx)
		if err != nil {
			continue
		}

		msgUuid := r.uncomMsgStorage.AddObj(msg)

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
			IsFile:      botDataDto.IsFile,
		}

		dataChan <- botData
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
