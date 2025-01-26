package broker

import (
	"context"
	"fmt"

	"github.com/DOs0x12/TeleBot/server/v2/internal/entities/broker"
	"github.com/DOs0x12/TeleBot/server/v2/internal/infrastructure/broker/consumer"
	"github.com/DOs0x12/TeleBot/server/v2/internal/infrastructure/broker/producer"
	"github.com/google/uuid"
)

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

func (b *KafkaBroker) StartReceivingData(ctx context.Context) (<-chan broker.DataFrom,
	<-chan broker.CommandFrom,
	<-chan error) {
	return b.cons.StartReceivingData(ctx)
}

func (b *KafkaBroker) Commit(ctx context.Context, msgUuid uuid.UUID) error {
	return b.cons.Commit(ctx, msgUuid)
}

func (b *KafkaBroker) TransmitData(ctx context.Context, data broker.DataTo) error {
	return b.prod.TransmitData(ctx, data)
}

func (b *KafkaBroker) Close() {
	b.prod.Close()
}
