package broker

import (
	"context"

	"github.com/DOs0x12/TeleBot/server/v3/internal/entities/broker"
	"github.com/google/uuid"
)

type MessageBroker interface {
	StartReceivingData(ctx context.Context) (<-chan broker.DataFrom, <-chan broker.CommandFrom, <-chan error)
	Commit(ctx context.Context, msgUuid uuid.UUID) error
	TransmitData(ctx context.Context, data broker.DataTo) error
	Close() error
}
