package service

import (
	serviceEnt "TeleBot/internal/entities/service"
	"context"
)

type DataTransmitter interface {
	StartTransmittingData(ctx context.Context) chan<- serviceEnt.OutData
}
