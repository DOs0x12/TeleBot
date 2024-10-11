package service

import (
	"context"

	"github.com/segmentio/kafka-go"
)

func RegisterCommand(ctx context.Context, w *kafka.Writer, commData []byte) error {
	return w.WriteMessages(ctx,
		kafka.Message{
			Topic: "botdata",
			Key:   []byte("command"),
			Value: commData,
		},
	)
}

func SendData() {

}

func StartGetData() {

}
