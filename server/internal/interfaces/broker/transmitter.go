package broker

import (
	"context"

	"github.com/Guise322/TeleBot/server/internal/entities/broker"
)

type DataTransmitter interface {
	TransmitData(ctx context.Context, data broker.OutData) error
	Close()
}
