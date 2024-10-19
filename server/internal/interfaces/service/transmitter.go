package service

import (
	"context"

	serviceEnt "github.com/Guise322/TeleBot/server/internal/entities/service"
)

type DataTransmitter interface {
	StartTransmittingData(ctx context.Context) chan<- serviceEnt.OutData
}
