package broker

import (
	"context"
	"fmt"

	"github.com/DOs0x12/TeleBot/client/broker/consumer"
	"github.com/DOs0x12/TeleBot/client/broker/producer"
	"github.com/google/uuid"
)

// Stores the data that is needed by the bot app.
type BrokerData struct {
	ChatID      int64
	Value       string
	MessageUuid uuid.UUID
}

type KafkaBroker struct {
	cons *consumer.KafkaConsumer
	prod producer.KafkaProducer
}

func NewKafkaBroker(address string) (*KafkaBroker, error) {
	cons, err := consumer.NewKafkaConsumer(address)
	if err != nil {
		return nil, fmt.Errorf("failed to create a kafka consumer: %w", err)
	}

	prod := producer.NewKafkaProducer(address)

	return &KafkaBroker{cons: cons, prod: prod}, nil
}

func (b *KafkaBroker) StartGetData(ctx context.Context) <-chan BrokerData {
	consMsgs := b.cons.StartGetData(ctx)
	brMsgs := make(chan BrokerData)

	go pipelineConsData(ctx, consMsgs, brMsgs)

	return brMsgs
}

func pipelineConsData(ctx context.Context,
	consMsgs <-chan consumer.KafkaConsumerData,
	brMsgs chan<- BrokerData) {

	for {
		select {
		case <-ctx.Done():
			return
		case consData := <-consMsgs:
			brMsgs <- BrokerData{
				ChatID:      consData.ChatID,
				Value:       consData.Value,
				MessageUuid: consData.MessageUuid,
			}
		}
	}
}

func (b *KafkaBroker) Commit(ctx context.Context, msgUuid uuid.UUID) error {
	return b.cons.Commit(ctx, msgUuid)
}

func (b *KafkaBroker) SendData(ctx context.Context, data BrokerData) error {
	prData := producer.KafkaProducerData{
		ChatID:      data.ChatID,
		Value:       data.Value,
		MessageUuid: data.MessageUuid,
	}
	return b.prod.SendData(ctx, prData)
}

func (b *KafkaBroker) Stop() {
	b.prod.Stop()
}
