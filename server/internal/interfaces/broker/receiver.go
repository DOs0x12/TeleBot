package broker

import (
	"context"

	"github.com/Guise322/TeleBot/server/internal/entities/broker"
)

type DataReceiver interface {
	StartReceivingData(ctx context.Context) (<-chan broker.InData, error)
}
