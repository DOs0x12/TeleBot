package broker

import (
	"context"

	"github.com/DOs0x12/TeleBot/server/internal/entities/broker"
)

type DataTransmitter interface {
	TransmitData(ctx context.Context, data broker.DataTo) error
	Close()
}
