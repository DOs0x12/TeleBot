package broker

import (
	"context"

	"github.com/DOs0x12/TeleBot/server/internal/entities/broker"
	"github.com/google/uuid"
)

type DataReceiver interface {
	StartReceivingData(ctx context.Context) (<-chan broker.InData, error)
	Commit(ctx context.Context, msgUuid uuid.UUID) error
}
