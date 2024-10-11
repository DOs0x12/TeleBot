package service

import (
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

type CommandData struct {
	Name        string
	Description string
}

type BotData struct {
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

func SendData(ctx context.Context, w *kafka.Writer, botData BotData) error {
	data, err := json.Marshal(botData)
	if err != nil {
		return err
	}
	return w.WriteMessages(ctx,
		kafka.Message{
			Topic: "botdata",
			Key:   []byte("data"),
			Value: data,
		},
	)
}

func StartGetData(ctx context.Context, topicName, address string) <-chan BotData {
	r := kafka.NewReader(kafka.ReaderConfig{
		GroupID:     "regdfgd1",
		Brokers:     []string{address},
		Topic:       topicName,
		Partition:   0,
		MaxBytes:    10e6,
		StartOffset: kafka.LastOffset,
	})

	dataChan := make(chan BotData)
	go consumeMessages(ctx, dataChan, r)

	return dataChan
}

func consumeMessages(ctx context.Context, dataChan chan<- BotData, r *kafka.Reader) {
	for {
		if ctx.Err() != nil {
			break
		}

		msg, err := r.FetchMessage(ctx)
		if err != nil {
			continue
		}

		var botData BotData
		if err = json.Unmarshal(msg.Value, &botData); err != nil {
			logrus.Error("failed to unmarshal an incoming data object", err)

			continue
		}

		dataChan <- botData
		r.CommitMessages(ctx, msg)
	}

	if err := r.Close(); err != nil {
		logrus.Fatal("failed to close reader:", err)
	}
}
