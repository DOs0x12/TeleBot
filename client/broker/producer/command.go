package producer

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/DOs0x12/TeleBot/client/broker/topic"
	"github.com/DOs0x12/TeleBot/server/system"
	"github.com/segmentio/kafka-go"
)

// CommandData is the data that represents a command to the bot app.
type CommandData struct {
	Name        string
	Description string
}

type commandDto struct {
	Name        string
	Description string
	Token       string
}

// Register a command in the bot app.
func (s KafkaProducer) RegisterCommand(ctx context.Context, commData CommandData) error {
	comToken, err := system.GetOrCreateTopicToken(commData.Name)
	if err != nil {
		return fmt.Errorf("failed to get or create a command token: %w", err)
	}

	commName := castToTelegramCommandName(commData.Name)

	dto := commandDto{Name: commName, Description: commData.Description, Token: comToken}
	data, err := json.Marshal(dto)
	if err != nil {
		return fmt.Errorf("failed to marshal a command data to json: %w", err)
	}

	err = topic.CreateTopicIfNotExist(ctx, comToken, s.w.Addr)
	if err != nil {
		return fmt.Errorf("failed to create topic %v: %w", comToken, err)
	}

	dataTopicName, err := system.GetOrCreateTopicToken("botdata")
	if err != nil {
		return fmt.Errorf("failed to get a topic token: %w", err)
	}

	err = topic.CreateTopicIfNotExist(ctx, dataTopicName, s.w.Addr)
	if err != nil {
		return fmt.Errorf("failed to create topic %v: %w", comToken, err)
	}

	msg := kafka.Message{
		Topic: dataTopicName,
		Key:   []byte("command"),
		Value: data,
	}

	err = s.w.WriteMessages(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to register a command: %w", err)
	}

	return nil
}

func castToTelegramCommandName(commName string) string {
	tCommName := strings.ToLower(commName)

	return "/" + tCommName
}
