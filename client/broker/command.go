package broker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/segmentio/kafka-go"
)

// CommandData is the data that represents a command to the bot app.
type CommandData struct {
	Name        string
	Description string
}

// Register a command in the bot app.
func (s Sender) RegisterCommand(ctx context.Context, commData CommandData) error {
	data, err := json.Marshal(commData)
	if err != nil {
		return fmt.Errorf("failed to marshal a command data to json: %w", err)
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
