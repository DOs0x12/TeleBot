package consumer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/DOs0x12/TeleBot/server/v2/broker_data"
	"github.com/DOs0x12/TeleBot/server/v2/internal/common/retry"
	"github.com/DOs0x12/TeleBot/server/v2/internal/entities/broker"
	"github.com/DOs0x12/TeleBot/server/v2/internal/infrastructure/broker/topic"
	"github.com/DOs0x12/TeleBot/server/v2/system"

	kafka "github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

type KafkaConsumer struct {
	offsetService             broker_data.OffsetService
	uncommittedMessageService broker_data.UncommitedMessageService
	reader                    *kafka.Reader
}

func NewKafkaConsumer(address string) (*KafkaConsumer, error) {
	dataTopicName, err := system.GetDataToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get the data token: %w", err)
	}

	err = topic.CreateDataTopic(dataTopicName, address)
	if err != nil {
		return nil, fmt.Errorf("failed to create the data topic: %w", err)
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		GroupID:     "TeleBotWorker",
		Brokers:     []string{address},
		Topic:       dataTopicName,
		MaxBytes:    10e6,
		StartOffset: kafka.LastOffset,
	})

	cons := KafkaConsumer{
		offsetService:             broker_data.NewOffsetService(),
		uncommittedMessageService: broker_data.NewUncommitedMessageService(),
		reader:                    reader,
	}

	return &cons, nil
}

func (kr KafkaConsumer) StartReceivingData(ctx context.Context) (
	<-chan broker.DataFrom,
	<-chan broker.CommandFrom,
	<-chan error) {
	dataChan := make(chan broker.DataFrom)
	commChan := make(chan broker.CommandFrom)
	errChan := make(chan error)

	go kr.consumeMessages(ctx, dataChan, commChan, errChan)
	kr.uncommittedMessageService.StartCleanupUncommittedMessages(ctx)
	kr.offsetService.StartCleanupOffsets(ctx)

	return dataChan, commChan, nil
}

func (kr KafkaConsumer) consumeMessages(ctx context.Context,
	dataChan chan<- broker.DataFrom,
	commChan chan<- broker.CommandFrom,
	errChan chan<- error) {
	for {
		if ctx.Err() != nil {
			break
		}

		msg, err := kr.fetchMesWithRetries(ctx)
		if err != nil {
			logrus.Error("Failed to get a message from the broker: ", err)

			continue
		}

		msgUuid := kr.uncommittedMessageService.AddMsgToUncommitted(msg)
		commandKey := "command"
		isCommand := string(msg.Key) == commandKey
		if isCommand {
			comm, err := unmarshalBotCommand(msg.Value)
			if err != nil {
				errChan <- err

				continue
			}

			comm.MsgUuid = msgUuid
			commChan <- comm
		} else {
			botData, err := unmarshalBotData(msg.Value)
			if err != nil {
				errChan <- err

				continue
			}

			botData.MsgUuid = msgUuid
			dataChan <- botData
		}
	}

	if err := kr.reader.Close(); err != nil {
		logrus.Error("Failed to close the reader: ", err)
	}
}

type CommandDto struct {
	Name,
	Description,
	Token string
}

func unmarshalBotCommand(rawComm []byte) (broker.CommandFrom, error) {
	var commDto CommandDto
	err := json.Unmarshal(rawComm, &commDto)
	if err != nil {
		return broker.CommandFrom{}, fmt.Errorf("failed to unmarshal a command object: %w", err)
	}

	return broker.CommandFrom{
			Name:        commDto.Name,
			Description: commDto.Description,
			Token:       commDto.Token,
		},
		nil
}

type BotDataDto struct {
	ChatID int64
	Value  []byte
	IsFile bool
}

func unmarshalBotData(rawBotData []byte) (broker.DataFrom, error) {
	var botData BotDataDto
	err := json.Unmarshal([]byte(rawBotData), &botData)
	if err != nil {
		return broker.DataFrom{}, err
	}

	return broker.DataFrom{ChatID: botData.ChatID, Value: botData.Value, IsFile: botData.IsFile}, nil
}

func (kr KafkaConsumer) fetchMesWithRetries(ctx context.Context) (kafka.Message, error) {
	var msg kafka.Message

	act := func(ctx context.Context) error {
		var err error
		msg, err = kr.reader.FetchMessage(ctx)

		return err
	}

	err := retry.ExecuteWithRetries(ctx, act)

	return msg, err
}
