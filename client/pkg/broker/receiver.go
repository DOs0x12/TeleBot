package broker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

func (b Broker) StartGetData(ctx context.Context, topicName, address string) <-chan BotData {
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
		logrus.Fatal("failed to close the reader:", err)
	}
}

func (b Broker) Stop() error {
	if err := b.w.Close(); err != nil {
		return fmt.Errorf("failed to stop the broker: %w", err)
	}

	return nil
}
