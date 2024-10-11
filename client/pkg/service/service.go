package service

import (
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"
)

type CommandData struct {
	Name        string
	Description string
}

func RegisterCommand(ctx context.Context, w *kafka.Writer, commData CommandData) error {
	data, err := json.Marshal(commData)
	if err != nil {
		return err
	}

	return w.WriteMessages(ctx,
		kafka.Message{
			Topic: "botdata",
			Key:   []byte("command"),
			Value: data,
		},
	)
}

func SendData(ctx context.Context, w *kafka.Writer) {

}

func StartGetData() {

}
