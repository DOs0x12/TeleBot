package broker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/DOs0x12/TeleBot/client/token"
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
func (s Sender) RegisterCommand(ctx context.Context, commData CommandData) error {
	comToken, err := token.GetOrCreateCommandToken(commData.Name)
	if err != nil {
		return fmt.Errorf("failed to get or create a command token: %w", err)
	}

	dto := commandDto{Name: commData.Name, Description: commData.Description, Token: comToken}
	data, err := json.Marshal(dto)
	if err != nil {
		return fmt.Errorf("failed to marshal a command data to json: %w", err)
	}

	err = s.createTopicIfNotExist()
	if err != nil {
		return fmt.Errorf("failed to process topic data: %w", err)
	}

	msg := kafka.Message{
		Topic: "botdata",
		Key:   []byte("command"),
		Value: data,
	}

	err = s.w.WriteMessages(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to register a command: %w", err)
	}

	return nil
}
