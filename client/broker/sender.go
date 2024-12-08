package broker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/segmentio/kafka-go"
)

// Sender works with the Kafka to send data to the bot app.
type Sender struct {
	w *kafka.Writer
}

// The method creates a sender to send data to a Kafka instance.
func NewSender(address string) Sender {
	w := &kafka.Writer{
		Addr:     kafka.TCP(address),
		Balancer: &kafka.LeastBytes{},
	}

	return Sender{w: w}
}

// Send data to the bot app via a Kafka instance.
func (s Sender) SendData(ctx context.Context, botData BotData) error {
	data, err := json.Marshal(botData)
	if err != nil {
		return err
	}

	err = s.w.WriteMessages(ctx,
		kafka.Message{
			Topic: "botdata",
			Key:   []byte("data"),
			Value: data,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to send data: %v", err)
	}

	return nil
}

// Stop working with the client.
func (s Sender) Stop() error {
	if err := s.w.Close(); err != nil {
		return fmt.Errorf("failed to stop the sender: %w", err)
	}

	return nil
}
