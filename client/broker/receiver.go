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

type offcetWithTimeStamp struct {
	value     int64
	timeStamp time.Time
}

// Receiver gets data from the bot app. It implements a way to store
// the uncommitted messages and offsets to commit them later.
type Receiver struct {
	mu                  *sync.Mutex
	reader              *kafka.Reader
	uncommittedMessages map[uuid.UUID]uncommittedMessage
	offcets             map[int]offcetWithTimeStamp
}

// Create a receiver to read data from a Kafka instance.
func NewReceiver(address string, command string) *Receiver {
	r := kafka.NewReader(kafka.ReaderConfig{
		GroupID:     "TeleBotClient",
		Brokers:     []string{address},
		Topic:       strings.Trim(command, "/"),
		MaxBytes:    10e6,
		StartOffset: kafka.LastOffset,
	})

	rec := Receiver{
		uncommittedMessages: make(map[uuid.UUID]uncommittedMessage),
		offcets:             make(map[int]offcetWithTimeStamp),
		mu:                  &sync.Mutex{},
		reader:              r,
	}

	return &rec
}

// Start receiving data from a Kafka instance. The received data is written to
// the return channel. The received messages are stored in a map.
func (r Receiver) StartGetData(ctx context.Context) <-chan BotData {
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

		msgUuid := uuid.New()
		r.mu.Lock()
		r.uncommittedMessages[msgUuid] = uncommittedMessage{msg: msg, timeStamp: time.Now()}
		r.mu.Unlock()

		var botData BotData
		if err = json.Unmarshal(msg.Value, &botData); err != nil {
			logrus.Error("failed to unmarshal an incoming data object", err)

			continue
		}

		botData.MessageUuid = msgUuid

		dataChan <- botData
	}

	if err := r.reader.Close(); err != nil {
		logrus.Fatal("failed to close the reader:", err)
	}
}
