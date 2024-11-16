package broker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/segmentio/kafka-go"
)

func (b Broker) SendData(ctx context.Context, botData BotData) error {
	data, err := json.Marshal(botData)
	if err != nil {
		return err
	}

	err = b.w.WriteMessages(ctx,
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
