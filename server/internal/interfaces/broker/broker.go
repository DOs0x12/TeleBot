package broker

import (
	"context"

	"github.com/DOs0x12/TeleBot/server/internal/entities/broker"
	"github.com/google/uuid"
)

type MessageBroker interface {
	StartReceivingData(ctx context.Context) (<-chan broker.DataFrom, error)
	Commit(ctx context.Context, msgUuid uuid.UUID) error
	TransmitData(ctx context.Context, data broker.DataTo) error
	Close()
}