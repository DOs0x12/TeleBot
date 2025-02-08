package broker

import (
	"context"
	"fmt"

	"github.com/DOs0x12/TeleBot/client/v2/broker/consumer"
	"github.com/DOs0x12/TeleBot/client/v2/broker/producer"
	"github.com/google/uuid"
)

// BrokerData stores the data that is needed by the bot application.
type BrokerData struct {
	CommName    string
	ChatID      int64
	Value       string
	MessageUuid uuid.UUID
}

// BrokerCommandData is the data that represents a command to the bot application.
type BrokerCommandData struct {
	Name        string
	Description string
}

// KafkaBroker is the main struct of the bot application client.
type KafkaBroker struct {
	cons *consumer.KafkaConsumer
	prod producer.KafkaProducer
}

// NewKafkaBroker creates a Kafka broker to work with application data.
func NewKafkaBroker(ctx context.Context, address, serviceName string) (*KafkaBroker, error) {
	cons, err := consumer.NewKafkaConsumer(address, serviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to create a kafka consumer: %w", err)
	}

	prod, err := producer.NewKafkaProducer(ctx, address)
	if err != nil {
		return nil, fmt.Errorf("failed to create a Kafka producer: %w", err)
	}

	return &KafkaBroker{cons: cons, prod: prod}, nil
}

// StartGetData starts geting application data from a Kafka broker.
func (b *KafkaBroker) StartGetData(ctx context.Context) <-chan BrokerData {
	consMsgs := b.cons.StartGetData(ctx)
	brMsgs := make(chan BrokerData)

	go pipelineConsData(ctx, consMsgs, brMsgs)

	return brMsgs
}

func pipelineConsData(
	ctx context.Context,
	consMsgs <-chan consumer.KafkaConsumerData,
	brMsgs chan<- BrokerData,
) {

	for {
		select {
		case <-ctx.Done():
			return
		case consData := <-consMsgs:
			brMsgs <- BrokerData{
				CommName:    consData.CommName,
				ChatID:      consData.ChatID,
				Value:       consData.Value,
				MessageUuid: consData.MessageUuid,
			}
		}
	}
}

// Commit commits a processed message in the Kafka broker.
func (b *KafkaBroker) Commit(ctx context.Context, msgUuid uuid.UUID) error {
	return b.cons.Commit(ctx, msgUuid)
}

// SendData sends application data to the kafka broker.
func (b *KafkaBroker) SendData(ctx context.Context, data BrokerData) error {
	prData := producer.KafkaProducerData{
		ChatID:      data.ChatID,
		Value:       data.Value,
		MessageUuid: data.MessageUuid,
	}
	return b.prod.SendData(ctx, prData)
}

// RegisterCommand registers a command in the bot application server.
func (s *KafkaBroker) RegisterCommand(
	ctx context.Context,
	commData BrokerCommandData,
	serviceName string,
) error {
	prodCommData := producer.CommandData{Name: commData.Name, Description: commData.Description}

	return s.prod.RegisterCommand(ctx, prodCommData, serviceName)
}

// Stop stops the Kafka broker.
func (b *KafkaBroker) Stop() {
	b.prod.Stop()
}
