package producer

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/DOs0x12/TeleBot/server/v2/system"
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

// RegisterCommand registers a command in the bot application.
func (s KafkaProducer) RegisterCommand(
	ctx context.Context,
	commData CommandData,
	serviceName string,
) error {
	servToken, err := system.GetServiceToken(serviceName)
	if err != nil {
		return fmt.Errorf("failed to get or create a service token: %w", err)
	}

	commName := castToTelegramCommandName(commData.Name)

	dto := commandDto{Name: commName, Description: commData.Description, Token: servToken}
	data, err := json.Marshal(dto)
	if err != nil {
		return fmt.Errorf("failed to marshal a command data to json: %w", err)
	}

	dataTopicName, err := system.GetDataToken()
	if err != nil {
		return fmt.Errorf("failed to get a data token: %w", err)
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
