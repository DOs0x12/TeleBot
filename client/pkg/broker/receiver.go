package broker

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

type uncommittedMessage struct {
	msg       kafka.Message
	timeStamp time.Time
}

type Receiver struct {
	mu                  *sync.Mutex
	reader              *kafka.Reader
	uncommittedMessages map[uuid.UUID]uncommittedMessage
}

func NewReceiver(address string, command string) Receiver {
	r := kafka.NewReader(kafka.ReaderConfig{
		GroupID:     "TeleBotClient",
		Brokers:     []string{address},
		Topic:       strings.Trim(command, "/"),
		MaxBytes:    10e6,
		StartOffset: kafka.LastOffset,
	})

	rec := Receiver{
		uncommittedMessages: make(map[uuid.UUID]uncommittedMessage),
		mu:                  &sync.Mutex{},
		reader:              r,
	}

	return rec
}

func (r Receiver) StartGetData(ctx context.Context, address string) <-chan BotData {
	dataChan := make(chan BotData)

	go r.consumeMessages(ctx, dataChan)

	return dataChan
}

func (r Receiver) consumeMessages(ctx context.Context, dataChan chan<- BotData) {
	for {
		if ctx.Err() != nil {
			break
		}

		msg, err := r.reader.FetchMessage(ctx)
		if err != nil {
			continue
		}

		var botData BotData
		if err = json.Unmarshal(msg.Value, &botData); err != nil {
			logrus.Error("failed to unmarshal an incoming data object", err)

			continue
		}

		dataChan <- botData
	}

	if err := r.reader.Close(); err != nil {
		logrus.Fatal("failed to close the reader:", err)
	}
}
