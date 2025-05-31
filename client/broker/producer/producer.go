package producer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/DOs0x12/TeleBot/client/v2/broker/topic"
	"github.com/DOs0x12/TeleBot/server/v3/system"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

// KafkaProducerData stores the data for the producer.
type KafkaProducerData struct {
	ChatID      int64
	Value       []byte
	MessageUuid uuid.UUID
	IsFile      bool
}

// Producer works with the Kafka to send data to the bot application.
type KafkaProducer struct {
	w *kafka.Writer
}

// The method creates a producer to send data to a Kafka instance.
func NewKafkaProducer(ctx context.Context, address string) (KafkaProducer, error) {
	topicName, err := system.GetDataToken()
	if err != nil {
		return KafkaProducer{}, fmt.Errorf("failed to get a data token: %w", err)
	}

	err = topic.CreateTopicIfNotExist(ctx, topicName, kafka.TCP(address))
	if err != nil {
		return KafkaProducer{}, fmt.Errorf("failed to create topic %v: %w", topicName, err)
	}

	w := &kafka.Writer{
		Addr:     kafka.TCP(address),
		Balancer: &kafka.LeastBytes{},
	}

	return KafkaProducer{w: w}, nil
}

// SendData sends data to the bot application via a Kafka instance.
func (s KafkaProducer) SendData(ctx context.Context, botData KafkaProducerData) error {
	data, err := json.Marshal(botData)
	if err != nil {
		return err
	}

	dataTopicName, err := system.GetDataToken()
	if err != nil {
		return fmt.Errorf("failed to get a topic token: %w", err)
	}

	err = s.w.WriteMessages(ctx,
		kafka.Message{
			Topic: dataTopicName,
			Key:   []byte("data"),
			Value: data,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to send data: %v", err)
	}

	return nil
}

// Stop stops working with the client.
func (s KafkaProducer) Stop() error {
	if err := s.w.Close(); err != nil {
		return fmt.Errorf("failed to stop the sender: %w", err)
	}

	return nil
}
