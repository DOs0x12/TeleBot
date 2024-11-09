package service

import (
	"context"

	"github.com/Guise322/TeleBot/server/internal/entities/service"
)

type DataTransmitter interface {
	TransmitData(ctx context.Context, data service.OutData) error
	Close()
}
