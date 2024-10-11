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

type BotDataDto struct {
	ChatID int64
	Value  string
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

func SendData(ctx context.Context, w *kafka.Writer, chatID int64, data string) error {
	botData, err := json.Marshal(BotDataDto{ChatID: chatID, Value: data})
	if err != nil {
		return err
	}
	return w.WriteMessages(ctx,
		kafka.Message{
			Topic: "botdata",
			Key:   []byte("data"),
			Value: botData,
		},
	)
}

func StartGetData() {

}
